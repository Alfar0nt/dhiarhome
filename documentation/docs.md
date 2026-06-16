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
- **Disk Usage** - Root filesystem storage metrics
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

### 5. Utility Widgets
- **Weather** — Open-Meteo API integration (free, no API key), temperature, condition, wind speed, configurable caching, mock mode
- **Date/Time** — Configurable timezone, 12h/24h format, client-side live clock (updates every second)
- **System Info** — Hostname, OS name, system uptime, Go runtime stats (goroutines, memory)
- **Custom Text** — Configurable title and content, HTML-sanitized for security
- Responsive grid layout (1-4 columns based on screen size)
- Glassmorphism card styling matching the dashboard theme

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

### APIs & Protocols
- **Proxmox VE API** - RESTful API for node status
- **Docker Engine API** - Unix socket or TCP socket communication
- **HTTP/HTTPS** - Service health checks

### Deployment
- **Docker** - Multi-stage build (golang:1.21-alpine → alpine:latest)
- **Binary** - Single executable file (~10MB)

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
│   │   └── config.go           # YAML configuration loader + AppearanceConfig
│   ├── docker/
│   │   └── client.go           # Docker API client
│   ├── monitor/
│   │   └── http.go             # HTTP service health checker
│   ├── proxmox/
│   │   └── client.go           # Proxmox API client
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

2. **Background Polling**
   - Goroutine polls configured services every 10 seconds
   - Stores results in thread-safe linked list cache (max 100 entries)

3. **HTTP Requests**
   - User accesses `http://localhost:8080/`
   - `index.html` rendered as Go template with appearance config injected
   - HTMX auto-refresh polls `/status` endpoint every 5s
   - Server fetches fresh data from Proxmox API and Docker socket
   - Renders `status.html` template with current metrics
   - Returns HTML fragment to browser
   - `/background` serves local background image file (if configured)
   - `/api/background` returns JSON with background config

4. **Real-time Updates**
   - HTMX swaps HTML content with fade transition
   - No full page reload required

### Key Components

#### Configuration Package (`internal/config`)
- Parses YAML configuration file
- Defines structs for Proxmox, Docker, Services, and Appearance config
- `setDefaults()` applies sensible defaults for omitted appearance fields
- Validates required fields

#### Proxmox Client (`internal/proxmox`)
- Connects to Proxmox VE API endpoint
- Authenticates using API token (PVEAPIToken header)
- Fetches node status (CPU, memory, disk, uptime)
- Supports self-signed certificates (TLS skip verify)
- Mock mode generates random realistic data

#### Docker Client (`internal/docker`)
- Communicates via Unix socket (`/var/run/docker.sock`) or TCP
- Uses Docker Engine API (`/containers/json?all=1`)
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

#### Template Rendering
- Go `html/template` with custom functions:
  - `percent(used, total)` - Calculate percentage
  - `mult(a, b)` - Multiplication
  - `gb(bytes)` - Convert bytes to gigabytes
- `index.html` rendered as Go template with appearance config data
- `status.html` rendered with current metrics and service states
- Conditional rendering for online/offline states
- Progress bars for CPU/memory/disk usage
- Glassmorphism card styling via CSS variables

#### HTTP Handlers
- `GET /` — Renders `index.html` as Go template (or serves static files)
- `GET /status` — Returns HTMX HTML fragment with current metrics
- `GET /background` — Serves local background image file with MIME type + 1h cache
- `GET /api/background` — Returns JSON with background source, opacity, blur

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
```

### Docker Section
```yaml
docker:
  socket: "unix:///var/run/docker.sock"         # Docker socket path
  monitor_containers:                           # Optional filter (empty = all)
    - "nginx"
    - "pihole"
```

### Services Section
```yaml
services:
  - name: "Personal Website"
    url: "https://example.com"
  - name: "Nextcloud"
    url: "https://nextcloud.example.com"
```

### Appearance Section
```yaml
appearance:
  background_image: ""                    # Local file path (relative to working dir)
  background_url: "https://..."           # Remote URL (overrides background_image)
  background_opacity: 0.4                 # Dark overlay opacity (0.0 - 1.0, default: 0.3)
  background_blur: 3                      # Background blur in px (0 - 20, default: 5)
  theme: "dark"                           # Theme (default: "dark")
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

---

## Security Considerations

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
- **Binary Size**: ~10 MB (statically compiled)
- **Startup Time**: <1 second
- **Concurrent Users**: Limited by Go HTTP server (thousands)
- **Cache Size**: 100 service states × ~100 bytes = ~10 KB

---

## Limitations

1. **Single Node** - Monitors only one Proxmox node per instance
2. **No Historical Data** - No persistent storage or graphs
3. **No Authentication** - Dashboard is publicly accessible
4. **No Alerts** - No email/webhook notifications
5. **HTTP Only** - Service checks limited to HTTP/HTTPS
6. **No HTTPS** - Dashboard itself doesn't support TLS (use reverse proxy)

---

## Future Enhancement Ideas

See [to-do.md](to-do.md) for the full phased implementation plan (33 steps). Key upcoming features:

- Weather widget (Open-Meteo API, free, no key)
- Date/time widget with timezone support
- System info widget (hostname, OS, uptime)
- Network interface monitoring (speed, RX/TX)
- Custom bookmarks and web links with icon support
- Service integration framework (Plex, Radarr, Sonarr, Portainer)
- Generic HTTP API widget for custom services
- Multi-node Proxmox support
- Historical metrics with SQLite/InfluxDB
- Alert notifications (email, Discord, Telegram)
- HTTPS support with Let's Encrypt
- User authentication

---

## License & Credits

This is a personal learning project for homelab monitoring. Feel free to modify and customize for your own use.
