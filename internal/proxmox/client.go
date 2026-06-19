package proxmox

import (
	"math/rand"
)

// ── Data Types ──────────────────────────────────────────────────────────────

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
	Swap struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
		Free  int64 `json:"free"`
	} `json:"swap"`
	RootFS struct {
		Total int64 `json:"total"`
		Used  int64 `json:"used"`
	} `json:"rootfs"`
	Disks         []DiskInfo `json:"-"`
	Uptime        int64      `json:"uptime"`
	CPUInfo       CPUInfo    `json:"-"`
	LoadAvg       [3]float64 `json:"-"`
	PVEVersion    string     `json:"-"`
	KernelVersion string     `json:"-"`
}

type CPUInfo struct {
	ModelName string
	Cores     int // physical cores
	Threads   int // logical CPUs (siblings)
}

type ResourceInfo struct {
	VMID   int    `json:"vmid"`
	Name   string `json:"name"`
	Status string `json:"status"` // "running", "stopped"
}

type VirtualizationInfo struct {
	VMRunning  int
	VMTotal    int
	LXCRunning int
	LXCTotal   int
	VMs        []ResourceInfo
	LXCs       []ResourceInfo
}

// ── Client (mock-only) ─────────────────────────────────────────────────────

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) GetNodeStatus() (NodeStatus, error) {
	return getMockStatus(), nil
}

func (c *Client) GetVirtualization() (VirtualizationInfo, error) {
	return getMockVirtualization(), nil
}

// ── Mock Data ───────────────────────────────────────────────────────────────

func getMockVirtualization() VirtualizationInfo {
	return VirtualizationInfo{
		VMRunning:  2,
		VMTotal:    3,
		LXCRunning: 5,
		LXCTotal:   7,
		VMs: []ResourceInfo{
			{VMID: 100, Name: "pfsense", Status: "running"},
			{VMID: 101, Name: "windows11", Status: "running"},
			{VMID: 102, Name: "ubuntu-dev", Status: "stopped"},
		},
		LXCs: []ResourceInfo{
			{VMID: 200, Name: "nginx-proxy", Status: "running"},
			{VMID: 201, Name: "pihole", Status: "running"},
			{VMID: 202, Name: "grafana", Status: "running"},
			{VMID: 203, Name: "mariadb", Status: "running"},
			{VMID: 204, Name: "redis", Status: "running"},
			{VMID: 205, Name: "vaultwarden", Status: "stopped"},
			{VMID: 206, Name: "homeassistant", Status: "stopped"},
		},
	}
}

func getMockStatus() NodeStatus {
	var status NodeStatus
	status.CPU = 0.15 + rand.Float64()*(0.65-0.15) // 15-65%

	status.Memory.Total = 16 * 1024 * 1024 * 1024
	status.Memory.Used = int64(float64(status.Memory.Total) * (0.4 + rand.Float64()*0.2))
	status.Memory.Free = status.Memory.Total - status.Memory.Used

	status.RootFS.Total = 256 * 1024 * 1024 * 1024
	status.RootFS.Used = 120 * 1024 * 1024 * 1024
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

	status.Swap.Total = 4 * 1024 * 1024 * 1024
	status.Swap.Used = int64(float64(status.Swap.Total) * (0.1 + rand.Float64()*0.15))
	status.Swap.Free = status.Swap.Total - status.Swap.Used

	status.LoadAvg = [3]float64{
		0.5 + rand.Float64()*2.0,
		0.8 + rand.Float64()*1.5,
		1.0 + rand.Float64()*1.0,
	}

	status.PVEVersion = "pve-manager/8.2.2/935536e9"
	status.KernelVersion = "6.8.12-1-pve"

	return status
}
