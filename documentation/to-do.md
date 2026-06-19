# dhiarhome - Feature Implementation To-Do

## Overview
Step-by-step implementation plan to transform dhiarhome into a comprehensive homelab monitoring solution inspired by `gethomepage/homepage`. Each step is a discrete, trackable task.

> **Status Legend:**
> - `[ ]` Not started
> - `[~]` In progress
> - `[x]` Complete
> - `[!]` Blocked / Issue

---

## Phase 1: Visual Enhancements

### Step 1.1 — Extend Config with Appearance Settings
- [x] Add `AppearanceConfig` struct to `internal/config/config.go`
  ```go
  type AppearanceConfig struct {
      BackgroundImage   string  `yaml:"background_image"`
      BackgroundURL     string  `yaml:"background_url"`
      BackgroundOpacity float64 `yaml:"background_opacity"`
      BackgroundBlur    int     `yaml:"background_blur"`
      Theme             string  `yaml:"theme"`
      CardOpacity       float64 `yaml:"card_opacity"`
      CardBlur          int     `yaml:"card_blur"`
      AccentColor       string  `yaml:"accent_color"`
  }
  ```
- [x] Add `Appearance AppearanceConfig` field to the main `Config` struct
- [x] Set sensible defaults when fields are omitted (opacity 0.3, blur 5, theme "dark")
- [x] Add appearance section to `config-example.yaml`
- [x] Test: verify config loads with and without the new section (backward compat)

### Step 1.2 — Implement Custom Background Image
- [x] Create `static/backgrounds/` directory with a default dark gradient image
- [x] Modify `static/index.html` to:
  - Render background image from config via a Go template variable
  - Apply CSS `background-size: cover; background-position: center`
  - Add dark overlay div with configurable opacity
  - Add optional CSS blur filter
- [x] Add a `/api/background` endpoint (or pass via template) to serve the configured background path
- [x] Support both local file paths and remote URLs
- [x] Test: load dashboard with a custom local image, verify overlay and blur work

### Step 1.3 — Glassmorphism Card Styling
- [x] Define CSS variables in `static/index.html` `<style>` block:
  ```css
  :root {
    --card-bg: rgba(30, 41, 59, 0.6);
    --card-border: rgba(255, 255, 255, 0.1);
    --card-blur: 12px;
    --text-primary: #f1f5f9;
    --text-secondary: #cbd5e1;
    --accent-blue: #3b82f6;
    --accent-green: #10b981;
    --accent-red: #ef4444;
  }
  ```
- [x] Update card classes in `templates/status.html`:
  - Replace solid `bg-gray-800` with `var(--card-bg)` + `backdrop-filter: blur()`
  - Add semi-transparent borders
  - Add subtle inner shadow
- [x] Add hover effect: `translateY(-2px)` + glow shadow
- [x] Ensure text remains readable over any background (test with light/dark images)
- [x] Test: verify glass effect renders in Chrome, Firefox, Safari

### Step 1.4 — Typography & Spacing Improvements
- [x] Add Inter font via Google Fonts CDN (with `display=swap`)
- [x] Set base font family to `Inter, system-ui, -apple-system, sans-serif`
- [x] Increase card padding from `p-6` to `p-6` with improved internal spacing
- [x] Use `tracking-tight` for headings, `tracking-normal` for body text
- [x] Add consistent spacing scale (4px base unit)
- [x] Improve metric label typography (smaller, uppercase, letter-spacing)
- [x] Test: compare before/after readability

### Step 1.5 — Smooth Animations & Transitions
- [x] Add CSS transitions to all interactive elements (cards, buttons, status badges)
- [x] Add HTMX swap transition: fade-out old content, fade-in new content
- [x] Add subtle pulse animation to "Live" indicator
- [x] Add loading skeleton/shimmer for initial data fetch
- [x] Respect `prefers-reduced-motion` media query (disable animations)
- [x] Test: verify animations feel smooth at 60fps

### Step 1.6 — Accessibility Audit (Visual Layer)
- [x] Check all text meets WCAG 2.1 AA contrast ratio (4.5:1 minimum)
- [x] Add visible focus rings for keyboard navigation
- [x] Ensure status indicators have text labels (not just color)
- [x] Add `aria-label` attributes to icon-only elements
- [x] Test with screen reader (basic navigation check)
- [x] Test: run Lighthouse accessibility audit, target score >90

---

## Phase 2: Utility & Information Widgets

### Step 2.1 — Create Widgets Package Structure
- [x] Create `internal/widgets/` directory
- [x] Create `internal/widgets/widget.go` with base interfaces:
  ```go
  type WidgetData struct {
      Type    string
      Label   string
      Values  map[string]interface{}
      Icon    string
  }
  ```
