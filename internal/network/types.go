package network

// InterfaceStats holds the current stats for a single network interface.
type InterfaceStats struct {
	Name    string  `json:"name"`
	Label   string  `json:"label"`
	Status  string  `json:"status"` // "up" or "down"
	RxBytes uint64  `json:"rx_bytes"`
	TxBytes uint64  `json:"tx_bytes"`
	RxRate  float64 `json:"rx_rate"`  // bytes/sec
	TxRate  float64 `json:"tx_rate"`  // bytes/sec
	RxTotal string  `json:"rx_total"` // human-readable total received
	TxTotal string  `json:"tx_total"` // human-readable total transmitted
	RxSpeed string  `json:"rx_speed"` // human-readable speed
	TxSpeed string  `json:"tx_speed"` // human-readable speed
}

// rawSample holds a raw reading from /proc/net/dev.
type rawSample struct {
	RxBytes uint64
	TxBytes uint64
	Time    int64 // unix nanoseconds
}
