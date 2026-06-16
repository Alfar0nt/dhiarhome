package main

import (
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"dhiarhome/internal/cache"
	"dhiarhome/internal/config"
	"dhiarhome/internal/docker"
	"dhiarhome/internal/monitor"
	"dhiarhome/internal/proxmox"
)

type DashboardData struct {
	Proxmox    proxmox.NodeStatus
	Containers []docker.Container
	Services   []cache.ServiceState
}

var (
	appConfig    *config.Config
	historyCache *cache.HistoryCache
	pxClient     *proxmox.Client
	dkClient     *docker.Client
	tmpl         *template.Template
	indexTmpl    *template.Template
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
	}).ParseFiles("templates/status.html"))

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

	data := DashboardData{
		Proxmox:    pxStatus,
		Containers: containers,
		Services:   latestServices,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}
