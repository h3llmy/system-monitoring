package service

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/shirou/gopsutil/v4/docker"
)

type DockerService interface {
	CollectMetrics()
	GetDockerMetrics() ([]byte, error)
}

type dockerService struct {
}

// NewDockerService creates a new DockerService instance.
//
// The returned service is an implementation of the DockerService interface.
// It provides a way to collect metrics from the Docker daemon.
func NewDockerService() DockerService {
	return &dockerService{}
}

var (
	dockerHistory []docker.CgroupDockerStat
)

// CollectMetrics is an infinite loop that collects Docker metrics at 1 second intervals.
//
// It collects the metrics by calling docker.GetDockerStatWithContext and stores the result in the dockerHistory field.
// If an error occurs while collecting the metrics, an error message is logged.
// The function is intended to be run in a goroutine.
func (ds *dockerService) CollectMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	dockerStats, err := docker.GetDockerStatWithContext(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	mu.Lock()
	dockerHistory = dockerStats
	mu.Unlock()

	time.Sleep(1 * time.Second)
}

// GetDockerMetrics returns the most recent Docker metrics as a JSON payload.
//
// The returned payload contains a slice of CgroupDockerStat, which holds the Docker container statistics.
// The function acquires a mutex lock to ensure thread-safe access to the dockerHistory.
func (ds *dockerService) GetDockerMetrics() ([]byte, error) {
	mu.Lock()
	defer mu.Unlock()
	return json.Marshal(dockerHistory)
}
