# dhiarhome — Demo Branch Documentation

> **Demo branch** — all data is mock/cosmetic. No real APIs, no deployment, no credentials needed.

## Overview

**dhiarhome** is a lightweight homelab monitoring dashboard built with Go, HTMX, Alpine.js, and Tailwind CSS. This demo branch showcases the UI, layout, and interactive features using entirely mock data.

The dashboard renders as a single-page app with a glassmorphism theme — transparent cards with backdrop blur, dark/light themes, and auto-refreshing content via HTMX polling.

---

## Running the Demo

```bash
go run .
```

Or build first:

```bash
go build -o dhiarhome .
./dhiarhome
```

Then open `http://localhost:8080`.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `config.yaml` | Path to YAML configuration file |
| `--addr` | `:8080` | Listen address (host:port) |

---

## Widgets & Sections

### Proxmox Server Status

The main card showing CPU usage, RAM, swap, load average, disk usage (3 mock disks), CPU model/core/thread info, and VM/LXC enumeration (3 VMs + 7 LXCs with mixed states). All data is randomly generated with realistic ranges.

**What it demonstrates:** Progress bars with color-coded thresholds (green/yellow/red), multi-disk display, virtualization overview.

### Docker Containers

Shows 5 hardcoded containers (nginx, pihole, portainer, plex, nextcloud) with running/exited states. Each container has a **toggle button** to switch between Up and Down, triggering a toast notification.

**What it demonstrates:** Interactive state toggling, toast notifications, container status badges.

### Web Services

Services from `config.yaml` are "monitored" with simulated response times (50-300ms). Services randomly bounce between Online and Offline during polling. Each service has a **toggle button** to force Online/Offline transitions.

**What it demonstrates:** Service health UI, response time display, state transition toasts, green ping indicators.

### Media Services

Mock stats for Sonarr (45 series, 3 wanted), Radarr (120 movies, 8 wanted), and Overseerr (95 total, 5 pending, 1520 available). Displayed as a clickable card with per-service stat boxes.

**What it demonstrates:** Multi-service card layout, stat boxes, WebUI link placeholders.

### Bookmarks

Configurable web links organized into groups. Icons support Lucide SVG names (server, globe, tv, etc.), custom image paths, or auto-fetched favicons. Responsive grid (2-6 columns).

**What it demonstrates:** Group-based link organization, icon rendering, favicon caching.

### Weather + Time

Combined card: live clock (client-side JS, updates every second) and mock weather data (4 rotating conditions, cached for 5 minutes). Supports Celsius/Fahrenheit.

**What it demonstrates:** Widget combining, client-side clock, cache TTL.

### System Info

Hostname, OS name, system uptime, and Go runtime stats (goroutines, memory). Reads from `/proc/cpuinfo`, `/etc/os-release`, `/proc/uptime`, and `runtime.MemStats`.

**What it demonstrates:** Local system data display, compact card layout.

### Network

Simulated RX/TX speeds per configured interface. Mock traffic with moving average smoothing. Display-cached (10s TTL) to prevent rapid HTML changes.

**What it demonstrates:** Speed formatting (b/s → Gbit/s), per-interface status, display caching.

### To-Do List

Interactive Alpine.js widget with 5 demo items. Check/uncheck toggles completion. Full-screen modal via expand button shows date metadata ("Added today", "Done yesterday"). Adding/deleting disabled in demo.

**What it demonstrates:** Alpine.js reactivity, modal UI, JSON persistence, date tracking.

---

## Interactive Features

### Toggle Buttons (Demo Only)

Each service and container has a refresh icon button that toggles its status. States persist in server-side override maps (in memory, cleared on restart).

### Toast Notifications

Popup alerts in the top-right corner on state transitions. Green for recovery, red for failure. Auto-dismiss after 4 seconds. Driven by a `TransitionEvent` buffer embedded in the HTMX response.

### Info Panels

Each widget section has an ⓘ button revealing a description. The header **"What am I looking at?"** button opens/closes all panels simultaneously via a `toggle-info-open` custom event.

### Dark/Light Theme

Sun/moon toggle in the header. Persisted to `localStorage`. Properly contrasted text and card backgrounds in both modes.

### Visitor Counter

localStorage-based page visit counter in the footer.

---

## Configuration

All settings are in `config.yaml`. Key sections:

### Services
```yaml
services:
  - name: "My Website"
    url: "https://example.com"
```

### Appearance
```yaml
appearance:
  theme: "dark"            # "dark" or "light"
  accent_color: "#3b82f6"  # Accent color hex
  card_opacity: 0.6        # 0.0 - 1.0
  card_blur: 12            # 0 - 30 (px)
  background_image: ""     # Local file path
  background_opacity: 0.3  # Dark overlay opacity
  background_blur: 5       # Background blur (px)
```

