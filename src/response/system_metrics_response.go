package response

// SystemMetrics represents the collected system performance data.
type SystemMetrics struct {
	Timestamp string        `json:"timestamp,omitempty"`
	CPU       *float64      `json:"cpu,omitempty"`
	Memory    *MemoryStats  `json:"memory,omitempty"`
	Disk      *[]DiskStats  `json:"disk,omitempty"`
	Network   *NetworkStats `json:"network,omitempty"`
}

// MemoryStats holds memory usage information.
type MemoryStats struct {
	Used  int64 `json:"used"`  // MB
	Total int64 `json:"total"` // MB
}

// DiskStats holds disk partition statistics.
type DiskStats struct {
	Name       string  `json:"name"`
	Mount      string  `json:"mount"`
	Type       string  `json:"type"`
	Used       float64 `json:"used"`  // GB
	Total      float64 `json:"total"` // GB
	ReadBytes  uint64  `json:"readBytes"`
	WriteBytes uint64  `json:"writeBytes"`
	ReadBps    float64 `json:"readBps"`
	WriteBps   float64 `json:"writeBps"`
}

// NetworkStats holds network traffic statistics.
type NetworkStats struct {
	Up   int64 `json:"up"`   // Mbps
	Down int64 `json:"down"` // Mbps
}
