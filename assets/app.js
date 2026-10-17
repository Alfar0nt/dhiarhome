// ============================================================
// dhiarhome static DEMO — client-side mock data engine.
// No backend, no fetch/XHR. Everything is generated in-browser.
// Timers pause while the tab is hidden to save resources.
// ============================================================

const ICONS = {
  server: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>',
  globe: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>',
  monitor: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>',
  'play-circle': '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>',
  tv: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>',
  container: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>',
  database: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2.21 3.582 4 8 4s8-1.79 8-4V7M4 7c0 2.21 3.582 4 8 4s8-1.79 8-4M4 7c0-2.21 3.582-4 8-4s8 1.79 8 4m0 5c0 2.21-3.582 4-8 4s-8-1.79-8-4"/>',
  home: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>',
  settings: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>',
  film: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 4v16M17 4v16M3 8h4m10 0h4M3 12h18M3 16h4m10 0h4M4 20h16a1 1 0 001-1V5a1 1 0 00-1-1H4a1 1 0 00-1 1v14a1 1 0 001 1z"/>',
  link: '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>'
};

function rnd(min, max) { return Math.random() * (max - min) + min; }
function irnd(min, max) { return Math.floor(rnd(min, max + 1)); }
function clamp(v, lo, hi) { return Math.max(lo, Math.min(hi, v)); }
function pick(arr) { return arr[Math.floor(Math.random() * arr.length)]; }
function nowTime() { return new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }); }

