package service

import (
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	"github.com/h3llmy/system-monitoring/src/response"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gopsutil_net "github.com/shirou/gopsutil/v4/net"
)

type MonitoringService interface {
	CollectMetrics()
	GetHistory() ([]byte, error)
	GetCpuHistory() ([]byte, error)
	GetMemoryHistory() ([]byte, error)
	GetDiskHistory() ([]byte, error)
	GetNetworkHistory() ([]byte, error)
}

type SystemMonitor struct{}

// NewSystemMonitorService returns a new SystemMonitor instance that implements the MonitoringService interface.
//
// The SystemMonitor type has a CollectMetrics method that periodically collects system metrics such as CPU usage, memory usage, disk usage, and network speed
// and stores them in the history slice. It also has a GetHistory method that retrieves the collected history as a JSON byte array.
func NewSystemMonitorService() MonitoringService {
	return &SystemMonitor{}
}

var (
	history          []response.SystemMetrics
	mu               sync.Mutex
	prevNetStats     = make(map[string]gopsutil_net.IOCountersStat)
	prevDiskCounters = make(map[string]disk.IOCountersStat)
	prevTime         time.Time
)

// init initializes the previous network and disk counters to the current values.
// This ensures that the first iteration of CollectMetrics will not result in
// zero values for network and disk speed.
func init() {
	prevTime = time.Now()
	if stats, err := gopsutil_net.IOCounters(true); err == nil {
		for _, st := range stats {
			prevNetStats[st.Name] = st
		}
	}
	if counters, err := disk.IOCounters(); err == nil {
		for name, cnt := range counters {
			prevDiskCounters[name] = cnt
		}
	}
}

// CollectMetrics periodically collects system metrics such as CPU usage, memory usage, disk usage, and network speed
// and stores them in the history slice. It runs in an infinite loop and sleeps for 1 second between each iteration.
// The loop is driven by a for loop, which keeps running until the program exits. The function collects metrics by
// calling the respective functions, and then appends the new metrics to the history slice. If the history slice has
// more than 60 items, the first item is discarded to keep the slice size constant. The function then waits for 1 second
// before collecting metrics again. The metrics are collected in a separate goroutine.
func (sm *SystemMonitor) CollectMetrics() {
	for {
		now := time.Now()
		elapsed := now.Sub(prevTime).Seconds()

		cpuPct := getCpuMetrics()
		memStats := getMemoryMetrics()
		diskStats := getDiskMetrics(elapsed)
		netStats := getNetworkMetrics(elapsed)

		metrics := response.SystemMetrics{
			Timestamp: now.Format(time.RFC3339),
			CPU:       &cpuPct,
			Memory:    &memStats,
			Disk:      &diskStats,
			Network:   &netStats,
		}

		mu.Lock()
		history = append(history, metrics)
		if len(history) > 60 {
			history = history[1:]
		}
		mu.Unlock()

		prevTime = now
		time.Sleep(time.Until(now.Add(1 * time.Second)))
	}
}

// GetHistory returns the collected system metrics history as a JSON byte array.
// It locks the history slice while marshaling to prevent concurrent modification.
// If an error occurs while marshaling the history, it is returned instead.
func (sm *SystemMonitor) GetHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return json.Marshal(history)
}

// GetCpuHistory returns the collected CPU usage metrics history as a JSON byte array.
// It locks the history slice while marshaling to prevent concurrent modification.
// If an error occurs while marshaling the history, it is returned instead.
func (sm *SystemMonitor) GetCpuHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	var cpuMetrics []response.SystemMetrics
	for _, metrics := range history {
		cpuMetrics = append(cpuMetrics, response.SystemMetrics{
			CPU:       metrics.CPU,
			Timestamp: metrics.Timestamp,
		})
	}
	return json.Marshal(cpuMetrics)
}

// GetMemoryHistory returns the collected memory usage metrics history as a JSON byte array.
// It locks the history slice while marshaling to prevent concurrent modification.
// If an error occurs while marshaling the history, it is returned instead.
func (sm *SystemMonitor) GetMemoryHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	var memoryMetrics []response.SystemMetrics
	for _, metrics := range history {
		memoryMetrics = append(memoryMetrics, response.SystemMetrics{
			Memory:    metrics.Memory,
			Timestamp: metrics.Timestamp,
		})
	}
	return json.Marshal(memoryMetrics)
}

