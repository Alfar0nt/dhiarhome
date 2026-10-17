# DESIGN — dhiarhome

> UI/UX design system. Product context: [`PRD.md`](PRD.md) §5.10. Technical: [`TRD.md`](TRD.md). Flow: [`APP-FLOW.md`](APP-FLOW.md).

## 1. Design Language

**Glassmorphism**: semi-transparent blurred cards over a custom background. Implemented with CSS variables in `static/index.html`.

### CSS tokens (dark theme defaults)
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
- `.glass-card` — backdrop blur, semi-transparent bg, subtle border + inner shadow.
- `.glass-inner` — nested glass surface (used by pills, sub-cards).
- Hover: `translateY(-2px)` + glow shadow, 200 ms ease.
- GPU compositing forced via `translateZ(0)` / `will-change: transform` to avoid backdrop blur flicker.

## 2. Theming

- **Dark (default)** and **Light** via config `appearance.theme`.
- **Toggle**: sun/moon button in header; persisted to `localStorage` (`dhiarhome-theme`); user override wins over config.
- **Light mode overrides**: body bg `#e2e8f0`; text `text-white`/`text-gray-300` → dark (`#1e293b`/`#334155`); status colors use darker tones; `bg-white/*` & `border-white/*` → dark-on-light; scrollbars + skeletons adapted; background overlay `rgba(15,23,42,0.15)` (keeps bg image visible).
- **Todo modal light mode**: overlay `bg-gray-900/95` → `rgba(226,232,240,0.95)`; inputs/borders/buttons adapted; button text dark amber `#b45309`.

## 3. Typography

- Font: **Inter** (Google Fonts CDN, `display=swap`); stack `Inter, system-ui, -apple-system, sans-serif`.
- Headings `tracking-tight`; numeric values `tabular-nums` (no jitter).
- `.metric-label` — uppercase, letter-spacing, muted, `0.75rem`.
- `.metric-value` — tight letter-spacing, bold.
- Widget font scale: values `text-base`, section titles `text-xl`/`text-2xl`, labels `text-xs`.

> Note: CSP currently blocks Google Fonts stylesheet → Inter falls back to system stack after cache expiry (backlog B.1).

## 4. Layout

### 4.1 Page shell (`static/index.html`)
```
#scroll-container (overflow-y auto, 100dvh)
 ├─ #bg-image, #bg-overlay        # background + dark overlay
 ├─ header: [logo] dhiarhome · "Lightweight homelab monitoring dashboard"
 │          right: [GitHub btn] [theme toggle] [Live pill]
 ├─ #dashboard-content (hx-ext="merge-swap", aria-live="polite")
 │    └─ polled from /status
 └─ footer (glass): "Built with dhiarhome · Go · HTMX · Alpine.js" + GitHub/author links
```

### 4.2 Dashboard grid (`templates/status.html`)
- **Widget row** (top): todo | weather+time | system info | network — 2-col mobile (2×2), 4-col desktop. All `min-h-[190px]`.
- **Main grid**: Proxmox (CPU&Memory | Virtualization | Disk) · Bookmarks+Services (combined, col-span-2) | Docker · Media Management (col-span-3).
- Responsive breakpoints: mobile `grid-cols-2`, desktop `lg:grid-cols-4`.
- Desktop width: `max-w-6xl lg:max-w-none` (fills screen).

## 5. Components

### Cards
`glass-card` container, `rounded-2xl`, section header with icon (`w-7 h-7`) + `text-2xl` title + `mb-5`.

### Progress bars
- CPU/Memory/Swap/Disk bars; memory & swap side-by-side.
- Threshold colors: green <60%, yellow 60–80%, red >80%.
- Height 8px (`h-2`), `cubic-bezier` transitions; `aria-valuenow`.

### Status indicators
- Pulsing dot: green (running/online) / gray (stopped) / red (offline).
- "Live" indicator: glass pill (`px-3 py-1.5 rounded-full`) with `live-pulse` animation.
- Status badges always include text label (not color-only) for accessibility.

### Bookmarks
- Flattened grid (groups merged), 2–6 cols; icon box `w-10 h-10` with `w-5 h-5` SVG; internal scroll `max-h-[200px]` beyond ~10 links.

### To-Do
- Compact widget: scrollable list `max-h-[72px]` (~2 items), header + input fixed.
- Full-screen modal: `w-full h-full`, larger text/touch targets, `Added`/`Done` date metadata, close via X / backdrop / Escape.
- Inline edit: pencil icon → input (amber border); Enter saves, Escape/click-away saves or cancels; optimistic update.

### Toasts
Top-right; green (recovered) / red (down); name, type, old→new state, timestamp; auto-dismiss 4 s; Alpine `x-init` reads embedded JSON.

### Footer
Glass surface, responsive `max-w-lg sm:max-w-2xl`.

## 6. Widget Specs

| Widget | Content |
|---|---|
| Weather+Time | live clock top (client JS, 1 s), weather condition below divider |
| System Info | hostname, OS, uptime, Go runtime (goroutines, mem) |
| Network summary | per-interface status + live RX/TX |
| Custom Text | title + escaped HTML content (model widget; hidden from main row) |

Standalone fallbacks: weather / datetime render individually if only one enabled.

## 7. Background & Branding

- Background: local path (`/background`) or remote URL; overlay opacity + blur configurable; solid dark bg when unset.
- Logo: `appearance.logo` (path/URL) → favicon + header; falls back to inline SVG (data URI).
- Accent color: `appearance.accent_color` default `#3b82f6`.

## 8. Animation

- Card hover transitions (200 ms), HTMX swap fade-out 180 ms + fade-in 250 ms.
- `live-pulse` keyframe; loading skeleton shimmer; progress bar easing.
- `prefers-reduced-motion` → all animations disabled.

## 9. Accessibility (WCAG 2.1 AA)

- Contrast ≥4.5:1 (both themes).
- `aria-label` on all meters (CPU/RAM/Disk/swap), icon-only buttons.
- `aria-hidden` on decorative icons/bars; `role="status"` on live indicator + skeleton.
- Visible `focus-visible` rings; `tabindex="0"` on service/container items.
- Text labels on status (not color-only).
- Lighthouse Accessibility target >90.

## 10. File Map

| Concern | Files |
|---|---|
| Page shell, CSS vars, theme, merge-swap JS | `static/index.html` |
| Main status grid | `templates/status.html` |
| Widgets | `templates/widgets/widgets.html` |
| Bookmarks | `templates/bookmarks.html` |
| To-do (widget + modal) | `templates/todo.html` |
| Network card | `templates/network.html` |
| Media services | `templates/mediaservices.html` |
