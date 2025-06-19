package service

import (
	"encoding/json"
	"log"
	"math"
	"path"
	"sync"
	"time"

	"maps"

	"github.com/h3llmy/system-monitoring/src/response"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gopsutil_net "github.com/shirou/gopsutil/v4/net"
)

type MonitoringService interface {
	CollectMetrics()
	GetHistory() ([]byte, error)
	GetCpuMetrics() ([]byte, error)
	GetMemoryMetrics() ([]byte, error)
	GetDiskMetrics() ([]byte, error)
	GetNetworkMetrics() ([]byte, error)
}

type SystemMonitor struct{}

// NewSystemMonitorService creates a new SystemMonitor instance which is responsible for collecting system metrics.
//
// The provided service is started in a goroutine to collect system metrics in the background.
func NewSystemMonitorService() MonitoringService {
	return &SystemMonitor{}
}

var (
	history          response.SystemMetrics
	mu               sync.Mutex
	prevNetStats     = make(map[string]gopsutil_net.IOCountersStat)
	prevDiskCounters = make(map[string]disk.IOCountersStat)
	prevTime         time.Time
	maxHistory       = 60
)

// init initializes the history and network/disk counters with the current values.
// This is done to calculate the difference between the current and previous values.
func init() {
	prevTime = time.Now()
	if stats, err := gopsutil_net.IOCounters(true); err == nil {
		for _, st := range stats {
			prevNetStats[st.Name] = st
		}
	}
	if counters, err := disk.IOCounters(); err == nil {
		maps.Copy(prevDiskCounters, counters)
	}
	history.Matrics = &[]response.Matrics{}
}

// CollectMetrics is an infinite loop that collects system metrics at 1 second intervals.
//
// It collects the following metrics:
// - CPU usage percentage
// - Memory usage statistics (used, total)
// - Disk usage statistics (used, total)
// - Network traffic statistics (up, down)
//
// The collected metrics are stored in the history field of the service.
// The history field is a slice of Matrics, which is a struct that holds the collected metrics.
// The history field is trimmed to a maximum length of maxHistory (default 60).
//
// The function is intended to be run in a goroutine.
func (sm *SystemMonitor) CollectMetrics() {
	for {
		now := time.Now()
		elapsed := now.Sub(prevTime).Seconds()

		cpuPct := getCpuMetrics()
		memStats := getMemoryMetrics()
		diskStats := getDiskMetrics(elapsed)
		netStats := getNetworkMetrics(elapsed)

		mu.Lock()

		// Update disk and timestamp
		history.Timestamp = now.Format(time.RFC3339)
		history.Disk = &diskStats

		// Append to history and trim if needed
		mat := response.Matrics{
			CPU:     &cpuPct,
			Memory:  &memStats,
			Network: &netStats,
		}
		*history.Matrics = append(*history.Matrics, mat)
		if len(*history.Matrics) > maxHistory {
			*history.Matrics = (*history.Matrics)[len(*history.Matrics)-maxHistory:]
		}

		mu.Unlock()

		prevTime = now
		time.Sleep(time.Until(now.Add(1 * time.Second)))
	}
}

// GetHistory returns the collected system metrics history as a JSON payload.
//
// The returned payload is a slice of Matrics, which is a struct that holds the collected metrics.
// The Matrics struct contains the following fields:
// - Timestamp: the ISO 8601 timestamp of when the metrics were collected
// - CPU: the CPU usage percentage
// - Memory: memory usage statistics (used, total)
// - Disk: disk usage statistics (used, total)
// - Network: network traffic statistics (up, down)
//
// The payload is ordered by timestamp, with the most recent metrics first.
func (sm *SystemMonitor) GetHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return json.Marshal(history)
}

// GetCpuMetrics returns the most recent CPU usage metric as a JSON payload.
//
// The returned payload contains the following fields:
// - Timestamp: the ISO 8601 timestamp of when the metrics were collected
// - CPU: the CPU usage percentage
func (sm *SystemMonitor) GetCpuMetrics() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	var last *float64
	if len(*history.Matrics) > 0 {
		last = (*history.Matrics)[len(*history.Matrics)-1].CPU
	}

	return json.Marshal(struct {
		Timestamp string   `json:"timestamp"`
		CPU       *float64 `json:"cpu"`
	}{
		Timestamp: history.Timestamp,
		CPU:       last,
	})
}

