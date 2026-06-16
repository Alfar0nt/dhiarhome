package widgets

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"dhiarhome/internal/config"
)

// SystemInfoWidget displays host system information.
type SystemInfoWidget struct {
	cfg config.SystemInfoWidgetConfig
}

// NewSystemInfoWidget creates a new system info widget from config.
func NewSystemInfoWidget(cfg config.SystemInfoWidgetConfig) *SystemInfoWidget {
	return &SystemInfoWidget{cfg: cfg}
}

func (s *SystemInfoWidget) Name() string { return "System Info" }
func (s *SystemInfoWidget) Type() string { return "system_info" }

func (s *SystemInfoWidget) Fetch() (*WidgetData, error) {
	hostname, _ := os.Hostname()
	osName := readOSRelease()
	uptime := readUptime()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &WidgetData{
		Type:  "system_info",
		Label: "System",
		Icon:  "🖥️",
		Values: map[string]interface{}{
			"hostname":   hostname,
			"os":         osName,
			"uptime":     uptime,
			"goroutines": runtime.NumGoroutine(),
			"go_mem_mb":  fmt.Sprintf("%.1f MB", float64(memStats.Alloc)/(1024*1024)),
		},
	}, nil
}

// readOSRelease parses /etc/os-release and returns the PRETTY_NAME.
func readOSRelease() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Unknown OS"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			name := strings.TrimPrefix(line, "PRETTY_NAME=")
			name = strings.Trim(name, "\"")
			return name
		}
	}
	return "Linux"
}

// readUptime reads /proc/uptime and returns a human-readable string.
func readUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "Unknown"
	}
	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return "Unknown"
	}
	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return "Unknown"
	}

	dur := time.Duration(seconds) * time.Second
	days := int(dur.Hours()) / 24
	hours := int(dur.Hours()) % 24
	minutes := int(dur.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
