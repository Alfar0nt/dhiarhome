package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dhiarhome/internal/cache"
	"dhiarhome/internal/config"
	"dhiarhome/internal/docker"
	"dhiarhome/internal/monitor"
	"dhiarhome/internal/network"
	"dhiarhome/internal/proxmox"
	"dhiarhome/internal/todo"
	"dhiarhome/internal/widgets"
)

type DashboardData struct {
	Proxmox      proxmox.NodeStatus
	CPUInfo      proxmox.CPUInfo
	Containers   []docker.Container
	Services     []cache.ServiceState
	Widgets      []widgets.WidgetData
	DateTime24h  bool
	DateTimezone string
	Network      []network.InterfaceStats
	NetShowSpeed bool
	NetShowTotal bool
	Todos        []todo.Todo
	TodoEnabled  bool
	TodoTitle    string
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
	localCPUInfo   proxmox.CPUInfo
)

func main() {
	var err error
	appConfig, err = config.LoadConfig("config.yaml")
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

	dkClient = docker.NewClient(appConfig.Docker.Socket)

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
	}).ParseFiles("templates/status.html", "templates/widgets/widgets.html", "templates/network.html", "templates/todo.html"))

	// Parse index.html as a template for dynamic appearance injection
	indexTmpl = template.Must(template.New("index.html").ParseFiles("static/index.html"))

	// Background poller for services
	go pollServices()

	// Serve static files but handle index.html as a template
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			indexHandler(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	})
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/api/background", backgroundHandler)
	http.HandleFunc("/background", backgroundServeHandler)

	// Todo API endpoints
	if appConfig.Todos.Enabled {
		http.HandleFunc("/api/todos", todoAPIHandler)
		http.HandleFunc("/api/todos/", todoItemHandler) // trailing slash for /api/todos/{id}
	}

	log.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
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

func doPoll() {
	for _, svc := range appConfig.Services {
		status, duration := monitor.CheckService(svc.URL)
		state := cache.ServiceState{
			Name:         svc.Name,
			Status:       status,
			ResponseTime: duration,
			Timestamp:    time.Now(),
		}
		historyCache.Add(state)
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

	data := struct {
		BackgroundSrc     string
		BackgroundOpacity float64
		BackgroundBlur    int
		CardOpacity       float64
		CardBlur          int
		AccentColor       string
		Theme             string
	}{
		BackgroundSrc:     bgSrc,
		BackgroundOpacity: appConfig.Appearance.BackgroundOpacity,
		BackgroundBlur:    appConfig.Appearance.BackgroundBlur,
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

	data, err := os.ReadFile(imgPath)
	if err != nil {
		log.Printf("Background image error: %v", err)
		http.NotFound(w, r)
		return
	}

	// Detect content type from extension
	ext := filepath.Ext(imgPath)
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

	// Fetch Docker containers
	containers, err := dkClient.GetContainers()
	if err != nil {
		log.Printf("Docker API Error: %v", err)
		// Fake containers for mock UI testing if docker socket fails
		if appConfig.Proxmox.Mock {
			containers = []docker.Container{
				{Names: []string{"/nginx"}, State: "running", Status: "Up 2 days"},
				{Names: []string{"/pihole"}, State: "exited", Status: "Exited (0) 5 hours ago"},
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

	data := DashboardData{
		Proxmox:      pxStatus,
		CPUInfo:      cpuInfo,
		Containers:   containers,
		Services:     latestServices,
		Widgets:      combineWidgets(widgetRegistry.FetchAll()),
		DateTime24h:  appConfig.Widgets.DateTime.Format24h,
		DateTimezone: appConfig.Widgets.DateTime.Timezone,
		Network:      netStats,
		NetShowSpeed: appConfig.Network.ShowSpeed,
		NetShowTotal: appConfig.Network.ShowTotal,
		Todos:        todos,
		TodoEnabled:  appConfig.Todos.Enabled,
		TodoTitle:    appConfig.Todos.Title,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

// combineWidgets merges weather + datetime into a single "weather_time" widget
// and reorders so custom_text comes first, then weather_time, then system_info.
func combineWidgets(raw []widgets.WidgetData) []widgets.WidgetData {
	var weather, datetime, sysinfo, custom *widgets.WidgetData
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
			custom = &raw[i]
		default:
			others = append(others, raw[i])
		}
	}

	var result []widgets.WidgetData

	// 1. Custom text (left side, compact)
	if custom != nil {
		result = append(result, *custom)
	}

	// 2. Combined weather + time card
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

	// 3. System info
	if sysinfo != nil {
		result = append(result, *sysinfo)
	}

	// 4. Any other widgets
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
		t := todoStore.Add(strings.TrimSpace(body.Text))
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
