# Changelogs - dhiarhome

All notable changes to this project are documented in this file.

---

## [1.0.1] - 2026-06-19 - Skip TLS Option for Services

### Added
- **`skip_tls` config option** for monitored web services — allows health checks against services with self-signed TLS certificates
  - `ServiceConfig.SkipTLS bool` in `internal/config/config.go`
  - When `true`, `CheckService()` creates a custom `http.Transport` with `InsecureSkipVerify: true`
  - Default is `false` (secure by default)
- **Bookmarks configuration reference** added to `documentation/docs.md`
  - Full YAML example with group/link structure, icon modes, and notes
  - Includes example groups: Infrastructure, Media

### Files Modified
- `internal/config/config.go` — Added `SkipTLS bool` field to `ServiceConfig`
- `internal/monitor/http.go` — Added `skipTLS` parameter, `crypto/tls` import, custom transport
- `main.go` — Passes `svc.SkipTLS` to `monitor.CheckService()`
- `config-example.yaml` — Added `skip_tls` example service, fixed `bookmarks: []` → `bookmarks:`
- `documentation/docs.md` — Added Bookmarks config reference, updated Services section with `skip_tls`

---

## [1.0.0] - 2026-06-19 - First Stable Release

### Overview
dhiarhome v1.0.0 marks the first stable public release. All core features are complete, tested, and documented — ready for production homelab use.

### What's New Since 0.10.1
No new features — this is a stability and documentation release. All functionality from 0.1.0 through 0.10.1 has been tested and verified across Docker and bare-metal deployments.

### Complete Feature Set
- **Proxmox VE** — CPU model + cores/threads, RAM, multi-disk, VM/LXC tracking, uptime
- **Docker** — container status with name filtering
- **Web services** — HTTP health checks with response times
- **Media services** — Sonarr, Radarr, Overseerr stats with WebUI links
- **Network** — per-interface RX/TX speeds (/proc/net/dev)
- **To-do list** — Alpine.js interactive, persisted to JSON
- **Weather + time** — Open-Meteo forecast, live clock, timezone support
- **System info** — hostname, OS, uptime, Go runtime
- **Bookmarks** — custom links with auto-fetched favicons
- **Glassmorphism UI** — blur cards, custom backgrounds, accent color
- **DOM diff swap** — zero flicker on 5s auto-refresh
- **Config validation** — startup warnings, graceful fallbacks
- **Security hardening** — CSP headers, rate limiting, path traversal protection
- **Mock mode** — test everything without real credentials
- **Single binary** — ~14MB, zero database, no external dependencies

### Files Modified
- `documentation/changelogs.md` — v1.0.0 release entry
- `documentation/docs.md` — Updated binary size, project structure, feature status
- `documentation/to-do.md` — Phase 6 verified complete
- `documentation/deployment.md` — Updated Go version references, fixed old naming
- `README.md` — Updated roadmap, bookmarks marked complete

---

## [0.10.1] - 2026-06-19 - Security Hardening

### Added
- **Security headers middleware** — applied to all responses via `securityHeaders()` wrapper:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
  - `Referrer-Policy: same-origin`
  - `Content-Security-Policy` — allows Tailwind CDN, HTMX/Alpine CDN, Google Fonts; blocks everything else
- **Per-IP rate limiter** — `rateLimiter` struct with sliding window (30 requests/min per IP)
  - Applied to `/api/todos` and `/api/todos/` endpoints
  - Respects `X-Forwarded-For` and `X-Real-IP` headers for reverse-proxy setups
  - Returns HTTP 429 Too Many Requests when exceeded
- **Path traversal protection** — `backgroundServeHandler` now:
  - Uses `filepath.Clean()` to normalize paths
  - Rejects paths containing `..`
  - Logs `[SECURITY]` warning on blocked attempts
- **Input length validation** — todo text capped at 500 characters (returns HTTP 400)

### Changed
- **Dockerfile** — copies `config-example.yaml` (safe placeholder) instead of real `config.yaml` into the Docker image
  - Prevents accidental credential leakage when publishing images
  - Added runtime volume-mount instruction in Dockerfile comment
- **HTTP routing** — switched from `http.HandleFunc` (global DefaultServeMux) to explicit `http.ServeMux` + `securityHeaders()` wrapper
  - Server `Handler` field now points to the secured handler chain

### Security Audit Results
- No hardcoded secrets, API keys, or passwords in source code
- `config.yaml` never committed to git (verified via git history)
- All example configs use placeholder values (`YOUR-SECRET-UUID-HERE`)
- `InsecureSkipVerify: true` retained for Proxmox (necessary for self-signed homelab certs)
- Binary size: 14MB (unchanged)

### Files Modified
- `main.go` — Security headers, rate limiter, path traversal protection, input validation, ServeMux refactor
- `Dockerfile` — Use `config-example.yaml` instead of `config.yaml`
- `documentation/changelogs.md` — This entry
- `documentation/prompt-history.md` — Session 22 added
- `documentation/to-do.md` — Security hardening step marked complete

---

## [0.10.0] - 2026-06-19 - Phase 6: Polish, Performance & Documentation

### Added
- **Graceful shutdown**: Server handles SIGINT/SIGTERM, stops network monitor, gracefully shuts down HTTP server with 5s timeout
- **HTTP server timeouts**: `ReadTimeout: 10s`, `WriteTimeout: 30s`, `IdleTimeout: 60s`
- **Config validation** (`internal/config/config.go`):
  - `Validate()` method called after loading config
  - Numeric range validation (opacity 0-1, blur 0-30, update interval 1-60)
  - URL format validation (background, services, media, bookmarks)
  - Proxmox fallback to mock mode when credentials are missing
  - Weather auto-disable when lat/long not set
  - Timezone validation with fallback to Local
  - Feature summary logged on startup (`Active features: Proxmox (mock), Weather, Todos, Bookmarks (11 links)`)

