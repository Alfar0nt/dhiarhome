# IMPLEMENTATION-PLAN — dhiarhome

> Refactored + compressed from former `documentation/to-do.md`. Pending work first, completed work compressed below.
> Legend: `[ ]` not started · `[~]` in progress · `[x]` done · `[!]` blocked.

---

# PART A — PENDING / NOT DONE

## Phase 12: Historical Graphs & Long-Term Monitoring
**Status: Pending.** No historical data. Add time-series storage + charts.

### Step 12.1 — Graph Data Storage (SQLite)
- [ ] Use SQLite (single-file, no external deps); add `modernc.org/sqlite` (pure Go, no CGO) — preferred over `go-sqlite3`.
- [ ] Create `internal/history/store.go`: `Open(dbPath)`, `Record(metric, value, timestamp)`, `Query(metric, from, to)`, `Prune(retention)`.
- [ ] Schema `metrics (id, name, value, recorded_at)` + index on `(name, recorded_at)`.
- [ ] Config `HistoryConfig{ Enabled, DBPath (default data/history.db), Interval (default 300s), RetentionDays (default 30) }`.
- [ ] Record CPU, memory, swap, disk, network RX/TX at interval; background goroutine for record + prune.

### Step 12.2 — Graph UI
- [ ] Add lightweight chart lib: **uPlot** (tiny/fast) or inline SVG sparklines.
- [ ] `templates/graphs.html`: range buttons 1h/6h/24h/7d/30d; line charts CPU%/Mem%/Swap%/Disk%/Net RX-TX; hover tooltip (value + timestamp).
- [ ] JSON endpoints: `GET /api/graphs/{cpu,memory,disk,network}?range=...`.
- [ ] Collapsible section below Proxmox metrics (or dedicated tab); responsive (full-width mobile, side-by-side desktop).

### Step 12.3 — Docs
- [ ] Update `.ai/TRD.md` (config + endpoints), `config-example.yaml` (history section).

**Files:** `internal/history/store.go`, `templates/graphs.html`, `main.go`, `go.mod`, `config-example.yaml`.

---

## Phase 18: Telegram Bot — Add/List/Complete Todos via Chat
**Status: Pending.** Currently todos require the web UI. Add chat commands via existing bot token (no new deps).

### Step 18.1 — Bot Poller
- [ ] `internal/notifications/telegram_bot.go` with `BotPoller` struct.
- [ ] Poll `getUpdates` every 3 s with `offset` tracking (avoid reprocessing).
- [ ] Initial blocking poll at startup to clear stale updates.
- [ ] Respect silent hours for non-command activity.

### Step 18.2 — Command Parser & Handlers
- [ ] `/add <text>` → add todo, reply confirmation + remaining list.
- [ ] `/done <id|text>` → mark done (by ID or text match), reply confirmation + remaining list.
- [ ] `/list` → current incomplete todos (or "🎉 All done!").
- [ ] `/help` → available commands.
- [ ] Ambiguous `/done` text match → reply numbered list, ask to pick by ID.

### Step 18.3 — Wire into main.go
- [ ] `go botPoller.Start(todoStore)` when Telegram enabled (store is thread-safe).

### Step 18.4 — Docs
- [ ] Update `.ai/TRD.md` (bot commands), `config-example.yaml` (`bot_enabled`, default false).

**Files:** `internal/notifications/telegram_bot.go`, `main.go`, `config-example.yaml`, `.ai/TRD.md`.

---

## Phase 5: Service Integration Framework — **DEFERRED** (remaining)
**Status: Deferred.** 5.4 (Radarr/Sonarr/Overseerr) + partial 5.6 (media services) already shipped and working. Remainder below.

### Step 5.1 — Widget Interface & Registry
- [ ] `internal/services/` package: `ServiceWidget` interface (`Name`, `Icon`, `Fetch`), `WidgetData{Metrics, Status, Details}`, `Metric{Label, Value}`.
- [ ] `registry.go`: `Register(type, factory)`, `Create(type, config)`; register built-ins at init.

### Step 5.2 — Generic HTTP API Widget
- [ ] `generic.go`: configurable URL/method/headers, JSON path extraction, auth (API key / Bearer / Basic), response cache TTL.

### Step 5.3 — Plex Widget
- [ ] `plex.go`: `{url}/status/sessions` with `X-Plex-Token`; parse now-playing / stream count / title; template + mock.

### Step 5.5 — Portainer Widget
- [ ] `portainer.go`: `/api/endpoints/{id}/docker/containers/json`; JWT/API key; running/stopped/total counts.

