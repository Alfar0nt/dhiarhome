package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"dhiarhome/internal/bookmarks"
	"dhiarhome/internal/cache"
	"dhiarhome/internal/config"
	"dhiarhome/internal/docker"
	"dhiarhome/internal/mediaservices"
	"dhiarhome/internal/monitor"
	"dhiarhome/internal/network"
	"dhiarhome/internal/notifications"
	"dhiarhome/internal/proxmox"
	"dhiarhome/internal/todo"
	"dhiarhome/internal/widgets"
)

// ── Security: simple per-IP rate limiter ──────────────────────────────────────

type rateLimiter struct {
	mu     sync.Mutex
	visits map[string][]time.Time
	limit  int
	window time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		visits: make(map[string][]time.Time),
		limit:  limit,
		window: window,
	}
}

func (rl *rateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Prune old entries
	var recent []time.Time
	for _, t := range rl.visits[ip] {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}

	if len(recent) >= rl.limit {
		rl.visits[ip] = recent
		return false
	}

	rl.visits[ip] = append(recent, now)
	return true
}

var apiLimiter = newRateLimiter(30, 1*time.Minute) // 30 requests/min per IP

type TransitionEvent struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	OldState  string `json:"old_state"`
	NewState  string `json:"new_state"`
	Timestamp string `json:"timestamp"`
}

type DashboardData struct {
	Proxmox        proxmox.NodeStatus
	CPUInfo        proxmox.CPUInfo
	VirtInfo       proxmox.VirtualizationInfo
	Containers     []docker.Container
	Services       []cache.ServiceState
	Widgets        []widgets.WidgetData
	DateTime24h    bool
	DateTimezone   string
	Network        []network.InterfaceStats
	NetShowSpeed   bool
	NetShowTotal   bool
	Todos          []todo.Todo
	TodoEnabled    bool
	TodoTitle      string
	MediaServices  []mediaservices.MediaServiceStats
	BookmarkGroups []bookmarks.DisplayGroup
	Transitions    []TransitionEvent
}

var (
	appConfig      *config.Config
	historyCache   *cache.HistoryCache
	pxClient       *proxmox.Client
	dkClient       *docker.Client
	tmpl           *template.Template
	indexTmpl      *template.Template
	widgetRegistry *widgets.Registry
	netMonitor     *network.Monitor
	todoStore      *todo.Store
	bookmarkStore  *bookmarks.Store
	localCPUInfo   proxmox.CPUInfo
	mediaStats     []mediaservices.MediaServiceStats
	mediaStatsMu   sync.RWMutex

	telegramNotifier    *notifications.Notifier
	prevServiceStates   map[string]string
	prevContainerStates map[string]string
	stateMu             sync.Mutex

	transitionBuffer []TransitionEvent
	transitionMu     sync.Mutex
)

// ── Security: middleware ───────────────────────────────────────────────────────

// securityHeaders adds standard security headers to every response.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.tailwindcss.com https://unpkg.com https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com")
		next.ServeHTTP(w, r)
	})
}

// clientIP extracts the client IP from the request (respects X-Forwarded-For).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Strip port from RemoteAddr
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}

