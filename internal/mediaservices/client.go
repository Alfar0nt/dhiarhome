package mediaservices

import "fmt"

type MediaServiceStats struct {
	Name           string
	WebUI          string
	Online         bool
	TotalCount     int
	WantedCount    int
	PendingCount   int
	AvailableCount int
}

// MockStats returns hardcoded demo data for media services.
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
