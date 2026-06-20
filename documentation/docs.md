# dhiarhome - Full Project Documentation

## Project Overview

**dhiarhome** is an ultra-lightweight, self-hosted web dashboard designed for monitoring homelab servers running Proxmox VE. It provides real-time visibility into server health metrics, Docker container status, and web service uptime—all from a single, beautiful dark-mode interface.

### Purpose

Home servers typically have limited resources (CPU/RAM). Many existing dashboard solutions are heavy, requiring databases or complex setups. This project solves that by providing:

- **Zero database** - All data is fetched in real-time or cached in-memory
- **Minimal resource usage** - Single statically-compiled Go binary
- **Configuration-driven** - No code changes needed for customization
- **Real-time updates** - Auto-refreshing UI using HTMX

---

## Core Features

### 1. Proxmox Server Monitoring
- **CPU Usage** - Real-time CPU utilization percentage
- **Memory Usage** - RAM usage with detailed breakdown (used/total)
- **Swap Usage** - Swap memory with color-coded bar (green <60%, yellow 60-80%, red >80%)
- **Load Average** - 1-minute, 5-minute, and 15-minute load averages from the Proxmox API
- **Disk Usage** - Root filesystem and multi-disk storage metrics
- **Extra Disks** - Monitor additional filesystem mountpoints (auto-detect via statfs) or remote/unmounted disks (manual total/used sizes)
- **Version Info** - PVE manager version and kernel version displayed in card footer with `font-semibold` labels for readability
- **VM/LXC Enumeration** - Individual VM and LXC containers listed with running/stopped status indicators, VMID, and type labels
- **Uptime Tracking** - Server uptime duration

### 2. Docker Container Monitoring
- Lists all Docker containers on the host
- Shows container state (running/exited/stopped)
- Displays container status and uptime
- Optional: Filter to monitor only specific containers

### 3. Web Service Health Checks
- HTTP/HTTPS endpoint monitoring
- Response time tracking
- Status indicators (Online/Offline/Warning)
- Configurable service list

### 4. Mock Mode
- Built-in mock data generation for UI testing
- No real credentials required for development
- Random varying metrics for realistic testing

### 5. Utility Widgets & Interactive Features
- **To-Do List** — Interactive Alpine.js widget with add, toggle, delete. Persisted to `data/todos.json`. Optimistic updates via `fetch()` API calls. Compact widget shows scrollable list capped at ~2 visible items (`max-h-[72px]`). **Full-screen modal** opens via expand button for better interaction on mobile and desktop (full-viewport `bg-gray-900/95` overlay with larger text and touch targets). **Date tracking** shows "Added [date]" and "Done [date]" in expanded mode with smart formatting (Today/Yesterday/date). Inline `x-data` definition (no external function dependency).
- **Weather + Time** — Combined into a single compact card: live clock at top (client-side JS updates every second), weather condition below a divider. Mock weather cached for 5 minutes.
- **System Info** — Hostname, OS name, system uptime, Go runtime stats (goroutines, memory) in compact card
- **Network Summary** — Compact card showing per-interface status with live RX/TX speeds
- **Standalone fallbacks** — Weather and DateTime render individually if only one is enabled
- **Mobile layout** — 2-column grid with 4 widgets (perfect 2x2): todo, weather+time, system info, network
- **Desktop layout** — 4-column row: todo, weather+time, system info, network
- All cards share `min-h-[190px]` for consistent row height. Grid's `align-items: stretch` matches same-row cards.
- Increased font sizes (time `text-2xl`, hostname `text-base`) and `flex flex-col justify-between` to fill cards
- Glassmorphism card styling matching the dashboard theme

### 6. Media Services Monitoring
- **Sonarr** — Fetches series count (`GET /api/v3/series`) and wanted count (`GET /api/v3/wanted/missing`)
- **Radarr** — Fetches movie count (`GET /api/v3/movie`) and wanted count (`GET /api/v3/wanted/missing`)
- **Overseerr** — Fetches pending request count (`GET /api/v1/request?take=1&filter=pending`) and available media count (`GET /api/v1/media/count`)
- All use `X-Api-Key` header authentication with 5s HTTP timeout
- Graceful failure: `Online: false` when API is unreachable, config errors, or non-2xx response
- Mock mode (`MockStats()`) returns hardcoded test data when `proxmox.mock: true` and no services configured
- Polled every 30 seconds via `pollMediaServices()` goroutine with mutex-protected shared state
- Rendered as a clickable card in the main grid (col-span-3) with per-service stat boxes
- Each service shows name, status dot (green pulsing / red), WebUI link, and type-specific stats

