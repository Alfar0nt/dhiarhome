package network

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Monitor tracks network interface statistics.
type Monitor struct {
	mu           sync.RWMutex
	interfaces   map[string]string      // name → label
	samples      map[string][]rawSample // name → last N samples
	stats        map[string]*InterfaceStats
	interval     time.Duration
	maxSamples   int
	stopCh       chan struct{}
	mock         bool
	displayMu    sync.Mutex
	cachedStats  []InterfaceStats
	statsCacheAt time.Time
	displayTTL   time.Duration // how long to cache display stats
}

// NewMonitor creates a new network monitor.
// interfaces maps interface name → display label.
func NewMonitor(interfaces map[string]string, intervalSec int, mock bool) *Monitor {
	if intervalSec < 1 {
		intervalSec = 3
	}
	return &Monitor{
		interfaces: interfaces,
		samples:    make(map[string][]rawSample),
		stats:      make(map[string]*InterfaceStats),
		interval:   time.Duration(intervalSec) * time.Second,
		maxSamples: 3,
		stopCh:     make(chan struct{}),
		mock:       mock,
		displayTTL: 10 * time.Second, // cache display output for 10s to reduce HTMX swap flicker
	}
}

// Start begins the background sampling goroutine.
func (m *Monitor) Start() {
	go m.sampleLoop()
}

// Stop terminates the background sampling.
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// GetStats returns a snapshot of all monitored interface stats.
// Results are cached for displayTTL (10s) to prevent rapid HTML changes during HTMX swaps.
func (m *Monitor) GetStats() []InterfaceStats {
	// Check display cache first
	m.displayMu.Lock()
	if m.cachedStats != nil && time.Since(m.statsCacheAt) < m.displayTTL {
		cached := m.cachedStats
		m.displayMu.Unlock()
		return cached
	}
	m.displayMu.Unlock()

	m.mu.RLock()
	result := make([]InterfaceStats, 0, len(m.interfaces))
	for name, label := range m.interfaces {
		if s, ok := m.stats[name]; ok {
			result = append(result, *s)
		} else {
			result = append(result, InterfaceStats{
				Name:   name,
				Label:  label,
				Status: "down",
			})
		}
	}
	m.mu.RUnlock()

	// Update display cache
	m.displayMu.Lock()
	m.cachedStats = result
	m.statsCacheAt = time.Now()
	m.displayMu.Unlock()

	return result
}

func (m *Monitor) sampleLoop() {
	// Initial sample
	m.takeSample()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.takeSample()
		case <-m.stopCh:
			return
		}
	}
}

func (m *Monitor) takeSample() {
	if m.mock {
		m.takeMockSample()
		return
	}

	raw, err := readProcNetDev()
	if err != nil {
		log.Printf("Network monitor: %v", err)
		return
	}

	now := time.Now().UnixNano()

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, label := range m.interfaces {
		counts, exists := raw[name]
		if !exists {
			// Interface not found
			m.stats[name] = &InterfaceStats{
				Name:   name,
				Label:  label,
				Status: "down",
			}
			continue
		}

		sample := rawSample{
			RxBytes: counts.RxBytes,
			TxBytes: counts.TxBytes,
			Time:    now,
		}

		// Append sample, keep last maxSamples
		m.samples[name] = append(m.samples[name], sample)
		if len(m.samples[name]) > m.maxSamples {
			m.samples[name] = m.samples[name][1:]
		}

		// Calculate rate using moving average
		rxRate, txRate := m.calcRate(name)

		m.stats[name] = &InterfaceStats{
			Name:    name,
			Label:   label,
			Status:  "up",
			RxBytes: counts.RxBytes,
			TxBytes: counts.TxBytes,
			RxRate:  rxRate,
			TxRate:  txRate,
			RxTotal: formatBytes(counts.RxBytes),
			TxTotal: formatBytes(counts.TxBytes),
			RxSpeed: formatSpeed(rxRate),
			TxSpeed: formatSpeed(txRate),
		}
	}
}

