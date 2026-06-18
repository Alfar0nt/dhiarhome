package mediaservices

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type MediaService struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
	WebUI  string `yaml:"webui"`
}

type MediaServiceStats struct {
	Name          string
	WebUI         string
	Online        bool
	TotalCount    int
	WantedCount   int
	PendingCount  int
	AvailableCount int
}

func FetchStats(svc MediaService) MediaServiceStats {
	stats := MediaServiceStats{
		Name:  svc.Name,
		WebUI: svc.WebUI,
	}

	switch svc.Name {
	case "Sonarr":
		fetchSonarr(svc, &stats)
	case "Radarr":
		fetchRadarr(svc, &stats)
	case "Overseerr":
		fetchOverseerr(svc, &stats)
	default:
		stats.Online = false
	}

	return stats
}

func doGet(url, apiKey string, dest interface{}) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return false
	}
	return true
}

func doGetOverseerr(url, apiKey string, dest interface{}) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return false
	}
	return true
}

// ── Sonarr ──────────────────────────────────────────────

func fetchSonarr(svc MediaService, stats *MediaServiceStats) {
	// Total series count
	var series []interface{}
	if !doGet(svc.URL+"/api/v3/series", svc.APIKey, &series) {
		return
	}
	stats.Online = true
	stats.TotalCount = len(series)

	// Wanted/missing count
	var missing struct {
		TotalRecords int `json:"totalRecords"`
	}
	if doGet(svc.URL+"/api/v3/wanted/missing?pageSize=1", svc.APIKey, &missing) {
		stats.WantedCount = missing.TotalRecords
	}
}

// ── Radarr ──────────────────────────────────────────────

func fetchRadarr(svc MediaService, stats *MediaServiceStats) {
	var movies []interface{}
	if !doGet(svc.URL+"/api/v3/movie", svc.APIKey, &movies) {
		return
	}
	stats.Online = true
	stats.TotalCount = len(movies)

	var missing struct {
		TotalRecords int `json:"totalRecords"`
	}
	if doGet(svc.URL+"/api/v3/wanted/missing?pageSize=1", svc.APIKey, &missing) {
		stats.WantedCount = missing.TotalRecords
	}
}

// ── Overseerr ───────────────────────────────────────────

type overseerrPageInfo struct {
	Total int `json:"results"`
}

type overseerrMediaCounts struct {
	Movies int `json:"movies"`
	Series int `json:"series"`
	Total  int `json:"total"`
}

func fetchOverseerr(svc MediaService, stats *MediaServiceStats) {
	// Pending requests
	type requestResponse struct {
		PageInfo overseerrPageInfo `json:"pageInfo"`
	}
	var pendingReq requestResponse
	if !doGetOverseerr(svc.URL+"/api/v1/request?take=1&filter=pending", svc.APIKey, &pendingReq) {
		return
	}
	stats.Online = true
	stats.PendingCount = pendingReq.PageInfo.Total

	// Total requests
	var allReq requestResponse
	if doGetOverseerr(svc.URL+"/api/v1/request?take=1", svc.APIKey, &allReq) {
		stats.TotalCount = allReq.PageInfo.Total
	}

	// Available media count
	type mediaResponse struct {
		PageInfo overseerrPageInfo `json:"pageInfo"`
	}
	var media mediaResponse
	if doGetOverseerr(svc.URL+"/api/v1/media?take=1", svc.APIKey, &media) {
		stats.AvailableCount = media.PageInfo.Total
	}

	// Also try to get movie/series breakdown
	var counts overseerrMediaCounts
	if doGetOverseerr(svc.URL+"/api/v1/media/count", svc.APIKey, &counts) {
		stats.AvailableCount = counts.Total
	}
}

// MockStats returns mock data for testing.
func MockStats() []MediaServiceStats {
	return []MediaServiceStats{
		{Name: "Sonarr", WebUI: "https://sonarr.example.com", Online: true, TotalCount: 45, WantedCount: 3},
		{Name: "Radarr", WebUI: "https://radarr.example.com", Online: true, TotalCount: 120, WantedCount: 8},
		{Name: "Overseerr", WebUI: "https://overseerr.example.com", Online: true, TotalCount: 95, PendingCount: 5, AvailableCount: 1520},
	}
}

func StatsSummary(s MediaServiceStats) string {
	switch s.Name {
	case "Sonarr":
		return fmt.Sprintf("%d series · %d wanted", s.TotalCount, s.WantedCount)
	case "Radarr":
		return fmt.Sprintf("%d movies · %d wanted", s.TotalCount, s.WantedCount)
	case "Overseerr":
		return fmt.Sprintf("%d available · %d pending", s.AvailableCount, s.PendingCount)
	default:
		if s.Online {
			return "Connected"
		}
		return "Offline"
	}
}
