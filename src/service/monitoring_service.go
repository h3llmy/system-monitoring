package service

import (
	"encoding/json"
	"log"
	"math"
	"sync"
	"system-monitoring/src/response"
	"time"

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

// CollectMetrics collects system metrics and maintains the latest 60 records.
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

// GetHistory returns the collected metrics history in JSON format.
func (sm *SystemMonitor) GetHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return json.Marshal(history)
}

func getCpuMetrics() float64 {
	cpuPercentages, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("Error getting CPU usage:", err)
	}
	return math.Floor(cpuPercentages[0]*100) / 100
}

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
