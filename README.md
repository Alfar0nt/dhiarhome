# dhiarhome

A lightweight, self-hosted homelab monitoring dashboard for Proxmox VE, Docker containers, web services, media services (Sonarr/Radarr/Overseerr), and network interfaces. Built with Go, HTMX, Alpine.js, and Tailwind CSS.

| Dark Mode | Light Mode |
|-----------|-----------|
| ![Dark Mode](screenshots/dark.png) | ![Light Mode](screenshots/light.png) |

---

## What is dhiarhome?

**dhiarhome** is an ultra-lightweight web dashboard for homelab servers. Real-time visibility into:

- **Proxmox VE** — CPU, RAM, multi-disk usage, CPU model/core/thread info
- **Docker containers** — status and state
- **Web services** — uptime with response times
- **Media services** — Sonarr/Radarr/Overseerr stats with WebUI links
- **Bookmarks** — custom links with auto-fetched favicons
- **Network interfaces** — RX/TX speeds per interface
- **Interactive to-do list** — add, toggle, delete (persisted to JSON)
- **Weather + time** — combined card with live clock, Open-Meteo weather
- **System info** — hostname, OS, uptime, Go runtime stats
- **Glassmorphism UI** — transparent cards, custom backgrounds, live indicator
- **Telegram alerts** — service/Docker state transitions with cooldown, silent hours
- **Auto-refreshing** — HTMX polling with DOM diff (no flicker)
- **Mock mode** — test without real credentials
- **Single binary** — ~15MB, zero database

---

## Features

- **Proxmox** — CPU model + cores/threads, RAM, multi-disk, VM/LXC, uptime
- **Docker** — all containers with up/down status
- **Web services** — health checks with response times
- **Media services** — Sonarr/Radarr/Overseerr stats, clickable WebUI
- **Network** — per-interface RX/TX speeds (via /proc/net/dev)
- **Bookmarks** — custom links with auto-fetched favicons
- **To-do list** — Alpine.js interactive, persisted to JSON
- **Weather + time** — live clock, Open-Meteo forecast, timezone support
- **System info** — hostname, OS, uptime, Go memory
- **Glassmorphism UI** — blur cards, custom backgrounds, accent color, dark/light toggle
- **DOM diff swap** — no backdrop flicker on refresh
- **5s auto-refresh** — HTMX polling with merge-swap
- **Toast notifications** — real-time popup alerts on service/Docker state changes
- **Telegram notifications** — service/Docker up/down alerts with cooldown & silent hours
- **Security hardening** — CSP headers, SSRF protection, rate limiting, path traversal protection
- **Responsive** — 2-col mobile, 4-col desktop widget grid
- **YAML config** — no code changes needed
- **Mock mode** — test everything without real servers

---

## Tech Stack

- **Backend:** Go 1.26 (statically compiled, single binary)
- **Frontend:** HTML5 + Tailwind CSS + HTMX 1.9.10 + Alpine.js 3.x
- **Config:** YAML
- **Deploy:** Docker multi-stage or bare metal

---

## Quick Start

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/Alfar0nt/dhiarhome.git
cd dhiarhome

# Create your config file
cp config-example.yaml config.yaml
nano config.yaml  # Edit with your settings

# Build and run
docker build -t dhiarhome .
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  dhiarhome
```

Open `http://localhost:8080` in your browser.

### Option 2: Build from Source

```bash
# Install Go from https://go.dev/

# Clone and build
git clone https://github.com/Alfar0nt/dhiarhome.git
cd dhiarhome
cp config-example.yaml config.yaml
go build -o dhiarhome main.go
./dhiarhome               # uses config.yaml, :8080
./dhiarhome --config myconfig.yaml --addr :9090
```

For detailed deployment instructions, see [documentation/deployment.md](documentation/deployment.md).

---

## Configuration

All customization happens in `config.yaml`—no code changes needed!

### 1. Copy the example config
```bash
cp config-example.yaml config.yaml
```

### 2. Edit your settings

**Add websites to monitor:**
```yaml
services:
  - name: "My Website"
    url: "https://example.com"
  - name: "Nextcloud"
    url: "https://nextcloud.example.com"
```

