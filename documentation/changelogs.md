# Changelogs - dhiarhome

All notable changes to this project are documented in this file.

---

## [Unreleased] - Pre-Feature Roadmap

### Planned
- Custom bookmarks and web links with icon support
- Service integration framework (Plex, Radarr, Sonarr, Portainer)
- Generic HTTP API widget for custom services

See [to-do.md](to-do.md) for the full 33-step implementation plan.

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
| 0.8.0 | Planned | Bookmarks and custom links |
| 0.9.0 | Planned | Service integration framework |
| 1.0.0 | Planned | First stable release with all planned features |
