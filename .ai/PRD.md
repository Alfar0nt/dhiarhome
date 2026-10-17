# PRD — dhiarhome

> Product Requirements Document. Source-refactored from former `documentation/docs.md` + `README.md`. Technical details: [`TRD.md`](TRD.md). Flow: [`APP-FLOW.md`](APP-FLOW.md). UI: [`DESIGN.md`](DESIGN.md). Roadmap: [`IMPLEMENTATION-PLAN.md`](IMPLEMENTATION-PLAN.md).

## 1. Overview

**dhiarhome** is an ultra-lightweight, self-hosted homelab monitoring dashboard. It gives real-time visibility into server health (Proxmox VE), Docker containers, web service uptime, media services, network interfaces, bookmarks, and an interactive to-do list — from a single binary, zero database.

## 2. Problem & Purpose

Home servers have limited CPU/RAM. Existing dashboards are heavy (databases, complex setup). dhiarhome targets:

- **Zero database** — data fetched in real-time or cached in-memory.
- **Minimal resources** — single statically-compiled Go binary.
- **Configuration-driven** — customize via YAML, no code changes.
- **Real-time** — auto-refreshing UI via HTMX polling.

## 3. Target Users

- Homelab / self-hosting enthusiasts running Proxmox VE (single node).
- Users who want a fast, low-footprint dashboard deployable as one binary or Docker container.

## 4. Goals / Non-Goals

**Goals**
- Single-node Proxmox + Docker + service monitoring with <1% CPU, ~10-20 MB RAM.
- All background data cached so HTTP request path never blocks on external APIs.
- Rich but configurable UI (glassmorphism, dark/light, widgets).

**Non-Goals (current)**
- Multi-node Proxmox, historical metrics/graphs, built-in auth, non-Telegram notifications, dashboard TLS (see §9 Limitations).

## 5. Functional Requirements

### 5.1 Proxmox VE Monitoring
- CPU usage %, CPU model name + cores/threads (from `/proc/cpuinfo`).
- RAM usage (used/total bar), Swap usage (color-coded bar, hidden when no swap).
- Load average (1m/5m/15m).
- Disk usage: root FS + multi-disk from Proxmox API; **extra disks** via auto-detect (`statfs`) or manual `total`/`used` strings.
- VM/LXC enumeration with running/stopped status, VMID, type; **sorted by VMID** (stable order).
- Uptime, PVE version, kernel version.
- Mock mode (random realistic data) for testing without credentials.

### 5.2 Docker Monitoring
- List containers with state (running/exited/stopped) + status/uptime.
- Optional name filter (`monitor_containers`).
- Connections: local Unix socket, remote TCP/TLS (mTLS, `skip_tls`), or **Portainer API**. Priority: Portainer > Remote Docker > Local socket.

### 5.3 Web Service Health Checks
- HTTP/HTTPS endpoint monitoring; response time; status Online/Offline/Warning.
- `skip_tls` per service for self-signed certs.

### 5.4 Media Services
- Sonarr (series count, wanted), Radarr (movie count, wanted), Overseerr (pending requests, available media).
- `X-Api-Key` auth; graceful failure (`Online: false`); clickable WebUI link; mock stats.

### 5.5 Network Monitoring
- Parse `/proc/net/dev`; per-interface RX/TX speed (moving-average smoothed) + cumulative totals.
- Configurable interfaces with labels; mock mode.

### 5.6 Bookmarks
- Grouped links (flattened in UI) with icons: Lucide SVG name, custom image path, or auto-fetched favicon (cached to `data/icons/`).
- `new_tab` option; internal scrolling beyond ~10 links.

### 5.7 Interactive To-Do
- Add, toggle, delete, **inline edit**; persisted to JSON (`data/todos.json`).
- Optimistic updates via `fetch()`; compact widget + full-screen modal with `Added`/`Done` date tracking.
- 500-char cap; per-IP rate limiting (30 req/min).

### 5.8 Utility Widgets
- Weather (Open-Meteo, no key), Date/Time (client-side live clock, timezone, 12h/24h), System Info (hostname/OS/uptime/Go runtime), Custom Text (HTML-escaped).
- Weather + DateTime combined into one card; per-widget `enabled` flag.

### 5.9 Notifications
- **Telegram**: service + Docker state transitions (up/down), todo add/complete, cooldown, silent hours, mock/dry-run, `/api/notifications/test`.
- **Web UI toasts**: top-right popups (green=recovered, red=down), auto-dismiss 4s, driven by same transition buffer.

### 5.10 UI / UX
- Glassmorphism; custom background (local/URL) with overlay opacity + blur; accent color.
- Configurable logo (favicon + header); dark/light theme toggle persisted to `localStorage`.
- Responsive: 2-col mobile → 4-col desktop widget grid; DOM-diff swap (no flicker) on 5s refresh.
- Accessibility: WCAG 2.1 AA contrast, ARIA labels, focus rings, `prefers-reduced-motion`.

### 5.11 Security
- Security headers (CSP, nosniff, X-Frame-Options DENY, X-XSS-Protection, Referrer-Policy).
- SSRF protection (favicon fetcher), path-traversal protection, JSON-injection-safe encoding, rate limiting, input validation.

## 6. Non-Functional Requirements

| Aspect | Target |
|---|---|
| Memory | ~10–20 MB typical |
| CPU | <1% idle, brief spikes on poll |
| Binary | ~14 MB static (stripped ~9.6 MB), zero DB |
| Startup | <1 s |
| `/status` response | <5 ms (cache reads only) |
| Config | Single YAML file |
| Deploy | Docker (multi-stage) or bare-metal binary |

## 7. Configuration Surface

All customization via `config.yaml` (see [`TRD.md`](TRD.md) §7 for full reference): `proxmox`, `docker`, `services`, `appearance`, `widgets`, `network`, `todos`, `media_services`, `bookmarks`, `notifications.telegram`.

## 8. Success Metrics

- Lighthouse Performance >90, Accessibility >90.
- Dashboard usable with mock mode (no real credentials).
- Backward-compatible config (old configs still load).

## 9. Limitations

1. Single Proxmox node per instance.
2. No historical data / graphs (planned — Phase 12).
3. No built-in authentication (use reverse proxy).
4. Notifications limited to Telegram (no email/Discord/webhook).
5. Service checks HTTP/HTTPS only.
6. Dashboard itself has no TLS (use reverse proxy).

## 10. Roadmap

Shipped: Phases 1–17 (see [`IMPLEMENTATION-PLAN.md`](IMPLEMENTATION-PLAN.md)). Pending: historical graphs (Phase 12), Telegram bot (Phase 18), remaining service-integration framework (Phase 5), backlog fixes. Full detail in [`IMPLEMENTATION-PLAN.md`](IMPLEMENTATION-PLAN.md).

## 11. License

MIT — Copyright (c) 2026 Dhiar Harianto.