### Widgets
```yaml
widgets:
  weather:
    enabled: true
    units: "celsius"       # "celsius" or "fahrenheit"
  datetime:
    enabled: true
    timezone: "Asia/Jakarta"
    format_24h: true
  system_info:
    enabled: true
  custom_text:
    enabled: false
    title: "Note"
    content: "Welcome!"
```

### Network
```yaml
network:
  enabled: true
  show_speed: true
  show_total_transfer: true
  update_interval: 3       # Seconds between mock samples
  interfaces:
    - name: "eth0"
      label: "Primary"
```

### To-Do
```yaml
todos:
  enabled: true
  file_path: "data/todos.json"
  title: "To-Do"
```

### Bookmarks
```yaml
bookmarks:
  - group: "Development"
    links:
      - name: "GitHub"
        url: "https://github.com"
        icon: "github"
        new_tab: true
  - group: "Infrastructure"
    links:
      - name: "Proxmox"
        url: "https://192.168.1.100:8006"
        icon: "server"
```

---

## Architecture

### Data Flow

1. Application loads `config.yaml`
2. Widget registry initializes enabled widgets
3. Service polling goroutine runs every 10s (mock response times)
4. Network monitor goroutine samples mock traffic every 3s
5. Browser loads `index.html` → HTMX polls `/status` every 5s
6. Custom `merge-swap` HTMX extension diffs DOM without destroying backdrop-filter layers
7. Toggle handlers update server-side override maps and flush transition events

### Key Components

| Package | Purpose |
|---------|---------|
| `internal/config` | YAML configuration loader and structs |
| `internal/proxmox` | Mock Proxmox data (CPU, RAM, disks, VMs, LXCs) |
| `internal/docker` | Mock Docker containers |
| `internal/cache` | In-memory service state cache (linked list) |
| `internal/widgets` | Weather, datetime, sysinfo, custom_text widgets |
| `internal/network` | Mock network traffic monitor |
| `internal/mediaservices` | Mock Sonarr/Radarr/Overseerr stats |
| `internal/bookmarks` | Bookmark store with favicon caching |
| `internal/todo` | Persistent to-do list (JSON file) |

### Frontend

- **HTMX 1.9.10** — Server-side rendering with 5s auto-refresh polling
- **Custom merge-swap extension** — DOM diffing that preserves `backdrop-filter` GPU layers (no flicker)
- **Alpine.js 3.x** — Client-side interactivity (toasts, info panels, to-do, theme toggle)
- **Tailwind CSS** (CDN) — Utility-first styling with glassmorphism cards

### HTTP Routes

| Route | Method | Description |
|-------|--------|-------------|
| `/` | GET | Dashboard page |
| `/status` | GET | HTMX status fragment (auto-refresh) |
| `/api/todos` | GET | To-do list JSON |
| `/api/todos/{id}` | PUT | Toggle to-do item |
| `/api/services/toggle` | PUT | Toggle service status (demo) |
| `/api/containers/toggle` | PUT | Toggle container status (demo) |
| `/bookmarks/icons/` | GET | Favicon file cache |

---

## Project Structure

```
dhiarhome/
├── main.go                      # Application entry point
├── config.yaml                  # Demo configuration
├── go.mod / go.sum              # Go module + dependencies
├── internal/
│   ├── bookmarks/store.go       # Bookmark processing + favicon cache
│   ├── cache/history.go         # Service state cache (linked list)
│   ├── config/config.go         # YAML loader + config structs
│   ├── docker/client.go         # Docker mock containers
│   ├── mediaservices/client.go  # Mock media service stats
│   ├── network/
│   │   ├── types.go             # InterfaceStats struct
│   │   └── monitor.go           # Mock network monitor
│   ├── proxmox/client.go        # Proxmox mock data + types
│   ├── todo/store.go            # Persistent to-do store
│   └── widgets/
│       ├── widget.go            # Widget interface
│       ├── registry.go          # Widget registry
│       ├── weather.go           # Mock weather widget
│       ├── datetime.go          # Date/time widget
│       ├── sysinfo.go           # System info widget
│       └── custom_text.go       # Custom text widget
├── static/
│   └── index.html               # Dashboard page template
├── templates/
│   ├── status.html              # Status template (HTMX fragment)
│   ├── mediaservices.html       # Media services card
│   ├── todo.html                # To-do widget
│   ├── network.html             # Network card
│   └── bookmarks.html           # Bookmarks card
└── documentation/
    └── docs.md                  # This file
```

---

## Performance

- **Memory:** ~10-20 MB typical
- **CPU:** <1% (mostly idle, brief spikes during polling)
- **Binary size:** ~14 MB (statically compiled)
- **Startup:** <1 second

---

## License

MIT License — see [LICENSE](../LICENSE) for details.

Copyright (c) 2026 Dhiar Harianto