- [x] Create `internal/widgets/registry.go` to manage enabled widgets
- [x] Add `WidgetsConfig` struct to `internal/config/config.go`:
  ```go
  type WidgetsConfig struct {
      Weather    WeatherWidgetConfig    `yaml:"weather"`
      DateTime   DateTimeWidgetConfig   `yaml:"datetime"`
      SystemInfo SystemInfoWidgetConfig `yaml:"system_info"`
      CustomText CustomTextWidgetConfig `yaml:"custom_text"`
  }
  ```
- [x] Add widgets section to `config-example.yaml`
- [x] Test: verify config loads correctly

### Step 2.2 — Implement Weather Widget (Open-Meteo)
- [x] Create `internal/widgets/weather.go`
- [x] Implement Open-Meteo API client:
  - Endpoint: `https://api.open-meteo.com/v1/forecast?latitude=X&longitude=Y&current=temperature_2m,weather_code`
  - No API key required
  - Parse temperature, weather code, wind speed
- [x] Map weather codes to icons and descriptions (WMO code table)
- [x] Implement caching: store last response + timestamp, refresh every N minutes
- [x] Add mock weather data for testing (`mock: true`)
- [x] Create `templates/widgets/widgets.html` template:
  - Show temperature, weather icon, condition label
  - Match Homepage screenshot style (e.g., "Current, 13.1°F Clear")
- [x] Test: fetch real data from Open-Meteo, verify caching works

### Step 2.3 — Implement Date/Time Widget
- [x] Create `internal/widgets/datetime.go`
- [x] Use Go `time` package with configurable timezone (`time.LoadLocation`)
- [x] Support 12h and 24h format via config
- [x] Format: time, day of week, full date
- [x] Create `templates/widgets/widgets.html` template
- [x] For live updates: use client-side JavaScript (lightweight `<script>` tag) instead of server polling
  - Update every second using `setInterval`
- [x] Test: verify timezone handling, format switching

### Step 2.4 — Implement System Info Widget
- [x] Create `internal/widgets/sysinfo.go`
- [x] Read hostname: `os.Hostname()`
- [x] Read OS info: parse `/etc/os-release` (Linux)
- [x] Read system uptime: parse `/proc/uptime`
- [x] Read Go runtime info: `runtime.NumGoroutine()`, `runtime.MemStats`
- [x] Create `templates/widgets/widgets.html` template
- [x] Test: verify data displays correctly on Linux

### Step 2.5 — Implement Custom Text Widget
- [x] Create `internal/widgets/custom_text.go`
- [x] Read content string from config
- [x] Support basic HTML rendering (sanitized)
- [x] Create `templates/widgets/widgets.html` template
- [x] Test: verify HTML rendering and sanitization

### Step 2.6 — Integrate Widgets into Dashboard Layout
- [x] Add widget data to `DashboardData` struct in `main.go`:
  ```go
  type DashboardData struct {
      Proxmox    proxmox.NodeStatus
      Containers []docker.Container
      Services   []cache.ServiceState
      Widgets    []widgets.WidgetData  // NEW
  }
  ```
- [x] Create a `/widgets` HTMX endpoint (or include in `/status`)
- [x] Modify `static/index.html` header to include widget container row
- [x] Create responsive grid for widgets (1-4 columns based on screen size)
- [x] Render widgets above the main dashboard content
- [x] Test: full page render with all widgets enabled, then with some disabled

---

## Phase 3: Network Monitoring

### Step 3.1 — Create Network Package
- [x] Create `internal/network/` directory
- [x] Create `internal/network/types.go` with data structures:
  ```go
  type InterfaceStats struct {
      Name      string
      Label     string
      Status    string  // "up" or "down"
      RxBytes   uint64
      TxBytes   uint64
      RxRate    float64 // bytes/sec
      TxRate    float64 // bytes/sec
  }
  ```
- [x] Create `internal/network/monitor.go` with `Monitor` struct

### Step 3.2 — Parse `/proc/net/dev` for Interface Stats
- [x] Implement `readProcNetDev()` function:
  - Open and parse `/proc/net/dev`
  - Extract bytes received/transmitted per interface
  - Skip loopback by default
- [x] Handle file read errors gracefully
- [x] Return map of interface name → byte counts
- [x] Test: parse sample `/proc/net/dev` output, verify values

### Step 3.3 — Calculate Network Speed
- [x] Implement speed calculation using two samples:
  - Store previous reading with timestamp
  - Calculate `rate = (current_bytes - previous_bytes) / elapsed_seconds`
  - Smooth with moving average (last 3 samples)
- [x] Run background goroutine that samples every N seconds (configurable)
- [x] Format speeds as human-readable: b/s, Kbit/s, Mbit/s, Gbit/s
- [x] Format totals as human-readable: KB, MB, GB, TB
- [x] Test: verify speed calculation accuracy with known values

