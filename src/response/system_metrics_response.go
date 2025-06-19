package response

type (
	// SystemMetrics represents the collected system performance data.
	SystemMetrics struct {
		Timestamp string       `json:"timestamp,omitempty"`
		Disk      *[]DiskStats `json:"disk,omitempty"`
		Matrics   *[]Matrics   `json:"matrics,omitempty"`
	}

	Matrics struct {
		CPU     *float64      `json:"cpu,omitempty"`
		Memory  *MemoryStats  `json:"memory,omitempty"`
		Network *NetworkStats `json:"network,omitempty"`
	}

	// MemoryStats holds memory usage information.
	MemoryStats struct {
		Used  int64 `json:"used"`  // byte
		Total int64 `json:"total"` // byte
	}

	// DiskStats holds disk partition statistics.
	DiskStats struct {
		Name        string  `json:"name"`
		Mount       string  `json:"mount"`
		Type        string  `json:"type"`
		UsedPercent float64 `json:"usedPercent"`
		Used        uint64  `json:"used"`  // byte
		Total       uint64  `json:"total"` // byte
		ReadBytes   uint64  `json:"readBytes"`
		WriteBytes  uint64  `json:"writeBytes"`
		ReadBps     float64 `json:"readBps"`
		WriteBps    float64 `json:"writeBps"`
	}

	// NetworkStats holds network traffic statistics.
	NetworkStats struct {
		Up   int64 `json:"up"`   // Mbps
		Down int64 `json:"down"` // Mbps
	}
)
