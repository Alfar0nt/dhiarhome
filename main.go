package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"proxmox-dashboard/internal/cache"
	"proxmox-dashboard/internal/config"
	"proxmox-dashboard/internal/docker"
	"proxmox-dashboard/internal/monitor"
	"proxmox-dashboard/internal/proxmox"
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

	// Background poller for services
	go pollServices()

	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/status", statusHandler)

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

func statusHandler(w http.ResponseWriter, r *http.Request) {
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