### 7. Network Monitoring
- **`/proc/net/dev` Parsing** — Reads Linux kernel interface byte counts directly
- **Speed Calculation** — Two-sample rate with moving average smoothing (last 3 samples)
- **Human-Readable Formatting** — Speeds: b/s, Kbit/s, Mbit/s, Gbit/s; Totals: KB, MB, GB, TB
- **Configurable Interfaces** — Monitor specific interfaces by name with custom labels
- **Background Sampling** — Goroutine polls at configurable interval (default 3 seconds)
- **Mock Mode** — Simulates network traffic for UI testing without real interfaces
- **Compact Summary Card** — Displayed in top widget row with per-interface status and live speeds

### 8. UI Refinements
- **Inline SVG favicon** — Dashboard icon embedded as data URI in `<link rel="icon">`, works in all browsers with no external files
- **Header logo** — Server/cluster SVG icon next to the "dhiarhome" title for visual branding
- **Configurable logo** — `appearance.logo` supports local file path (e.g. `static/logo.png`) or remote URL; used as both favicon and header logo; falls back to inline SVG when empty
- **Dark/Light theme toggle** — Sun/moon button in page header toggles between themes; persisted to `localStorage`; defaults to `appearance.theme` from config; overrides CSS variables for card backgrounds (`#e2e8f0` body), text colors (dark text on light cards), accent colors, and scrollbars
- **Light mode readability** — Text colors properly contrasted for light backgrounds (`text-white` → dark `#1e293b`, status colors use darker tones); all `bg-white/*` and `border-white/*` classes overridden to dark-on-light equivalents; skeleton loaders use gray tones
- **Light mode background overlay** — Uses `rgba(15, 23, 42, 0.15)` instead of bright gray so background images remain visible
- **Todo modal light mode** — Overlay switches from `bg-gray-900/95` to light `rgba(226, 232, 240, 0.95)`; modal content backgrounds, borders, text, and buttons adapt to light theme
- **Bigger widget text** — All primary metric values increased from `text-sm` to `text-base`, section titles from `text-lg`/`text-xl` to `text-xl`/`text-2xl`, labels from `text-[10px]`/`text-[11px]` to `text-xs`
- **Metric label CSS** — `.metric-label` font-size increased from `0.6875rem` to `0.75rem`
- **Bookmarks & Services alignment** — Section titles bumped to `text-2xl` with `w-7 h-7` icons and `mb-5` spacing to match Proxmox/Docker/Media section headings; Services item spacing/padding (`space-y-3`/`p-3`) aligned with Docker cards
- **GitHub button** — GitHub icon button in the page header (top-right, before theme toggle) linking to the dhiarhome repository with glassmorphism styling
- **Footer** — Glassmorphism footer below the dashboard content with tech stack credits ("Built with dhiarhome · Go · HTMX · Alpine.js"), "Star on GitHub" link with icon, and "About the author" link. Responsive `max-w-lg sm:max-w-2xl` container

### 9. Security
- **Security headers middleware** — `securityHeaders()` wrapper applied to all HTTP responses:
  - `Content-Security-Policy` — restricts script/style/img/font/connect sources; `'unsafe-eval'` included for Alpine.js runtime expression evaluation
  - `X-Content-Type-Options: nosniff` — prevents MIME-type sniffing
  - `X-Frame-Options: DENY` — prevents clickjacking via iframe embedding
  - `X-XSS-Protection: 1; mode=block` — legacy XSS filter for older browsers
  - `Referrer-Policy: same-origin` — limits referrer information sent to external sites
