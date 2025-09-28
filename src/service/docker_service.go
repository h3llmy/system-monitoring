package service

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/h3llmy/system-monitoring/src/response"
	"github.com/h3llmy/system-monitoring/src/utils/httpClient"
)

type DockerService interface {
	CollectMetrics()
	GetDockerMetrics() ([]byte, error)
	GetDockerStatus() ([]byte, error)
}

type dockerService struct {
	client *httpClient.Client
}

func NewDockerService(client *httpClient.Client) DockerService {
	client.SetBaseURL("http://unix")
	socketPath := os.Getenv("DOCKER_SOCK_PATH")
	if socketPath == "" {
		socketPath = "/var/run/docker.sock"
	}

	client.HTTPClient = &http.Client{
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}
	return &dockerService{client: client}
}

var (
	dockerHistory []response.DockerContainerHistory
	rwMutex       sync.RWMutex
)

func (ds *dockerService) fetchContainers() ([]response.DockerContainerHistory, error) {
	res, err := ds.client.Get("/containers/json")
	if err != nil {
		return nil, err
	}

	var containers []response.DockerContainerHistory
	if err := json.Unmarshal(res, &containers); err != nil {
		return nil, err
	}

	return containers, nil
}

func (ds *dockerService) fetchStats(id string) (response.DockerStats, error) {
	endpoint := fmt.Sprintf("/containers/%s/stats?stream=false", id)
	data, err := ds.client.Get(endpoint)
	if err != nil {
		return response.DockerStats{}, err
	}

	var stats response.DockerStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return response.DockerStats{}, err
	}

	return stats, nil
}

func updateHistory(container response.DockerContainerHistory, stats response.DockerStats) {
	rwMutex.Lock()
	defer rwMutex.Unlock()

	for i, existing := range dockerHistory {
		if existing.ID == container.ID {
			dockerHistory[i].DockerStats = append(existing.DockerStats, stats)

			// trim to maxHistory
			if len(dockerHistory[i].DockerStats) > maxHistory {
				dockerHistory[i].DockerStats = dockerHistory[i].DockerStats[len(dockerHistory[i].DockerStats)-maxHistory:]
			}
			return
		}
	}

	// first time we see this container
	container.DockerStats = []response.DockerStats{stats}
	dockerHistory = append(dockerHistory, container)
}

func (ds *dockerService) CollectMetrics() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		containers, err := ds.fetchContainers()
		if err != nil {
			slog.Error("Failed to get containers", "error", err)
			continue
		}

		var wg sync.WaitGroup
		for _, c := range containers {
			wg.Add(1)
			go func(container response.DockerContainerHistory) {
				defer wg.Done()

				stats, err := ds.fetchStats(container.ID)
				if err != nil {
					slog.Error("Failed to get stats", "id", container.ID, "error", err)
					return
				}

				updateHistory(container, stats)
			}(c)
		}
		wg.Wait()
	}
}

func (ds *dockerService) GetDockerMetrics() ([]byte, error) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	return json.Marshal(dockerHistory)
}

func (ds *dockerService) GetDockerStatus() ([]byte, error) {
	rwMutex.RLock()
	defer rwMutex.RUnlock()
	view := make([]response.DockerContainerHistory, len(dockerHistory))
	for i, c := range dockerHistory {
		view[i] = response.DockerContainerHistory{
			ID:      c.ID,
			Names:   c.Names,
			Image:   c.Image,
			ImageID: c.ImageID,
			Created: c.Created,
			Ports:   c.Ports,
			Labels:  c.Labels,
			State:   c.State,
			Status:  c.Status,
			HostConfig: struct {
				NetworkMode string `json:"networkMode"`
			}{NetworkMode: c.HostConfig.NetworkMode},
			NetworkSettings: struct {
				Networks map[string]response.Network `json:"networks"`
			}{Networks: c.NetworkSettings.Networks},
			Mounts: c.Mounts,
		}
	}

	return json.Marshal(view)
}
