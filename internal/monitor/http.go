package monitor

import (
	"crypto/tls"
	"net/http"
	"time"
)

// CheckService performs an HTTP GET request to check the service status.
// Set skipTLS to true to skip TLS certificate verification (e.g., for self-signed certs).
func CheckService(url string, skipTLS bool) (status string, responseTime time.Duration) {
	start := time.Now()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if skipTLS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
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
