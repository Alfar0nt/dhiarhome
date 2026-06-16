# Changelogs - dhiarhome

All notable changes to this project are documented in this file.

---

## [Unreleased] - Pre-Feature Roadmap

### Planned
- Network interface monitoring (speed, RX/TX, connections)
- Custom bookmarks and web links with icon support
- Service integration framework (Plex, Radarr, Sonarr, Portainer)
- Generic HTTP API widget for custom services

See [to-do.md](to-do.md) for the full 33-step implementation plan.

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
| 0.5.0 | Planned | Network monitoring |
| 0.6.0 | Planned | Bookmarks and custom links |
| 0.7.0 | Planned | Service integration framework |
| 1.0.0 | Planned | First stable release with all planned features |
