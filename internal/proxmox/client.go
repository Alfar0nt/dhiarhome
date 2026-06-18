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

type DiskInfo struct {
	Mountpoint string
	Total      int64
	Used       int64
}

type NodeStatus struct {
	CPU    float64 `json:"cpu"`
	Memory struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
		Free  int64 `json:"free"`
	} `json:"memory"`
	RootFS struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
	} `json:"rootfs"`
	Disks   []DiskInfo `json:"-"`
	Uptime  int64      `json:"uptime"`
	CPUInfo CPUInfo    `json:"-"`
}

type CPUInfo struct {
	ModelName string
	Cores     int // physical cores
	Threads   int // logical CPUs (siblings)
}

type VirtualizationInfo struct {
	VMRunning  int
	VMTotal    int
	LXCRunning int
	LXCTotal   int
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

	ns := result.Data
	// Populate Disks from RootFS
	ns.Disks = append(ns.Disks, DiskInfo{
		Mountpoint: "/",
		Total:      ns.RootFS.Total,
		Used:       ns.RootFS.Used,
	})

	// Try to fetch additional disks from disk list endpoint
	ns.fetchDiskList(c)

	return ns, nil
}

func (ns *NodeStatus) fetchDiskList(c *Client) {
	endpoint := fmt.Sprintf("%s/nodes/%s/disks/list", c.url, url.PathEscape(c.nodeName))
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var diskResult struct {
		Data []struct {
			Devpath    string `json:"devpath"`
			Used       int64  `json:"used,string"`
			Size       int64  `json:"size,string"`
			Mountpoint string `json:"mountpoint"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&diskResult); err != nil {
		return
	}

	for _, d := range diskResult.Data {
		if d.Mountpoint == "" || d.Mountpoint == "/" {
			continue // root already added from RootFS
		}
		ns.Disks = append(ns.Disks, DiskInfo{
			Mountpoint: d.Mountpoint,
			Total:      d.Size,
			Used:       d.Used,
		})
	}
}

func (c *Client) GetVirtualization() (VirtualizationInfo, error) {
	if c.mock {
		return getMockVirtualization(), nil
	}

	var info VirtualizationInfo

	// Fetch QEMU VMs
	vmEndpoint := fmt.Sprintf("%s/nodes/%s/qemu", c.url, url.PathEscape(c.nodeName))
	vms, err := c.fetchResourceList(vmEndpoint)
	if err == nil {
		info.VMTotal = len(vms)
		for _, vm := range vms {
			if vm.Status == "running" {
				info.VMRunning++
			}
		}
	}

	// Fetch LXC containers
	lxcEndpoint := fmt.Sprintf("%s/nodes/%s/lxc", c.url, url.PathEscape(c.nodeName))
	lxcs, err := c.fetchResourceList(lxcEndpoint)
	if err == nil {
		info.LXCTotal = len(lxcs)
		for _, lxc := range lxcs {
			if lxc.Status == "running" {
				info.LXCRunning++
			}
		}
	}

	return info, nil
}

type proxmoxResource struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"` // "running", "stopped"
}

func (c *Client) fetchResourceList(endpoint string) ([]proxmoxResource, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.tokenID, c.tokenSecret))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("proxmox API returned status: %d", resp.StatusCode)
	}

	var result struct {
		Data []proxmoxResource `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Data, nil
}

func getMockVirtualization() VirtualizationInfo {
	return VirtualizationInfo{
		VMRunning:  2,
		VMTotal:    3,
		LXCRunning: 5,
		LXCTotal:   7,
	}
}

func getMockStatus() NodeStatus {
	// Generate some random varying stats for testing the UI
	var status NodeStatus
	status.CPU = 0.15 + rand.Float64()*(0.65-0.15) // Random CPU between 15% and 65%

	status.Memory.Total = 16 * 1024 * 1024 * 1024                                         // 16GB
	status.Memory.Used = int64(float64(status.Memory.Total) * (0.4 + rand.Float64()*0.2)) // 40-60% used
	status.Memory.Free = status.Memory.Total - status.Memory.Used

	status.RootFS.Total = 256 * 1024 * 1024 * 1024 // 256GB
	status.RootFS.Used = 120 * 1024 * 1024 * 1024  // ~120GB used
	status.Disks = []DiskInfo{
		{Mountpoint: "/", Total: 256 * 1024 * 1024 * 1024, Used: 120 * 1024 * 1024 * 1024},
		{Mountpoint: "/mnt/storage", Total: 2 * 1024 * 1024 * 1024 * 1024, Used: 800 * 1024 * 1024 * 1024},
		{Mountpoint: "/mnt/backup", Total: 4 * 1024 * 1024 * 1024 * 1024, Used: 1500 * 1024 * 1024 * 1024},
	}

	status.Uptime = 3600 * 24 * 7 // 7 days

	status.CPUInfo = CPUInfo{
		ModelName: "Intel Core i7-12700K",
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
				info.ModelName = cleanCPUName(val)
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

func cleanCPUName(name string) string {
	name = strings.TrimSpace(name)
	// Remove vendor branding marks
	name = strings.ReplaceAll(name, "(TM)", "")
	name = strings.ReplaceAll(name, "(R)", "")
	name = strings.ReplaceAll(name, "  ", " ")
	// Strip verbose suffixes
	suffixes := []string{
		" with Radeon Graphics",
		" with Iris Xe Graphics",
		" with UHD Graphics",
		" with HD Graphics",
	}
	for _, s := range suffixes {
		if idx := strings.Index(name, s); idx >= 0 {
			name = name[:idx]
		}
	}
	// Strip " CPU @ ..." pattern
	if idx := strings.Index(name, " CPU @"); idx >= 0 {
		name = name[:idx]
	}
	// Strip "-Core Processor" pattern
	if idx := strings.Index(name, "-Core Processor"); idx >= 0 {
		name = name[:idx]
	}
	// Strip trailing " Processor"
	if idx := strings.Index(name, " Processor"); idx >= 0 {
		name = name[:idx]
	}
	return strings.TrimSpace(name)
}
