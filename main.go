package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
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
	"dhiarhome/internal/network"
	"dhiarhome/internal/proxmox"
	"dhiarhome/internal/todo"
	"dhiarhome/internal/widgets"
)

// ── Types ───────────────────────────────────────────────────────────────────

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
	Demo           bool // always true in demo branch
}

// ── Globals ─────────────────────────────────────────────────────────────────

var (
	appConfig      *config.Config
	historyCache   *cache.HistoryCache
	pxClient       *proxmox.Client
	tmpl           *template.Template
	indexTmpl      *template.Template
	widgetRegistry *widgets.Registry
	netMonitor     *network.Monitor
	todoStore      *todo.Store
	bookmarkStore  *bookmarks.Store

	transitionBuffer []TransitionEvent
	transitionMu     sync.Mutex

	serviceOverrides   map[string]string
	containerOverrides map[string]string
	overrideMu         sync.RWMutex
)

// ── Main ────────────────────────────────────────────────────────────────────

func main() {
	var err error

	configPath := flag.String("config", "config.yaml", "Path to config file")
	listenAddr := flag.String("addr", ":8080", "Listen address")
	flag.Parse()

	appConfig, err = config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	historyCache = cache.NewHistoryCache(100)
	pxClient = proxmox.NewClient()

	// Initialize override maps for demo toggle
	serviceOverrides = make(map[string]string)
	containerOverrides = make(map[string]string)

	// Initialize widget registry
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

	// Initialize todo store
	if appConfig.Todos.Enabled {
		if dir := filepath.Dir(appConfig.Todos.FilePath); dir != "" {
			os.MkdirAll(dir, 0755)
		}
		todoStore = todo.NewStore(appConfig.Todos.FilePath)
		log.Println("Todo store initialized:", appConfig.Todos.FilePath)
	}

	// Initialize bookmarks
	if len(appConfig.Bookmarks) > 0 {
		bookmarkStore = bookmarks.NewStore(appConfig.Bookmarks, "data/icons")
		log.Printf("Bookmarks initialized: %d groups", len(appConfig.Bookmarks))
	}

	// Initialize network monitor
	if appConfig.Network.Enabled && len(appConfig.Network.Interfaces) > 0 {
		ifaces := make(map[string]string)
		for _, iface := range appConfig.Network.Interfaces {
			ifaces[iface.Name] = iface.Label
		}
		netMonitor = network.NewMonitor(ifaces, appConfig.Network.UpdateInterval)
		netMonitor.Start()
		log.Printf("Network monitor started for %d interfaces (mock)", len(appConfig.Network.Interfaces))
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
		"trimPrefix": strings.TrimPrefix,
	}).ParseFiles("templates/status.html", "templates/widgets/widgets.html", "templates/network.html", "templates/todo.html", "templates/mediaservices.html", "templates/bookmarks.html"))

	indexTmpl = template.Must(template.New("index.html").ParseFiles("static/index.html"))

	// Seed initial service data
	go pollServices()

	// HTTP routes
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
	mux.Handle("/bookmarks/icons/", http.StripPrefix("/bookmarks/icons/", http.FileServer(http.Dir("data/icons"))))

	// Todo API
	if appConfig.Todos.Enabled {
		mux.HandleFunc("/api/todos", todoListHandler)
		mux.HandleFunc("/api/todos/", todoToggleHandler)
	}

	// Demo toggle endpoints
	mux.HandleFunc("/api/services/toggle", serviceToggleHandler)
	mux.HandleFunc("/api/containers/toggle", containerToggleHandler)

	log.Println("Demo server listening on", *listenAddr)

	// Wrap mux with security headers middleware
	handler := securityHeaders(mux)

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	if netMonitor != nil {
		netMonitor.Stop()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("HTTP server forced shutdown: %v", err)
	}
	log.Println("Server stopped")
}

// ── Polling ─────────────────────────────────────────────────────────────────

func pollServices() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	doPoll()
	for range ticker.C {
		doPoll()
	}
}

func doPoll() {
	for _, svc := range appConfig.Services {
		// Check for demo override
		overrideMu.RLock()
		overriddenStatus, hasOverride := serviceOverrides[svc.Name]
		overrideMu.RUnlock()

		var status string
		var duration time.Duration

		if hasOverride {
			status = overriddenStatus
			duration = 150 * time.Millisecond
		} else {
			// Simulate a healthy service with realistic response time
			status = "Online"
			duration = time.Duration(80+time.Now().UnixNano()%120) * time.Millisecond
		}

		state := cache.ServiceState{
			Name:         svc.Name,
			Status:       status,
			ResponseTime: duration,
			Timestamp:    time.Now(),
		}
		historyCache.Add(state)
	}
}

// ── Transition Events (for toast notifications) ─────────────────────────────

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

// ── Security Headers Middleware ──────────────────────────────────────────────

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")
		// Prevent MIME-type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// XSS filter (legacy browsers)
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Control referrer information
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Prevent search engine indexing
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
		// Content Security Policy
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.tailwindcss.com https://unpkg.com; "+
				"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; "+
				"font-src https://fonts.gstatic.com; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// ── Handlers ────────────────────────────────────────────────────────────────