// GetDiskHistory returns the collected disk usage metrics history as a JSON byte array.
// It locks the history slice while marshaling to prevent concurrent modification.
// If an error occurs while marshaling the history, it is returned instead.
func (sm *SystemMonitor) GetDiskHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	var diskMetrics []response.SystemMetrics
	for _, metrics := range history {
		diskMetrics = append(diskMetrics, response.SystemMetrics{
			Disk:      metrics.Disk,
			Timestamp: metrics.Timestamp,
		})
	}
	return json.Marshal(diskMetrics)
}

// GetNetworkHistory returns the collected network usage metrics history as a JSON byte array.
// It locks the history slice while marshaling to prevent concurrent modification.
// If an error occurs while marshaling the history, it is returned instead.
func (sm *SystemMonitor) GetNetworkHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	var networkMetrics []response.SystemMetrics
	for _, metrics := range history {
		networkMetrics = append(networkMetrics, response.SystemMetrics{
			Network:   metrics.Network,
			Timestamp: metrics.Timestamp,
		})
	}
	return json.Marshal(networkMetrics)
}

// getCpuMetrics retrieves the current CPU usage statistics from the system.
// It returns the CPU usage as a float between 0 and 100. If an error occurs
// while retrieving the CPU statistics, 0 is returned instead.
func getCpuMetrics() float64 {
	pct, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("Error getting CPU usage:", err)
		return 0
	}
	return math.Round(pct[0]*100) / 100
}

// getMemoryMetrics retrieves the current memory usage statistics from the system.
// It returns a response.MemoryStats object containing the used and total memory
// in megabytes. If an error occurs while retrieving the memory statistics, an
// empty response.MemoryStats object is returned.
func getMemoryMetrics() response.MemoryStats {
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Println("Error getting memory usage:", err)
		return response.MemoryStats{}
	}
	return response.MemoryStats{
		Used:  int64(m.Used / 1024 / 1024),
		Total: int64(m.Total / 1024 / 1024),
	}
}

// getDiskMetrics retrieves disk partition statistics from the system.
// It returns a slice of response.DiskStats objects containing the name, mount
// point, type, used and total capacity, read and write bytes, and read and write
// bandwidth for each partition. The function also keeps track of the previous
// disk counters and calculates the read and write bandwidth differences between
// the current and previous counters. If an error occurs while retrieving the
// counters, an empty slice is returned.
func getDiskMetrics(elapsed float64) []response.DiskStats {
	parts, err := disk.Partitions(true)
	if err != nil {
		log.Println("Error getting disk partitions:", err)
		return nil
	}
	counters, _ := disk.IOCounters()

	var stats []response.DiskStats
	for _, p := range parts {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		curr := counters[p.Device]
		prev := prevDiskCounters[p.Device]

		var rbps, wbps float64
		if elapsed > 0 {
			rbps = float64(curr.ReadBytes-prev.ReadBytes) / elapsed
			wbps = float64(curr.WriteBytes-prev.WriteBytes) / elapsed
		}

		stats = append(stats, response.DiskStats{
			Name:       p.Device,
			Mount:      p.Mountpoint,
			Type:       p.Fstype,
			Used:       float64(usage.Used) / (1024 * 1024 * 1024),
			Total:      float64(usage.Total) / (1024 * 1024 * 1024),
			ReadBytes:  curr.ReadBytes,
			WriteBytes: curr.WriteBytes,
			ReadBps:    rbps,
			WriteBps:   wbps,
		})
		prevDiskCounters[p.Device] = curr
	}
	return stats
}

// getNetworkMetrics retrieves network traffic statistics from the system.
// It returns a response.NetworkStats object containing the total upload and
// download speeds in Mbps. The function also keeps track of the previous
// network counters and calculates the upload and download speed differences
// between the current and previous counters. If an error occurs while retrieving
// the counters, an empty response.NetworkStats object is returned.
func getNetworkMetrics(elapsed float64) response.NetworkStats {
	counters, err := gopsutil_net.IOCounters(true)
	if err != nil {
		log.Println("Error getting network stats:", err)
		return response.NetworkStats{}
	}

	var up, down float64
	for _, c := range counters {
		if prev, found := prevNetStats[c.Name]; found {
			up += (float64(c.BytesSent-prev.BytesSent) * 8 / 1e6) / elapsed
			down += (float64(c.BytesRecv-prev.BytesRecv) * 8 / 1e6) / elapsed
		}
		prevNetStats[c.Name] = c
	}

	return response.NetworkStats{
		Up:   int64(math.Round(up)),
		Down: int64(math.Round(down)),
	}
}