func (m *Monitor) calcRate(name string) (rxRate, txRate float64) {
	samples := m.samples[name]
	if len(samples) < 2 {
		return 0, 0
	}

	// Moving average over available samples (up to maxSamples)
	var rxSum, txSum float64
	count := 0
	for i := 1; i < len(samples); i++ {
		elapsed := float64(samples[i].Time-samples[i-1].Time) / 1e9 // seconds
		if elapsed <= 0 {
			continue
		}
		rxDelta := float64(samples[i].RxBytes - samples[i-1].RxBytes)
		txDelta := float64(samples[i].TxBytes - samples[i-1].TxBytes)
		rxSum += rxDelta / elapsed
		txSum += txDelta / elapsed
		count++
	}
	if count == 0 {
		return 0, 0
	}
	return rxSum / float64(count), txSum / float64(count)
}

func (m *Monitor) takeMockSample() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().UnixNano()

	for name, label := range m.interfaces {
		prev := m.samples[name]

		var rxBasis, txBasis uint64
		if len(prev) > 0 {
			rxBasis = prev[len(prev)-1].RxBytes
			txBasis = prev[len(prev)-1].TxBytes
		}

		// Simulate some traffic
		rxIncrement := uint64(500000 + time.Now().UnixNano()%2000000) // 0.5-2.5 MB
		txIncrement := uint64(100000 + time.Now().UnixNano()%500000)  // 0.1-0.6 MB

		sample := rawSample{
			RxBytes: rxBasis + rxIncrement,
			TxBytes: txBasis + txIncrement,
			Time:    now,
		}

		m.samples[name] = append(m.samples[name], sample)
		if len(m.samples[name]) > m.maxSamples {
			m.samples[name] = m.samples[name][1:]
		}

		rxRate, txRate := m.calcRate(name)

		m.stats[name] = &InterfaceStats{
			Name:    name,
			Label:   label,
			Status:  "up",
			RxBytes: sample.RxBytes,
			TxBytes: sample.TxBytes,
			RxRate:  rxRate,
			TxRate:  txRate,
			RxTotal: formatBytes(sample.RxBytes),
			TxTotal: formatBytes(sample.TxBytes),
			RxSpeed: formatSpeed(rxRate),
			TxSpeed: formatSpeed(txRate),
		}
	}
}

// readProcNetDev parses /proc/net/dev and returns interface byte counts.
func readProcNetDev() (map[string]rawSample, error) {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return nil, fmt.Errorf("open /proc/net/dev: %w", err)
	}
	defer f.Close()

	result := make(map[string]rawSample)
	scanner := bufio.NewScanner(f)

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		// Skip header lines
		if lineNum <= 2 {
			continue
		}

		line := scanner.Text()
		// Format: "  iface: rx_bytes rx_packets ... tx_bytes tx_packets ..."
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		fields := strings.Fields(parts[1])
		if len(fields) < 16 {
			continue
		}

		// fields[0] = rx_bytes, fields[8] = tx_bytes
		rx, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		tx, err := strconv.ParseUint(fields[8], 10, 64)
		if err != nil {
			continue
		}

		result[name] = rawSample{RxBytes: rx, TxBytes: tx}
	}

	return result, scanner.Err()
}

// formatBytes converts bytes to human-readable string.
func formatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(TB))
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// formatSpeed converts bytes/sec to human-readable speed string.
func formatSpeed(bytesPerSec float64) string {
	const (
		Kbit = 1000
		Mbit = Kbit * 1000
		Gbit = Mbit * 1000
	)
	bitsPerSec := bytesPerSec * 8
	switch {
	case bitsPerSec >= Gbit:
		return fmt.Sprintf("%.2f Gbit/s", bitsPerSec/Gbit)
	case bitsPerSec >= Mbit:
		return fmt.Sprintf("%.2f Mbit/s", bitsPerSec/Mbit)
	case bitsPerSec >= Kbit:
		return fmt.Sprintf("%.1f Kbit/s", bitsPerSec/Kbit)
	default:
		return fmt.Sprintf("%.0f b/s", bitsPerSec)
	}
}
