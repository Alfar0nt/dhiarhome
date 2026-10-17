# CHANGELOGS — dhiarhome

> Compressed from former `documentation/changelogs.md` (dropped per-version "Files Modified" lists and verbose internals). Newest first. See [`../.ai/IMPLEMENTATION-PLAN.md`](../.ai/IMPLEMENTATION-PLAN.md) for roadmap.

## Version History Summary

| Version | Date | Description |
|---------|------|-------------|
| 1.5.3 | 2026-06-20 | Phase 17: Todo Telegram notifications |
| 1.5.2 | 2026-06-20 | Phase 16: VM/LXC sort fix |
| 1.5.1 | 2026-06-20 | Phase 15: font size & spacing consistency |
| 1.5.0 | 2026-06-20 | Phase 14: to-do edit |
| 1.4.5 | 2026-06-20 | Phase 13: background polling & instant rendering |
| 1.4.4 | 2026-06-20 | GitHub button, footer, security headers, SSRF |
| 1.4.3 | 2026-06-20 | CLI flags `--config`, `--addr` |
| 1.4.2 | 2026-06-19 | Security hardening: stripped binaries, config perms |
| 1.4.1 | 2026-06-19 | Web UI toast notifications, Telegram URL |
| 1.4.0 | 2026-06-19 | Phase 11: Telegram notifications |
| 1.3.2 | 2026-06-19 | Phase 10: logo, theme toggle, bigger text |
| 1.3.1 | 2026-06-19 | Todo modal, CPU/Memory widget, date tracking |
| 1.3.0 | 2026-06-19 | Phase 9: remote Docker TLS + Portainer |
| 1.2.0 | 2026-06-19 | Phase 8: extra disk monitoring |
| 1.1.0 | 2026-06-19 | Phase 7: Proxmox API enrichment |
| 1.0.1 | 2026-06-18 | `skip_tls` for services |
| 1.0.0 | 2026-06-18 | First stable release |
| 0.10.1 | 2026-06-18 | Security hardening |
| 0.10.0 | 2026-06-18 | Phase 6: polish, performance, docs |
| 0.9.2 | 2026-06-18 | Scrollbar theme unification |
| 0.9.1 | 2026-06-18 | Internal scrolling & UI refinements |
| 0.9.0 | 2026-06-18 | Phase 4: custom links & bookmarks |
| 0.8.3 | 2026-06-18 | Proxmox layout consolidation, LXC/VM monitoring |
| 0.8.2 | 2026-06-17 | Media services, widget row redesign, live badge |
| 0.8.1 | 2026-06-17 | Widget scroll & sizing fixes |
| 0.8.0 | 2026-06-17 | Widget row & media services |
| 0.7.2 | 2026-06-17 | Todo reactivity, multi-disk, CPU name, mobile fixes |
| 0.7.1 | 2026-06-16 | Todo bug fix & compact redesign |
| 0.7.0 | 2026-06-16 | Interactive to-do list & CPU core/thread |
| 0.6.0 | 2026-06-16 | Backdrop-flicker elimination (DOM diff swap) |
| 0.5.3 | 2026-06-16 | Network display caching & alignment |
| 0.5.2 | 2026-06-16 | Widget stability & layout |
| 0.5.1 | 2026-06-16 | Dashboard layout refinements |
| 0.5.0 | 2026-06-16 | Phase 3: network monitoring |
| 0.4.0 | 2026-06-16 | Phase 2: utility & info widgets |
| 0.3.0 | 2026-06-16 | Phase 1: visual enhancements |
| 0.2.0 | 2026-06-16 | Project rebrand to "dhiarhome" |
| 0.1.0 | Pre-2026-06-16 | Original "Selfhosted Proxmox Dashboard" |

---

## [1.5.3] — 2026-06-20 — Phase 17: Todo Telegram Notifications
- Added `notify_todo_add` / `notify_todo_complete`; messages include task text, timestamp, and remaining incomplete tasks list.
- Respects silent hours; **no cooldown** (user-initiated actions).

## [1.5.2] — 2026-06-20 — Phase 16: VM/LXC Sort Fix
- Fixed VM/LXC lists shuffling on each refresh — now sorted by VMID ascending in `GetVirtualization()`.

## [1.5.1] — 2026-06-20 — Phase 15: Font & Spacing Consistency
- Bookmarks/Services section titles `text-xl → text-2xl`, icons `w-6 → w-7`, `mb-4 → mb-5`; Services item spacing/padding aligned with Docker.

