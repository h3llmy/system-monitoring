package response

// SystemMetrics represents the collected system performance data.
type SystemMetrics struct {
	Timestamp string       `json:"timestamp"`
	CPU       float64      `json:"cpu"`
	Memory    MemoryStats  `json:"memory"`
	Disk      []DiskStats  `json:"disk"`
	Network   NetworkStats `json:"network"`
}

// MemoryStats holds memory usage information.
type MemoryStats struct {
	Used  int64 `json:"used"`
	Total int64 `json:"total"`
}

// DiskStats holds disk partition statistics.
type DiskStats struct {
	Name  string  `json:"name"`
	Mount string  `json:"mount"`
	Type  string  `json:"type"`
	Used  float64 `json:"used"`
	Total float64 `json:"total"`
}

// NetworkStats holds network traffic statistics.
type NetworkStats struct {
	Up   int64 `json:"up"`   // Upload speed in Mbps
	Down int64 `json:"down"` // Download speed in Mbps
}