func indexHandler(w http.ResponseWriter, _ *http.Request) {
	bgSrc := ""
	if appConfig.Appearance.BackgroundURL != "" {
		bgSrc = appConfig.Appearance.BackgroundURL
	} else if appConfig.Appearance.BackgroundImage != "" {
		bgSrc = appConfig.Appearance.BackgroundImage
	}

	logoSrc := ""
	if appConfig.Appearance.Logo != "" {
		if strings.HasPrefix(appConfig.Appearance.Logo, "http://") || strings.HasPrefix(appConfig.Appearance.Logo, "https://") {
			logoSrc = appConfig.Appearance.Logo
		} else {
			logoSrc = appConfig.Appearance.Logo
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
		LogoSrc:           logoSrc,
		LogoHasFile:       false,
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
		bgSrc = appConfig.Appearance.BackgroundImage
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"src":     bgSrc,
		"opacity": appConfig.Appearance.BackgroundOpacity,
		"blur":    appConfig.Appearance.BackgroundBlur,
	})
}

func statusHandler(w http.ResponseWriter, _ *http.Request) {
	// Always use mock data
	pxStatus, _ := pxClient.GetNodeStatus()

	// Get mock containers and apply overrides
	containers := docker.MockContainers()
	overrideMu.RLock()
	for i := range containers {
		cName := strings.TrimPrefix(containers[i].Names[0], "/")
		if overriddenState, ok := containerOverrides[cName]; ok {
			containers[i].State = overriddenState
			if overriddenState == "running" {
				containers[i].Status = "Up (demo)"
			} else {
				containers[i].Status = "Exited (demo)"
			}
		}
	}
	overrideMu.RUnlock()

	// Filter containers if specified
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

	// Get latest services from cache
	var latestServices []cache.ServiceState
	for _, svc := range appConfig.Services {
		if state, found := historyCache.GetLatest(svc.Name); found {
			latestServices = append(latestServices, state)
		}
	}

	// Network stats (mock)
	var netStats []network.InterfaceStats
	if netMonitor != nil {
		netStats = netMonitor.GetStats()
	}

	// Virtualization info (mock)
	virtInfo, _ := pxClient.GetVirtualization()

	// Todos
	var todos []todo.Todo
	if todoStore != nil {
		todos = todoStore.GetAll()
	}

	// Media services (always mock)
	medias := mediaservices.MockStats()

	// Bookmarks
	var bookmarkGroups []bookmarks.DisplayGroup
	if bookmarkStore != nil {
		bookmarkGroups = bookmarkStore.GetGroups()
	}

	data := DashboardData{
		Proxmox:        pxStatus,
		CPUInfo:        pxStatus.CPUInfo,
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
		Demo:           true,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// combineWidgets merges weather + datetime into a single "weather_time" widget.
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
			// replaced by todo widget in template
		default:
			others = append(others, raw[i])
		}
	}

	var result []widgets.WidgetData

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
		if weather != nil {
			result = append(result, *weather)
		}
		if datetime != nil {
			result = append(result, *datetime)
		}
	}

	if sysinfo != nil {
		result = append(result, *sysinfo)
	}

	result = append(result, others...)
	return result
}

// ── Todo Handlers ───────────────────────────────────────────────────────────

func todoListHandler(w http.ResponseWriter, r *http.Request) {
	if todoStore == nil {
		http.Error(w, "Todos not enabled", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todoStore.GetAll())
}

func todoToggleHandler(w http.ResponseWriter, r *http.Request) {
	if todoStore == nil {
		http.Error(w, "Todos not enabled", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	if ok := todoStore.Toggle(id); !ok {
		http.Error(w, "Todo not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// ── Demo Toggle Handlers ────────────────────────────────────────────────────

func serviceToggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing service name", http.StatusBadRequest)
		return
	}

	// Validate service exists
	var svcExists bool
	for _, svc := range appConfig.Services {
		if svc.Name == name {
			svcExists = true
			break
		}
	}
	if !svcExists {
		http.Error(w, "Service not found", http.StatusNotFound)
		return
	}

	// Get current state
	overrideMu.RLock()
	currentState, hasOverride := serviceOverrides[name]
	overrideMu.RUnlock()

	if !hasOverride {
		if cached, found := historyCache.GetLatest(name); found {
			currentState = cached.Status
		} else {
			currentState = "Online"
		}
	}

	// Toggle
	var newStatus string
	if currentState == "Online" {
		newStatus = "Offline"
	} else {
		newStatus = "Online"
	}

	overrideMu.Lock()
	serviceOverrides[name] = newStatus
	overrideMu.Unlock()

	// Update cache
	historyCache.Add(cache.ServiceState{
		Name:         name,
		Status:       newStatus,
		ResponseTime: 150 * time.Millisecond,
		Timestamp:    time.Now(),
	})

	recordTransition(name, "service", currentState, newStatus)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":        true,
		"name":      name,
		"type":      "service",
		"old_state": currentState,
		"new_state": newStatus,
	})
}

func containerToggleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing container name", http.StatusBadRequest)
		return
	}

	// Get current state from overrides or mock containers
	overrideMu.RLock()
	currentState, hasOverride := containerOverrides[name]
	overrideMu.RUnlock()

	if !hasOverride {
		for _, c := range docker.MockContainers() {
			cName := strings.TrimPrefix(c.Names[0], "/")
			if cName == name {
				currentState = c.State
				break
			}
		}
		if currentState == "" {
			http.Error(w, "Container not found", http.StatusNotFound)
			return
		}
	}

	// Toggle
	var newState string
	if currentState == "running" {
		newState = "exited"
	} else {
		newState = "running"
	}

	overrideMu.Lock()
	containerOverrides[name] = newState
	overrideMu.Unlock()

	recordTransition(name, "container", currentState, newState)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ok":        true,
		"name":      name,
		"type":      "container",
		"old_state": currentState,
		"new_state": newState,
	})
}