### Step 3.4 — Add Network Config & Integration
- [x] Add `NetworkConfig` struct to `internal/config/config.go`:
  ```go
  type NetworkConfig struct {
      Enabled        bool              `yaml:"enabled"`
      Interfaces     []NetIfConfig     `yaml:"interfaces"`
      ShowSpeed      bool              `yaml:"show_speed"`
      ShowTotal      bool              `yaml:"show_total_transfer"`
      UpdateInterval int               `yaml:"update_interval"`
  }
  type NetIfConfig struct {
      Name  string `yaml:"name"`
      Label string `yaml:"label"`
  }
  ```
- [x] Initialize network monitor in `main.go` (if enabled)
- [x] Start background sampling goroutine
- [x] Add network stats to `DashboardData` struct
- [x] Test: verify monitor starts and collects data

### Step 3.5 — Create Network Stats UI
- [x] Create `templates/network.html` template
- [x] Design cards matching the Homepage screenshot style:
  - Interface name and label (e.g., "Internal: vmbr1")
  - RX/TX speeds with directional arrows (↓ ↑)
  - Total transferred data
  - Status indicator (up/down)
- [x] Add to main dashboard grid (below Proxmox metrics)
- [x] Responsive: stack vertically on mobile
- [x] Test: verify UI updates via HTMX polling

---

## Phase 4: Custom Links & Web Bookmarks

### Step 4.1 — Add Bookmark Config Structures
- [x] Add to `internal/config/config.go`:
  ```go
  type BookmarkGroup struct {
      Group string         `yaml:"group"`
      Links []BookmarkLink `yaml:"links"`
  }
  type BookmarkLink struct {
      Name        string `yaml:"name"`
      URL         string `yaml:"url"`
      Icon        string `yaml:"icon"`
      Description string `yaml:"description"`
      NewTab      bool   `yaml:"new_tab"`
  }
  ```
- [x] Add `Bookmarks []BookmarkGroup` field to main `Config` struct
- [x] Add bookmarks section to `config-example.yaml`
- [x] Test: verify config parsing

### Step 4.2 — Add Icon Support
- [x] Support three icon modes:
  1. **Built-in icon name**: map string to inline SVG (e.g., `"globe"`, `"server"`)
  2. **Custom image path**: serve from `static/icons/` directory
  3. **Favicon fetch**: auto-fetch from `url/favicon.ico` and cache
- [x] Create favicon cache in `data/icons/` with MD5-hashed filenames
- [x] Implement favicon fetcher: HTTP GET `url + /favicon.ico`, save asynchronously
- [x] Test: verify all three icon modes render

### Step 4.3 — Create Bookmarks UI Template
- [x] Create `templates/bookmarks.html`:
  - Render groups as labeled sections
  - Each link as a card with icon, name, description
  - Click opens URL (new tab if configured)
  - Hover effect matching glassmorphism theme
- [x] Add bookmarks section to `templates/status.html` (below widgets, above dashboard)
- [x] Responsive grid: 2-6 columns based on screen size
- [x] Add group headings with subtle separators
- [x] Test: render sample bookmarks, verify layout

### Step 4.4 — Optional: Link Health Checking
- [ ] Reuse existing `internal/monitor/http.go` `CheckService()` for bookmark URLs
- [ ] Show small status dot (green/red) on each bookmark card
- [ ] Poll bookmark URLs on a slower interval (60s) to avoid hammering
- [ ] Add config option `check_health: true/false` per group
- [ ] Test: verify health status updates

---

## Phase 5: Service Integration Framework — **DEFERRED (Future Work)**

> **Status:** Deferred. Steps 5.4 and parts of 5.6 (Radarr/Sonarr/Overseerr media services) are already implemented and working.
> Remaining steps (5.1–5.3, 5.5, rest of 5.6) are deferred to a future release.
> The project works fully without this phase.

### Step 5.1 — Design Widget Interface & Registry
- [ ] Create `internal/services/` directory
- [ ] Create `internal/services/widget.go`:
  ```go
  type ServiceWidget interface {
      Name() string
      Icon() string
      Fetch() (*WidgetData, error)
  }
  type WidgetData struct {
      Metrics []Metric
      Status  string  // "running", "stopped", "error"
      Details map[string]string
  }
  type Metric struct {
      Label string
      Value string
  }
  ```
- [ ] Create `internal/services/registry.go`:
  - `Register(type string, factory func(config) ServiceWidget)`
  - `Create(type string, config) (ServiceWidget, error)`
  - Register all built-in widget types at init
- [ ] Test: register a dummy widget, create it, fetch data

### Step 5.2 — Implement Generic HTTP API Widget
- [ ] Create `internal/services/generic.go`
- [ ] Support configurable:
  - URL, HTTP method, headers
  - JSON path extraction (use `encoding/json` + manual path walking, or add `gjson`)
  - Metric label/value mapping
- [ ] Handle auth: API key header, Bearer token, Basic auth
- [ ] Add response caching with configurable TTL
- [ ] Test: fetch data from a sample API, extract JSON paths

### Step 5.3 — Implement Plex Widget
- [ ] Create `internal/services/plex.go`
- [ ] Use Plex API:
  - Endpoint: `{url}/status/sessions` (active streams)
  - Auth: `X-Plex-Token` header
  - Parse: now playing, stream count, media title
