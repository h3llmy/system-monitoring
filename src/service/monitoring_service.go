package service

import (
	"encoding/json"
	"log"
	"math"
	"sync"
	"time"

	"github.com/h3llmy/system-monitoring/src/response"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	gopsutil_net "github.com/shirou/gopsutil/v3/net"
)

type MonitoringService interface {
	CollectMetrics()
	GetHistory() ([]byte, error)
}

// SystemMonitor implements MonitoringService
type SystemMonitor struct{}

func NewSystemMonitorService() MonitoringService {
	return &SystemMonitor{}
}

var (
	history       []response.SystemMetrics
	mu            sync.Mutex
	prevNetStats  = make(map[string]gopsutil_net.IOCountersStat)
	prevTime      time.Time
)


// CollectMetrics continuously collects system metrics and stores them in a ring buffer of a limited size.
// The metrics are collected every second and stored in the order of collection.
func (sm *SystemMonitor) CollectMetrics() {
	for {
		startTime := time.Now()
		cpuPercentages := getCpuMetrics()
		memStats := getMemoryMetrics()
		disks := getDiskMetrics()
		netStats := getNetworkMetrics()

		metrics := response.SystemMetrics{
			Timestamp: startTime.Format(time.RFC3339),
			CPU: cpuPercentages,
			Memory: memStats,
			Disk: disks,
			Network: netStats,
		}

		mu.Lock()
		history = append(history, metrics)
		if len(history) > 60 {
			history = history[1:]
		}
		mu.Unlock()

		time.Sleep(1 * time.Second)
	}
}

// GetHistory returns the collected system metrics in JSON format.
// The returned metrics are from the latest 60 seconds, or as many as have been collected.
// The metrics are sorted in the order of collection.
func (sm *SystemMonitor) GetHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return json.Marshal(history)
}

// getCpuMetrics returns the current CPU usage as a float64 in the range [0.00, 100.00].
func getCpuMetrics() float64 {
	cpuPercentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("Error getting CPU usage:", err)
	}
	return math.Floor(cpuPercentages[0]*100) / 100
}

// getMemoryMetrics retrieves the current memory usage statistics.
// It returns a MemoryStats struct containing the used and total memory,
// both measured in megabytes. In case of an error, the used and total
// values will default to zero, and an error message will be logged.
func getMemoryMetrics() response.MemoryStats {
	memStats, err := mem.VirtualMemory()
	if err != nil {
		log.Println("Error getting memory usage:", err)
	}
	return response.MemoryStats{
		Used:  int64(memStats.Used / 1024 / 1024),
		Total: int64(memStats.Total / 1024 / 1024),
	}
}

// getDiskMetrics retrieves the current disk usage statistics.
// It returns a slice of DiskStats containing the device name, mount point, file system type, used and total disk space,
// both measured in gigabytes. In case of an error, the returned slice will be empty, and an error message will be logged.
func getDiskMetrics() []response.DiskStats {
	diskStats, err := disk.Partitions(true)
	if err != nil {
		log.Println("Error getting disk partitions:", err)
	}

	var disks []response.DiskStats
	for _, d := range diskStats {
		usage, err := disk.Usage(d.Mountpoint)
		if err == nil {
			disks = append(disks, response.DiskStats{
				Name:  d.Device,
				Mount: d.Mountpoint,
				Type:  d.Fstype,
				Used:  float64(usage.Used) / (1024 * 1024 * 1024),
				Total: float64(usage.Total) / (1024 * 1024 * 1024),
			})
		}
	}
	return disks
}

// getNetworkMetrics retrieves the current network usage statistics.
// It returns a NetworkStats struct containing the number of megabits per second sent and received.
// In case of an error, the up and down values will default to zero, and an error message will be logged.
func getNetworkMetrics() response.NetworkStats {
	counters, err := gopsutil_net.IOCounters(true)
	if err != nil {
		log.Println("Error getting network stats:", err)
		return response.NetworkStats{}
	}

	currentTime := time.Now()
	elapsed := currentTime.Sub(prevTime).Seconds()
	if elapsed <= 0 {
		prevTime = currentTime
		return response.NetworkStats{}
	}

	var sentRate, recvRate float64
	for _, currStat := range counters {
		prevStat, exists := prevNetStats[currStat.Name]
		if exists {
			bytesSent := currStat.BytesSent - prevStat.BytesSent
			bytesRecv := currStat.BytesRecv - prevStat.BytesRecv

			sentRate += (float64(bytesSent) * 8) / 1_000_000 / elapsed
			recvRate += (float64(bytesRecv) * 8) / 1_000_000 / elapsed
		}
		prevNetStats[currStat.Name] = currStat
	}

	prevTime = currentTime
	return response.NetworkStats{
		Up:   int64(sentRate),
		Down: int64(recvRate),
	}
}
