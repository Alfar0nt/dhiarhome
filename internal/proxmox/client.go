package proxmox

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type NodeStatus struct {
	CPU    float64 `json:"cpu"`
	Memory struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
		Free  int64 `json:"free"`
	} `json:"memory"`
	Disk struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
	} `json:"rootfs"`
	Uptime  int64   `json:"uptime"`
	CPUInfo CPUInfo `json:"-"` // not from API, read from /proc/cpuinfo or mock
}

type CPUInfo struct {
	ModelName string
	Cores     int // physical cores
	Threads   int // logical CPUs (siblings)
}

type Client struct {
	url         string
	nodeName    string
	tokenID     string
	tokenSecret string
	mock        bool
	httpClient  *http.Client
}

func NewClient(url, nodeName, tokenID, tokenSecret string, mock bool) *Client {
	// Proxmox often uses self-signed certificates
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	return &Client{
		url:         url,
		nodeName:    nodeName,
		tokenID:     tokenID,
		tokenSecret: tokenSecret,
		mock:        mock,
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   5 * time.Second,
		},
	}
}

func (c *Client) GetNodeStatus() (NodeStatus, error) {
	if c.mock {
		return getMockStatus(), nil
	}

	endpoint := fmt.Sprintf("%s/nodes/%s/status", c.url, url.PathEscape(c.nodeName))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return NodeStatus{}, err
	}

	// Set Proxmox API Token header
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return NodeStatus{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return NodeStatus{}, fmt.Errorf("proxmox API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data NodeStatus `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return NodeStatus{}, err
	}

	return result.Data, nil
}

func getMockStatus() NodeStatus {
	// Generate some random varying stats for testing the UI
	var status NodeStatus
	status.CPU = 0.15 + rand.Float64()*(0.65-0.15) // Random CPU between 15% and 65%

	status.Memory.Total = 16 * 1024 * 1024 * 1024                                         // 16GB
	status.Memory.Used = int64(float64(status.Memory.Total) * (0.4 + rand.Float64()*0.2)) // 40-60% used
	status.Memory.Free = status.Memory.Total - status.Memory.Used

	status.Disk.Total = 256 * 1024 * 1024 * 1024 // 256GB
	status.Disk.Used = 120 * 1024 * 1024 * 1024  // ~120GB used

	status.Uptime = 3600 * 24 * 7 // 7 days

	status.CPUInfo = CPUInfo{
		ModelName: "Mock CPU (Intel Core i7-12700K)",
		Cores:     12,
		Threads:   20,
	}

	return status
}

// ReadLocalCPUInfo parses /proc/cpuinfo to get physical cores and logical threads.
func ReadLocalCPUInfo() CPUInfo {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return CPUInfo{ModelName: "Unknown CPU", Cores: 0, Threads: 0}
	}
	defer f.Close()

	var info CPUInfo
	var threads int
	physicalIDs := make(map[string]bool)
	coreIDs := make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "model name":
			if info.ModelName == "" {
				info.ModelName = val
			}
		case "processor":
			threads++
		case "physical id":
			physicalIDs[val] = true
		case "core id":
			coreIDs[val] = true
		case "cpu cores":
			if info.Cores == 0 {
				if n, err := strconv.Atoi(val); err == nil {
					// cpu cores = cores per socket, multiply by sockets later
					info.Cores = n * len(physicalIDs)
					if info.Cores == 0 {
						// physicalIDs not populated yet, store per-socket count
						info.Cores = n
					}
				}
			}
		}
	}

	info.Threads = threads

	// If we couldn't determine cores from "cpu cores" field, use unique physical_id + core_id combos
	if info.Cores == 0 {
		info.Cores = len(coreIDs)
		if info.Cores == 0 {
			info.Cores = threads // fallback: assume no hyperthreading distinction
		}
	}

	// Re-calculate cores properly after we know all physical IDs
	if len(physicalIDs) > 0 {
		// Re-read to get per-socket core count
		f2, err := os.Open("/proc/cpuinfo")
		if err == nil {
			defer f2.Close()
			scanner2 := bufio.NewScanner(f2)
			var coresPerSocket int
			for scanner2.Scan() {
				line := scanner2.Text()
				parts := strings.SplitN(line, ":", 2)
				if len(parts) != 2 {
					continue
				}
				if strings.TrimSpace(parts[0]) == "cpu cores" {
					if n, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
						coresPerSocket = n
						break
					}
				}
			}
			if coresPerSocket > 0 {
				info.Cores = coresPerSocket * len(physicalIDs)
			}
		}
	}

	if info.ModelName == "" {
		info.ModelName = "Unknown CPU"
	}

	return info
}