// GetMemoryMetrics returns the most recent memory usage metric as a JSON payload.
//
// The returned payload contains the following fields:
// - Timestamp: the ISO 8601 timestamp of when the metrics were collected
// - Memory: memory usage statistics (used, total)
func (sm *SystemMonitor) GetMemoryMetrics() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	var last *response.MemoryStats
	if len(*history.Matrics) > 0 {
		last = (*history.Matrics)[len(*history.Matrics)-1].Memory
	}

	return json.Marshal(struct {
		Timestamp string                `json:"timestamp"`
		Memory    *response.MemoryStats `json:"memory"`
	}{
		Timestamp: history.Timestamp,
		Memory:    last,
	})
}

// GetDiskMetrics returns the most recent disk usage metric as a JSON payload.
//
// The returned payload contains the following fields:
// - Timestamp: the ISO 8601 timestamp of when the metrics were collected
// - Disk: disk usage statistics (used, total)
func (sm *SystemMonitor) GetDiskMetrics() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	return json.Marshal(struct {
		Timestamp string                `json:"timestamp"`
		Disk      *[]response.DiskStats `json:"disk"`
	}{
		Timestamp: history.Timestamp,
		Disk:      history.Disk,
	})
}

// GetNetworkMetrics returns the most recent network traffic metrics as a JSON payload.
//
// The returned payload contains the following fields:
// - Timestamp: the ISO 8601 timestamp of when the metrics were collected
// - Network: network traffic statistics (up, down)
func (sm *SystemMonitor) GetNetworkMetrics() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	var last *response.NetworkStats
	if len(*history.Matrics) > 0 {
		last = (*history.Matrics)[len(*history.Matrics)-1].Network
	}

	return json.Marshal(struct {
		Timestamp string                 `json:"timestamp"`
		Network   *response.NetworkStats `json:"network"`
	}{
		Timestamp: history.Timestamp,
		Network:   last,
	})
}

// getCpuMetrics returns the current CPU usage as a percentage.
// It returns 0 if an error occurs.
func getCpuMetrics() float64 {
	pct, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("Error getting CPU usage:", err)
		return 0
	}
	return math.Round(pct[0]*100) / 100
}

// getMemoryMetrics returns the current memory usage statistics.
//
// The returned struct contains the following fields:
// - Used: the current amount of memory used, in bytes
// - Total: the total amount of memory, in bytes
//
// If an error occurs, the function logs the error and returns an empty
// response.MemoryStats struct.
func getMemoryMetrics() response.MemoryStats {
	m, err := mem.VirtualMemory()
	if err != nil {
		log.Println("Error getting memory usage:", err)
		return response.MemoryStats{}
	}
	return response.MemoryStats{
		Used:  int64(m.Used),
		Total: int64(m.Total),
	}
}

// getDiskMetrics returns disk usage statistics for all partitions.
//
// It calculates the read and write bytes per second (Bps) based on the elapsed time
// since the last call. The function retrieves partition information and I/O counters
// for each disk and computes statistics including the used and total space, read bytes,
// write bytes, and Bps values.
//
// Parameters:
// - elapsed: Time duration in seconds since the last invocation, used to calculate Bps.
//
// Returns:
// A slice of response.DiskStats containing the disk usage statistics for each partition.
// If an error occurs while retrieving partition information, an empty slice is returned
// and the error is logged.
func getDiskMetrics(elapsed float64) []response.DiskStats {
	parts, err := disk.Partitions(false)
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

		diskName := path.Base(p.Device)
		stats = append(stats, response.DiskStats{
			Name:        diskName,
			Mount:       p.Mountpoint,
			Type:        p.Fstype,
			UsedPercent: usage.UsedPercent,
			Used:        usage.Used,
			Total:       usage.Total,
			ReadBytes:   curr.ReadBytes,
			WriteBytes:  curr.WriteBytes,
			ReadBps:     rbps,
			WriteBps:    wbps,
		})
		prevDiskCounters[p.Device] = curr
	}
	return stats
}

// getNetworkMetrics returns network traffic statistics.
//
// It calculates the uplink and downlink values in Mbps based on the elapsed time
// since the last call. The function retrieves I/O counters for all network interfaces
// and computes statistics including the current uplink and downlink values.
//
// Parameters:
// - elapsed: Time duration in seconds since the last invocation, used to calculate Mbps.
//
// Returns:
// A response.NetworkStats containing the network traffic statistics.
// If an error occurs while retrieving I/O counters, an empty struct is returned
// and the error is logged.
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
