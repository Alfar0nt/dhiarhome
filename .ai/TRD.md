# TRD — dhiarhome

> Technical Requirements Document. Product context: [`PRD.md`](PRD.md). Runtime flows: [`APP-FLOW.md`](APP-FLOW.md). UI: [`DESIGN.md`](DESIGN.md). Deploy: [`../docs/DEPLOYMENT.md`](../docs/DEPLOYMENT.md).

## 1. Tech Stack

| Layer | Choice |
|---|---|
| Language | Go 1.26 (statically compiled, `CGO_ENABLED=0`) |
| Target | linux/amd64 (cross-compile ARM64/ARM/Windows/macOS) |
| Backend deps | `gopkg.in/yaml.v3` (only external dep) |
| Frontend | HTML5 + Tailwind CSS (CDN) + HTMX 1.9.10 + Alpine.js 3.x (CDN) |
| Config | YAML |
| Storage | In-memory cache + JSON files (`data/`) — no database |
| Deploy | Docker multi-stage or single binary |

## 2. Repository Structure

```
dhiarhome/
├── main.go                     # Entry point: config, clients, pollers, HTTP handlers, routes
├── config-example.yaml         # Config template
├── Dockerfile                  # Multi-stage build
├── internal/
│   ├── cache/history.go        # In-memory service state cache (linked list, max 100)
│   ├── config/config.go        # YAML loader + all config structs + Validate() + ParseSize()
│   ├── docker/client.go        # Docker API client (socket/TCP/TLS/Portainer)
│   ├── mediaservices/client.go # Sonarr/Radarr/Overseerr clients
│   ├── bookmarks/store.go      # Bookmark processing + favicon cache (SSRF-safe)
│   ├── monitor/http.go         # HTTP service health checker
│   ├── network/{types,monitor}.go  # /proc/net/dev parser, speed calc, mock
│   ├── notifications/telegram.go   # Telegram notifier
│   ├── proxmox/client.go       # Proxmox API client + CPU info + disk usage
│   ├── todo/store.go           # Persistent to-do store (JSON)
│   └── widgets/{widget,registry,weather,datetime,sysinfo,custom_text}.go
├── static/
│   ├── index.html              # Dashboard page (Go template + HTMX + merge-swap JS)
│   └── backgrounds/            # Custom background images
├── templates/
│   ├── status.html             # Main status fragment (HTMX target)
│   ├── network.html, bookmarks.html, mediaservices.html, todo.html
│   └── widgets/widgets.html
├── .ai/                        # This documentation set
└── docs/                       # CHANGELOGS.md, DEPLOYMENT.md
```

## 3. Architecture

### 3.1 Data Flow
1. **Startup** — load `config.yaml` → `Validate()` → init Proxmox/Docker clients + history cache → **initial blocking poll** (Proxmox + Docker) to warm cache → start HTTP server + background goroutines.
2. **Background polling** (see §4) — all external API calls happen here; results stored in mutex-protected shared state.
3. **HTTP request** — `index.html` rendered as Go template with appearance config; HTMX polls `/status` every 5s; handler reads **cache only** (<5 ms) and renders `status.html`.
4. **DOM update** — custom `merge-swap` HTMX extension diffs DOM in place (no flicker).

### 3.2 Concurrency Model
- Every shared cache guarded by `sync.RWMutex` (`pxStatusMu`, `containersMu`, service/media/network state).
- Pollers write under write-lock; HTTP handlers read under read-lock.
- On poll error: keep previous values (stale data > no data).

## 4. Background Goroutines

| Goroutine | Interval | Work |
|---|---|---|
| `pollProxmox()` | 5 s | Node status, virtualization (VMs/LXCs), merge extra disks |
| `pollDocker()` | 5 s | Container list + name filter (mock fallback on error) |
| service monitor (`doPoll`) | 10 s | HTTP health checks + response times |
| network monitor | 3 s (config) | Sample `/proc/net/dev` (display cached 10s TTL) |
| media services | 30 s | Sonarr/Radarr/Overseerr stats |
| container state tracker | 15 s | Detect Docker state transitions → notifications |

HTMX auto-refresh: `/status` every 5 s (cache read only).

## 5. Component Details

### 5.1 `internal/config`
- Parses YAML; structs for all sections (see §7).
- `setDefaults()` applies omitted appearance/widget defaults.
- `Validate()`: range checks (opacity 0-1, blur 0-30, interval 1-60), URL checks, Proxmox→mock fallback when creds missing, weather auto-disable when lat/long unset, timezone fallback to Local. Logs active feature summary.
- `ParseSize()`: human-readable sizes → bytes (B/KB/MB/GB/TB + KiB/MiB/GiB/TiB, decimals).