## [1.5.0] — 2026-06-20 — Phase 14: To-Do Edit
- `Store.Update()` + `PATCH /api/todos/{id}`; inline edit (pencil icon) in compact widget and full-screen modal; Enter save / Escape cancel; optimistic updates with API fallback.

## [1.4.5] — 2026-06-20 — Phase 13: Background Polling
- Proxmox + Docker polling moved to background goroutines; `statusHandler` cache-only (<5 ms); initial blocking polls at startup warm the cache.

## [1.4.4] — 2026-06-20 — GitHub Button, Footer & Security Hardening
- GitHub header button; glassmorphism footer.
- Security headers middleware (CSP, X-Frame-Options, X-Content-Type-Options, X-XSS-Protection, Referrer-Policy); SSRF protection for favicon fetcher; JSON injection fix via `json.NewEncoder`.

## [1.4.3] — 2026-06-20 — CLI Flags
- `--config <path>` (default `config.yaml`) and `--addr <addr>` (default `:8080`).

## [1.4.2] — 2026-06-19 — Security Hardening
- Binaries stripped (14 MB → 9.6 MB); `config.yaml` permissions `0600`. Full audit: no hardcoded secrets; remaining findings theoretical for local homelab.

## [1.4.1] — 2026-06-19 — Toast Notifications & Telegram URL
- Web UI toasts for service/Docker transitions (green/red, auto-dismiss 4 s) via `TransitionEvent` buffer.
- `NotifyServiceChange()` includes service URL.

## [1.4.0] — 2026-06-19 — Phase 11: Telegram Notifications
- `Notifier.SendMessage()` (HTML); service + Docker up/down alerts; cooldown (default 5 min); silent hours; mock mode; `GET /api/notifications/test`.

## [1.3.2] — 2026-06-19 — Phase 10: Logo, Theme Toggle, Bigger Text
- Configurable logo (`appearance.logo`, file/URL) as favicon + header; `/logo` endpoint; inline SVG favicon fallback.
- Dark/light theme toggle (localStorage, default from config); light-mode readability overrides (text, borders, status colors, scrollbars, modal).
- Global font-size bumps across Proxmox/Docker/Services/Media/Bookmarks/Widgets.

## [1.3.1] — 2026-06-19 — Todo Modal, CPU/Memory Widget, Date Tracking
- Full-screen todo modal (larger text/touch targets, shared Alpine state); `Added`/`Done` date metadata with smart formatting.
- CPU/Memory card stretches to fill row; thicker bars; divider; `done_at` persistence fix.

## [1.3.0] — 2026-06-19 — Phase 9: Remote Docker & Portainer
- Remote Docker via TCP + TLS (mTLS, `skip_tls`); Portainer API integration; connection priority Portainer > Remote > Local.

## [1.2.0] — 2026-06-19 — Phase 8: Extra Disk Monitoring
- `ExtraDiskConfig` (auto-detect via `syscall.Statfs`, or manual total/used); `ParseSize()`; merge + dedup by mountpoint.

## [1.1.0] — 2026-06-19 — Phase 7: Proxmox API Enrichment
- Swap usage (color-coded), load average (1m/5m/15m), PVE + kernel version footer; VM/LXC enumeration (VMID/name/status, scrollable); mock data for all.

## [1.0.1] — 2026-06-18 — Skip TLS for Services
- `skip_tls` per service for self-signed certs. Bookmarks config reference added.

## [1.0.0] — 2026-06-18 — First Stable Release
- All core features complete/tested: Proxmox, Docker, services, media services, network, todo, weather+time, sysinfo, bookmarks, glassmorphism UI, DOM-diff swap, security hardening, mock mode, single ~14 MB binary, zero DB.

## [0.10.1] — 2026-06-18 — Security Hardening
- Security headers middleware; per-IP rate limiter (30 req/min on todo APIs); path-traversal protection; 500-char input cap. Dockerfile ships `config-example.yaml` (no real creds).

## [0.10.0] — 2026-06-18 — Phase 6: Polish, Performance & Docs
- Graceful shutdown (SIGINT/SIGTERM); HTTP server timeouts; config `Validate()` (ranges, URLs, mock fallback, weather/auto-disable, timezone fallback); startup feature summary. Phase 5 deferred.

## [0.9.2] — 2026-06-18 — Scrollbar Theme Unification
- Consistent thin scrollbars across todo, bookmarks, services, docker, and main page (Firefox + WebKit).

## [0.9.1] — 2026-06-18 — Internal Scrolling & UI Refinements
- Bookmark icon/text scale-up; internal scroll for bookmarks (`max-h-[200px]`), services (`max-h-[230px]`), docker (`max-h-[230px]`). Added mock bookmarks/services/containers.