- [ ] Create template showing: current media, playback progress, stream count
- [ ] Add mock data for testing
- [ ] Test: connect to real Plex instance (if available), verify parsing

### Step 5.4 — Implement Radarr/Sonarr Widgets
- [x] Create `internal/mediaservices/client.go` — Radarr, Sonarr, and Overseerr API clients
- [x] Use Radarr/Sonarr APIs:
  - Endpoint: `{url}/api/v3/movie` (Radarr) or `{url}/api/v3/series` (Sonarr)
  - Auth: `X-Api-Key` header
  - Parse: wanted count, total items, queue status
- [x] Create `templates/mediaservices.html` showing: wanted/total counts per service
- [x] Test: verify API parsing with sample responses

### Step 5.5 — Implement Portainer Widget
- [ ] Create `internal/services/portainer.go`
- [ ] Use Portainer API:
  - Endpoint: `{url}/api/endpoints/{id}/docker/containers/json`
  - Auth: JWT token or API key
  - Parse: running/stopped container counts
- [ ] Create template showing: running/stopped/total counts
- [ ] Test: verify data display

### Step 5.6 — Integrate Service Widgets into Dashboard
- [x] Add `MediaServiceConfig` to config (under `media_services`)
- [x] Initialize media service polling in `main.go` (30s goroutine)
- [x] Create `templates/mediaservices.html` with card grid
- [x] Each card: icon, name, status badge, key metrics per service type
- [x] Add to dashboard layout in main grid (col-span-3)
- [ ] Add `ServiceWidgetConfig` to config:
  ```go
  type ServiceWidgetConfig struct {
      Type    string            `yaml:"type"`
      Name    string            `yaml:"name"`
      URL     string            `yaml:"url"`
      Token   string            `yaml:"token"`
      APIKey  string            `yaml:"api_key"`
      Icon    string            `yaml:"icon"`
      Headers map[string]string `yaml:"headers"`
      // Generic API fields
      JSONPath []JSONPathConfig `yaml:"json_path"`
  }
  ```
- [ ] Initialize service widgets from config in `main.go`
- [ ] Add service widget data fetching to the polling loop
- [ ] Create `templates/services.html` with card grid
- [ ] Each card: icon, name, status badge, 2-4 key metrics
- [ ] Match Homepage screenshot style (dark cards, metric boxes, status badges)
- [ ] Add to dashboard layout between Proxmox metrics and Docker section
- [ ] Test: full render with multiple service widgets

---

## Phase 6: Polish, Performance & Documentation

### Step 6.1 — Performance Optimization
- [x] Profile memory usage with all features enabled
- [x] Ensure weather API caching works (no duplicate calls)
- [x] Ensure network sampling doesn't leak goroutines
- [x] Add request timeouts to all external HTTP calls (5s default)
- [x] Optimize template rendering (pre-parse templates at startup)
- [x] Add graceful shutdown (signal handling, stop goroutines)
- [x] Verify binary size stays under 15 MB
- [x] Target: <50 MB RAM, <2% CPU with all features active