### 5.2 `internal/proxmox`
- Auth via `PVEAPIToken` header; TLS skip-verify (self-signed).
- `GetNodeStatus()`: CPU, memory, swap, loadavg (`json.Number` for string-encoded floats), `pveversion`, `kversion`. Multi-disk from `/nodes/{node}/disks/list`.
- Extra disks: `ReadDiskUsage()` via `syscall.Statfs`; `mergeExtraDisks()` dedups by mountpoint.
- `GetVirtualization()`: QEMU/LXC lists from `/nodes/{node}/qemu`, `/nodes/{node}/lxc`; **sorted by VMID**.
- `ReadLocalCPUInfo()`: parses `/proc/cpuinfo`; `cleanCPUName()` strips vendor suffixes.
- Mock mode for all endpoints.

### 5.3 `internal/docker`
- Engine API `/containers/json?all=1`.
- TLS: client certs (mTLS), `skip_tls`; endpoint `tcp://` + TLS → `https://`.
- Portainer: `/api/endpoints/{env_id}/docker/containers/json` with `X-API-Key`.
- Connection priority: Portainer > Remote Docker > Local socket.

### 5.4 `internal/cache`
- Thread-safe doubly-linked list; last 100 service snapshots; `GetLatest()` per service.

### 5.5 `internal/monitor`
- HTTP GET, 5 s timeout; "Online" (2xx) / "Offline" (error) / "Warning" (non-2xx); measures response time.

### 5.6 `internal/network`
- Parse `/proc/net/dev`; two-sample rate + moving average (last 3); `formatSpeed()`, `formatBytes()`.
- `GetStats()` caches formatted output 10 s TTL to prevent HTML churn; raw sampling continues at full rate.

### 5.7 `internal/todo`
- Thread-safe CRUD, JSON persistence, auto-increment IDs.
- `GetAll`, `GetByID`, `Add`, `Update`, `Toggle` (records/clears `done_at` RFC3339), `Delete`. Saves on every mutation.

### 5.8 `internal/bookmarks`
- Icon resolution (Lucide name / image path / favicon) + favicon caching to `data/icons/` (MD5 filenames).
- SSRF-safe fetch: scheme validation, DNS resolve, block private ranges (loopback, RFC1918, link-local, IPv6 ULA).

### 5.9 `internal/mediaservices`
- Sonarr/Radarr/Overseerr clients, `X-Api-Key`, 5 s timeout, graceful failure, `MockStats()`.

### 5.10 `internal/notifications`
- Telegram `sendMessage` (HTML parse mode); cooldown, silent hours, mock mode.
- `NotifyServiceChange`, `NotifyTodoAdded`, `NotifyTodoCompleted`, `buildRemainingList()`.

### 5.11 `internal/widgets`
- `Widget` interface (`Name`, `Type`, `Fetch`) + registry (`Register`, `FetchAll`, `Count`).
- Weather (Open-Meteo + WMO code map, cached), DateTime (client clock), SysInfo, CustomText (escaped).

### 5.12 Template rendering
- `html/template` funcs: `percent(used,total)`, `mult(a,b)`, `gb(bytes)`, `roundDur(duration)`.
- `combineWidgets()` merges weather+datetime, reorders for layout.
- Templates pre-parsed at startup (`template.Must`).

## 6. HTTP API

| Method | Path | Purpose |
|---|---|---|
| GET | `/` | Dashboard page (Go template) |
| GET | `/status` | HTMX HTML fragment (cache reads, <5 ms) |
| GET | `/background` | Serve local background image (MIME + 1h cache) |
| GET | `/api/background` | JSON background config (source, opacity, blur) |
| GET | `/logo` | Serve local logo file (MIME + 1h cache) |
| GET | `/bookmarks/icons/` | Serve cached favicon |
| GET | `/api/todos` | List todos (JSON) |
| POST | `/api/todos` | Add todo (`{"text": "..."}`) |
| PATCH | `/api/todos/{id}` | Update todo text |
| PUT | `/api/todos/{id}` | Toggle done state |
| DELETE | `/api/todos/{id}` | Delete todo |
| GET | `/api/notifications/test` | Send test Telegram message |

Todos endpoints rate-limited per IP (30 req/min, sliding window, respects `X-Forwarded-For`/`X-Real-IP`).

