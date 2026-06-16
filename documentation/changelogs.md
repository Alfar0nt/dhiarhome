# Changelogs - dhiarhome

All notable changes to this project are documented in this file.

---

## [Unreleased] - Pre-Feature Roadmap

### Planned
- Customizable background images (config-based)
- Glassmorphism theme with transparent cards
- Weather widget (Open-Meteo API)
- Date/time widget with timezone support
- System info widget
- Network interface monitoring (speed, RX/TX, connections)
- Custom bookmarks and web links with icon support
- Service integration framework (Plex, Radarr, Sonarr, Portainer)
- Generic HTTP API widget for custom services

See [to-do.md](to-do.md) for the full 33-step implementation plan.

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
| 0.3.0 | Planned | Visual enhancements (background, glassmorphism, animations) |
| 0.4.0 | Planned | Utility widgets (weather, datetime, system info) |
| 0.5.0 | Planned | Network monitoring |
| 0.6.0 | Planned | Bookmarks and custom links |
| 0.7.0 | Planned | Service integration framework |
| 1.0.0 | Planned | First stable release with all planned features |