### Changed
- **Dockerfile**: Updated to `golang:alpine` (latest), added `data/icons` directory creation for favicon cache
- **config-example.yaml**: Updated bookmarks section with expanded examples (8 links), removed description fields, added icon reference and scrolling note
- **Phase 5 deferred**: Marked as future work in to-do.md. Steps 5.4 (Radarr/Sonarr) and partial 5.6 (media services) remain active

### Verified (already done)
- All HTTP clients have 5s timeouts (Proxmox, Docker, weather, services, media, bookmarks, monitor)
- Weather API caching works (RWMutex + configurable TTL)
- Templates pre-parsed at startup via `template.Must()`
- Binary size: 14MB (under 15MB target)

### Files Modified
- `main.go` — Graceful shutdown with signal handling, HTTP server timeouts
- `internal/config/config.go` — `Validate()` method with comprehensive checks + feature summary
- `Dockerfile` — `golang:alpine`, `data/icons` directory
- `config-example.yaml` — Bookmarks section expanded, descriptions removed
- `documentation/to-do.md` — Phase 5 deferred, Phase 6 marked complete, progress tracker updated

---

## [0.9.2] - 2026-06-19 - Scrollbar Theme Unification

### Changed
- **Todo widget scrollbar**: Added `scrollbar-width: thin; scrollbar-color: rgba(255,255,255,0.1) transparent` to match bookmarks, services, and docker widgets
- **Main page scrollbar (Firefox)**: `scrollbar-width: thin; scrollbar-color: rgba(255,255,255,0.15) transparent` on `#scroll-container`
- **Main page scrollbar (WebKit)**: Custom `::-webkit-scrollbar` with 6px width, transparent track, `rgba(255,255,255,0.15)` thumb (0.25 on hover), rounded corners

### Files Modified
- `templates/todo.html` — Scrollbar inline style on todo list container
- `static/index.html` — Global scrollbar CSS for `#scroll-container` (Firefox + WebKit)

---

## [0.9.1] - 2026-06-19 - Internal Scrolling & UI Refinements

### Changed
- **Bookmark icons/text scaled up**: `w-10 h-10` icon boxes with `w-5 h-5` SVGs, `text-[11px]` names, `p-2` padding, `gap-2` — better fills 5×2 grid
- **Internal scrolling for bookmarks**: `max-h-[200px]` scrollable container (unchanged)
- **Internal scrolling for Monitored Services**: `max-h-[230px]` with thin scrollbar (shows ~5 items, scroll for more)
- **Internal scrolling for Docker Services**: `max-h-[230px]` with thin scrollbar (shows ~3 items, scroll for more)

### Added
- **5 more mock bookmarks** in config.yaml (11 total: Proxmox, Portainer, Grafana, Uptime Kuma, Pi-hole, Plex, Sonarr, Radarr, Prowlarr, Jellyfin, Bazarr)
- **3 more monitored services** in config.yaml (6 total: Personal Website, Nextcloud, PDF Tools, Uptime Kuma, Home Assistant, Vaultwarden)
- **3 more mock Docker containers** in main.go (5 total: nginx, pihole, portainer, plex, nextcloud)

### Files Modified
- `templates/bookmarks.html` — Larger icons/text, adjusted max-height
- `templates/status.html` — Internal scroll for services and docker lists
- `config.yaml` — Additional mock bookmarks and services
- `main.go` — Additional mock Docker containers

---

## [0.9.0] - 2026-06-18 - Phase 4: Custom Links & Web Bookmarks

### Added
- **Bookmarks feature**: Configurable web bookmarks organized into groups
- **Bookmark config structures** (`internal/config/config.go`): `BookmarkGroup` and `BookmarkLink` structs with name, URL, icon, description, new_tab options
- **Bookmarks store** (`internal/bookmarks/store.go`): Processes bookmark groups, resolves icons, fetches and caches favicons
- **Icon support**: Three modes - Lucide icon name (SVG), custom image path, or auto-fetched favicon
- **Favicon caching**: Downloads and caches favicons from bookmark URLs to `data/icons/` with MD5-hashed filenames
- **Favicon endpoint**: `/bookmarks/icons/` serves cached favicon files

### Changed
- **Combined Bookmarks + Services widget**: Bookmarks (left) and Monitored Services (right) merged into a single col-span-2 card with vertical divider
- **Flat bookmark layout**: Removed group headers; all links displayed in a single flat section regardless of configured groups
- **Compact 5-column grid**: `sm:grid-cols-5` on desktop, `grid-cols-3` on mobile, with tighter spacing (`gap-1.5`, `p-1.5`)
- **Internal scrolling**: Bookmark container scrolls internally when exceeding ~10 items (`max-h-[180px]`, `overflow-y-auto`)
- **Compact service list**: Reduced padding, smaller status dots, smaller text for tighter layout
- **VM/LXC unified colors**: VM sub-card background changed to teal to match LXC; icon colors remain distinct (orange VM, teal LXC)
- **Responsive**: Side-by-side on desktop, stacked with horizontal divider on mobile

### Files Created
- `internal/bookmarks/store.go` — Bookmark processing and favicon caching
- `templates/bookmarks.html` — Bookmarks UI template (partial, embedded in combined card)

### Files Modified
- `internal/config/config.go` — `BookmarkGroup`, `BookmarkLink`, `Bookmarks` field
- `main.go` — `bookmarkStore` init, `BookmarkGroups` in DashboardData, favicon endpoint, template parsing
- `templates/status.html` — Combined Bookmarks + Services card, VM background color unified to teal
- `templates/bookmarks.html` — Flat list, 5-col grid, internal scroll, no group headers
- `config-example.yaml` — Bookmarks section with examples
- `config.yaml` — Sample bookmarks for testing

---

## [0.8.3] - 2026-06-18 - Proxmox Layout Consolidation & LXC/VM Monitoring

