# dhiarhome

A lightweight, self-hosted homelab monitoring dashboard for Proxmox VE, Docker containers, and web services. Built with Go, HTMX, and Tailwind CSS.

![Dashboard Screenshot](Screenshot.png)

---

## What is dhiarhome?

**dhiarhome** is an ultra-lightweight web dashboard designed for monitoring homelab servers. It provides real-time visibility into:

- **Proxmox VE** server metrics (CPU, RAM, disk usage)
- **Docker containers** status and state
- **Web services** uptime with response times
- Auto-refreshing dashboard (no manual reload needed)
- Mock mode for testing without real credentials
- Single binary deployment (~10MB)

This is my personal learning project for homelab monitoring—simple, fast, and easy to customize.

---

## Features

- Real-time Proxmox server monitoring (CPU, memory, disk)
- Docker container status tracking
- Web service health checks with response times
- Auto-refreshing UI using HTMX (5-second polling)
- Configuration-driven (no code changes needed)
- Mock mode for UI testing
- Dark mode design with Tailwind CSS
- Lightweight: ~10-20MB RAM, <1% CPU

---

## Tech Stack

- **Backend:** Go 1.26.3 (statically compiled, single binary)
- **Frontend:** HTML5 + Tailwind CSS + HTMX 1.9.10
- **Configuration:** YAML files
- **Deployment:** Docker multi-stage build or bare metal

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
./dhiarhome
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
├── main.go                 # Application entry point
├── config.yaml             # Your configuration (gitignored)
├── config-example.yaml     # Configuration template
├── Dockerfile              # Multi-stage Docker build
├── internal/               # Application logic
│   ├── cache/             # Service state history
│   ├── config/            # YAML config loader
│   ├── docker/            # Docker API client
│   ├── monitor/           # HTTP health checker
│   └── proxmox/           # Proxmox API client
├── static/
│   └── index.html         # Main dashboard page
├── templates/
│   └── status.html        # Status template
└── documentation/         # Project documentation
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

Planned features (see [to-do.md](documentation/to-do.md) for details):

- Customizable background images
- Enhanced glassmorphism theme
- Weather and date/time widgets
- Network interface monitoring
- Custom bookmarks and links
- Service integrations (Plex, Radarr, Sonarr, Portainer)

---

## Contributing

This is a personal learning project, but feel free to fork and customize for your own use. If you find bugs or have suggestions, open an issue!

---

## License

Personal project for homelab monitoring. Free to modify and use.
