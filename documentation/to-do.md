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
- [ ] Add to `internal/config/config.go`:
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
- [ ] Add `Bookmarks []BookmarkGroup` field to main `Config` struct
- [ ] Add bookmarks section to `config-example.yaml`
- [ ] Test: verify config parsing

### Step 4.2 — Add Icon Support
- [ ] Add Lucide Icons CDN to `static/index.html`:
  ```html
  <script src="https://unpkg.com/lucide@latest"></script>
  ```
- [ ] Support three icon modes:
  1. **Built-in icon name**: map string to Lucide icon (e.g., `"globe"`, `"server"`)
  2. **Custom image path**: serve from `static/icons/` directory
  3. **Favicon fetch**: auto-fetch from `url/favicon.ico` and cache
- [ ] Create favicon cache in `static/icons/cache/`
- [ ] Implement favicon fetcher: HTTP GET `url + /favicon.ico`, save with URL hash as filename
- [ ] Test: verify all three icon modes render

### Step 4.3 — Create Bookmarks UI Template
- [ ] Create `templates/bookmarks.html`:
  - Render groups as labeled sections
  - Each link as a card with icon, name, description
  - Click opens URL (new tab if configured)
  - Hover effect matching glassmorphism theme
- [ ] Add bookmarks section to `static/index.html` (below widgets, above dashboard)
- [ ] Responsive grid: 2-6 columns based on screen size
- [ ] Add group headings with subtle separators
- [ ] Test: render sample bookmarks, verify layout

### Step 4.4 — Optional: Link Health Checking
- [ ] Reuse existing `internal/monitor/http.go` `CheckService()` for bookmark URLs
- [ ] Show small status dot (green/red) on each bookmark card
- [ ] Poll bookmark URLs on a slower interval (60s) to avoid hammering
- [ ] Add config option `check_health: true/false` per group
- [ ] Test: verify health status updates

---

## Phase 5: Service Integration Framework

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
- [ ] Create `internal/services/radarr.go` and `internal/services/sonarr.go`
- [ ] Use Radarr/Sonarr APIs:
  - Endpoint: `{url}/api/v3/movie` (Radarr) or `{url}/api/v3/series` (Sonarr)
  - Auth: `X-Api-Key` header
  - Parse: wanted count, total items, queue status
- [ ] Create templates showing: wanted/total counts, queue info
- [ ] Test: verify API parsing with sample responses

### Step 5.5 — Implement Portainer Widget
- [ ] Create `internal/services/portainer.go`
- [ ] Use Portainer API:
  - Endpoint: `{url}/api/endpoints/{id}/docker/containers/json`
  - Auth: JWT token or API key
  - Parse: running/stopped container counts
- [ ] Create template showing: running/stopped/total counts
- [ ] Test: verify data display

### Step 5.6 — Integrate Service Widgets into Dashboard
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
- [ ] Profile memory usage with all features enabled
- [ ] Ensure weather API caching works (no duplicate calls)
- [ ] Ensure network sampling doesn't leak goroutines
- [ ] Add request timeouts to all external HTTP calls (5s default)
- [ ] Optimize template rendering (pre-parse templates at startup)
- [ ] Verify binary size stays under 15 MB
- [ ] Target: <50 MB RAM, <2% CPU with all features active

### Step 6.2 — Configuration Validation
- [ ] Add `Validate()` method to `Config` struct
- [ ] Check required fields per feature (e.g., weather needs lat/long)
- [ ] Validate URL formats
- [ ] Validate numeric ranges (opacity 0-1, blur 0-20)
- [ ] Print clear warnings on startup for invalid config
- [ ] Gracefully disable features with bad config (don't crash)
- [ ] Test: verify each validation rule

### Step 6.3 — Update config-example.yaml
- [ ] Add all new sections with inline comments
- [ ] Provide realistic example values
- [ ] Include commented-out optional features
- [ ] Add section headers and separators for readability
- [ ] Test: copy example to config.yaml, verify it loads

### Step 6.4 — Update Dockerfile
- [ ] Copy `static/backgrounds/` directory
- [ ] Copy `static/icons/` directory (if created)
- [ ] Ensure all new static assets are included
- [ ] Test: build Docker image, run, verify all features work

### Step 6.5 — Update Documentation
- [ ] Update `documentation/docs.md` with new features
- [ ] Update `documentation/deployment.md` with new config options
- [ ] Update `README.md` with new screenshots and feature list
- [ ] Add configuration examples for each new feature
- [ ] Update `documentation/prompt-history.md` with session log

### Step 6.6 — Final Testing & Bug Fixes
- [ ] Test on Chrome, Firefox, Safari (latest)
- [ ] Test mobile responsiveness (375px, 768px, 1024px, 1440px)
- [ ] Test with mock mode enabled (all features)
- [ ] Test with empty config (all features disabled)
- [ ] Test backward compat with old `config.yaml`
- [ ] Run Lighthouse audit (target: Performance >90, A11y >90)
- [ ] Fix any discovered bugs
- [ ] Final review of all new code

---

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

---

## Progress Tracker

| Phase | Steps | Done | Remaining |
|-------|-------|------|-----------|
| 1. Visual Enhancements | 6 | 6 | 0 |
| 2. Utility Widgets | 6 | 6 | 0 |
| 3. Network Monitoring | 5 | 5 | 0 |
| 4. Bookmarks & Links | 4 | 0 | 4 |
| 5. Service Widgets | 6 | 0 | 6 |
| 6. Polish & Docs | 6 | 0 | 6 |
| **Total** | **33** | **17** | **16** |