### Added
- **Virtualization Overview card**: New LXC/VM monitoring widget in Proxmox section showing active/total counts (e.g., "5/7" for LXC, "2/3" for VMs)
- **Proxmox API integration** (`internal/proxmox/client.go`): `GetVirtualization()` fetches QEMU VM and LXC container lists from `/nodes/{node}/qemu` and `/nodes/{node}/lxc` endpoints
- **Mock data**: `getMockVirtualization()` returns realistic test data (5/7 LXC, 2/3 VM) for development without a live Proxmox instance
- **VM/LXC sub-card backgrounds**: Each VM and LXC section now has its own colored background (orange tint for VMs, teal tint for LXC) with subtle borders for better visual separation

### Changed
- **CPU + Memory merged**: Combined into a single "CPU & Memory" card with dual progress bars, reducing vertical space usage
- **CPU & Memory label alignment**: Title now uses simple `<p>` like other cards, with CPU info (model name, cores/threads) on a separate line below for better vertical alignment with Virtualization and Disk cards
- **Proxmox grid layout**: Now 3 columns: CPU & Memory | Virtualization | Disk (was: CPU | Memory | Disk)

### Files Modified
- `internal/proxmox/client.go` — `VirtualizationInfo` struct, `GetVirtualization()`, `fetchResourceList()`, mock data
- `main.go` — `VirtInfo` field in `DashboardData`, fetch virtualization in `statusHandler`
- `templates/status.html` — Combined CPU+Memory card, new Virtualization card with icons

---

## [0.8.2] - 2026-06-17 - Media Services Integration, Widget Row Redesign, Live Badge

### Added
- **Media Services monitoring** (`internal/mediaservices/client.go`):
  - Sonarr: fetches series count and wanted count via `/api/v3/series` and `/api/v3/wanted/missing`
  - Radarr: fetches movie count and wanted count via `/api/v3/movie` and `/api/v3/wanted/missing`
  - Overseerr: fetches pending request count and available media count via `/api/v1/request` and `/api/v1/media`
  - All services use `X-Api-Key` header authentication with 5s timeout and graceful failure (`Online: false`)
  - Polled every 30 seconds via `pollMediaServices()` goroutine with mutex-protected shared state
  - `MockStats()` provides hardcoded test data when `proxmox.mock: true` and no services are configured
- **Live indicator pill badge**: "Live" text now wrapped in `px-3 py-1.5 rounded-full glass-inner` for consistent glass aesthetic
- **Widget card min-height + bigger text**: All top-row widgets set to `min-h-[190px]` with increased font sizes (time `text-2xl`, hostname `text-base`, network labels `text-xs`)

### Changed
- **Todo widget moved to widget row**: Replaces the Welcome (custom_text) card as the leftmost widget in the top row grid. Removed from standalone position between widget row and main grid.
- **Todo input overflow fix**: Added `min-w-0` to input to prevent native element intrinsic width from overflowing on mobile
- **Todo scrollable list**: Only the todo items list scrolls (`flex-1 min-h-0 overflow-y-auto max-h-[72px]`); the header and input field stay fixed. Card uses `flex flex-col h-full` to fill the grid row.
- **Todo limit**: Scroll area capped at `max-h-[72px]` (~2 visible items); 3rd+ items require scrolling
- **custom_text removed** from `combineWidgets()` in `main.go` — no longer rendered in the widget list
- **Inline Alpine.js component**: Todo `x-data` is now a JS object literal instead of `x-data="todoApp()"`, avoiding the script-evaluation issue in `data-preserve` divs during merge-swap
- **Widget text sizes increased**: Weather/time (time `text-2xl`, temp `text-xl`), system (hostname `text-base`, values `text-sm`), network (label `text-xs`, speeds `text-xs`, dot `h-3 w-3`)
- **Widget cards now use `flex flex-col justify-between`**: Weather/time, system info, and network cards stretch content to fill the card height

### Files Created
- `internal/mediaservices/client.go` — Sonarr/Radarr/Overseerr API clients
- `templates/mediaservices.html` — Media management card template

### Files Modified
- `internal/config/config.go` — `MediaServices []MediaServiceConfig` config field
- `main.go` — `mediaStats` + mutex, `pollMediaServices()` goroutine, `combineWidgets()` no longer includes custom_text
- `templates/todo.html` — Inline x-data, scrollable list with `max-h-[72px]`, `min-w-0` on input, `min-h-[190px]`
- `templates/widgets/widgets.html` — Todo rendered first in grid, removed custom_text section, bigger text/icons, `min-h-[190px]`
- `templates/status.html` — Removed standalone `{{ template "todo.html" . }}`, media services in grid
- `static/index.html` — Live indicator with `glass-inner` pill, adjusted styles
- `config-example.yaml` — Added `media_services` section

---

## [0.8.1] - 2026-06-17 - Widget Scroll & Sizing Fixes

### Fixed
- **Todo 3rd item disappearing**: Scroll area capped at `max-h-[72px]` (~2 visible items). Items beyond the 2nd require scrolling (usable scrollbar). Previously the `flex-1` container would expand to match content, but the grid row height constraint could clip items.
- **Mobile input overflow**: Added `min-w-0` to the todo `<input>` with `flex-1` — native input elements have an intrinsic minimum width that prevents shrinking in tight flex layouts.
- **Inconsistent widget heights**: All top-row widget cards now share `min-h-[190px]`. Grid's `align-items: stretch` makes same-row cards match.

### Changed
- `min-h-[160px]` → `min-h-[190px]` on all widget cards (weather, system, network)
- Todo card: `min-h-[190px]` added back (was removed), scroll area `max-h-[72px]`
- Weather/time, system, network cards: increased font sizes to fill the taller cards

---

## [0.8.0] - 2026-06-17 - Widget Row & Media Services

### Added
- **Media Services monitoring** (Sonarr, Radarr, Overseerr) with clickable WebUI links and stat boxes
- **Media Management card** in main grid with per-service status indicators