- **SSRF protection** — Favicon fetcher (`internal/bookmarks/store.go`) validates URL scheme (http/https only), resolves DNS before fetching, and blocks requests to private IP ranges (loopback, RFC 1918, link-local, IPv6 ULA)
- **JSON injection prevention** — `backgroundHandler` uses `json.NewEncoder` for safe JSON encoding instead of string formatting
- **Path traversal protection** — `backgroundServeHandler` uses `filepath.Clean()` and rejects paths containing `..`
- **Per-IP rate limiting** — 30 requests/min on `/api/todos` endpoints with sliding window
- **Input validation** — Todo text capped at 500 characters

---

## Technology Stack

### Backend
- **Language**: Go 1.26.3
- **Architecture**: Statically compiled binary (CGO_ENABLED=0)
- **Target**: Linux/amd64
- **Dependencies**:
  - `gopkg.in/yaml.v3` - YAML configuration parsing

### Frontend
- **HTML5** - Semantic markup
- **Tailwind CSS** (CDN) - Utility-first CSS framework for dark-mode design
- **HTMX 1.9.10** - Hypermedia-driven interactions (no JavaScript required)
- **Alpine.js 3.x** (CDN) - Lightweight reactive framework for interactive to-do widget

### APIs & Protocols
- **Proxmox VE API** - RESTful API for node status
- **Docker Engine API** - Unix socket or TCP socket communication
- **HTTP/HTTPS** - Service health checks

### Deployment
- **Docker** - Multi-stage build (golang:alpine → alpine:latest)
- **Binary** - Single executable file (~14MB)

---

## Project Structure

```
dhiarhome/
├── main.go                      # Application entry point
├── config.yaml                  # User configuration (gitignored)
├── config-example.yaml          # Configuration template
├── Dockerfile                   # Multi-stage Docker build
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
│
├── internal/                    # Private application code
│   ├── cache/
│   │   └── history.go          # In-memory service state cache (linked list)
│   ├── config/
│   │   └── config.go           # YAML configuration loader + all config structs
│   ├── docker/
│   │   └── client.go           # Docker API client
│   ├── mediaservices/
│   │   └── client.go           # Sonarr/Radarr/Overseerr API clients
│   ├── bookmarks/
│   │   └── store.go            # Bookmark processing + favicon cache
│   ├── monitor/
│   │   └── http.go             # HTTP service health checker
│   ├── network/
│   │   ├── types.go            # InterfaceStats struct + rawSample type
│   │   └── monitor.go          # /proc/net/dev parser, speed calc, mock mode
│   ├── proxmox/
│   │   └── client.go           # Proxmox API client
│   ├── todo/
│   │   └── store.go            # Persistent to-do store (JSON file)
│   └── widgets/
│       ├── widget.go           # Widget interface + WidgetData struct
│       ├── registry.go         # Widget registry manager
│       ├── weather.go          # Open-Meteo weather widget
│       ├── datetime.go         # Date/time widget with client-side clock
│       ├── sysinfo.go          # System information widget
│       └── custom_text.go      # Custom text widget
│
├── static/
│   ├── index.html              # Main dashboard page (Go template + HTMX)
│   └── backgrounds/            # Custom background images (local files)
├── templates/
│   ├── status.html             # Server-side rendered status template
│   ├── network.html            # Network interface cards template
│   ├── todo.html               # Interactive to-do list template (Alpine.js)
│   ├── mediaservices.html      # Media management card template
│   └── widgets/
│       └── widgets.html        # Widget rendering template (all types)
│
└── Screenshot.png              # Project screenshot
```

---

## Architecture Details

### Data Flow

1. **Configuration Loading**
   - Application starts and loads `config.yaml`
   - Initializes Proxmox client, Docker client, and history cache
   - Runs initial blocking Proxmox and Docker polls to warm the cache before accepting requests

2. **Background Polling**
   - Proxmox poller (5s interval): fetches node status, virtualization info, merges extra disks
   - Docker poller (5s interval): fetches container list, applies name filters
   - Service monitor polls configured services every 10 seconds
   - Stores results in thread-safe linked list cache (max 100 entries)
   - Network monitor goroutine samples `/proc/net/dev` at configurable interval (default 3s)
   - Media services poll every 30 seconds (Sonarr/Radarr/Overseerr stats)
   - All results are stored in thread-safe shared state (`sync.RWMutex`)

