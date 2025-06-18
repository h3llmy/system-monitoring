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

func (sm *SystemMonitor) GetHistory() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return json.Marshal(history)
}

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

func getCpuMetrics() float64 {
	pct, err := cpu.Percent(0, false)
	if err != nil {
		log.Println("Error getting CPU usage:", err)
		return 0
	}
	return math.Round(pct[0]*100) / 100
}

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
			Name:       diskName,
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
