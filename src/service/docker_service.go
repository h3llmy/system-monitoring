package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/docker"
)

type DockerService interface {
	CollectMetrics()
	GetDockerMetrics() ([]byte, error)
}

type dockerService struct{}

func NewDockerService() DockerService {
	return &dockerService{}
}

var (
	dockerHistory []docker.CgroupDockerStat
	mutex         sync.RWMutex
)

// CollectMetrics starts a loop that periodically collects Docker stats.
// This should be run as a single goroutine.
func (ds *dockerService) CollectMetrics() {
	ticker := time.NewTicker(5 * time.Second) // Collect every 5 seconds
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		dockerStats, err := docker.GetDockerStatWithContext(ctx)

		cancel()
		if err != nil {
			slog.Error("Failed to get Docker stats", err)
			continue
		}

		mutex.Lock()
		dockerHistory = dockerStats
		mutex.Unlock()
	}
}

// GetDockerMetrics returns the latest collected Docker stats in JSON format.
func (ds *dockerService) GetDockerMetrics() ([]byte, error) {
	mutex.RLock()
	defer mutex.RUnlock()
	return json.Marshal(dockerHistory)
}