## [0.9.0] — 2026-06-18 — Phase 4: Custom Links & Web Bookmarks
- Bookmark groups + links; icon modes (Lucide SVG / image path / auto favicon to `data/icons/`); combined Bookmarks+Services card (col-span-2); flat 5-col grid; internal scrolling.

## [0.8.3] — 2026-06-18 — Proxmox Layout Consolidation & LXC/VM Monitoring
- Virtualization Overview card (active/total counts); `GetVirtualization()` + mock; CPU+Memory merged into one card; Proxmox grid → 3 columns.

## [0.8.2] — 2026-06-17 — Media Services, Widget Row, Live Badge
- Sonarr/Radarr/Overseerr clients (30 s poll, mock stats); live indicator glass pill; widget cards `min-h-[190px]`; todo moved to widget row; custom_text removed from render.

## [0.8.1] — 2026-06-17 — Widget Scroll & Sizing Fixes
- Todo 3rd-item clipping + mobile input overflow fixes; uniform `min-h-[190px]` widget cards.

## [0.8.0] — 2026-06-17 — Widget Row & Media Services
- Media Management card; todo in widget row; optimistic todo updates; `:key` fix on `<template>` child; removed race condition in todo add.

## [0.7.2] — 2026-06-17 — Todo Reactivity, Multi-Disk, CPU Name, Mobile Fixes
- Todo `Array.concat()`/`map()` for Alpine reactivity; mobile background scroll fix (`#scroll-container`, 100dvh); CPU model name (`cleanCPUName()`); multi-disk (`Disks []DiskInfo`); scrollable disk widget; `roundDur` tempate func.

## [0.7.1] — 2026-06-16 — Todo Bug Fix & Compact Redesign
- `data-preserve` on todo card (merge-swap skipped) fixes add-after-poll; compact spacing redesign.

## [0.7.0] — 2026-06-16 — Interactive To-Do & CPU Core/Thread
- Interactive todo widget (add/toggle/delete) persisted to `data/todos.json`; CRUD API; CPU core/thread count from `/proc/cpuinfo`; Alpine.js added.

## [0.6.0] — 2026-06-16 — Backdrop-Flicker Elimination
- Custom HTMX `merge-swap` extension (`mergeDOM`/`patchChildren`/`syncAttrs`) updates only text nodes + changed attrs → preserves `backdrop-filter` GPU layers, no flicker.

## [0.5.3] — 2026-06-16 — Network Display Caching & Alignment
- `GetStats()` caches formatted output 10 s; network card mobile alignment fixes.

## [0.5.2] — 2026-06-16 — Widget Stability & Layout
- Mock weather cached 5 min; server date includes weekday (matches JS clock); network moved to widget row; grid `grid-cols-2 lg:grid-cols-4`.

## [0.5.1] — 2026-06-16 — Dashboard Layout Refinements
- Weather+DateTime combined card; custom text first; compact widget grid; network repositioned; standalone weather/datetime fallbacks.

## [0.5.0] — 2026-06-16 — Phase 3: Network Monitoring
- `internal/network` package; `/proc/net/dev` parser; two-sample speed + moving average; human-readable formatting; mock mode; per-interface cards with RX/TX.

## [0.4.0] — 2026-06-16 — Phase 2: Utility & Info Widgets
- Widgets package (interface + registry); weather (Open-Meteo, WMO codes, cache), datetime (client clock, timezone, 12/24h), sysinfo (hostname/OS/uptime/runtime), custom text (escaped); responsive widget grid.

## [0.3.0] — 2026-06-16 — Phase 1: Visual Enhancements
- `AppearanceConfig`; custom background (local/URL, overlay, blur, `/background` + `/api/background`); glassmorphism (`glass-card`/`glass-inner`); Inter typography; animations (`prefers-reduced-motion` respected); WCAG 2.1 AA accessibility pass.

## [0.2.0] — 2026-06-16 — Project Rebrand
- Rebrand "Selfhosted Proxmox Dashboard" → **dhiarhome** (module, imports, image/container/binary names, service name, repo URL). Added `documentation/` set + rewritten README.

## [0.1.0] — Pre-Rebrand — Original State
- "Selfhosted Proxmox Dashboard": Proxmox monitoring (CPU/RAM/disk/uptime, mock), Docker container listing + filter, web service health checks (10 s poll), dark Tailwind UI with HTMX 5 s polling, YAML config.