// rateLimitMiddleware rejects requests that exceed the per-IP rate limit.
func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		if !apiLimiter.Allow(ip) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	var err error

	configPath := flag.String("config", "config.yaml", "Path to config file")
	listenAddr := flag.String("addr", ":8080", "Listen address (e.g. :8080, :9090)")
	flag.Parse()

	appConfig, err = config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	historyCache = cache.NewHistoryCache(100) // Keep last 100 service states

	pxClient = proxmox.NewClient(
		appConfig.Proxmox.URL,
		appConfig.Proxmox.NodeName,
		appConfig.Proxmox.TokenID,
		appConfig.Proxmox.TokenSecret,
		appConfig.Proxmox.Mock,
	)

	dkClient = docker.NewClientWithOptions(docker.Options{
		Endpoint:     appConfig.Docker.Socket,
		SkipTLS:      appConfig.Docker.SkipTLS,
		CACert:       appConfig.Docker.TLSCACert,
		Cert:         appConfig.Docker.TLSCert,
		Key:          appConfig.Docker.TLSKey,
		PortainerURL: appConfig.Docker.PortainerURL,
		PortainerKey: appConfig.Docker.PortainerKey,
		PortainerEnv: appConfig.Docker.PortainerEnvID,
	})

	// Initialize widget registry and register enabled widgets
	widgetRegistry = widgets.NewRegistry()
	if appConfig.Widgets.Weather.Enabled {
		widgetRegistry.Register(widgets.NewWeatherWidget(appConfig.Widgets.Weather))
	}
	if appConfig.Widgets.DateTime.Enabled {
		widgetRegistry.Register(widgets.NewDateTimeWidget(appConfig.Widgets.DateTime))
	}
	if appConfig.Widgets.SystemInfo.Enabled {
		widgetRegistry.Register(widgets.NewSystemInfoWidget(appConfig.Widgets.SystemInfo))
	}
	if appConfig.Widgets.CustomText.Enabled {
		widgetRegistry.Register(widgets.NewCustomTextWidget(appConfig.Widgets.CustomText))
	}

	// Initialize todo store if enabled
	if appConfig.Todos.Enabled {
		// Ensure data directory exists
		if dir := filepath.Dir(appConfig.Todos.FilePath); dir != "" {
			os.MkdirAll(dir, 0755)
		}
		todoStore = todo.NewStore(appConfig.Todos.FilePath)
		log.Println("Todo store initialized:", appConfig.Todos.FilePath)
	}

	// Read local CPU info once at startup
	localCPUInfo = proxmox.ReadLocalCPUInfo()

	// Initialize bookmarks store if configured
	if len(appConfig.Bookmarks) > 0 {
		bookmarkStore = bookmarks.NewStore(appConfig.Bookmarks, "data/icons")
		log.Printf("Bookmarks initialized: %d groups", len(appConfig.Bookmarks))
	}

	// Initialize Telegram notifier if enabled
	if appConfig.Notifications.Telegram.Enabled {
		telegramNotifier = notifications.NewNotifier(
			appConfig.Notifications.Telegram.BotToken,
			appConfig.Notifications.Telegram.ChatID,
			appConfig.Notifications.Telegram.MessageThreadID,
			appConfig.Notifications.Telegram.NotifyUp,
			appConfig.Notifications.Telegram.NotifyDown,
			appConfig.Notifications.Telegram.Cooldown,
			appConfig.Notifications.Telegram.SilentHours,
			appConfig.Notifications.Telegram.Mock,
		)
		prevServiceStates = make(map[string]string)
		prevContainerStates = make(map[string]string)
		log.Printf("Telegram notifier initialized (mock=%v, notify_up=%v, notify_down=%v)",
			appConfig.Notifications.Telegram.Mock,
			appConfig.Notifications.Telegram.NotifyUp,
			appConfig.Notifications.Telegram.NotifyDown)
	}

	// Initialize network monitor if enabled
	if appConfig.Network.Enabled && len(appConfig.Network.Interfaces) > 0 {
		ifaces := make(map[string]string)
		for _, iface := range appConfig.Network.Interfaces {
			ifaces[iface.Name] = iface.Label
		}
		netMonitor = network.NewMonitor(ifaces, appConfig.Network.UpdateInterval, appConfig.Network.Mock)
		netMonitor.Start()
		log.Printf("Network monitor started for %d interfaces", len(appConfig.Network.Interfaces))
	}

	tmpl = template.Must(template.New("status.html").Funcs(template.FuncMap{
		"percent": func(used, total int64) float64 {
			if total == 0 {
				return 0
			}
			return float64(used) / float64(total) * 100
		},
		"mult": func(a float64, b float64) float64 {
			return a * b
		},
		"gb": func(bytes int64) float64 {
			return float64(bytes) / (1024 * 1024 * 1024)
		},
		"roundDur": func(d time.Duration) string {
			ms := d.Seconds() * 1000
			if ms < 1000 {
				return fmt.Sprintf("%.0f ms", ms)
			}
			return fmt.Sprintf("%.2f s", d.Seconds())
		},
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	}).ParseFiles("templates/status.html", "templates/widgets/widgets.html", "templates/network.html", "templates/todo.html", "templates/mediaservices.html", "templates/bookmarks.html"))

	// Parse index.html as a template for dynamic appearance injection
	indexTmpl = template.Must(template.New("index.html").ParseFiles("static/index.html"))

	// Background poller for services
	go pollServices()

	// Background poller for media services
	if len(appConfig.MediaServices) > 0 {
		go pollMediaServices()
	}

	// Background poller for Docker container state tracking
	if appConfig.Notifications.Telegram.Enabled && dkClient != nil {
		go pollContainers()
	}

	// Serve static files but handle index.html as a template
	fs := http.FileServer(http.Dir("static"))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			indexHandler(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})
	mux.HandleFunc("/status", statusHandler)
	mux.HandleFunc("/api/background", backgroundHandler)
	mux.HandleFunc("/background", backgroundServeHandler)
	mux.HandleFunc("/logo", logoServeHandler)

	// Bookmarks favicon cache endpoint
	mux.Handle("/bookmarks/icons/", http.StripPrefix("/bookmarks/icons/", http.FileServer(http.Dir("data/icons"))))

	// Notification test endpoint
	if appConfig.Notifications.Telegram.Enabled {
		mux.Handle("/api/notifications/test", rateLimitMiddleware(http.HandlerFunc(notificationsTestHandler)))
	}

	// Todo API endpoints (rate-limited)
	if appConfig.Todos.Enabled {
		todoHandler := rateLimitMiddleware(http.HandlerFunc(todoAPIHandler))
		todoItemRoute := rateLimitMiddleware(http.HandlerFunc(todoItemHandler))
		mux.Handle("/api/todos", todoHandler)
		mux.Handle("/api/todos/", todoItemRoute) // trailing slash for /api/todos/{id}
	}

	// Wrap entire mux with security headers
	handler := securityHeaders(mux)

	log.Println("Server listening on", *listenAddr)

	// Start HTTP server in a goroutine for graceful shutdown
	srv := &http.Server{
		Addr:         *listenAddr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Stop background goroutines
	if netMonitor != nil {
		netMonitor.Stop()
		log.Println("Network monitor stopped")
	}

	// Graceful HTTP shutdown (wait up to 5s for in-flight requests)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}

func pollMediaServices() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	doPollMedia()

	for range ticker.C {
		doPollMedia()
	}
}

