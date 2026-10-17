# APP-FLOW — dhiarhome

> Runtime flows: startup, pollers, request handling, notifications, todos. Components: [`TRD.md`](TRD.md) §5. Product: [`PRD.md`](PRD.md).

## 1. Startup Flow

```
main()
 ├─ parse flags (--config, --addr)
 ├─ config.Load(path) → Validate()  → log active features
 ├─ init clients: proxmox, docker, network, media, bookmarks, widgets, todo store, notifier
 ├─ read local CPU info (/proc/cpuinfo) once
 ├─ load todos from JSON file
 ├─ INITIAL BLOCKING POLL: doPollProxmox() + doPollDocker()   # warm cache
 ├─ start goroutines: pollProxmox, pollDocker, doPoll(services),
 │                    networkMonitor, pollMediaServices, pollContainers  (as enabled)
 ├─ register HTTP routes (securityHeaders wrapper, ServeMux)
 └─ http.Server{ ReadTimeout 10s, WriteTimeout 30s, IdleTimeout 60s }.ListenAndServe(addr)
        └─ SIGINT/SIGTERM → graceful shutdown (stop network monitor, 5s drain)
```

Why initial blocking poll: user already sees skeleton during HTML load; by the time HTMX fires `/status`, cache is warm → instant first render.

## 2. Background Polling Flows

Each poller loops on `time.Ticker`, writes to mutex-guarded shared state. On error → keep previous value.

- **Proxmox (5 s)** `pollProxmox()`:
  `GetNodeStatus()` → `mergeExtraDisks()` (statfs/manual) → `GetVirtualization()` (VMs/LXCs sorted by VMID) → store under `pxStatusMu`.
- **Docker (5 s)** `pollDocker()`:
  `GetContainers()` (local/TCP/TLS/Portainer) → apply `monitor_containers` filter → store under `containersMu`. On error: mock fallback if `proxmox.mock` else keep previous.
- **Services (10 s)** `doPoll()`:
  For each service: HTTP `CheckService()` → append to `historyCache`. Detect Online↔Offline transitions → `prevServiceStates` map → `recordTransition()` + Telegram `NotifyServiceChange()`.
- **Network (3 s)**:
  sample `/proc/net/dev` → compute RX/TX rate (moving avg) → store. `GetStats()` returns 10s-cached formatted strings.
- **Media (30 s)** `pollMediaServices()`:
  fetch Sonarr/Radarr/Overseerr → store under media mutex.
- **Container tracker (15 s)** `pollContainers()`:
  detect running↔exited → `recordTransition()` + Telegram.

## 3. HTTP Request Flows

### 3.1 Page load — `GET /`
```
Browser → GET /  → indexHandler
  index.html (Go template) rendered with appearance config (bg, logo, theme, accent)
  → HTML shell (skeleton) + HTMX script + merge-swap extension + Alpine
  → HTMX immediately GETs /status
```

### 3.2 Auto-refresh — `GET /status` (every 5 s)
```
HTMX → GET /status → statusHandler
  read cachedPxStatus / cachedVirtInfo / cachedContainers / services / network / media / widgets
  (+ flush pending TransitionEvents into response for toasts)
  → render status.html fragment  (<5 ms, no external API calls)
  → merge-swap diffs DOM in place (text nodes + changed attrs only)
```

### 3.3 Static asset / media endpoints
- `GET /background` → read local image file, 1h cache, correct MIME.
- `GET /api/background` → JSON `{source, opacity, blur}`.
- `GET /logo` → serve local logo (fallback inline SVG).
- `GET /bookmarks/icons/` → serve cached favicon from `data/icons/`.

### 3.4 Todo API (Alpine.js `fetch()`, optimistic)
```
POST   /api/todos          {text}  → Add()   → save JSON → (if enabled) NotifyTodoAdded
PATCH  /api/todos/{id}     {text}  → Update()→ save JSON
PUT    /api/todos/{id}             → Toggle()→ save JSON → if now done: NotifyTodoCompleted
DELETE /api/todos/{id}             → Delete()→ save JSON
GET    /api/todos                  → JSON array
```
UI updates instantly; API sync in background; on failure revert + console log. Rate-limited 30/min/IP.

### 3.5 Notification test
`GET /api/notifications/test` → send a Telegram test message (verify config).

## 4. State Transition → Notification Flow

```
poller detects state change (service Online↔Offline, container running↔exited, todo add/complete)
  → recordTransition(event)             # buffered for web UI
  → flushTransitions()                  # next /status response embeds events
        → Alpine x-init reads JSON → toast (green=up / red=down, dismiss 4s)
  → notifier.Notify*()  (if enabled + flag set)
        → skip if silent hour
        → skip if within cooldown (services/containers only)
        → POST api.telegram.org/bot{token}/sendMessage (HTML)
```
Todos fire immediately (no cooldown — user-initiated).

## 5. Widget Rendering Flow

```
registry.FetchAll() → []WidgetData (weather, datetime, sysinfo)
  → combineWidgets()  # merge weather+datetime into one card; reorder for grid
  → template widgets/widgets.html renders per type
  → datetime uses client-side setInterval(1s) clock (no server polling)
  → weather cached (config cache_minutes, default 15)
```

## 6. DOM Diff Swap Flow (merge-swap)

```
First load: skeleton HTML → standard innerHTML swap (no glass-cards yet)
Subsequent /status polls:
  mergeDOM(current, new) recursive walk
    patchChildren()  → skip blank text nodes; compare tag/type; update or replace
    syncAttrs()      → class, style, aria-valuenow/aria-label, role
    data-preserve elements (e.g. todo widget root) → SKIPPED entirely (protect Alpine state)
  → glass-card elements preserved → backdrop-filter GPU layer intact → no flicker
```

## 7. Config Load & Validation Flow

```
read YAML → unmarshal structs → setDefaults()
  → Validate():
      numeric ranges (opacity, blur, interval)
      URL formats
      Proxmox: creds missing → force mock mode
      Weather: lat/long missing → disable
      Timezone: invalid → Local
      extra_disks: mountpoint required; ParseSize validation
  → log "Active features: ..." summary → continue (never crash on bad config)
```
