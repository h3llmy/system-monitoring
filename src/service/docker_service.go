package service

import (
	"encoding/json"
	"log"

	"github.com/shirou/gopsutil/v4/docker"
)

type DockerService interface {
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

func (ds *dockerService) GetDockerMetrics() ([]byte, error) {
	dockerStats, err := docker.GetDockerStat()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return json.Marshal(dockerStats)
}
