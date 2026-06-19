package network

import (
	"fmt"
	"sync"
	"time"
)

// Monitor tracks mock network interface statistics (demo only).
type Monitor struct {
	mu           sync.RWMutex
	interfaces   map[string]string
	samples      map[string][]rawSample
	stats        map[string]*InterfaceStats
	interval     time.Duration
	maxSamples   int
	stopCh       chan struct{}
	displayMu    sync.Mutex
	cachedStats  []InterfaceStats
	statsCacheAt time.Time
	displayTTL   time.Duration
}

func NewMonitor(interfaces map[string]string, intervalSec int) *Monitor {
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
		displayTTL: 10 * time.Second,
	}
}

func (m *Monitor) Start() {
	go m.sampleLoop()
}

func (m *Monitor) Stop() {
	close(m.stopCh)
}

func (m *Monitor) GetStats() []InterfaceStats {
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
			result = append(result, InterfaceStats{Name: name, Label: label, Status: "down"})
		}
	}
	m.mu.RUnlock()

	m.displayMu.Lock()
	m.cachedStats = result
	m.statsCacheAt = time.Now()
	m.displayMu.Unlock()

	return result
}

func (m *Monitor) sampleLoop() {
	m.takeMockSample()
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.takeMockSample()
		case <-m.stopCh:
			return
		}
	}
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

		rxIncrement := uint64(500000 + time.Now().UnixNano()%2000000)
		txIncrement := uint64(100000 + time.Now().UnixNano()%500000)

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

func (m *Monitor) calcRate(name string) (rxRate, txRate float64) {
	samples := m.samples[name]
	if len(samples) < 2 {
		return 0, 0
	}
	var rxSum, txSum float64
	count := 0
	for i := 1; i < len(samples); i++ {
		elapsed := float64(samples[i].Time-samples[i-1].Time) / 1e9
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
