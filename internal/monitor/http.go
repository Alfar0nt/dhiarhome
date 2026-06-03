package monitor

import (
	"net/http"
	"time"
)

// CheckService performs an HTTP GET request to check the service status.
func CheckService(url string) (status string, responseTime time.Duration) {
	start := time.Now()
	
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	duration := time.Since(start)

	if err != nil {
		return "Offline", duration
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "Online", duration
	}

	return "Warning", duration
}