### Changed
- **Todo widget moved to widget row** replacing the Welcome (custom_text) card — leftmost position in the top 4-card grid
- **custom_text removed** from `combineWidgets()` — no longer rendered
- **Standalone todo** removed from `status.html` (was between widget row and main grid)
- **Desktop width**: `max-w-6xl mx-auto lg:max-w-none` — no width constraint on desktop, fills screen
- **Card min-heights**: `md:min-h-[320px]` on Monitored Services (col-span-2) and Docker Containers (col-span-1)
- **Media services** inside main grid (after Proxmox, col-span-3) with proper `gap-6`
- **Inline Alpine x-data**: Todo component defined as JS object literal instead of external `todoApp()` function (scripts inside `data-preserve` divs are never executed by merge-swap)
- **Optimistic updates**: Todo add/toggle/delete immediately reflect in UI, API syncs in background

### Fixed
- **`:key` on `<template>` element**: Moved to child `<div>` — Alpine.js 3 requires `:key` on the first child, not on `<template>` itself. This was the root cause of todo add not working.
- **Race condition in todo add**: `init()`→`refresh()` (GET /api/todos) was overwriting POST response. Removed `init()`/`refresh()` entirely — data comes from Go template, no race condition.

---

## [0.7.2] - 2026-06-17 - Todo Reactivity Fix, Multi-Disk, CPU Name, Mobile Fixes

### Fixed
- **Todo add not working** — Root cause: Alpine.js reactivity does not reliably detect `Array.push()` (in-place mutation). Changed to `Array.concat()` which creates a new array reference that Alpine.js reactivity detects. Also fixed toggle mutation using `Array.map()` instead of in-place mutation.
- **Mobile background image scrolling** — On mobile, `position: fixed` elements shift during overscroll rubber-band. Fixed by setting `html, body { overflow: hidden; height: 100%; }` and wrapping all content in a `#scroll-container` div with `overflow-y: auto` and `100dvh` height. Background moved from `body::before`/`body::after` pseudo-elements to real `<div>` elements for better mobile browser reliability.

### Added
- **CPU model name** in CPU widget alongside core/thread count (e.g., "Intel Core i7-12700K" above "12C / 20T"). `cleanCPUName()` strips verbose suffixes like " with Radeon Graphics", " CPU @ X.XXGHz", "-Core Processor".
- **Multi-disk support** — `NodeStatus.Disk` replaced with `NodeStatus.Disks []DiskInfo` with mountpoint, total, used per disk. Real Proxmox API fetches disk list from `/nodes/{node}/disks/list` endpoint. Mock mode returns 3 disks (`/`, `/mnt/storage`, `/mnt/backup`).
- **Scrollable disk widget** — Disk area now has `max-h-[210px] overflow-y-auto` with thin scrollbar. Multiple disks scroll internally without expanding the card.
- **`roundDur` template function** — Response times now display as clean "150 ms" or "1.23 s" instead of Go's default verbose "150.123456ms".

### Changed
- `templates/todo.html`: Removed duplicate `@click.prevent` and `@keydown.enter.prevent` handlers — form's `@submit.prevent` handles both naturally. In-place mutations replaced with immutable array operations for Alpine.js compatibility.
- `templates/widgets/widgets.html`: Network interface display restructured to vertical stacking — each interface shows name/label on top, speed (↓ ↑) indented below. Prevents horizontal squeeze on mobile 2-column grid.
- `templates/status.html`: Response time display uses `{{ roundDur .ResponseTime }}` instead of raw `{{ .ResponseTime }}`.
- `static/index.html`: Background moved to real `<div>` elements (`#bg-image`, `#bg-overlay`), `body` padding moved to `#scroll-container`, added `overscroll-behavior: none` on `html`.

### Files Modified
- `internal/proxmox/client.go` — `DiskInfo` struct, `Disks` field, `fetchDiskList()`, `cleanCPUName()`, mock returns 3 disks
- `templates/todo.html` — Immutable array ops + simplified event handlers
- `templates/status.html` — CPU model name, multi-disk scrollable widget, `roundDur` function
- `templates/widgets/widgets.html` — Network card vertical stack layout
- `static/index.html` — `#scroll-container` scroll model, real background divs
- `main.go` — `roundDur` template function added
- All documentation files updated

---

## [0.7.1] - 2026-06-16 - Todo Widget Bug Fix & Compact Redesign

### Fixed
- **Cannot add new todos** — Root cause: the merge-swap DOM diff was patching the todo card's `<script>` tag content on each HTMX poll, disrupting Alpine.js state. Fixed by adding `data-preserve` attribute to the todo card, which tells the merge-swap to skip the element entirely (including its children and script).

### Changed
- `static/index.html`: Added `data-preserve` check in `patchChildren()` — elements with this attribute are skipped during DOM diffing
- `templates/todo.html`: Added `data-preserve` attribute to todo card root element
- **Compact redesign**: `p-6` → `p-4`, `mb-6` → `mb-4`, `text-xl` → `text-base`, `text-sm` → `text-xs`, `py-2` → `py-1.5`, checkbox `w-5` → `w-4`, icons smaller, `space-y-2` → `space-y-1`, tighter spacing throughout

### Files Modified
- `static/index.html` — `data-preserve` skip logic in merge-swap
- `templates/todo.html` — Preserve attribute + compact styling
- All documentation files updated

---

## [0.7.0] - 2026-06-16 - Interactive To-Do List & CPU Core/Thread Display

### Added
- **Interactive To-Do List widget** — Users can add, check off, and delete tasks directly on the dashboard. Persists to `data/todos.json`.
  - `internal/todo/store.go`: Thread-safe CRUD store with JSON file persistence (`NewStore`, `GetAll`, `Add`, `Toggle`, `Delete`)
  - `templates/todo.html`: Alpine.js-powered interactive UI with add form, checkboxes, delete buttons, done counter, sorted display (active first, then done)
  - API endpoints: `GET /api/todos` (list), `POST /api/todos` (add), `PUT /api/todos/{id}` (toggle), `DELETE /api/todos/{id}` (delete)
  - Config: `todos.enabled`, `todos.file_path`, `todos.title`