3. **HTTP Requests**
   - User accesses `http://localhost:8080/`
   - `index.html` rendered as Go template with appearance config injected
   - HTMX auto-refresh polls `/status` endpoint every 5s
   - Server reads cached data from background pollers — no direct API calls
   - Renders `status.html` template with current metrics in <5ms
   - Returns HTML fragment to browser
   - `/background` serves local background image file (if configured)
   - `/api/background` returns JSON with background config

4. **Real-time Updates**
   - HTMX swaps HTML content with fade transition
   - No full page reload required

### Key Components

#### Configuration Package (`internal/config`)
- Parses YAML configuration file
- Defines structs for Proxmox, Docker, Services, Appearance, Widgets, and Network config
- `setDefaults()` applies sensible defaults for omitted appearance fields
- Validates required fields

#### Proxmox Client (`internal/proxmox`)
- Connects to Proxmox VE API endpoint
- Authenticates using API token (PVEAPIToken header)
- Fetches node status (CPU, memory, disk, uptime)
- **Multi-disk support**: `Disks []DiskInfo` with mountpoint, total, used per disk. Fetches additional disks from `/nodes/{node}/disks/list` endpoint. Mock mode returns 3 disks.
- **Extra disk monitoring**: `ExtraDisks []ExtraDiskConfig` in config supports two modes:
  - **Auto-detect**: reads real disk usage from local filesystem via `syscall.Statfs` (requires mountpoint to exist on dashboard host)
  - **Manual override**: accepts static `total`/`used` values as human-readable strings (e.g. "8TB", "500GB") for remote or unmounted disks
  - Deduplicates by mountpoint against Proxmox API-reported disks
  - `ReadDiskUsage()` function uses `statfs` to read block-level disk stats
  - `ParseSize()` in config package converts human-readable size strings to bytes (supports B, KB, MB, GB, TB, KiB, MiB, GiB, TiB)
- **Virtualization monitoring**: `GetVirtualization()` fetches QEMU VM and LXC container lists from `/nodes/{node}/qemu` and `/nodes/{node}/lxc`. Returns `VirtualizationInfo` with running/total counts and individual resource lists (`VMs []ResourceInfo`, `LXCs []ResourceInfo`) including VMID, name, and status (running/stopped). VMs and LXCs are sorted by VMID ascending for stable display order (prevents shuffling on every poll). Mock mode returns 3 VMs and 7 LXCs with mixed states.
- **API enrichment**: Parses swap usage (total/used/free), load average (1m/5m/15m from `loadavg` array), PVE version (`pveversion`), and kernel version (`kversion`) directly from the `/nodes/{node}/status` response — no extra API calls needed.
- Uses `json.Number` for load average parsing since Proxmox returns these as string-encoded floats.
- Supports self-signed certificates (TLS skip verify)
- Mock mode generates random realistic data

#### Docker Client (`internal/docker`)
- Communicates via Unix socket, TCP, or TLS-secured TCP
- Uses Docker Engine API (`/containers/json?all=1`)
- **Portainer integration**: fetches containers via Portainer API when configured
- **TLS support**: client certificates (mTLS) and `skip_tls` for self-signed certs
- Connection priority: Portainer > Remote Docker (TCP/TLS) > Local socket
- Lists all containers with state and status
- Supports filtering by container name

#### Cache System (`internal/cache`)
- Thread-safe doubly-linked list implementation
- Stores last 100 service state snapshots
- Provides `GetLatest()` for most recent state per service
- Uses `sync.RWMutex` for concurrent access

#### HTTP Monitor (`internal/monitor`)
- Performs HTTP GET requests to service URLs
- 5-second timeout per request
- Returns status: "Online" (2xx), "Offline" (error), or "Warning" (non-2xx)
- Measures response time