func doPollMedia() {
	var stats []mediaservices.MediaServiceStats
	for _, svc := range appConfig.MediaServices {
		ms := mediaservices.MediaService{
			Name:   svc.Name,
			URL:    svc.URL,
			APIKey: svc.APIKey,
			WebUI:  svc.WebUI,
		}
		stats = append(stats, mediaservices.FetchStats(ms))
	}
	mediaStatsMu.Lock()
	mediaStats = stats
	mediaStatsMu.Unlock()
}

func pollServices() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Initial poll
	doPoll()

	for range ticker.C {
		doPoll()
	}
}

// pollContainers periodically fetches Docker containers and detects state transitions.
func pollContainers() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Initial check
	checkContainerStates()

	for range ticker.C {
		checkContainerStates()
	}
}

func checkContainerStates() {
	if dkClient == nil || telegramNotifier == nil {
		return
	}

	containers, err := dkClient.GetContainers()
	if err != nil {
		log.Printf("Docker API Error (container monitoring): %v", err)
		return
	}

	stateMu.Lock()
	defer stateMu.Unlock()

	// Initialize prev states map on first run
	if prevContainerStates == nil {
		prevContainerStates = make(map[string]string)
	}

	for _, c := range containers {
		name := strings.TrimPrefix(c.Names[0], "/")
		prev, exists := prevContainerStates[name]

		// Only track running/exited states for meaningful transitions
		if c.State == "running" || c.State == "exited" {
			prevContainerStates[name] = c.State

			if exists && prev != c.State {
				go telegramNotifier.NotifyContainerChange(name, prev, c.State)
				recordTransition(name, "container", prev, c.State)
			}
		}
	}
}

func recordTransition(name, typ, oldState, newState string) {
	transitionMu.Lock()
	defer transitionMu.Unlock()
	t := TransitionEvent{
		Name:      name,
		Type:      typ,
		OldState:  oldState,
		NewState:  newState,
		Timestamp: time.Now().Format("15:04:05"),
	}
	transitionBuffer = append(transitionBuffer, t)
	if len(transitionBuffer) > 20 {
		transitionBuffer = transitionBuffer[len(transitionBuffer)-20:]
	}
}

func flushTransitions() []TransitionEvent {
	transitionMu.Lock()
	defer transitionMu.Unlock()
	events := transitionBuffer
	transitionBuffer = nil
	return events
}

