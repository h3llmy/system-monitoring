package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"path"
	"strings"
	"time"

	"maps"

	"github.com/h3llmy/system-monitoring/src/response"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	gopsutil_net "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/sensors"
)

type MonitoringService interface {
	CollectMetrics()
	GetHistory() ([]byte, error)
	GetCpuMetrics() ([]byte, error)
	GetMemoryMetrics() ([]byte, error)
	GetDiskMetrics() ([]byte, error)
	GetNetworkMetrics() ([]byte, error)
	GetSensorsMetrics() ([]byte, error)
}

type SystemMonitor struct{}

// NewSystemMonitorService creates a new SystemMonitor instance which is responsible for collecting system metrics.
//
// The provided service is started in a goroutine to collect system metrics in the background.
func NewSystemMonitorService() MonitoringService {
	return &SystemMonitor{}
}

var (
	systemHistory    response.SystemMetrics
	prevNetStats     = make(map[string]gopsutil_net.IOCountersStat)
	prevDiskCounters = make(map[string]disk.IOCountersStat)
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
	systemHistory.Matrics = &[]response.Matrics{}
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
	ctx := context.Background()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for now := range ticker.C {
		elapsed := now.Sub(prevTime).Seconds()

		cpuPct := getCpuMetrics(ctx)
		memStats := getMemoryMetrics(ctx)
		diskStats := getDiskMetrics(ctx, elapsed)
		netStats := getNetworkMetrics(ctx, elapsed)
		temp, err := getTemperatureSensors()
		if err != nil {
			log.Println(err)
		}

		mu.Lock()

		// Update disk and timestamp
		systemHistory.Timestamp = now.Format(time.RFC3339)
		systemHistory.Disk = &diskStats
		systemHistory.Temprature = temp

		// Append to history and trim if needed
		mat := response.Matrics{
			CPU:     &cpuPct,
			Memory:  &memStats,
			Network: &netStats,
		}
		*systemHistory.Matrics = append(*systemHistory.Matrics, mat)
		if len(*systemHistory.Matrics) > maxHistory {
			*systemHistory.Matrics = (*systemHistory.Matrics)[len(*systemHistory.Matrics)-maxHistory:]
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
	return json.Marshal(systemHistory)
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
	if len(*systemHistory.Matrics) > 0 {
		last = (*systemHistory.Matrics)[len(*systemHistory.Matrics)-1].CPU
	}

	return json.Marshal(struct {
		Timestamp string   `json:"timestamp"`
		CPU       *float64 `json:"cpu"`
	}{
		Timestamp: systemHistory.Timestamp,
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
	if len(*systemHistory.Matrics) > 0 {
		last = (*systemHistory.Matrics)[len(*systemHistory.Matrics)-1].Memory
	}

	return json.Marshal(struct {
		Timestamp string                `json:"timestamp"`
		Memory    *response.MemoryStats `json:"memory"`
	}{
		Timestamp: systemHistory.Timestamp,
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
		Timestamp: systemHistory.Timestamp,
		Disk:      systemHistory.Disk,
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
	if len(*systemHistory.Matrics) > 0 {
		last = (*systemHistory.Matrics)[len(*systemHistory.Matrics)-1].Network
	}

	return json.Marshal(struct {
		Timestamp string                 `json:"timestamp"`
		Network   *response.NetworkStats `json:"network"`
	}{
		Timestamp: systemHistory.Timestamp,
		Network:   last,
	})
}

// GetSensorsMetrics returns the most recent sensors metrics as a JSON payload.
//
// The returned payload contains the following fields:
// - Timestamp: the ISO 8601 timestamp of when the metrics were collected
// - CPU: a slice of CoreTemperatureStats for each CPU core
// - Ambient: the ambient temperature
// - GPU: a slice of CoreTemperatureStats for each GPU core
// - Core: a slice of CoreTemperatureStats for each core (both CPU and GPU)
func (sm *SystemMonitor) GetSensorsMetrics() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()

	return json.Marshal(struct {
		Timestamp string                           `json:"timestamp"`
		CPU       []*response.CoreTemperatureStats `json:"cpu,omitempty"`
		Ambient   *float64                         `json:"ambient,omitempty"`
		GPU       []*response.CoreTemperatureStats `json:"gpu,omitempty"`
		Core      []*response.CoreTemperatureStats `json:"core,omitempty"`
	}{
		Timestamp: systemHistory.Timestamp,
		CPU:       systemHistory.Temprature.CPU,
		Ambient:   systemHistory.Temprature.Ambient,
		GPU:       systemHistory.Temprature.GPU,
		Core:      systemHistory.Temprature.Core,
	})
}

// getCpuMetrics returns the current CPU usage as a percentage.
// It returns 0 if an error occurs.
func getCpuMetrics(ctx context.Context) float64 {
	pct, err := cpu.PercentWithContext(ctx, 100*time.Millisecond, false)

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
func getMemoryMetrics(ctx context.Context) response.MemoryStats {
	m, err := mem.VirtualMemoryWithContext(ctx)
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
func getDiskMetrics(ctx context.Context, elapsed float64) []response.DiskStats {
	parts, err := disk.PartitionsWithContext(ctx, false)
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
func getNetworkMetrics(ctx context.Context, elapsed float64) response.NetworkStats {
	counters, err := gopsutil_net.IOCountersWithContext(ctx, true)
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

// getTemperatureSensors returns a slice of response.TemperatureStats that contains
// the temperatures of all found sensors. The returned slice is sorted by sensor name.
//
// The function maps the sensor key to a friendly label, using a hardcoded mapping
// for well-known substrings. If no mapping is found, it uses a fallback that
// title-cases the words in the sensor key.
//
// If an error occurs while retrieving the temperatures, an empty slice is returned
// and the error is logged.
func getTemperatureSensors() (*response.TemperatureStats, error) {
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return nil, err
	}

	result := &response.TemperatureStats{}
	var cpuTemps []*response.CoreTemperatureStats
	var gpuTemps []*response.CoreTemperatureStats
	var coreTemps []*response.CoreTemperatureStats
	var ambientTemp *float64

	for _, t := range temps {
		key := strings.ToLower(t.SensorKey)

		switch {
		case strings.Contains(key, "package"):
			// CPU package temperature
			cpuTemps = append(cpuTemps, &response.CoreTemperatureStats{
				Name:     getSensorsName(t.SensorKey),
				Temp:     t.Temperature,
				High:     t.High,
				Critical: t.Critical,
			})
		case strings.Contains(key, "gpu"):
			// GPU temperature
			gpuTemps = append(gpuTemps, &response.CoreTemperatureStats{
				Name:     getSensorsName(t.SensorKey),
				Temp:     t.Temperature,
				High:     t.High,
				Critical: t.Critical,
			})
		case strings.Contains(key, "core"):
			// CPU core temperature
			coreTemps = append(coreTemps, &response.CoreTemperatureStats{
				Name:     getSensorsName(t.SensorKey),
				Temp:     t.Temperature,
				High:     t.High,
				Critical: t.Critical,
			})
		case strings.Contains(key, "acpitz"):
			// Ambient temperature - keep the highest one
			if ambientTemp == nil || t.Temperature > *ambientTemp {
				ambientTemp = &t.Temperature
			}
		default:
			// Try to categorize other sensors
			friendlyName := getSensorsName(t.SensorKey)
			if strings.Contains(strings.ToLower(friendlyName), "cpu") {
				cpuTemps = append(cpuTemps, &response.CoreTemperatureStats{
					Name:     friendlyName,
					Temp:     t.Temperature,
					High:     t.High,
					Critical: t.Critical,
				})
			} else {
				// Default to core temps for unknown sensors
				coreTemps = append(coreTemps, &response.CoreTemperatureStats{
					Name:     friendlyName,
					Temp:     t.Temperature,
					High:     t.High,
					Critical: t.Critical,
				})
			}
		}
	}

	// Assign non-empty slices to result
	if len(cpuTemps) > 0 {
		result.CPU = cpuTemps
	}
	if len(gpuTemps) > 0 {
		result.GPU = gpuTemps
	}
	if len(coreTemps) > 0 {
		result.Core = coreTemps
	}
	if ambientTemp != nil {
		result.Ambient = ambientTemp
	}

	// Check if we found any temperatures
	if result.CPU == nil && result.GPU == nil && result.Core == nil && result.Ambient == nil {
		return nil, fmt.Errorf("no temperature sensors found")
	}

	return result, nil
}

// getSensorsName returns a human-readable name for the given sensorKey.
// It uses simple pattern matching for common patterns and falls back to title-casing
// the words in sensorKey if no match is found.
func getSensorsName(sensorKey string) string {
	// Simple mapping for common patterns
	key := strings.ToLower(sensorKey)

	if strings.Contains(key, "package") {
		return "CPU"
	}
	if strings.Contains(key, "gpu") {
		return "GPU"
	}
	if strings.Contains(key, "acpitz") {
		return "Ambient"
	}

	// Fallback: title-case words in sensorKey
	parts := strings.Split(sensorKey, "_")
	caser := cases.Title(language.English)
	for i := range parts {
		parts[i] = caser.String(parts[i])
	}
	return strings.Join(parts, " ")
}