#### Network Monitor (`internal/network`)
- Parses `/proc/net/dev` for per-interface byte counts
- Two-sample rate calculation with moving average smoothing (last 3 samples)
- Background goroutine samples at configurable interval
- `formatSpeed()`: b/s → Kbit/s → Mbit/s → Gbit/s
- `formatBytes()`: B → KB → MB → GB → TB
- Mock mode generates random traffic for UI testing
- Thread-safe access via `sync.RWMutex`
- **Display caching** (10s TTL): `GetStats()` caches formatted output to prevent rapid HTML changes during HTMX swaps. Raw sampling continues at full rate for accuracy.

#### Template Rendering
- Go `html/template` with custom functions:
  - `percent(used, total)` - Calculate percentage
  - `mult(a, b)` - Multiplication
  - `gb(bytes)` - Convert bytes to gigabytes
  - `roundDur(duration)` - Format response time as "150 ms" or "1.23 s"
- `index.html` rendered as Go template with appearance config data
- `status.html` rendered with current metrics and service states
- `combineWidgets()` post-processes widget data: merges weather+datetime into combined card, reorders for layout
- Conditional rendering for online/offline states
- Progress bars for CPU/memory/disk usage
- Glassmorphism card styling via CSS variables

#### Client-Side DOM Diff Swap
- Custom HTMX extension (`merge-swap`) replaces default `innerHTML` swap
- `mergeDOM()` recursively walks current vs new DOM trees
- Updates only text nodes and dynamic attributes (class, style, aria-valuenow)
- Preserves glass-card elements so `backdrop-filter: blur()` GPU compositing layers aren't destroyed
- `data-preserve` attribute: marks interactive zones (e.g., Alpine.js todo widget) that must be skipped entirely during diff
- Eliminates backdrop-filter flickering during 5s HTMX polling
- Falls back to normal `innerHTML` on first load (skeleton → full render)

#### To-Do Store (`internal/todo`)
- Thread-safe CRUD store with JSON file persistence
- `NewStore(filePath)` loads existing data, auto-increments IDs
- `GetAll()`, `GetByID(id)`, `Add(text)`, `Update(id, text)`, `Toggle(id)`, `Delete(id)` with `sync.RWMutex`
- `Update()` modifies the text of an existing item (trims whitespace, 500 char limit enforced by handlers)
- `Toggle()` records `done_at` timestamp (RFC3339) when completing, clears when uncompleting
- Saves to `data/todos.json` on every mutation
- Interactive UI via Alpine.js (client-side `fetch()` to REST API)
- Full-screen modal with date metadata display (created_at, done_at)
- Inline edit: pencil icon replaces text with editable input; Enter to save, Escape to cancel

#### Bookmarks Store (`internal/bookmarks`)
- Configurable web bookmarks organized into named groups
- `NewStore(groups, cacheDir)` processes config and initializes favicon cache
- Icon resolution: Lucide SVG name, image path, or auto-fetched favicon
- Favicon caching: downloads from URL's `/favicon.ico` and saves to `data/icons/` with MD5-hashed filenames
- Responsive grid UI (2-6 columns) with glassmorphism cards and group headings
- Built-in SVG icons for common services: server, globe, monitor, play-circle, tv, container, database, home, settings, film

#### CPU Info (`internal/proxmox`)
- `CPUInfo` struct: `ModelName`, `Cores` (physical), `Threads` (logical)
- `ReadLocalCPUInfo()` parses `/proc/cpuinfo` for accurate core/thread counts
- `cleanCPUName()` strips verbose suffixes (` with Radeon Graphics`, ` CPU @ X.XXGHz`, `-Core Processor`, `(TM)`, `(R)` branding marks)
- Handles multi-socket systems, hyperthreading, and various CPU topologies
- Read once at startup (static hardware data, no polling needed)
- Mock mode provides simulated info (i7-12700K, 12C/20T)
- Displayed in CPU widget as model name above core/thread count (e.g., "Intel Core i7-12700K" above "12C / 20T")

#### HTTP Handlers
- `GET /` — Renders `index.html` as Go template (or serves static files)
- `GET /status` — Returns HTMX HTML fragment with current metrics (reads from background cache, <5ms response)
- `GET /background` — Serves local background image file with MIME type + 1h cache
- `GET /api/background` — Returns JSON with background source, opacity, blur
- `GET /api/todos` — Returns all todos as JSON array
- `POST /api/todos` — Creates a new todo (body: `{"text": "..."}`)
- `PATCH /api/todos/{id}` — Updates todo text (body: `{"text": "..."}`), returns updated Todo as JSON
- `PUT /api/todos/{id}` — Toggles todo done state
- `DELETE /api/todos/{id}` — Deletes a todo