func doPoll() {
	for _, svc := range appConfig.Services {
		status, duration := monitor.CheckService(svc.URL, svc.SkipTLS)
		state := cache.ServiceState{
			Name:         svc.Name,
			Status:       status,
			ResponseTime: duration,
			Timestamp:    time.Now(),
		}
		historyCache.Add(state)

		// Notify on state transition
		if telegramNotifier != nil {
			stateMu.Lock()
			prev, exists := prevServiceStates[svc.Name]
			prevServiceStates[svc.Name] = status
			stateMu.Unlock()

			if exists && prev != status {
				telegramNotifier.NotifyServiceChange(svc.Name, svc.URL, prev, status)
				recordTransition(svc.Name, "service", prev, status)
			}
		}
	}
}

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	// Determine background source
	bgSrc := ""
	if appConfig.Appearance.BackgroundURL != "" {
		// Remote URL: use directly in CSS
		bgSrc = appConfig.Appearance.BackgroundURL
	} else if appConfig.Appearance.BackgroundImage != "" {
		// Local file: serve via /background endpoint
		bgSrc = "/background"
	}

	// Determine logo source
	LogoSrc := ""
	logoHasFile := false
	if appConfig.Appearance.Logo != "" {
		if strings.HasPrefix(appConfig.Appearance.Logo, "http://") || strings.HasPrefix(appConfig.Appearance.Logo, "https://") {
			LogoSrc = appConfig.Appearance.Logo
		} else {
			LogoSrc = "/logo"
			logoHasFile = true
		}
	}

	data := struct {
		BackgroundSrc     string
		BackgroundOpacity float64
		BackgroundBlur    int
		LogoSrc           string
		LogoHasFile       bool
		CardOpacity       float64
		CardBlur          int
		AccentColor       string
		Theme             string
	}{
		BackgroundSrc:     bgSrc,
		BackgroundOpacity: appConfig.Appearance.BackgroundOpacity,
		BackgroundBlur:    appConfig.Appearance.BackgroundBlur,
		LogoSrc:           LogoSrc,
		LogoHasFile:       logoHasFile,
		CardOpacity:       appConfig.Appearance.CardOpacity,
		CardBlur:          appConfig.Appearance.CardBlur,
		AccentColor:       appConfig.Appearance.AccentColor,
		Theme:             appConfig.Appearance.Theme,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := indexTmpl.Execute(w, data); err != nil {
		log.Printf("Index template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func backgroundHandler(w http.ResponseWriter, _ *http.Request) {
	bgSrc := ""
	if appConfig.Appearance.BackgroundURL != "" {
		bgSrc = appConfig.Appearance.BackgroundURL
	} else if appConfig.Appearance.BackgroundImage != "" {
		bgSrc = "/background"
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"src":"%s","opacity":%.2f,"blur":%d}`, bgSrc, appConfig.Appearance.BackgroundOpacity, appConfig.Appearance.BackgroundBlur)
}

func backgroundServeHandler(w http.ResponseWriter, r *http.Request) {
	imgPath := appConfig.Appearance.BackgroundImage
	if imgPath == "" {
		http.NotFound(w, r)
		return
	}

	// Security: reject path traversal attempts
	clean := filepath.Clean(imgPath)
	if strings.Contains(clean, "..") || filepath.IsAbs(clean) && !strings.HasPrefix(clean, "/app/") {
		log.Printf("[SECURITY] Blocked path traversal attempt: %s", imgPath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		log.Printf("Background image error: %v", err)
		http.NotFound(w, r)
		return
	}

	// Detect content type from extension
	ext := filepath.Ext(clean)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "image/png"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(data)
}

func statusHandler(w http.ResponseWriter, _ *http.Request) {
	// Fetch Proxmox status
	pxStatus, err := pxClient.GetNodeStatus()
	if err != nil {
		log.Printf("Proxmox API Error: %v", err)
	}

	// Merge extra disks from config into Proxmox disks
	mergeExtraDisks(&pxStatus, appConfig.Proxmox.ExtraDisks)

	// Fetch Docker containers
	containers, err := dkClient.GetContainers()
	if err != nil {
		log.Printf("Docker API Error: %v", err)
		// Fake containers for mock UI testing if docker socket fails
		if appConfig.Proxmox.Mock {
			containers = []docker.Container{
				{Names: []string{"/nginx"}, State: "running", Status: "Up 2 days"},
				{Names: []string{"/pihole"}, State: "exited", Status: "Exited (0) 5 hours ago"},
				{Names: []string{"/portainer"}, State: "running", Status: "Up 14 days"},
				{Names: []string{"/plex"}, State: "running", Status: "Up 7 days"},
				{Names: []string{"/nextcloud"}, State: "running", Status: "Up 3 days"},
			}
		}
	} else {
		// Filter containers if specified in config
		if len(appConfig.Docker.MonitorContainers) > 0 {
			var filtered []docker.Container
			for _, c := range containers {
				for _, allowed := range appConfig.Docker.MonitorContainers {
					for _, name := range c.Names {
						if name == "/"+allowed || name == allowed {
							filtered = append(filtered, c)
						}
					}
				}
			}
			containers = filtered
		}
	}

	// Get latest services from cache
	var latestServices []cache.ServiceState
	for _, svc := range appConfig.Services {
		if state, found := historyCache.GetLatest(svc.Name); found {
			latestServices = append(latestServices, state)
		}
	}

	// Fetch network stats
	var netStats []network.InterfaceStats
	if netMonitor != nil {
		netStats = netMonitor.GetStats()
	}

	// Get CPU info (use local /proc/cpuinfo, or mock data from proxmox)
	cpuInfo := localCPUInfo
	if appConfig.Proxmox.Mock && cpuInfo.Cores == 0 {
		cpuInfo = pxStatus.CPUInfo
	}

	// Get todos
	var todos []todo.Todo
	if todoStore != nil {
		todos = todoStore.GetAll()
	}

	// Fetch virtualization info (VMs and LXCs)
	virtInfo, err := pxClient.GetVirtualization()
	if err != nil {
		log.Printf("Proxmox Virtualization Error: %v", err)
	}

	// Get media service stats
	var medias []mediaservices.MediaServiceStats
	mediaStatsMu.RLock()
	if len(mediaStats) > 0 {
		medias = make([]mediaservices.MediaServiceStats, len(mediaStats))
		copy(medias, mediaStats)
	} else if len(appConfig.MediaServices) > 0 {
		// Services configured but no data yet (first poll)
		medias = mediaservices.MockStats()
	} else if appConfig.Proxmox.Mock {
		// Mock mode: show demo data
		medias = mediaservices.MockStats()
	}
	mediaStatsMu.RUnlock()

	// Get bookmark groups
	var bookmarkGroups []bookmarks.DisplayGroup
	if bookmarkStore != nil {
		bookmarkGroups = bookmarkStore.GetGroups()
	}

	data := DashboardData{
		Proxmox:        pxStatus,
		CPUInfo:        cpuInfo,
		VirtInfo:       virtInfo,
		Containers:     containers,
		Services:       latestServices,
		Widgets:        combineWidgets(widgetRegistry.FetchAll()),
		DateTime24h:    appConfig.Widgets.DateTime.Format24h,
		DateTimezone:   appConfig.Widgets.DateTime.Timezone,
		Network:        netStats,
		NetShowSpeed:   appConfig.Network.ShowSpeed,
		NetShowTotal:   appConfig.Network.ShowTotal,
		Todos:          todos,
		TodoEnabled:    appConfig.Todos.Enabled,
		TodoTitle:      appConfig.Todos.Title,
		MediaServices:  medias,
		BookmarkGroups: bookmarkGroups,
		Transitions:    flushTransitions(),
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// combineWidgets merges weather + datetime into a single "weather_time" widget
// and reorders so weather_time comes first, then system_info.
func combineWidgets(raw []widgets.WidgetData) []widgets.WidgetData {
	var weather, datetime, sysinfo *widgets.WidgetData
	var others []widgets.WidgetData

	for i := range raw {
		switch raw[i].Type {
		case "weather":
			weather = &raw[i]
		case "datetime":
			datetime = &raw[i]
		case "system_info":
			sysinfo = &raw[i]
		case "custom_text":
			// custom_text replaced by todo widget in the template
		default:
			others = append(others, raw[i])
		}
	}

	var result []widgets.WidgetData

	// 1. Combined weather + time card
	if weather != nil && datetime != nil {
		combined := widgets.WidgetData{
			Type:  "weather_time",
			Label: "Weather & Time",
			Icon:  weather.Icon,
			Values: map[string]interface{}{
				"temperature": weather.Values["temperature"],
				"condition":   weather.Values["condition"],
				"wind_speed":  weather.Values["wind_speed"],
				"time":        datetime.Values["time"],
				"date":        datetime.Values["date"],
				"timezone":    datetime.Values["timezone"],
				"day":         datetime.Values["day"],
			},
		}
		result = append(result, combined)
	} else {
		// If only one is present, keep it standalone
		if weather != nil {
			result = append(result, *weather)
		}
		if datetime != nil {
			result = append(result, *datetime)
		}
	}

	// 2. System info
	if sysinfo != nil {
		result = append(result, *sysinfo)
	}

	// 3. Any other widgets
	result = append(result, others...)

	return result
}

// todoAPIHandler handles GET (list all) and POST (add new) todos.
func todoAPIHandler(w http.ResponseWriter, r *http.Request) {
	if todoStore == nil {
		http.Error(w, "Todos not enabled", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(todoStore.GetAll())

	case http.MethodPost:
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Text) == "" {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// Security: limit input length to prevent abuse
		text := strings.TrimSpace(body.Text)
		if len(text) > 500 {
			http.Error(w, "Text too long (max 500 characters)", http.StatusBadRequest)
			return
		}
		t := todoStore.Add(text)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// todoItemHandler handles PUT (toggle) and DELETE on /api/todos/{id}.
func todoItemHandler(w http.ResponseWriter, r *http.Request) {
	if todoStore == nil {
		http.Error(w, "Todos not enabled", http.StatusNotFound)
		return
	}

	// Extract ID from URL: /api/todos/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/todos/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Missing todo ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPut:
		if ok := todoStore.Toggle(id); !ok {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	case http.MethodDelete:
		if ok := todoStore.Delete(id); !ok {
			http.Error(w, "Todo not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func notificationsTestHandler(w http.ResponseWriter, r *http.Request) {
	if telegramNotifier == nil {
		http.Error(w, "Telegram notifier not initialized", http.StatusNotFound)
		return
	}
	if err := telegramNotifier.NotifyTest(); err != nil {
		log.Printf("Test notification error: %v", err)
		http.Error(w, fmt.Sprintf("Failed to send test notification: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status":"ok","message":"Test notification sent"}`)
}

func logoServeHandler(w http.ResponseWriter, r *http.Request) {
	logoPath := appConfig.Appearance.Logo
	if logoPath == "" {
		http.NotFound(w, r)
		return
	}

	clean := filepath.Clean(logoPath)
	if strings.Contains(clean, "..") || filepath.IsAbs(clean) && !strings.HasPrefix(clean, "/app/") {
		log.Printf("[SECURITY] Blocked path traversal attempt: %s", logoPath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	data, err := os.ReadFile(clean)
	if err != nil {
		log.Printf("Logo image error: %v", err)
		http.NotFound(w, r)
		return
	}

	ext := filepath.Ext(clean)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "image/png"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(data)
}

// mergeExtraDisks appends configured extra disks to the Proxmox disk list.
// Deduplicates by mountpoint. Supports auto-detect (statfs) and manual (static total/used) modes.
func mergeExtraDisks(status *proxmox.NodeStatus, extraDisks []config.ExtraDiskConfig) {
	// Build set of existing mountpoints for deduplication
	existing := make(map[string]bool)
	for _, d := range status.Disks {
		existing[d.Mountpoint] = true
	}

	for _, ed := range extraDisks {
		if ed.Mountpoint == "" {
			continue
		}
		// Skip if mountpoint already in the disk list
		if existing[ed.Mountpoint] {
			log.Printf("[INFO] extra_disk %s: mountpoint already present, skipping", ed.Mountpoint)
			continue
		}

		disk := proxmox.DiskInfo{
			Mountpoint: ed.Mountpoint,
		}

		// Determine disk size source
		hasStaticTotal := ed.Total != "" && ed.Used != ""
		if hasStaticTotal {
			// Manual override: use static values from config
			totalBytes, _ := config.ParseSize(ed.Total)
			usedBytes, _ := config.ParseSize(ed.Used)
			disk.Total = totalBytes
			disk.Used = usedBytes
		} else if ed.AutoDetect {
			// Auto-detect: read from filesystem via statfs
			totalBytes, usedBytes, err := proxmox.ReadDiskUsage(ed.Mountpoint)
			if err != nil {
				log.Printf("[WARN] extra_disk %s: statfs failed (%v), skipping", ed.Mountpoint, err)
				continue
			}
			disk.Total = totalBytes
			disk.Used = usedBytes
		} else {
			// No static values and auto-detect disabled: skip
			log.Printf("[WARN] extra_disk %s: no total/used provided and auto_detect is false, skipping", ed.Mountpoint)
			continue
		}

		status.Disks = append(status.Disks, disk)
		existing[ed.Mountpoint] = true
		log.Printf("[INFO] extra_disk added: %s (total: %d, used: %d)", ed.Mountpoint, disk.Total, disk.Used)
	}
}