## 7. Configuration Reference

### Proxmox
```yaml
proxmox:
  url: "https://192.168.1.100:8006/api2/json"
  node_name: "pve"
  token_id: "root@pam!dashboard"
  token_secret: "YOUR-SECRET-UUID"
  mock: false
  extra_disks:
    - mountpoint: "/mnt/data"    # required
      auto_detect: true          # read via statfs (default true)
    - mountpoint: "/mnt/nas"
      label: "NAS Storage"
      total: "8TB"               # manual (B/KB/MB/GB/TB/KiB/MiB/GiB/TiB)
      used: "3.2TB"
      auto_detect: false
```

### Docker
```yaml
docker:
  socket: "unix:///var/run/docker.sock"   # or tcp://host:2375|2376
  monitor_containers: []                   # empty = all
  # skip_tls: true
  # tls_ca_cert: "/path/ca.pem"
  # tls_cert: "/path/cert.pem"
  # tls_key: "/path/key.pem"
  # portainer_url: "https://portainer.example.com"
  # portainer_api_key: "ptr_XXX"
  # portainer_env_id: 1
```

### Services / Appearance / Widgets / Network / Todos / Media / Bookmarks / Notifications
```yaml
services:
  - name: "Nextcloud"
    url: "https://nextcloud.example.com"
    # skip_tls: true

appearance:
  background_image: ""
  background_url: ""
  background_opacity: 0.4
  background_blur: 3
  logo: "static/logo.png"       # path or URL; empty = default SVG
  theme: "dark"                 # "dark" | "light"
  card_opacity: 0.6
  card_blur: 12
  accent_color: "#3b82f6"

widgets:
  weather: { enabled: false, latitude: 40.7128, longitude: -74.0060, units: "celsius", cache_minutes: 15, mock: false }
  datetime: { enabled: false, timezone: "America/New_York", format_24h: false }
  system_info: { enabled: false }
  custom_text: { enabled: false, title: "Note", content: "Welcome!" }

network:
  enabled: false
  show_speed: true
  show_total_transfer: true
  update_interval: 3
  mock: false
  interfaces:
    - { name: "eth0", label: "Primary" }

todos:
  enabled: false
  file_path: "data/todos.json"
  title: "To-Do"

media_services:
  - { name: "Sonarr", url: "...", api_key: "...", webui: "..." }

bookmarks:
  - group: "Infrastructure"
    links:
      - { name: "Proxmox VE", url: "https://...", icon: "server", new_tab: true }

notifications:
  telegram:
    enabled: false
    bot_token: "YOUR_BOT_TOKEN"
    chat_id: "YOUR_CHAT_ID"
    notify_up: true
    notify_down: true
    notify_todo_add: true
    notify_todo_complete: true
    cooldown: 5
    silent_hours: []
    mock: false
```

## 8. Security

- **Headers** (`securityHeaders()` wrapper on all responses): CSP (`'unsafe-eval'` for Alpine), `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `X-XSS-Protection: 1; mode=block`, `Referrer-Policy: same-origin`.
- **SSRF**: favicon fetcher validates scheme + blocks private IPs.
- **Path traversal**: `backgroundServeHandler` uses `filepath.Clean()`, rejects `..`.
- **JSON injection**: `json.NewEncoder` (no string formatting).
- **Rate limiting**: 30 req/min per IP on todo APIs.
- **Input validation**: todo text cap 500 chars.
- Secrets live only in gitignored `config.yaml`; Docker image ships `config-example.yaml`.
- Proxmox client uses `InsecureSkipVerify` (necessary for self-signed homelab certs).

## 9. Performance

- Memory ~10–20 MB; CPU <1%; binary ~14 MB (stripped ~9.6 MB); startup <1 s.
- Cache: 100 service states × ~100 B ≈ 10 KB.
- `statusHandler` cache-only reads → <5 ms; background pollers absorb latency.

## 10. Build & Deploy

See [`../docs/DEPLOYMENT.md`](../docs/DEPLOYMENT.md). Quick build:
```bash
go build -o dhiarhome main.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dhiarhome main.go
```
CLI flags: `--config <path>` (default `config.yaml`), `--addr <addr>` (default `:8080`).

## 11. Known Technical Debt

- CSP currently blocks Google Fonts stylesheet (backlog B.1).
- Proxmox disk API skips LVM-thin pools (no mountpoint) (backlog B.2).
- Phase 5 service-integration framework partially deferred (generic/Plex/Portainer widgets).