#### Background Goroutines
- **Proxmox poller**: every 5 seconds — fetches node status, virtualization info (VMs/LXCs), merges extra disks. On error, keeps previous values (stale > no data)
- **Docker poller**: every 5 seconds — fetches container list, applies name filters. Mock fallback on error when `proxmox.mock: true`
- **Service monitor**: every 10 seconds — response times, online/offline status
- **Network monitor**: samples `/proc/net/dev` every 3 seconds (RX/TX rates)
- **Media services**: every 30 seconds (Sonarr/Radarr/Overseerr stats)
- **Container state tracker**: every 15 seconds — detects container state transitions for Telegram alerts
- HTMX auto-refresh polls `/status` every 5 seconds (reads from cache, no API calls in request path)

---

## Configuration Reference

### Proxmox Section
```yaml
proxmox:
  url: "https://192.168.1.100:8006/api2/json"  # Proxmox API endpoint
  node_name: "pve"                              # Node to monitor
  token_id: "root@pam!dashboard"                # API token ID
  token_secret: "YOUR-SECRET-UUID"              # API token secret
  mock: false                                   # Use mock data (true/false)
  # Extra disks to monitor beyond Proxmox API disks
  extra_disks:
    - mountpoint: "/mnt/data"                   # Required: filesystem mountpoint
      auto_detect: true                         # Read real usage via statfs (default: true)
    - mountpoint: "/mnt/nas"                    # Remote/unmounted disk
      label: "NAS Storage"                      # Optional friendly name
      total: "8TB"                              # Manual total size (supports B/KB/MB/GB/TB/KiB/MiB/GiB/TiB)
      used: "3.2TB"                             # Manual used size
      auto_detect: false                        # Disable auto-detect for manual mode
```

> **Note:** Extra disks are merged with Proxmox API-reported disks. Duplicate mountpoints are automatically skipped. Auto-detect mode requires the mountpoint to exist on the dashboard host (uses `syscall.Statfs`). Manual mode is useful for remote NAS, network shares, or disks not directly mounted on the dashboard host.

### Docker Section
```yaml
docker:
  # Local socket (default)
  socket: "unix:///var/run/docker.sock"
  monitor_containers:                           # Optional filter (empty = all)
    - "nginx"
    - "pihole"

  # Remote Docker with TLS (optional)
  # socket: "tcp://docker.example.com:2376"
  # skip_tls: true                              # Skip TLS verification
  # tls_ca_cert: "/path/to/ca.pem"              # CA certificate
  # tls_cert: "/path/to/cert.pem"               # Client certificate
  # tls_key: "/path/to/key.pem"                # Client key

  # Portainer API (optional, takes priority over socket/TCP)
  # portainer_url: "https://portainer.example.com"
  # portainer_api_key: "ptr_XXXXXXXXXXXX"       # Portainer access token
  # portainer_env_id: 1                         # Endpoint ID
```

### Services Section
```yaml
services:
  - name: "Personal Website"
    url: "https://example.com"
  - name: "Nextcloud"
    url: "https://nextcloud.example.com"
  - name: "Self-Signed App"
    url: "https://192.168.1.100:8443"
    skip_tls: true        # Skip TLS verification for self-signed certificates
```

### Appearance Section
```yaml
appearance:
  background_image: ""                    # Local file path (relative to working dir)
  background_url: "https://..."           # Remote URL (overrides background_image)
  background_opacity: 0.4                 # Dark overlay opacity (0.0 - 1.0, default: 0.3)
  background_blur: 3                      # Background blur in px (0 - 20, default: 5)
  logo: "static/logo.png"                 # Logo path or URL (favicon + header); empty = default SVG
  theme: "dark"                           # Theme: "dark" or "light" (user can toggle, persisted)
  card_opacity: 0.6                       # Card background opacity (0.0 - 1.0, default: 0.6)
  card_blur: 12                           # Card backdrop blur in px (0 - 30, default: 12)
  accent_color: "#3b82f6"                 # Accent color hex (default: "#3b82f6")
```