- **CPU core/thread count** displayed in Proxmox CPU card (e.g., "8C / 16T")
  - `proxmox.CPUInfo` struct with `ModelName`, `Cores`, `Threads`
  - `ReadLocalCPUInfo()` parses `/proc/cpuinfo` for physical cores and logical threads
  - Mock mode provides simulated CPU info (i7-12700K, 12C/20T)
- **Alpine.js** (3.x CDN) for interactive todo widget

### Changed
- `internal/config/config.go`: Added `TodoConfig` struct with defaults
- `main.go`: Todo store init, API routes, `DashboardData` fields, `localCPUInfo` read at startup
- `templates/status.html`: Todo template inclusion, CPU card shows core/thread count
- `.gitignore`: Added `data/` directory
- `config.yaml` / `config-example.yaml`: Added `todos` section

### Files Added
- `internal/todo/store.go` — Persistent to-do store
- `templates/todo.html` — Interactive todo template

### Files Modified
- `internal/proxmox/client.go` — CPUInfo struct, ReadLocalCPUInfo()
- `internal/config/config.go` — TodoConfig
- `main.go` — Todo integration, CPU info, API handlers
- `templates/status.html` — Todo inclusion, CPU cores
- `static/index.html` — Alpine.js CDN
- `config.yaml`, `config-example.yaml` — Todos section
- `.gitignore` — data/ directory
- All documentation files updated

---

## [0.6.0] - 2026-06-16 - Backdrop-Flicker Elimination via DOM Diff Swap

### Fixed
- **Persistent backdrop-filter flickering on data refresh** — Replaced HTMX's default `innerHTML` swap with a custom `merge-swap` extension that performs in-place DOM diffing. Instead of destroying and recreating all glass-card elements every 5 seconds, only text nodes and dynamic attributes (class, style, aria-valuenow) are updated. This preserves the browser's GPU compositing layers for `backdrop-filter: blur()`, eliminating the flicker entirely.

### Added
- **Custom HTMX swap extension** (`merge-swap`) in `static/index.html`:
  - `mergeDOM()` — recursive tree walker that patches current DOM against new server HTML
  - `patchChildren()` — filters blank text nodes, compares node types/tags, updates or replaces as needed
  - `syncAttrs()` — syncs class, style, aria-valuenow, aria-label, role attributes
  - Falls back to normal `innerHTML` on first load (skeleton → first render)

### Changed
- `static/index.html`: Added `hx-ext="merge-swap"` to `#dashboard-content` div, 100 lines of custom swap JS

### Technical Details
- First load: skeleton → full render uses standard `innerHTML` (no glass-cards exist yet)
- Subsequent polls: DOM diff only touches text nodes and changed attributes
- Script elements stay in DOM (not recreated), so client-side clock interval is not duplicated
- Conditional rendering (service up/down transitions) correctly replaces the changed subtree

### Files Modified
- `static/index.html` — Custom merge-swap extension
- All documentation files updated

---

## [0.5.3] - 2026-06-16 - Network Display Caching & Alignment Fix

### Fixed
- **Network speed flickering** — `GetStats()` now caches formatted display output for 10 seconds (`displayTTL`), preventing speed values from changing on every HTMX 5-second poll. Raw sampling continues at 3s interval for accuracy, but displayed strings stay stable.
- **Network card text misalignment on mobile** — Increased spacing between interface name and speed values: `space-x-1.5` → `space-x-2`, added `ml-3` minimum gap to speed container, `flex-shrink-0` on dots and speed text to prevent squishing, `min-w-0` on name container for proper truncation, space added between arrow and value (`↓ 1.23 Mbit/s`).

### Changed
- `internal/network/monitor.go`: Added `displayMu`, `cachedStats`, `statsCacheAt`, `displayTTL` fields to `Monitor` struct. `GetStats()` returns cached output within 10-second window.
- `templates/widgets/widgets.html`: Network card alignment improved with better spacing and flex constraints

### Files Modified
- `internal/network/monitor.go` — Display cache with 10s TTL
- `templates/widgets/widgets.html` — Network card alignment fixes
- All documentation files updated

---

## [0.5.2] - 2026-06-16 - Widget Stability & Layout Fixes

### Fixed
- **Mock weather randomizing every HTMX poll** — `mockData()` now caches result for 5 minutes via `mockCache` struct, preventing weather conditions from changing every 5 seconds during auto-refresh
- **Date format inconsistency** — Server-rendered date now includes weekday (`"Monday, January 2, 2006"`) to match client-side JS clock format (`weekday:'long'`), eliminating visible format jump
- **Backdrop-filter flickering on widget swap** — Root cause was data instability from above bugs causing different HTML on every HTMX swap; fixing data caching eliminates most visible flicker

### Changed
- **Network monitor moved to widget row** — Compact network summary card added as 4th widget in top row (alongside custom_text, weather_time, system_info)
- **Widget grid updated** to `grid-cols-2 lg:grid-cols-4` — Mobile: perfect 2x2 grid (4 widgets); Desktop: 4-column row
- **Network summary card** shows per-interface status with live RX/TX speeds in compact format
- Removed full network card (`{{ template "network.html" . }}`) from bottom grid section

### Files Modified
- `internal/widgets/weather.go` — Added `mockCache` struct with 5-minute TTL for mock data
- `internal/widgets/datetime.go` — Date format changed to include weekday
- `templates/widgets/widgets.html` — Grid updated to 4-col, network summary card added
- `templates/status.html` — Removed network template from bottom grid
- All documentation files updated

---

## [0.5.1] - 2026-06-16 - Dashboard Layout Refinements

### Changed
- **Weather + DateTime combined** into a single compact `weather_time` card via `combineWidgets()` in `main.go`
  - Time shown at top (with live client-side clock), weather below a divider
  - Saves vertical space by eliminating a full card
- **Custom Text widget moved to left** (first position in widget row) with compact card styling
- **Widget grid compacted** from `grid-cols-1 sm:grid-cols-2 lg:grid-cols-4` to `grid-cols-2 lg:grid-cols-3`
  - Mobile: 2-column grid (3 widgets fit without covering monitoring cards below)
  - Desktop: 3-column grid (custom_text, weather_time, system_info)