### Step 6.2 — Configuration Validation
- [x] Add `Validate()` method to `Config` struct
- [x] Check required fields per feature (e.g., weather needs lat/long)
- [x] Validate URL formats
- [x] Validate numeric ranges (opacity 0-1, blur 0-20)
- [x] Print clear warnings on startup for invalid config
- [x] Gracefully disable features with bad config (don't crash)
- [x] Test: verify each validation rule

### Step 6.3 — Update config-example.yaml
- [x] Add all new sections with inline comments
- [x] Provide realistic example values
- [x] Include commented-out optional features
- [x] Add section headers and separators for readability
- [x] Add bookmarks examples
- [x] Test: copy example to config.yaml, verify it loads

### Step 6.4 — Update Dockerfile
- [x] Copy `static/backgrounds/` directory
- [x] Ensure `data/` directory is created for favicon cache and todos
- [x] Ensure all new static assets are included
- [x] Update Go version to 1.24
- [x] Test: build Docker image, run, verify all features work

### Step 6.5 — Update Documentation
- [x] Update `documentation/docs.md` with new features
- [x] Update `documentation/deployment.md` with new config options
- [x] Update `README.md` with new screenshots and feature list
- [x] Add configuration examples for each new feature
- [x] Update `documentation/prompt-history.md` with session log

### Step 6.6 — Final Testing & Bug Fixes
- [x] Test on Chrome, Firefox, Safari (latest)
- [x] Test mobile responsiveness (375px, 768px, 1024px, 1440px)
- [x] Test with mock mode enabled (all features)
- [x] Test with empty config (all features disabled)
- [x] Test backward compat with old `config.yaml`
- [x] Run Lighthouse audit (target: Performance >90, A11y >90)
- [x] Fix any discovered bugs
- [x] Final review of all new code

### Step 6.7 — Security Hardening
- [x] Audit source code for hardcoded secrets, API keys, passwords
- [x] Verify `config.yaml` never committed to git history
- [x] Dockerfile: copy `config-example.yaml` instead of real `config.yaml` to prevent credential leakage in published images
- [x] Add security response headers (`X-Content-Type-Options`, `X-Frame-Options`, `X-XSS-Protection`, `Referrer-Policy`, CSP)
- [x] Add per-IP rate limiting to API endpoints (30 req/min)
- [x] Add path traversal protection to file-serving endpoints
- [x] Add input length validation to user-submitted data (500 char cap)

---

## Phase 7: Proxmox API Enrichment (Swap, Load, Kernel & Version)

### Step 7.1 — Parse Additional Fields from Proxmox API
- [x] Add `Swap` struct (Total, Used, Free) to `NodeStatus` in `internal/proxmox/client.go`:
  ```go
  Swap struct {
      Total int64 `json:"total"`
      Used  int64 `json:"used"`
      Free  int64 `json:"free"`
  } `json:"swap"`
  ```
- [x] Add `LoadAvg` field — the API returns `loadavg` as `[1min, 5min, 15min]` array
- [x] Add `PVEVersion` string field (`json:"pveversion"`) — Proxmox VE version
- [x] Add `KernelVersion` string field (`json:"kversion"`) — running kernel version
- [x] Add all new fields to `DashboardData` in `main.go`
- [x] Add mock data for swap, load average, kernel, and PVE version in `getMockStatus()`

### Step 7.2 — Display Swap, Load & Version in UI
- [x] Update `templates/status.html` to show swap usage bar below Memory section (same style as memory bar)
- [x] Add load average display (1m / 5m / 15m) near CPU section — useful at a glance
- [x] Add kernel version and PVE version as a subtle info line (e.g., below CPU model or in card footer)
- [x] Responsive layout — stack swap under memory on mobile
- [x] ARIA labels for swap meter and load average
- [x] Color-code swap bar (green → yellow → red) matching memory bar thresholds

### Step 7.3 — Update Documentation
- [x] Update `documentation/docs.md` with swap/load/version feature description
- [x] Update `config-example.yaml` with any new config options (if added)
- [x] Update `documentation/to-do.md` — Phase 7 marked complete

---

## Phase 8: Manual & Filesystem Disk Monitoring

### Step 8.1 — Add Extra Disks Config
- [x] Add `ExtraDisks []ExtraDiskConfig` to `config.go`:
  ```go
  type ExtraDiskConfig struct {
      Mountpoint string `yaml:"mountpoint"` // required: e.g. "/mnt/data"
      Label      string `yaml:"label"`      // optional: friendly name (defaults to mountpoint)
      Total      string `yaml:"total"`      // optional: manual override, e.g. "500GB", "1TB"
      Used       string `yaml:"used"`       // optional: manual override, e.g. "200GB"
      AutoDetect bool   `yaml:"auto_detect"` // if true (default), read from filesystem via statfs
  }
  ```
- [x] Parse human-readable sizes (GB, TB) into bytes in config validation
- [x] When `AutoDetect` is true and `Total`/`Used` are empty, use `syscall.Statfs` to read actual disk usage from the mount point at runtime
- [x] When `Total`/`Used` are provided manually, use static values (for remote/unmounted disks)
- [x] Merge extra disks into `Proxmox.Disks` alongside API-fetched disks (skip duplicates by mountpoint)
- [x] Add `skip_tls` support for the Proxmox API connection (already present via `InsecureSkipVerify: true`)

### Step 8.2 — Update Mock Data
- [x] Add sample extra disks to mock config for testing (both auto-detect and static modes)
- [x] Ensure merge logic works when both API disks and extra disks are present
- [x] Ensure deduplication: if an extra disk mountpoint matches an API-reported disk, skip it

### Step 8.3 — Update Documentation
- [x] Update `documentation/docs.md` with `extra_disks` config reference (explain auto-detect vs manual)
- [x] Update `config-example.yaml` with example extra disks section
- [x] Update `documentation/to-do.md` — Phase 8 marked complete

---

## Phase 9: Remote Docker & Portainer Support

> **Note:** The existing Docker client (`internal/docker/client.go`) already supports `unix://`, `tcp://`, and `http(s)://` endpoints.
> This phase adds the missing pieces: TLS client certificates, `skip_tls` option, and Portainer API integration.

### Step 9.1 — Add Docker Connection Config
- [x] Extend `DockerConfig` in `internal/config/config.go`:
  ```go
  type DockerConfig struct {
      Socket string `yaml:"socket"`  // existing: "unix:///var/run/docker.sock"
      // NEW options:
      SkipTLS        bool   `yaml:"skip_tls"`         // skip TLS verification for remote Docker
      TLSCACert      string `yaml:"tls_ca_cert"`      // path to CA cert (optional)
      TLSCert        string `yaml:"tls_cert"`          // path to client cert
      TLSKey         string `yaml:"tls_key"`           // path to client key
      PortainerURL   string `yaml:"portainer_url"`     // e.g. "https://portainer.example.com"
      PortainerKey   string `yaml:"portainer_api_key"` // Portainer API key
      PortainerEnvID int    `yaml:"portainer_env_id"`  // Portainer environment/endpoint ID
  }
  ```
- [x] Update `internal/docker/client.go` to support:
  - TLS client certificates (load CA + client cert/key when configured)
  - `skip_tls` flag for self-signed certs on remote Docker daemons
  - Portainer API proxy (`GET /api/endpoints/{env_id}/docker/containers/json` with `X-API-Key` header)
- [x] Add connection priority: Portainer > Remote Docker (TCP/TLS) > Local socket
- [x] Add mock data support for Portainer responses

### Step 9.2 — Update Documentation
- [x] Update `documentation/docs.md` with remote Docker and Portainer config reference
- [x] Update `config-example.yaml` with examples for each connection method (socket, TCP, TLS, Portainer)
- [x] Update `documentation/to-do.md` — Phase 9 marked complete

---

## Phase 10: UI Refinements & Theme Toggle

### Step 10.1 — Favicon & Header Logo
- [x] Add a built-in SVG favicon to `static/` (or `data/icons/`)
- [x] Update `static/index.html` to add `<link rel="icon" href="/favicon.svg" type="image/svg+xml">`
- [x] Optionally add config option `appearance.favicon` to allow custom favicon path/URL
- [x] Add a small logo next to the "dhiarhome" text in the page header (optional)

### Step 10.2 — Bigger Widget Text & Readability
- [x] Increase font sizes across `templates/status.html`:
  - CPU & Memory: title, percentage, GB values
  - Virtualization: VM/LXC counts
  - Disk Usage: mountpoint, percentage, used/total
  - Services, Docker, Media Services: names, stats, labels
- [x] Adjust card padding and spacing to accommodate bigger text
- [x] Test responsiveness — ensure no overflow on mobile

### Step 10.3 — Dark/Light Theme Toggle
- [x] Add CSS variables for a light theme alongside existing dark theme
- [x] Add a toggle button (sun/moon icon) in the page header
- [x] Persist theme choice in `localStorage` (client-side)
- [x] Default to `appearance.theme` from config, allow user override via toggle
- [x] Ensure all glassmorphism effects look good in both themes

### Step 10.4 — Update Documentation
- [x] Update `documentation/docs.md` with UI changes and theme toggle
- [x] Update `config-example.yaml` with logo config
- [x] Update `documentation/to-do.md` — Phase 10 marked complete

---

## Phase 11: Telegram Notifications (Service & Container Alerts)

### Step 11.1 — Add Notifications Config
- [x] Add `Notifications` section to `internal/config/config.go`:
  ```go
  type NotificationsConfig struct {
      Telegram TelegramConfig `yaml:"telegram"`
  }
  type TelegramConfig struct {
      Enabled     bool   `yaml:"enabled"`
      BotToken    string `yaml:"bot_token"`      // Telegram bot token
      ChatID      string `yaml:"chat_id"`        // Telegram chat/group/channel ID
      NotifyUp    bool   `yaml:"notify_up"`      // notify when service recovers (default: true)
      NotifyDown  bool   `yaml:"notify_down"`    // notify when service goes down (default: true)
      Cooldown    int    `yaml:"cooldown"`       // minutes between repeat alerts (default: 5)
      SilentHours []int  `yaml:"silent_hours"`   // optional: hours to suppress (e.g., [23,0,1] for night)
  }
  ```

### Step 11.2 — Implement Telegram Notifier
- [x] Create `internal/notifications/telegram.go`:
  - `SendMessage(botToken, chatID, message string)` — HTTP POST to `https://api.telegram.org/bot{token}/sendMessage`
  - Support `parse_mode: HTML` for formatted messages (bold service name, status emoji)
  - Format message with: service name, status (up/down), response time, timestamp
- [x] Integrate into `doPoll()` in `main.go`:
  - Track previous service states in a `map[string]string` (name → last known status)
  - When a service transitions **Online → Offline**, send a down alert (if `notify_down`)
  - When a service transitions **Offline → Online**, send a recovery alert (if `notify_up`)
  - Rate-limit: respect `cooldown` — don't resend within N minutes for the same service
  - Optional: suppress notifications during `silent_hours`
- [x] Also monitor Docker container state transitions (running → exited, exited → running)
- [x] Mock/dry-run mode for testing without real Telegram tokens (log messages to stdout)
- [x] Add a `/api/notifications/test` endpoint to send a test message manually

### Step 11.3 — Update Documentation
- [x] Update `documentation/docs.md` with Telegram notification config reference
- [x] Update `config-example.yaml` with Telegram section (commented out)
- [x] Update `documentation/to-do.md` — Phase 11 marked complete

---

## Phase 12: Historical Graphs & Long-Term Monitoring

### Step 12.1 — Design & Implement Graph Data Storage (SQLite)
- [ ] Use **SQLite** as the time-series store (best fit: single binary, no external dependencies, file-based)
- [ ] Add `go-sqlite3` or `modernc.org/sqlite` (pure Go, no CGO) as a dependency
- [ ] Create `internal/history/store.go`:
  - `Open(dbPath string)` — create/open SQLite database
  - `Record(metric string, value float64, timestamp time.Time)` — insert data point
  - `Query(metric string, from, to time.Time)` — fetch data points for a time range
  - `Prune(retention time.Duration)` — delete data older than retention window
- [ ] Schema: `metrics (id, name, value, recorded_at)` with index on `(name, recorded_at)`
- [ ] Config options:
  ```go
  type HistoryConfig struct {
      Enabled       bool   `yaml:"enabled"`
      DBPath        string `yaml:"db_path"`         // default: "data/history.db"
      Interval      int    `yaml:"interval"`        // seconds between snapshots (default: 300 = 5 min)
      RetentionDays int    `yaml:"retention_days"`  // default: 30
  }
  ```
- [ ] Record CPU, memory, swap, disk usage, and network speeds at configured interval
- [ ] Start background goroutine for periodic recording + pruning

### Step 12.2 — Implement Graph UI
- [ ] Add a lightweight charting library: **uPlot** (tiny, fast, no dependencies) or inline SVG sparklines
- [ ] Create `templates/graphs.html` template with:
  - Time-range selector buttons: 1h, 6h, 24h, 7d, 30d
  - Line charts for: CPU %, Memory %, Swap %, Disk %, Network RX/TX
  - Hover tooltip showing exact value + timestamp
- [ ] Add JSON API endpoints:
  - `GET /api/graphs/cpu?range=24h`
  - `GET /api/graphs/memory?range=7d`
  - `GET /api/graphs/disk?range=30d`
  - `GET /api/graphs/network?range=24h`
- [ ] Render graphs in a collapsible section below the Proxmox metrics (or as a dedicated tab)
- [ ] Responsive: full-width charts on mobile, side-by-side on desktop

### Step 12.3 — Update Documentation
- [ ] Update `documentation/docs.md` with graph/history config reference
- [ ] Update `config-example.yaml` with history section (commented out)
- [ ] Update `documentation/to-do.md` — Phase 12 marked complete

---

## Backlog / Future Improvements

### B.1 — Fix CSP to Allow Google Fonts
- [ ] Add `https://fonts.googleapis.com` to `style-src` in Content-Security-Policy header (`main.go:137`)
- [ ] Current CSP blocks Google Fonts stylesheet; Inter font falls back to system stack once cache expires
- [ ] Add `https://fonts.gstatic.com` to `font-src` (already present) — only `style-src` needs updating

### B.2 — Proxmox LVM/Local-Thin Disk Monitoring
- [ ] Current disk API endpoint (`/nodes/{node}/disks/list`) returns disks but skips entries with empty mountpoints
- [ ] LVM thin pools like `local-lvm` have no mountpoint, so they're filtered out
- [ ] Fix: use Proxmox storage API endpoint `/nodes/{node}/storage` or `/nodes/{node}/disks/lvmthin` to fetch LVM pool usage
- [ ] Parse `used` and `total` from storage status response, add to disk list
- [ ] Filter: only include LVM thin pools, skip already-listed mountpoints

## Quick Reference: File Changes by Step

| Step | Files Modified/Created |
|------|----------------------|
| 1.1 | `internal/config/config.go`, `config-example.yaml` |
| 1.2 | `static/index.html`, `static/backgrounds/` |
| 1.3 | `static/index.html`, `templates/status.html` |
| 1.4 | `static/index.html`, `templates/status.html` |
| 1.5 | `static/index.html`, `templates/status.html` |
| 1.6 | `static/index.html`, `templates/status.html` |
| 2.1 | `internal/widgets/widget.go`, `internal/widgets/registry.go`, `internal/config/config.go` |
| 2.2 | `internal/widgets/weather.go`, `templates/widgets/weather.html` |
| 2.3 | `internal/widgets/datetime.go`, `templates/widgets/datetime.html` |
| 2.4 | `internal/widgets/sysinfo.go`, `templates/widgets/sysinfo.html` |
| 2.5 | `internal/widgets/custom_text.go`, `templates/widgets/custom_text.html` |
| 2.6 | `main.go`, `static/index.html`, `templates/status.html` |
| 3.1 | `internal/network/types.go`, `internal/network/monitor.go` |
| 3.2 | `internal/network/monitor.go` |
| 3.3 | `internal/network/monitor.go` |
| 3.4 | `internal/config/config.go`, `main.go`, `config-example.yaml` |
| 3.5 | `templates/network.html`, `templates/status.html` |
| 4.1 | `internal/config/config.go`, `config-example.yaml` |
| 4.2 | `static/index.html`, `static/icons/` |
| 4.3 | `templates/bookmarks.html`, `static/index.html` |
| 4.4 | `internal/monitor/http.go`, `templates/bookmarks.html` |
| 5.1 | `internal/services/widget.go`, `internal/services/registry.go` |
| 5.2 | `internal/services/generic.go` |
| 5.3 | `internal/services/plex.go` |
| 5.4 | `internal/services/radarr.go`, `internal/services/sonarr.go` |
| 5.5 | `internal/services/portainer.go` |
| 5.6 | `main.go`, `config-example.yaml`, `templates/services.html` |
| 6.1 | All backend files (profiling/optimization) |
| 6.2 | `internal/config/config.go` |
| 6.3 | `config-example.yaml` |
| 6.4 | `Dockerfile` |
| 6.5 | `documentation/*.md`, `README.md` |
| 6.6 | All files (testing/fixes) |
| 6.7 | `main.go`, `Dockerfile` (security hardening) |
| 7.1 | `internal/proxmox/client.go`, `main.go` — swap, load avg, kernel, version parsing |
| 7.2 | `templates/status.html` — swap bar, load avg, kernel/version display |
| 7.3 | `documentation/docs.md`, `config-example.yaml` |
| 8.1 | `internal/config/config.go` — ExtraDisks config with auto-detect + `ParseSize()` |
| 8.2 | `main.go`, `internal/proxmox/client.go` — merge extra disks, `ReadDiskUsage()`, dedup |
| 8.3 | `documentation/docs.md`, `config-example.yaml` |
| 9.1 | `internal/config/config.go`, `internal/docker/client.go` — TLS, skip_tls, Portainer |
| 9.2 | `documentation/docs.md`, `config-example.yaml` |
| 10.1 | `static/index.html`, `static/favicon.svg` — favicon/logo in tab title |
| 10.2 | `templates/status.html` — bigger text across widgets |
| 10.3 | `static/index.html`, `templates/status.html` — dark/light theme toggle |
| 10.4 | `documentation/docs.md`, `config-example.yaml` |
| 11.1 | `internal/config/config.go` — Notifications/Telegram config |
| 11.2 | `internal/notifications/telegram.go`, `main.go` — up/down alerts, container alerts, cooldown |
| 11.3 | `documentation/docs.md`, `config-example.yaml` |
| 12.1 | `internal/history/store.go`, `go.mod` — SQLite time-series storage |
| 12.2 | `templates/graphs.html`, `main.go` — chart endpoints, uPlot/sparklines |
| 12.3 | `documentation/docs.md`, `config-example.yaml` |

---

## Progress Tracker

| Phase | Steps | Done | Remaining | Status |
|-------|-------|------|----------|--------|
| 1. Visual Enhancements | 6 | 6 | 0 | Complete |
| 2. Utility Widgets | 6 | 6 | 0 | Complete |
| 3. Network Monitoring | 5 | 5 | 0 | Complete |
| 4. Bookmarks & Links | 4 | 3 | 1 | Mostly done (4.4 optional) |
| 5. Service Widgets | 6 | 2 | 4 | **Deferred** (5.4 + partial 5.6 done) |
| 6. Polish & Docs | 7 | 7 | 0 | Complete (incl. 6.5–6.7) |
| 7. Proxmox API Enrichment | 3 | 3 | 0 | Complete |
| 8. Disk Monitoring | 3 | 3 | 0 | Complete |
| 9. Remote Docker & Portainer | 2 | 2 | 0 | Complete |
| 10. UI & Theme Toggle | 4 | 4 | 0 | **Complete** |
| 11. Telegram Notifications | 3 | 3 | 0 | **Complete** |
| 12. Historical Graphs | 3 | 0 | 3 | **Pending** |
| B. Backlog | 2 | 0 | 2 | **Pending** |
| **Total** | **60** | **44** | **16** |

> **v1.4.0:** Phase 11 complete — Telegram notifications for service and Docker container state transitions with cooldown, silent hours, mock mode, and test endpoint.
> **v1.3.2:** Phase 10 complete — inline SVG favicon, configurable logo (file/URL), header logo, bigger widget text across all templates, dark/light theme toggle with localStorage persistence.
> **v1.3.1:** UI refinements — full-screen todo modal with date tracking, CPU/Memory widget vertical stretch and divider.
> **v1.3.0:** Phase 9 complete — remote Docker TLS + Portainer API integration.
> **v1.2.0:** Phase 8 complete — extra disk monitoring with auto-detect (statfs) and manual override modes.
> **v1.1.0:** Phase 7 complete — swap usage, load average, PVE/kernel version display.
> **v1.0.1:** Added `skip_tls` option for services, fixed Overseerr pageInfo JSON tag, added Bookmarks Config reference to docs.
> **v1.0.0 Release:** All 29 completed steps are included in the first stable release. Remaining items (Phases 7–12) are planned for future versions.