> **Note:** If both `background_image` and `background_url` are empty, no background image is rendered (solid dark background). If `background_url` is set, it takes priority over `background_image`.

### Widgets Section
```yaml
widgets:
  weather:
    enabled: false                        # Enable weather widget
    latitude: 40.7128                     # Your latitude
    longitude: -74.0060                   # Your longitude
    units: "celsius"                      # "celsius" or "fahrenheit" (default: celsius)
    cache_minutes: 15                     # API cache duration (default: 15)
    mock: false                           # Use mock data (default: false)

  datetime:
    enabled: false                        # Enable date/time widget
    timezone: "America/New_York"          # IANA timezone (default: Local)
    format_24h: false                     # 24-hour format (default: false)

  system_info:
    enabled: false                        # Enable system info widget

  custom_text:
    enabled: false                        # Enable custom text widget
    title: "Note"                         # Widget title (default: "Note")
    content: "Welcome!"                   # Text content (HTML is escaped)
```

> **Note:** Each widget has an `enabled` flag. Disabled widgets are not registered and consume zero resources.

### Network Section
```yaml
network:
  enabled: false                          # Enable network monitoring
  show_speed: true                        # Show real-time RX/TX speed
  show_total_transfer: true               # Show cumulative total bytes
  update_interval: 3                      # Seconds between /proc/net/dev samples (default: 3)
  mock: false                             # Use mock data for testing
  interfaces:
    - name: "eth0"                        # Linux interface name (from /proc/net/dev)
      label: "Primary"                    # Human-friendly display label
    - name: "wlan0"
      label: "WiFi"
```

> **Note:** Interface names must match entries in `/proc/net/dev`. Use `cat /proc/net/dev` to list available interfaces. The `mock` flag enables simulated traffic for UI testing without real interfaces.

### Todos Section
```yaml
todos:
  enabled: false               # Enable interactive to-do list
  file_path: "data/todos.json" # JSON file for persistent storage
  title: "To-Do"               # Widget title displayed on the card
```

> **Note:** Todo data is persisted to the specified JSON file and survives server restarts. The `data/` directory is created automatically and gitignored.

### Media Services Section
```yaml
media_services:
  - name: "Sonarr"
    url: "http://192.168.1.100:8989"
    api_key: "YOUR_SONARR_API_KEY"
    webui: "http://192.168.1.100:8989"
  - name: "Radarr"
    url: "http://192.168.1.100:7878"
    api_key: "YOUR_RADARR_API_KEY"
    webui: "http://192.168.1.100:7878"
  - name: "Overseerr"
    url: "http://192.168.1.100:5055"
    api_key: "YOUR_OVERSEERR_API_KEY"
    webui: "http://192.168.1.100:5055"
```

> **Note:** Each service requires `name`, `url` (API endpoint), `api_key`, and `webui` (browser-accessible URL). Services are polled every 30 seconds. Mock stats are automatically shown when `proxmox.mock: true` and no `media_services` are configured.

### Bookmarks Section
```yaml
bookmarks:
  - group: "Infrastructure"
    links:
      - name: "Proxmox VE"
        url: "https://192.168.1.100:8006"
        icon: "server"          # Lucide icon name, image path, or "favicon"
        new_tab: true            # Open in new tab (default: false)
      - name: "Portainer"
        url: "https://192.168.1.100:9443"
        icon: "container"
        new_tab: true
  - group: "Media"
    links:
      - name: "Sonarr"
        url: "https://sonarr.example.com"
        icon: "tv"
        new_tab: true
      - name: "Radarr"
        url: "https://radarr.example.com"
        icon: "film"
        new_tab: true
```

> **Note:** Groups are flattened in the UI — all links appear in a single section. Internal scrolling activates when there are more than ~10 links. Icons support three modes: Lucide icon name (e.g., `"server"`, `"globe"`, `"tv"`), custom image path, or `"favicon"` to auto-fetch from the URL's favicon.ico.