- **All widget cards reduced padding** from `p-5` to `p-4`, font sizes reduced for compactness
- **Network Monitor repositioned** from below Proxmox metrics to below Monitored Services + Docker Containers
- **Standalone fallbacks**: Weather and DateTime still render individually if only one is enabled

### Fixed
- Clock JS `dateEl` null check added (prevents error when `widget-date` element doesn't exist)

### Files Modified
- `main.go` — Added `combineWidgets()` function, widget data post-processing in `statusHandler`
- `templates/widgets/widgets.html` — Full rewrite: combined weather_time card, compact styling, 2-col mobile grid
- `templates/status.html` — Moved `{{ template "network.html" . }}` below Services + Docker sections
- `documentation/changelogs.md` — This entry
- `documentation/prompt-history.md` — Session 11
- `documentation/docs.md` — Updated widget layout description

---

## [0.5.0] - 2026-06-16 - Network Monitoring (Phase 3 Complete)

### Added
- **Network Package** (`internal/network/`)
  - `types.go`: `InterfaceStats` struct with speed, total, and human-readable formatted fields
  - `monitor.go`: Background sampling goroutine with configurable interval

- **`/proc/net/dev` Parser** (`internal/network/monitor.go`)
  - `readProcNetDev()` parses Linux kernel interface byte counts
  - Skips header lines, handles malformed data gracefully
  - Returns map of interface name to byte counts

- **Speed Calculation**
  - Two-sample rate calculation: `rate = (current_bytes - previous_bytes) / elapsed_seconds`
  - Moving average smoothing over last 3 samples
  - Human-readable formatting: b/s, Kbit/s, Mbit/s, Gbit/s
  - Human-readable totals: KB, MB, GB, TB

- **Mock Mode**
  - Simulates network traffic with random increments (0.5-2.5 MB RX, 0.1-0.6 MB TX per sample)
  - Enables UI testing without real network interfaces

- **Network Config** (`internal/config/config.go`)
  - `NetworkConfig` struct: enabled, interfaces list, show_speed, show_total_transfer, update_interval, mock
  - `NetIfConfig` struct: interface name + human-friendly label
  - Default update interval: 3 seconds

- **Network Template** (`templates/network.html`)
  - Responsive grid: 1 col (mobile) → 2 cols (md) → N cols (lg) based on interface count
  - Per-interface card: name, label, up/down status indicator
  - RX/TX speeds with directional arrows (↓ ↑) in blue/emerald
  - Cumulative total bytes transferred
  - Glassmorphism styling consistent with dashboard theme
  - ARIA labels on interface cards

### Changed
- `main.go`: Added `netMonitor` global, network monitor initialization from config
- `main.go`: `DashboardData` struct now includes `Network`, `NetShowSpeed`, `NetShowTotal`
- `main.go`: Template parsing now includes `templates/network.html`
- `templates/status.html`: Includes network template via `{{ template "network.html" . }}`
- `config.yaml` / `config-example.yaml`: Added `network` section

### Files Created
- `internal/network/types.go` — InterfaceStats struct + rawSample internal type
- `internal/network/monitor.go` — Network monitor with /proc/net/dev parsing, speed calculation, mock mode
- `templates/network.html` — Network interface cards template

### Files Modified
- `internal/config/config.go` — NetworkConfig + NetIfConfig structs + defaults
- `main.go` — Network monitor init, template parsing, DashboardData fields
- `templates/status.html` — Network template inclusion
- `config-example.yaml` — Network section with comments
- `config.yaml` — Network section (enabled with mock data)

---

## [0.4.0] - 2026-06-16 - Utility & Information Widgets (Phase 2 Complete)

### Added
- **Widgets Package** (`internal/widgets/`)
  - `widget.go`: `WidgetData` struct and `Widget` interface
  - `registry.go`: Widget registry with `Register()`, `FetchAll()`, `Count()`
  - Each widget implements `Name()`, `Type()`, and `Fetch()`

- **Weather Widget** (`internal/widgets/weather.go`)
  - Open-Meteo API integration (free, no API key required)
  - Fetches temperature, weather code, wind speed
  - WMO weather code mapping to emoji icons + descriptions
  - Configurable caching (default 15 minutes)
  - Celsius/Fahrenheit support
  - Mock mode for testing without API calls
  - 5-second HTTP timeout

- **DateTime Widget** (`internal/widgets/datetime.go`)
  - Configurable timezone via IANA names (`time.LoadLocation`)
  - 12h/24h format toggle
  - Client-side JavaScript clock (updates every second, no server polling)
  - Uses `Intl.DateTimeFormat` for timezone-aware client rendering

- **System Info Widget** (`internal/widgets/sysinfo.go`)
  - Hostname via `os.Hostname()`
  - OS name from `/etc/os-release` PRETTY_NAME
  - System uptime from `/proc/uptime` (formatted as days/hours/minutes)
  - Go runtime stats: goroutine count, allocated memory

- **Custom Text Widget** (`internal/widgets/custom_text.go`)
  - Configurable title and content from YAML
  - HTML content sanitized via `html.EscapeString` to prevent XSS

- **Widget Templates** (`templates/widgets/widgets.html`)
  - Responsive grid: 1 col (mobile) → 2 cols (sm) → 4 cols (lg)
  - Type-specific rendering via conditional blocks
  - Glassmorphism card styling matching Phase 1 theme
  - ARIA labels on all widget cards

- **Widget Config Structs** (`internal/config/config.go`)
  - `WidgetsConfig`, `WeatherWidgetConfig`, `DateTimeWidgetConfig`, `SystemInfoWidgetConfig`, `CustomTextWidgetConfig`
  - Per-widget `enabled` flag
  - Sensible defaults (15-min cache, celsius, Local timezone)

### Changed
- `main.go`: Added `widgetRegistry` global, widget initialization from config
- `main.go`: `DashboardData` struct now includes `Widgets`, `DateTime24h`, `DateTimezone`
- `main.go`: Template parsing now includes `templates/widgets/widgets.html`
- `templates/status.html`: Includes widgets template via `{{ template "widgets.html" . }}`
- `config.yaml` / `config-example.yaml`: Added `widgets` section

### Files Created
- `internal/widgets/widget.go` — Widget interface and data struct
- `internal/widgets/registry.go` — Widget registry manager
- `internal/widgets/weather.go` — Open-Meteo weather widget
- `internal/widgets/datetime.go` — Date/time widget with client-side clock
- `internal/widgets/sysinfo.go` — System information widget
- `internal/widgets/custom_text.go` — Custom text widget
- `templates/widgets/widgets.html` — Combined widget template with responsive grid

### Files Modified
- `internal/config/config.go` — Widget config structs + defaults
- `main.go` — Widget registry init, template parsing, DashboardData fields
- `templates/status.html` — Widget template inclusion
- `config-example.yaml` — Widgets section with comments
- `config.yaml` — Widgets section (all enabled with mock data)

---

## [0.3.0] - 2026-06-16 - Visual Enhancements (Phase 1 Complete)

### Added
- **Appearance Config System** (`internal/config/config.go`)
  - New `AppearanceConfig` struct with fields: `background_image`, `background_url`, `background_opacity`, `background_blur`, `theme`, `card_opacity`, `card_blur`, `accent_color`
  - Sensible defaults applied automatically when fields are omitted
  - Full backward compatibility — old configs without `appearance` section still work

- **Custom Background Image** (`static/index.html`, `static/backgrounds/`)
  - Support for local file paths and remote URLs
  - Dark overlay with configurable opacity
  - CSS blur filter with configurable intensity
  - `/api/background` JSON endpoint
  - `/background` endpoint that reads local image files from disk and serves them via HTTP with proper content-type and cache headers
  - `static/index.html` is now a Go template for dynamic rendering

- **Glassmorphism UI** (`static/index.html`, `templates/status.html`)
  - `glass-card` and `glass-inner` CSS classes replacing solid `bg-gray-800`
  - `backdrop-filter: blur()` on all cards
  - Semi-transparent borders with `rgba(255,255,255,0.1)`
  - Hover effect: `translateY(-2px)` + glow shadow

- **Typography Improvements**
  - Inter font loaded via Google Fonts CDN with `display=swap`
  - Font stack: `Inter, system-ui, -apple-system, sans-serif`
  - `.metric-label` class: uppercase, letter-spacing, muted color
  - `.metric-value` class: tight letter-spacing, bold weight
  - `tabular-nums` for numeric values (no jitter)

- **Animations & Transitions**
  - Smooth card hover transitions (200ms ease)
  - HTMX swap transitions: fade-out (180ms) + fade-in (250ms)
  - `live-pulse` keyframe animation for Live indicator
  - Loading skeleton shimmer replacing spinner
  - Progress bar transitions with cubic-bezier easing
  - `prefers-reduced-motion` respected (all animations disabled)

- **Accessibility (WCAG 2.1 AA)**
  - `aria-label` on all meter widgets (CPU, RAM, Disk)
  - `aria-hidden="true"` on decorative icons and progress bars
  - `aria-live="polite"` on dashboard content region
  - Visible `focus-visible` rings on all interactive elements
  - `tabindex="0"` on service and container items
  - Status badges have text labels (not color-only)
  - `role="status"` on live indicator and loading skeleton

### Changed
- `main.go`: `index.html` now served via Go template engine (`indexHandler`), not plain static file
- `main.go`: New `/api/background` endpoint returning JSON config
- `main.go`: New `/background` endpoint that reads local image files from disk and serves them with correct MIME type and 1-hour cache
- `main.go`: Static file server scoped to non-index paths only
- `config.yaml` / `config-example.yaml`: Added `appearance` section

### Fixed
- Background image not displaying when using local file paths (e.g. `image.png`) — CSS `url()` cannot reference filesystem paths directly; now routed through `/background` HTTP handler
- Glassmorphism card backdrop-filter flickering on hover — fixed by forcing GPU compositing layer with `translateZ(0)` and `will-change: transform`

### Files Modified
- `internal/config/config.go` — `AppearanceConfig` struct + `setDefaults()`
- `main.go` — `indexHandler`, `backgroundHandler`, `backgroundServeHandler`, `indexTmpl` variable
- `static/index.html` — Full rewrite as Go template with CSS variables, glassmorphism, accessibility
- `templates/status.html` — Replaced solid cards with `glass-card`/`glass-inner`, added ARIA
- `config-example.yaml` — Added appearance section with comments
- `config.yaml` — Added appearance section with Unsplash background URL
- `static/backgrounds/` — New directory for custom background images

---

## [0.2.0] - 2026-06-16 - Project Rebrand

### Changed
- **Full project rebrand** from "Selfhosted Proxmox Dashboard" to **dhiarhome**
- Go module name: `proxmox-dashboard` -> `dhiarhome`
- All import paths updated to `dhiarhome/internal/...`
- Docker image name: `homelab-dash` -> `dhiarhome`
- Container name: `homelab-dashboard` -> `dhiarhome`
- Binary name: `dashboard` -> `dhiarhome`
- UI header: `HomelabDash` -> `dhiarhome`
- Page title: "Proxmox Dashboard" -> "dhiarhome"
- Page subtitle updated to "Lightweight homelab monitoring dashboard"
- Systemd service name: `homelab-dashboard` -> `dhiarhome`
- All cross-compile binary outputs renamed (e.g., `dhiarhome-arm64`, `dhiarhome.exe`)
- GitHub repo URL updated to `github.com/Alfar0nt/dhiarhome`

### Added
- `documentation/` folder with comprehensive docs:
  - `docs.md` - Full project documentation
  - `deployment.md` - Deployment guide (Docker + bare metal)
  - `to-do.md` - Feature implementation roadmap (33 steps)
  - `prompt-history.md` - Conversation log
  - `changelogs.md` - This file
- README.md rewritten with improved structure:
  - Feature list, tech stack, quick start guide
  - Configuration examples
  - Roadmap section linking to to-do.md
  - Links to all documentation files

### Files Modified
- `go.mod` - Module name
- `main.go` - Import paths
- `static/index.html` - Title, header, subtitle
- `Dockerfile` - Binary name in build, copy, and CMD
- `.gitignore` - Added `dhiarhome` binary entry
- `README.md` - Full rewrite
- `documentation/docs.md` - Project name and directory structure
- `documentation/deployment.md` - All 30+ name references
- `documentation/to-do.md` - Project name in title/overview

---

## [0.1.0] - Pre-Rebrand (Original State)

### Project Name
"Selfhosted Proxmox Dashboard" (Go module: `proxmox-dashboard`)

### Features
- **Proxmox Server Monitoring**
  - CPU usage percentage
  - Memory usage (used/total with GB display)
  - Disk usage (root filesystem)
  - Uptime tracking
  - Mock mode with random realistic data

- **Docker Container Monitoring**
  - Lists all containers via Docker Engine API
  - Shows container state (running/exited/stopped)
  - Container status and uptime display
  - Optional container filtering by name

- **Web Service Health Checks**
  - HTTP/HTTPS endpoint monitoring
  - Response time tracking
  - Status indicators (Online/Offline/Warning)
  - Configurable service list
  - Background polling every 10 seconds

- **UI**
  - Dark mode design (Tailwind CSS slate-900 theme)
  - Auto-refreshing dashboard via HTMX (5-second polling)
  - Progress bars for CPU/memory/disk
  - Animated "Live" indicator
  - Responsive grid layout (mobile-friendly)
  - Loading spinner for initial data fetch
  - Smooth HTMX swap transitions

- **Configuration**
  - YAML-based configuration (`config.yaml`)
  - Example config template (`config-example.yaml`)
  - Mock mode toggle for testing
  - No code changes needed for customization

### Tech Stack
- **Backend:** Go 1.26.3 (statically compiled, CGO_ENABLED=0, Linux/amd64)
- **Frontend:** HTML5 + Tailwind CSS (CDN) + HTMX 1.9.10
- **Dependencies:** `gopkg.in/yaml.v3` (YAML parsing only)
- **Deployment:** Multi-stage Docker build (golang:1.21-alpine -> alpine:latest)

### Architecture
- Single Go binary (~10MB)
- Zero database
- In-memory cache (thread-safe doubly-linked list, 100 entries max)
- Proxmox API client with TLS skip verify (self-signed certs)
- Docker API client supporting Unix socket and TCP endpoints
- HTTP monitor with 5-second timeout

### Project Structure (Original)
```
personalProject-Dashboard/
├── main.go
├── config.yaml
├── config-example.yaml
├── Dockerfile
├── go.mod
├── go.sum
├── .gitignore
├── Screenshot.png
├── README.md
├── internal/
│   ├── cache/history.go
│   ├── config/config.go
│   ├── docker/client.go
│   ├── monitor/http.go
│   └── proxmox/client.go
├── static/
│   └── index.html
└── templates/
    └── status.html
```

### Deployment Methods
- Docker build and run
- Docker Compose
- Bare metal (build from source)
- Systemd service
- Cross-compile for ARM64, ARM, Windows, macOS

### Known Limitations
- Single Proxmox node per instance
- No historical data or graphs
- No authentication mechanism
- No alert notifications
- HTTP-only service checks
- No HTTPS for the dashboard itself

---

## Version History Summary

| Version | Date | Description |
|---------|------|-------------|
| 0.1.0 | Pre-2026-06-16 | Original "Selfhosted Proxmox Dashboard" with core features |
| 0.2.0 | 2026-06-16 | Rebrand to "dhiarhome" + documentation system |
| 0.3.0 | 2026-06-16 | Visual enhancements: glassmorphism, background, animations, accessibility |
| 0.4.0 | 2026-06-16 | Utility widgets: weather, datetime, system info, custom text |
| 0.5.0 | 2026-06-16 | Network monitoring: /proc/net/dev parsing, speed calculation, interface cards |
| 0.5.1 | 2026-06-16 | Layout refinements: combined weather+time card, compact mobile grid, network repositioned |
| 0.5.2 | 2026-06-16 | Widget stability: mock weather caching, date format fix, network in widget row (2x2 mobile) |
| 0.5.3 | 2026-06-16 | Network display caching (10s TTL), mobile alignment fix |
| 0.6.0 | 2026-06-16 | Backdrop-flicker elimination: custom DOM diff swap extension |
| 0.7.0 | 2026-06-16 | Interactive to-do list (Alpine.js), CPU core/thread display |
| 0.7.1 | 2026-06-16 | Todo add bug fix (data-preserve), compact redesign |
| 0.7.2 | 2026-06-17 | Todo reactivity fix, CPU model name, multi-disk support, mobile bg scroll fix, network card vertical layout, response time formatting |
| 0.8.0 | 2026-06-17 | Media services (Sonarr/Radarr/Overseerr), todo moved to widget row, desktop layout, inline Alpine x-data, optimistic todo updates |
| 0.8.1 | 2026-06-17 | Widget scroll fix (max-h 72px), mobile input overflow fix (min-w-0), consistent min-h-[190px], bigger widget text/icons |
| 0.8.2 | 2026-06-17 | Live indicator glass pill, media services polling goroutine, Flex layout for widget cards, responsive sizing |
| 0.9.0 | Planned | Bookmarks and custom links |
| 0.10.0 | 2026-06-19 | Phase 6: graceful shutdown, config validation, Dockerfile hardening |
| 0.10.1 | 2026-06-19 | Security hardening: headers, rate limiting, path traversal, input validation |
| 1.0.0 | 2026-06-19 | First stable release with all planned features |