### Step 5.6 — Integrate Service Widgets (rest)
- [ ] Add `ServiceWidgetConfig` (type/name/url/token/api_key/icon/headers/json_path) to config.
- [ ] Init from config in `main.go`; fetch in poll loop; `templates/services.html` card grid (icon, name, status badge, 2–4 metrics); add to main grid.
- [x] Media services integration (done — see Phase 17 area / changelog 0.8.x).

**Files:** `internal/services/*.go`, `templates/services.html`, `main.go`, `config-example.yaml`.

---

## Phase 4.4 — Optional: Bookmark Link Health Checking
- [ ] Reuse `internal/monitor/http.go` `CheckService()` for bookmark URLs.
- [ ] Small green/red status dot per bookmark; poll at slower interval (60 s).
- [ ] Config `check_health: true/false` per group.

---

## Backlog / Future Improvements
- [ ] **B.1 — CSP allow Google Fonts**: add `https://fonts.googleapis.com` to `style-src` (`main.go` security headers). Currently blocks Inter stylesheet → fallback to system stack after cache expiry. `fonts.gstatic.com` already in `font-src`.
- [ ] **B.2 — Proxmox LVM / Local-Thin disk monitoring**: current `/nodes/{node}/disks/list` skips entries with empty mountpoints (e.g. `local-lvm`). Use `/nodes/{node}/storage` or `/disks/lvmthin` to fetch pool usage; parse `used`/`total`; filter to thin pools, skip already-listed mountpoints.

---

# PART B — COMPLETED (compressed)

| Phase | Title | Status | Essence |
|---|---|---|---|
| 1 | Visual Enhancements | Complete | Appearance config; custom background (local/URL, opacity, blur); glassmorphism; Inter typography; animations; WCAG AA accessibility |
| 2 | Utility Widgets | Complete | Widgets package (interface + registry); weather (Open-Meteo), datetime (client clock), sysinfo, custom text; dashboard integration |
| 3 | Network Monitoring | Complete | `internal/network`; `/proc/net/dev` parse; two-sample speed + moving avg; config + interface cards |
| 4 | Bookmarks & Links | Mostly done | Bookmark config; icon modes (Lucide/path/favicon cache); grid UI. **4.4 optional pending** |
| 5 | Service Framework | Deferred | 5.4 (media services) done; remainder pending (see Part A) |
| 6 | Polish, Perf & Docs | Complete | Profiling, 5s timeouts, graceful shutdown, config `Validate()`, Dockerfile, security hardening |
| 7 | Proxmox API Enrichment | Complete | Swap, load average, PVE/kernel version, VM/LXC enumeration |
| 8 | Disk Monitoring | Complete | Extra disks: statfs auto-detect + manual sizes; `ParseSize()`; merge + dedup |
| 9 | Remote Docker & Portainer | Complete | TLS (mTLS/`skip_tls`), Portainer API, connection priority |
| 10 | UI & Theme Toggle | Complete | Favicon/logo config, bigger text, dark/light toggle (localStorage) |
| 11 | Telegram Notifications | Complete | Notifier; service + container up/down alerts; cooldown; silent hours; mock; test endpoint |
| 12 | Historical Graphs | **Pending** | See Part A |
| 13 | Background Polling | Complete | Background Proxmox/Docker pollers, cache-only `statusHandler` (<5ms), initial blocking polls |
| 14 | To-Do Edit | Complete | Store `Update()`; PATCH endpoint; inline edit in widget + modal; optimistic |
| 15 | Font Consistency | Complete | Bookmarks/Services titles → `text-2xl`, icons `w-7`, spacing aligned with Docker |
| 16 | VM/LXC Sort Fix | Complete | Sort VMs/LXCs by VMID in `GetVirtualization()` |
| 17 | Todo Telegram Notifications | Complete | `NotifyTodoAdd`/`NotifyTodoComplete` + remaining task list; silent hours; no cooldown |
| Backlog | B.1, B.2 | **Pending** | See Part A |

### Progress Tracker

| Phase | Steps | Done | Remaining | Status |
|---|---|---|---|---|
| 1–3 | 17 | 17 | 0 | Complete |
| 4 | 4 | 3 | 1 | Mostly done (4.4 optional) |
| 5 | 6 | 2 | 4 | Deferred |
| 6–11 | 24 | 24 | 0 | Complete |
| 12 | 3 | 0 | 3 | Pending |
| 13–17 | 20 | 20 | 0 | Complete |
| B (backlog) | 2 | 0 | 2 | Pending |

**No fixed deadlines.** Next-up recommendation: **B.1 (trivial CSP fix)** → **Phase 12 (highest user value)** → **Phase 18 (convenience)** → Phase 5 remainder.