### Notifications (Telegram)
```yaml
notifications:
  telegram:
    enabled: false                # Enable Telegram alerts
    bot_token: "YOUR_BOT_TOKEN"   # Telegram bot token from @BotFather
    chat_id: "YOUR_CHAT_ID"       # Telegram chat/group/channel ID
    notify_up: true               # Notify when service/Docker recovers
    notify_down: true             # Notify when service/Docker goes down
    notify_todo_add: true         # Notify when a new to-do item is added
    notify_todo_complete: true    # Notify when a to-do item is marked complete
    cooldown: 5                   # Minutes between repeat alerts for the same service
    silent_hours: []              # Optional: suppress during certain hours (e.g., [23,0,1])
    mock: false                   # Dry-run: log to stdout instead of sending (for testing)
```

> **Note:** The notifier tracks state transitions for both monitored services (HTTP checks) and Docker containers (running ↔ exited). Cooldown prevents alert fatigue by suppressing repeat notifications within the configured window. Silent hours are specified as a list of hours (0-23) in server local time. Todo notifications (`notify_todo_add`, `notify_todo_complete`) fire immediately with no cooldown — they are user-initiated actions, not automated polling results. Each message includes the task name and the remaining incomplete tasks list.

### Toast Notifications (Web UI)

State transitions also appear as **toast popups** in the top-right corner of the dashboard:
- **Green toast** — Service recovered (Online) or container started
- **Red toast** — Service went down (Offline) or container stopped
- Auto-dismiss after 4 seconds
- Uses Alpine.js `x-init` with the transition data embedded in the HTMX response
- Styled with Tailwind CSS classes matching the dashboard theme

Toasts are driven by the same `TransitionEvent` buffer that powers Telegram notifications. No additional configuration needed.

---

1. **Proxmox API Token**
   - Create a read-only API token with minimal privileges
   - Token secret is stored in plaintext in `config.yaml` (keep secure)
   - `config.yaml` is gitignored to prevent accidental commits

2. **Docker Socket**
   - Mounting `/var/run/docker.sock` grants full Docker control
   - Consider using read-only mount: `-v /var/run/docker.sock:/var/run/docker.sock:ro`
   - Only run in trusted environments

3. **TLS Certificates**
   - Proxmox client skips TLS verification (self-signed certs)
   - Not suitable for untrusted networks

4. **Network Exposure**
   - Dashboard listens on all interfaces (`:8080`)
   - No authentication mechanism built-in
   - Use reverse proxy (nginx/Caddy) with auth for production

---

## Performance Characteristics

- **Memory Usage**: ~10-20 MB typical
- **CPU Usage**: <1% (mostly idle, brief spikes during polling)
- **Binary Size**: ~14 MB (statically compiled)
- **Startup Time**: <1 second
- **Concurrent Users**: Limited by Go HTTP server (thousands)
- **Cache Size**: 100 service states × ~100 bytes = ~10 KB

---

## Limitations

1. **Single Node** - Monitors only one Proxmox node per instance
2. **No Historical Data** - No persistent storage or graphs
3. **No Authentication** - Dashboard is publicly accessible
4. **Telegram Only** - Notifications are limited to Telegram (no email, Discord, webhook)
5. **HTTP Only** - Service checks limited to HTTP/HTTPS
6. **No HTTPS** - Dashboard itself doesn't support TLS (use reverse proxy)

---

## Future Enhancement Ideas

See [to-do.md](to-do.md) for the full phased implementation plan (58 steps). Key upcoming features:

- ~~Weather widget (Open-Meteo API, free, no key)~~
- ~~Date/time widget with timezone support~~
- ~~System info widget (hostname, OS, uptime)~~
- ~~Network interface monitoring (speed, RX/TX)~~
- ~~Custom bookmarks and web links with icon support~~
- ~~Service integration framework (Radarr, Sonarr, Overseerr)~~
- Service integration framework (Plex, Portainer, generic API widget)
- Multi-node Proxmox support
- Historical metrics with SQLite/InfluxDB
- ~~Alert notifications (Telegram)~~
- HTTPS support with Let's Encrypt
- User authentication

---

## License & Credits

This is a personal learning project for homelab monitoring. Feel free to modify and customize for your own use.
