# dhiarhome - Deployment Guide

This guide covers all deployment methods for dhiarhome, from Docker containers to bare metal installations.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Docker Deployment](#docker-deployment)
3. [Bare Metal / Building from Source](#bare-metal--building-from-source)
4. [Configuration](#configuration)
5. [Proxmox API Setup](#proxmox-api-setup)
6. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### For Docker Deployment
- **Docker** (version 20.10 or higher recommended)
- **Docker Compose** (optional, but recommended)
- **Git** (for cloning the repository)

### For Bare Metal Deployment
- **Go** (version 1.21 or higher; v1.0.0 built with Go 1.26)
- **Git**
- **Linux** (tested on Debian/Ubuntu, should work on any Linux distro)
- **Docker socket access** (if monitoring Docker containers)

### Network Requirements
- Access to Proxmox VE API endpoint (default port 8006)
- Access to Docker socket (`/var/run/docker.sock`) or TCP endpoint
- Outbound HTTP/HTTPS for service monitoring

---

## Docker Deployment

### Method 1: Standard Docker Build

#### Step 1: Clone the Repository
```bash
git clone https://github.com/Alfar0nt/dhiarhome.git
cd dhiarhome
```

#### Step 2: Create Configuration File
```bash
cp config-example.yaml config.yaml
```

Edit `config.yaml` with your settings:
```bash
nano config.yaml
```

See the [Configuration](#configuration) section for details.

#### Step 3: Build the Docker Image
```bash
docker build -t dhiarhome .
```

This creates a multi-stage build:
- Stage 1: Compiles Go binary using `golang:alpine`
- Stage 2: Creates minimal runtime image using `alpine:latest`

Final image size: ~20-25 MB

#### Step 4: Run the Container
```bash
docker run -d \
  --name dhiarhome \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  --restart unless-stopped \
  dhiarhome
```

**Flags Explained:**
- `-d` - Run in detached mode (background)
- `--name dhiarhome` - Container name
- `-p 8080:8080` - Map host port 8080 to container port 8080
- `-v $(pwd)/config.yaml:/app/config.yaml` - Mount config file
- `-v /var/run/docker.sock:/var/run/docker.sock:ro` - Mount Docker socket (read-only)
- `--restart unless-stopped` - Auto-restart on failure

#### Step 5: Access the Dashboard
Open your browser and navigate to:
```
http://localhost:8080
```

Or use your server's IP:
```
http://YOUR_SERVER_IP:8080
```

---

### Method 2: Docker Compose (Recommended)

#### Step 1: Create `docker-compose.yml`
```yaml
version: '3.8'

services:
  dashboard:
    build: .
    container_name: dhiarhome
    ports:
      - "8080:8080"
    volumes:
      - ./config.yaml:/app/config.yaml
      - /var/run/docker.sock:/var/run/docker.sock:ro
    restart: unless-stopped
```

#### Step 2: Start the Service
```bash
docker-compose up -d
```

#### Step 3: View Logs
```bash
docker-compose logs -f
```

#### Step 4: Stop the Service
```bash
docker-compose down
```

---

### Method 3: Pre-built Image (If Available)

Pre-built images are available via GitHub Container Registry:

```bash
docker pull ghcr.io/alfar0nt/dhiarhome:v1.0.0
```

```bash
docker run -d \
  --name dhiarhome \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  --restart unless-stopped \
  ghcr.io/alfar0nt/dhiarhome:v1.0.0
```

---

### Docker Security Hardening

For production deployments, add security flags:

```bash
docker run -d \
  --name dhiarhome \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  --user 1000:1000 \
  --read-only \
  --tmpfs /tmp \
  --cap-drop ALL \
  --security-opt no-new-privileges \
  --restart unless-stopped \
  dhiarhome
```

**Security Flags:**
- `--user 1000:1000` - Run as non-root user
- `--read-only` - Read-only root filesystem
- `--tmpfs /tmp` - Writable temp directory
- `--cap-drop ALL` - Drop all Linux capabilities
- `--security-opt no-new-privileges` - Prevent privilege escalation

---

## Bare Metal / Building from Source

### Method 1: Build from Source

#### Step 1: Install Go
Download and install Go from [golang.org](https://go.dev/dl/):

```bash
# Download Go (check for latest version at https://go.dev/dl/)
wget https://go.dev/dl/go1.26.0.linux-amd64.tar.gz

# Extract to /usr/local
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.26.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

#### Step 2: Clone the Repository
```bash
git clone https://github.com/Alfar0nt/dhiarhome.git
cd dhiarhome
```

#### Step 3: Create Configuration
```bash
cp config-example.yaml config.yaml
nano config.yaml
```

#### Step 4: Build the Binary
```bash
go build -o dhiarhome main.go
```

For optimized production build:
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o dhiarhome main.go
```

**Build Flags:**
- `CGO_ENABLED=0` - Disable C dependencies (static binary)
- `GOOS=linux` - Target Linux OS
- `GOARCH=amd64` - Target 64-bit x86
- `-ldflags="-w -s"` - Strip debug info (smaller binary)

#### Step 5: Run the Application
```bash
./dhiarhome
```

dhiarhome will start on port 8080 using `config.yaml` in the current directory.

**Command-line flags:**
```bash
./dhiarhome                                    # default: config.yaml, :8080
./dhiarhome --config /path/to/config.yaml      # custom config path
./dhiarhome --addr :9090                        # custom port
./dhiarhome --config demo.yaml --addr :9090     # both
```

Available flags: `--config` (config file path, default `config.yaml`), `--addr` (listen address, default `:8080`).

#### Step 6: Run as Background Service

**Option A: Using nohup**
```bash
nohup ./dhiarhome > dhiarhome.log 2>&1 &
```

**Option B: Using screen/tmux**
```bash
screen -S dhiarhome
./dhiarhome
# Press Ctrl+A, then D to detach
```

**Option C: Using systemd (Recommended)**

Create a systemd service file:
```bash
sudo nano /etc/systemd/system/dhiarhome.service
```

```ini
[Unit]
Description=dhiarhome - Homelab Dashboard
After=network.target docker.service

[Service]
Type=simple
User=your-username
WorkingDirectory=/path/to/dhiarhome
ExecStart=/path/to/dhiarhome/dhiarhome
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable dhiarhome
sudo systemctl start dhiarhome
```

Check status:
```bash
sudo systemctl status dhiarhome
```

View logs:
```bash
sudo journalctl -u dhiarhome -f
```

---

### Method 2: Cross-Compile for Different Architectures

**For ARM64 (Raspberry Pi, ARM servers):**
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o dhiarhome-arm64 main.go
```

**For ARM (32-bit):**
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-w -s" -o dhiarhome-arm main.go
```

**For Windows:**
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o dhiarhome.exe main.go
```

**For macOS:**
```bash
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o dhiarhome-mac main.go
```

---

## Configuration

### Step 1: Copy Example Configuration
```bash
cp config-example.yaml config.yaml
```

### Step 2: Edit Configuration

#### Proxmox Configuration

```yaml
proxmox:
  url: "https://192.168.1.100:8006/api2/json"
  node_name: "pve"
  token_id: "root@pam!dashboard"
  token_secret: "YOUR-SECRET-UUID-HERE"
  mock: false
```

**Field Descriptions:**
- `url` - Proxmox VE API endpoint (include `/api2/json`)
- `node_name` - Name of the Proxmox node to monitor
- `token_id` - API token ID (format: `user@realm!tokenname`)
- `token_secret` - API token secret (UUID)
- `mock` - Set to `true` for testing without real Proxmox server

#### Docker Configuration

**Local Docker Socket (default):**
```yaml
docker:
  socket: "unix:///var/run/docker.sock"
  monitor_containers:
    - "nginx"
    - "pihole"
```

**Remote Docker with TLS:**
```yaml
docker:
  socket: "tcp://docker.example.com:2376"
  skip_tls: false                     # set true for self-signed certs
  tls_ca_cert: "/path/to/ca.pem"      # CA certificate
  tls_cert: "/path/to/cert.pem"       # client certificate (mTLS)
  tls_key: "/path/to/key.pem"         # client key
```

**Portainer API (takes priority over socket/TCP):**
```yaml
docker:
  portainer_url: "https://portainer.example.com"
  portainer_api_key: "ptr_XXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  portainer_env_id: 1                  # Portainer environment/endpoint ID
```

**Field Descriptions:**
- `socket` - Docker socket path, TCP endpoint, or HTTPS URL
  - Unix socket: `unix:///var/run/docker.sock`
  - TCP: `tcp://192.168.1.100:2375` (no TLS) or `tcp://docker.example.com:2376` (with TLS)
- `skip_tls` - Skip TLS certificate verification (for self-signed certs)
- `tls_ca_cert` / `tls_cert` / `tls_key` - Paths to TLS certificates for mTLS
- `portainer_url` - Portainer instance URL (when set, containers are fetched via Portainer API)
- `portainer_api_key` - API access token (from Portainer > Account Settings > Access tokens)
- `portainer_env_id` - Portainer environment/endpoint ID (default: 1)
- `monitor_containers` - List of container names to monitor (leave empty for all)

> **Connection priority:** Portainer > Remote Docker (TCP/TLS) > Local socket

#### Services Configuration

```yaml
services:
  - name: "Personal Website"
    url: "https://example.com"
  - name: "Nextcloud"
    url: "https://nextcloud.example.com"
  - name: "PDF Tools"
    url: "https://pdftools.example.com"
```

Add as many services as you want to monitor. The dashboard will check each URL and display online/offline status.

### Step 3: Test Configuration

Start the application and check the logs:
```bash
./dhiarhome
```

Or view Docker logs:
```bash
docker logs dhiarhome
```

---

## Proxmox API Setup

### Step 1: Create API Token

1. Log in to Proxmox VE web interface
2. Navigate to **Datacenter** → **Permissions** → **API Tokens**
3. Click **Add**
4. Fill in the form:
   - **User**: Select user (e.g., `root@pam`)
   - **Token ID**: Enter a name (e.g., `dashboard`)
   - **Comment**: Optional description
   - **Expire**: Leave blank for no expiration
   - **Privilege Separation**: Uncheck (use user permissions)
5. Click **Add**
6. **Copy the token secret immediately** (shown only once)

Token format: `root@pam!dashboard=YOUR-SECRET-UUID`
- `token_id`: `root@pam!dashboard`
- `token_secret`: `YOUR-SECRET-UUID`

### Step 2: Set Permissions

1. Navigate to **Datacenter** → **Permissions**
2. Add permission for the API token user:
   - **Path**: `/nodes/YOUR_NODE_NAME`
   - **Role**: `PVEAuditor` (read-only)
   - **User/Group**: Select the API token user

Or use more granular permissions:
- `/nodes` - `PVEAuditor`
- `/vms` - `PVEAuditor` (if monitoring VMs)

### Step 3: Test API Access

Test the API from your server:
```bash
curl -k -H "Authorization: PVEAPIToken=root@pam!dashboard=YOUR-SECRET" \
  https://192.168.1.100:8006/api2/json/nodes/pve/status
```

You should receive JSON with CPU, memory, and disk stats.

---

## Troubleshooting

### Issue: "Error loading config"

**Solution:**
- Ensure `config.yaml` exists in the same directory as the binary
- Check YAML syntax (use a YAML validator)
- Verify file permissions

### Issue: "Proxmox API Error: status 401"

**Solution:**
- Check API token ID and secret in `config.yaml`
- Verify token has not expired
- Ensure user has proper permissions
- Test API access with curl (see above)

### Issue: "Docker API Error: connection refused"

**Solution:**
- Verify Docker is running: `systemctl status docker`
- Check socket path: `ls -la /var/run/docker.sock`
- Ensure user has permission to access socket
- For Docker deployment, verify socket is mounted:
  ```bash
  docker inspect dhiarhome | grep -A5 Mounts
  ```

### Issue: "No containers found"

**Solution:**
- Check if containers exist: `docker ps -a`
- Verify `monitor_containers` list in config (or leave empty for all)
- Check container name format (with or without leading `/`)

### Issue: Services showing as "Offline"

**Solution:**
- Verify URLs are accessible from the server:
  ```bash
  curl -I https://example.com
  ```
- Check firewall rules
- Ensure URLs include protocol (`https://`)

### Issue: Dashboard not accessible from other machines

**Solution:**
- Check if application is listening on all interfaces:
  ```bash
  netstat -tlnp | grep 8080
  ```
  Should show `0.0.0.0:8080` or `:::8080`
- Check firewall rules:
  ```bash
  sudo ufw status
  sudo iptables -L -n
  ```
- For Docker, ensure port mapping is correct: `-p 8080:8080`

### Issue: High CPU/Memory usage

**Solution:**
- This is unusual; the dashboard should use <1% CPU and ~20MB RAM
- Check if mock mode is enabled (generates random data constantly)
- Reduce number of monitored services
- Check for infinite loops in logs

### Issue: Template execution error

**Solution:**
- Ensure `templates/` and `static/` directories exist
- Verify file permissions
- Check that application is run from correct working directory

---

## Reverse Proxy Setup (Optional)

### Nginx

```nginx
server {
    listen 80;
    server_name dashboard.example.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Caddy

```
dashboard.example.com {
    reverse_proxy localhost:8080
}
```

### Apache

```apache
<VirtualHost *:80>
    ServerName dashboard.example.com
    
    ProxyPreserveHost On
    ProxyPass / http://localhost:8080/
    ProxyPassReverse / http://localhost:8080/
</VirtualHost>
```

---

## Monitoring and Maintenance

### View Logs

**Docker:**
```bash
docker logs dhiarhome
docker logs -f dhiarhome  # Follow mode
```

**Systemd:**
```bash
sudo journalctl -u dhiarhome
sudo journalctl -u dhiarhome -f  # Follow mode
```

**Manual:**
```bash
tail -f dhiarhome.log
```

### Update the Dashboard

**Docker:**
```bash
cd dhiarhome
git pull
docker-compose down
docker-compose build
docker-compose up -d
```

**Bare Metal:**
```bash
cd dhiarhome
git pull
go build -o dhiarhome main.go
sudo systemctl restart dhiarhome
```

### Backup Configuration

```bash
cp config.yaml config.yaml.backup
```

Store backups in a secure location (contains API secrets).

---

## Performance Tuning

### Reduce Polling Frequency

Edit `main.go` and change the ticker interval:
```go
ticker := time.NewTicker(30 * time.Second) // Change from 10s to 30s
```

### Increase Cache Size

Edit `main.go`:
```go
historyCache = cache.NewHistoryCache(500) // Change from 100 to 500
```

### Optimize Docker Image

The Dockerfile already uses multi-stage builds. For even smaller images, consider:
- Using `scratch` instead of `alpine` (no shell, minimal)
- UPX compression of the binary

---

## Support and Issues

For bugs, questions, or feature requests, please open an issue on GitHub.

---

## Quick Reference Commands

```bash
# Build Docker image
docker build -t dhiarhome .

# Run container
docker run -d -p 8080:8080 -v $(pwd)/config.yaml:/app/config.yaml -v /var/run/docker.sock:/var/run/docker.sock:ro dhiarhome

# Build binary
go build -o dhiarhome main.go

# Run binary
./dhiarhome

# Start systemd service
sudo systemctl start dhiarhome

# View Docker logs
docker logs -f dhiarhome

# Test Proxmox API
curl -k -H "Authorization: PVEAPIToken=TOKEN_ID=SECRET" https://PROXMOX_IP:8006/api2/json/nodes/NODE_NAME/status
```