function demoApp() {
  return {
    theme: localStorage.getItem('dhiarhome-theme') || 'dark',
    clock: '',
    clockDate: '',
    toasts: [],
    activity: [],
    _seq: 0,
    _timers: [],
    _running: false,

    // ── State ───────────────────────────────────────────
    proxmox: {
      cpu: 23, cpuModel: 'Intel Core i7-12700K', cores: 12, threads: 20,
      memTotal: 32 * 1024 ** 3, memUsed: 14 * 1024 ** 3,
      swapTotal: 4 * 1024 ** 3, swapUsed: 0.6 * 1024 ** 3,
      load: [0.52, 0.41, 0.33],
      pveVersion: 'pve-manager/8.2.2/935536e9',
      kernelVersion: '6.8.12-1-pve',
      disks: [
        { mountpoint: '/', total: 120 * 1024 ** 3, used: 58 * 1024 ** 3 },
        { mountpoint: '/mnt/storage', total: 2 * 1024 ** 4, used: 1.1 * 1024 ** 4 },
        { mountpoint: '/mnt/backup', total: 4 * 1024 ** 4, used: 2.3 * 1024 ** 4 }
      ]
    },
    virt: {
      vms: [
        { id: 100, name: 'windows-11', status: 'running' },
        { id: 101, name: 'ubuntu-dev', status: 'running' },
        { id: 102, name: 'macos-sonoma', status: 'stopped' }
      ],
      lxcs: [
        { id: 200, name: 'docker-host', status: 'running' },
        { id: 201, name: 'pihole', status: 'running' },
        { id: 202, name: 'nginx-proxy', status: 'running' },
        { id: 203, name: 'nextcloud', status: 'running' },
        { id: 204, name: 'postgres', status: 'running' },
        { id: 205, name: 'redis', status: 'running' },
        { id: 206, name: 'backup-agent', status: 'running' }
      ]
    },
    docker: [
      { name: 'nginx', state: 'running', status: 'Up 3 days' },
      { name: 'pihole', state: 'running', status: 'Up 12 days' },
      { name: 'portainer', state: 'running', status: 'Up 12 days' },
      { name: 'plex', state: 'running', status: 'Up 5 days' },
      { name: 'nextcloud', state: 'running', status: 'Up 2 days' }
    ],
    services: [
      { name: 'Personal Website', status: 'Online', rt: 142 },
      { name: 'Nextcloud', status: 'Online', rt: 88 },
      { name: 'PDF Tools', status: 'Online', rt: 210 },
      { name: 'Uptime Kuma', status: 'Online', rt: 35 },
      { name: 'Home Assistant', status: 'Online', rt: 64 },
      { name: 'Vaultwarden', status: 'Online', rt: 51 }
    ],
    media: [
      { name: 'Sonarr', webui: 'http://sonarr.lan:8989', online: true, stat1: 148, stat1Label: 'Series', stat2: 7, stat2Label: 'Wanted', stat2Class: 'text-amber-400' },
      { name: 'Radarr', webui: 'http://radarr.lan:7878', online: true, stat1: 812, stat1Label: 'Movies', stat2: 3, stat2Label: 'Wanted', stat2Class: 'text-amber-400' },
      { name: 'Overseerr', webui: 'http://overseerr.lan:5055', online: true, stat1: 1642, stat1Label: 'Available', stat2: 4, stat2Label: 'Pending', stat2Class: 'text-amber-400' }
    ],
    network: [
      { name: 'eth0', label: 'Primary', status: 'up', rx: 1.2e6, tx: 3.4e5, rxTotal: 542 * 1024 ** 3, txTotal: 128 * 1024 ** 3, rxSpeed: '', txSpeed: '' },
      { name: 'vmbr1', label: 'Internal', status: 'up', rx: 4.1e5, tx: 2.2e5, rxTotal: 210 * 1024 ** 3, txTotal: 96 * 1024 ** 3, rxSpeed: '', txSpeed: '' }
    ],
    weather: { temp: 27, condition: 'Partly Cloudy', icon: '⛅', wind: '12 km/h' },
    sysinfo: { hostname: 'dhiarhome-pve', os: 'Proxmox VE 8.2.2', uptime: '14d 6h 32m', goMem: 18.4 },
    bookmarks: [
      { name: 'Proxmox VE', url: '#', icon: ICONS.server },
      { name: 'Portainer', url: '#', icon: ICONS.container },
      { name: 'Grafana', url: '#', icon: ICONS.monitor },
      { name: 'Uptime Kuma', url: '#', icon: ICONS.link },
      { name: 'Pi-hole', url: '#', icon: ICONS.globe },
      { name: 'Plex', url: '#', icon: ICONS['play-circle'] },
      { name: 'Sonarr', url: '#', icon: ICONS.tv },
      { name: 'Radarr', url: '#', icon: ICONS.film },
      { name: 'Home Assistant', url: '#', icon: ICONS.home },
      { name: 'Vaultwarden', url: '#', icon: ICONS.settings }
    ],
    todos: [
      { id: 1, text: 'Renew Proxmox backup schedule', done: false, created_at: new Date(Date.now() - 864e5).toISOString(), done_at: '' },
      { id: 2, text: 'Update Docker containers', done: true, created_at: new Date(Date.now() - 2 * 864e5).toISOString(), done_at: new Date(Date.now() - 36e5).toISOString() },
      { id: 3, text: 'Add Plex to alerts', done: false, created_at: new Date().toISOString(), done_at: '' }
    ],

    newText: '', showModal: false, editingId: null, editText: '', _todoId: 100,

    // ── Derived ─────────────────────────────────────────
    get doneCount() { return this.todos.filter(t => t.done).length; },
    get vmRunning() { return this.virt.vms.filter(v => v.status === 'running').length; },
    get lxcRunning() { return this.virt.lxcs.filter(v => v.status === 'running').length; },
    get swapColor() {
      const p = this.pct(this.proxmox.swapUsed, this.proxmox.swapTotal);
      return p >= 80 ? 'bg-red-500' : p >= 60 ? 'bg-yellow-500' : 'bg-amber-400';
    },

    // ── Helpers ─────────────────────────────────────────
    pct(used, total) { return total > 0 ? (used / total) * 100 : 0; },
    gb(bytes) { return bytes / 1024 ** 3; },
    roundDur(ms) { return ms < 1000 ? Math.round(ms) + ' ms' : (ms / 1000).toFixed(2) + ' s'; },
    formatSpeed(bps) {
      const bits = bps * 8;
      if (bits < 1e3) return bits.toFixed(0) + ' b/s';
      if (bits < 1e6) return (bits / 1e3).toFixed(1) + ' Kbit/s';
      if (bits < 1e9) return (bits / 1e6).toFixed(2) + ' Mbit/s';
      return (bits / 1e9).toFixed(2) + ' Gbit/s';
    },
    formatDate(dateStr) {
      if (!dateStr) return '';
      const d = new Date(dateStr); if (isNaN(d.getTime())) return '';
      const now = new Date();
      const time = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
      if (d.toDateString() === now.toDateString()) return 'Today ' + time;
      const y = new Date(now); y.setDate(y.getDate() - 1);
      if (d.toDateString() === y.toDateString()) return 'Yesterday ' + time;
      return d.toLocaleDateString([], { month: 'short', day: 'numeric' }) + ' ' + time;
    },

    // ── Lifecycle ───────────────────────────────────────
    start() {
      this.applyTheme(this.theme);
      document.addEventListener('visibilitychange', () => {
        if (document.hidden) { this._stopTimers(); }
        else { this.tickClock(); this.tickFast(); this.tickMain(); this._startTimers(); }
      });
      this.tickClock();
      this.tickFast();
      this.tickMain();
      this._startTimers();
      this.logEvent('🟢', 'Demo session started — data is simulated', true);
    },

    _startTimers() {
      if (this._running) return;
      this._running = true;
      this._timers.push(
        setInterval(() => this.tickClock(), 1000),
        setInterval(() => this.tickFast(), 2000),
        setInterval(() => this.tickMain(), 4000),
        setInterval(() => this.triggerEvent(), 15000)
      );
    },
    _stopTimers() {
      this._timers.forEach(clearInterval);
      this._timers = [];
      this._running = false;
    },

    tickClock() {
      const n = new Date();
      this.clock = n.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      this.clockDate = n.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' });
    },

    tickFast() {
      this.network.forEach(n => {
        n.rx = clamp(n.rx * rnd(0.7, 1.35) + rnd(-8e4, 3e5), 2e4, 4e7);
        n.tx = clamp(n.tx * rnd(0.7, 1.35) + rnd(-4e4, 1.5e5), 1e4, 2e7);
        n.rxTotal += n.rx * 2; n.txTotal += n.tx * 2;
        n.rxSpeed = this.formatSpeed(n.rx);
        n.txSpeed = this.formatSpeed(n.tx);
      });
      this.proxmox.cpu = clamp(this.proxmox.cpu + rnd(-1.5, 1.8), 4, 92);
      this.proxmox.load = this.proxmox.load.map((v, i) => clamp(v + rnd(-0.06, 0.06), 0.05, 4 + i));
    },

    tickMain() {
      this.proxmox.memUsed = clamp(this.proxmox.memUsed + rnd(-3e8, 3.2e8), 6 * 1024 ** 3, 30 * 1024 ** 3);
      this.proxmox.swapUsed = clamp(this.proxmox.swapUsed + rnd(-4e7, 5e7), 0, 3.6 * 1024 ** 3);
      this.proxmox.disks.forEach(d => { d.used = clamp(d.used + rnd(-6e8, 9e8), d.total * 0.1, d.total * 0.95); });
      this.services.forEach(s => { if (s.status === 'Online') s.rt = clamp(s.rt * rnd(0.85, 1.2), 18, 480); });
      this.media.forEach(m => {
        if (Math.random() < 0.25) m.stat1 += irnd(1, 3);
        if (Math.random() < 0.2) m.stat2 = irnd(0, 9);
      });
      if (Math.random() < 0.15) {
        const w = pick([
          { condition: 'Sunny', icon: '☀️' }, { condition: 'Partly Cloudy', icon: '⛅' },
          { condition: 'Cloudy', icon: '☁️' }, { condition: 'Light Rain', icon: '🌧️' },
          { condition: 'Clear', icon: '🌙' }, { condition: 'Windy', icon: '💨' }
        ]);
        this.weather.condition = w.condition; this.weather.icon = w.icon;
        this.weather.temp = clamp(this.weather.temp + irnd(-1, 1), 18, 34);
        this.weather.wind = irnd(4, 28) + ' km/h';
      }
      this.sysinfo.goMem = Math.round((this.sysinfo.goMem + rnd(-0.6, 0.6)) * 10) / 10;
    },

    triggerEvent() {
      const roll = Math.random();
      if (roll < 0.4) this.flipService();
      else if (roll < 0.7) this.flipContainer();
      else this.flipVirt();
    },

    flipService() {
      const s = pick(this.services);
      const wasOnline = s.status === 'Online';
      s.status = wasOnline ? 'Offline' : 'Online';
      s.rt = wasOnline ? 0 : irnd(30, 260);
      this.pushTransition(s.name, 'service', wasOnline ? 'Online' : 'Offline', s.status);
    },

    flipContainer() {
      const c = pick(this.docker);
      const wasUp = c.state === 'running';
      c.state = wasUp ? 'exited' : 'running';
      c.status = wasUp ? 'Exited (0) ' + nowTime() : 'Up ' + irnd(0, 2) + ' hours';
      this.pushTransition(c.name, 'container', wasUp ? 'running' : 'exited', c.state);
    },

    flipVirt() {
      const pool = Math.random() < 0.5 ? this.virt.vms : this.virt.lxcs;
      const r = pick(pool);
      const wasRunning = r.status === 'running';
      r.status = wasRunning ? 'stopped' : 'running';
      this.pushTransition(r.name, wasRunning ? 'VM' : 'LXC', wasRunning ? 'running' : 'stopped', r.status);
    },

    pushTransition(name, type, oldState, newState) {
      const ok = newState === 'Online' || newState === 'running';
      const t = { id: ++this._seq, name, type, old_state: oldState, new_state: newState, ok, time: nowTime() };
      this.toasts = this.toasts.concat([t]);
      setTimeout(() => { this.toasts = this.toasts.filter(x => x.id !== t.id); }, 4000);
      this.logEvent(ok ? '🟢' : '🔴', `${name} (${type}): ${oldState} → ${newState}`, ok);
    },

    logEvent(icon, text, ok) {
      this.activity = [{ id: ++this._seq, icon, text, ok, time: nowTime() }, ...this.activity].slice(0, 40);
    },

    demoLinkEvent(name) {
      this.logEvent('🔗', `Opened ${name} (demo — links are non-functional)`, true);
    },

    // ── Theme ───────────────────────────────────────────
    toggleTheme() {
      this.theme = this.theme === 'light' ? 'dark' : 'light';
      this.applyTheme(this.theme);
    },
    applyTheme(t) {
      document.documentElement.classList.toggle('light-theme', t === 'light');
      localStorage.setItem('dhiarhome-theme', t);
    },

    // ── Todo (in-memory, demo only) ─────────────────────
    sortedTodos() {
      return this.todos.slice().sort((a, b) => {
        if (a.done !== b.done) return a.done ? 1 : -1;
        return a.id - b.id;
      });
    },
    addTodo() {
      const text = this.newText.trim();
      if (!text) return;
      this.todos = this.todos.concat([{ id: ++this._todoId, text, done: false, created_at: new Date().toISOString(), done_at: '' }]);
      this.newText = '';
      this.logEvent('📝', `Todo added: ${text}`, true);
    },
    editTodo(id) {
      this.editingId = id;
      const t = this.todos.find(x => x.id === id);
      if (t) this.editText = t.text;
      setTimeout(() => { const el = document.getElementById('edit-input-' + id); if (el) el.focus(); }, 60);
    },
    saveEdit(id) {
      const text = this.editText.trim();
      if (!text) { this.cancelEdit(); return; }
      this.todos = this.todos.map(t => t.id === id ? { ...t, text } : t);
      this.cancelEdit();
    },
    cancelEdit() { this.editingId = null; this.editText = ''; },
    toggleTodo(id) {
      let nowDone = false;
      this.todos = this.todos.map(t => {
        if (t.id !== id) return t;
        nowDone = !t.done;
        return { ...t, done: nowDone, done_at: nowDone ? new Date().toISOString() : '' };
      });
      const t = this.todos.find(x => x.id === id);
      if (t) this.logEvent('✅', `Todo ${nowDone ? 'completed' : 'reopened'}: ${t.text}`, true);
    },
    deleteTodo(id) { this.todos = this.todos.filter(t => t.id !== id); }
  };
}