**Configure Proxmox monitoring:**
```yaml
proxmox:
  url: "https://192.168.1.100:8006/api2/json"
  node_name: "pve"
  token_id: "root@pam!dashboard"
  token_secret: "YOUR-SECRET-UUID"
  mock: false  # Set to true for testing
```

**Filter Docker containers:**
```yaml
docker:
  socket: "unix:///var/run/docker.sock"
  monitor_containers:
    - "nginx"
    - "pihole"
```

### 3. Test without real data

Set `mock: true` in the Proxmox section to see fake bouncing data for UI testing.

---

## Documentation

Full documentation is available in the `/documentation` folder:

- **[docs.md](documentation/docs.md)** - Complete project documentation, architecture, and technical details
- **[deployment.md](documentation/deployment.md)** - Deployment guides (Docker, bare metal, systemd, reverse proxy)
- **[to-do.md](documentation/to-do.md)** - Feature implementation roadmap
- **[prompt-history.md](documentation/prompt-history.md)** - Development conversation log
- **[changelogs.md](documentation/changelogs.md)** - Version history and changes

---

## Project Structure

```
dhiarhome/
├── main.go                    # App entry point
├── config.yaml                # Your config (gitignored)
├── config-example.yaml        # Template
├── Dockerfile                 # Build
├── internal/
│   ├── bookmarks/             # Bookmark processing + favicon cache
│   ├── cache/                 # Service state cache
│   ├── config/                # YAML loader
│   ├── docker/                # Docker API client
│   ├── mediaservices/         # Sonarr/Radarr/Overseerr clients
│   ├── monitor/               # HTTP health checker
│   ├── network/               # /proc/net/dev monitor
│   ├── notifications/         # Telegram notifier
│   ├── proxmox/               # Proxmox API client
│   ├── todo/                  # Persistent to-do store
│   └── widgets/               # Weather, datetime, sysinfo, custom_text
├── static/
│   ├── index.html             # Dashboard page (Go template)
│   └── backgrounds/           # Custom bg images
├── templates/
│   ├── status.html            # Status page
│   ├── mediaservices.html     # Media services card
│   ├── todo.html              # To-do widget (Alpine.js)
│   ├── network.html           # Network card
│   └── widgets/               # Widget rendering
└── documentation/             # Docs
```

---

## Common Issues

**"No containers found"**
- Check if Docker socket is mounted: `-v /var/run/docker.sock:/var/run/docker.sock:ro`
- Verify container names in `monitor_containers` (or leave empty for all)

**"Proxmox API Error"**
- Verify API token ID and secret in `config.yaml`
- Test API access: `curl -k -H "Authorization: PVEAPIToken=TOKEN_ID=SECRET" https://PROXMOX_IP:8006/api2/json/nodes/pve/status`

**Services showing "Offline"**
- Test URLs from your server: `curl -I https://example.com`
- Check firewall rules and network connectivity

For more troubleshooting, see [documentation/deployment.md](documentation/deployment.md#troubleshooting).

---

## Why This Project?

Home servers often have limited resources. Many existing dashboards are heavy and require running databases or complex setups. dhiarhome provides:

- **Zero database** - All data fetched in real-time
- **Minimal resources** - ~10-20MB RAM, <1% CPU
- **Simple deployment** - Single binary or Docker container
- **Easy customization** - Edit YAML, not code

---

## Roadmap

- ✅ Background images, glassmorphism theme, accessibility
- ✅ Weather, datetime, system info, custom text widgets
- ✅ Network interface monitoring (RX/TX speeds)
- ✅ Interactive to-do list (Alpine.js)
- ✅ Media services (Sonarr, Radarr, Overseerr)
- ✅ Custom bookmarks and web links
- ✅ Telegram notifications (service/Docker alerts with cooldown & silent hours)
- ⬜ Additional service integrations (Plex, Portainer)
- ⬜ Generic HTTP API widget

> **v1.4.4 released** — all planned core features complete. Future work will focus on additional integrations.

---

## Contributing

This is a personal learning project, but feel free to fork and customize for your own use. If you find bugs or have suggestions, open an issue!

---

## License

This project is licensed under the [MIT License](LICENSE) — see the [LICENSE](LICENSE) file for details.

Copyright (c) 2026 Dhiar Harianto
