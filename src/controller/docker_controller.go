package controller

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/service"
	"github.com/valyala/fasthttp"
)

type DockerController struct {
	dockerService service.DockerService
}

// NewDockerController creates a new DockerController instance that is responsible for handling docker-related endpoints.
// It takes a service.DockerService dependency which is used to interact with the Docker daemon.
func NewDockerController(dockerService service.DockerService) *DockerController {
	go dockerService.CollectMetrics()
	return &DockerController{
		dockerService: dockerService,
	}
}

// streamHandler is a generic handler that streams data to the client by calling the provided fetchFunc every second.
// The data is sent as a series of events with the type "data" and the actual data as the payload.
// If fetchFunc returns an error, an event with the type "data" is sent with a payload of "Error retrieving metrics: <error message>".
// The event-stream connection is kept open until the client closes it.
func (controller *DockerController) streamHandler(c *fiber.Ctx, fetchFunc func() ([]byte, error)) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")
	c.Status(fiber.StatusOK)

	// capture context to detect client disconnect
	ctx := c.Context()

	c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				data, err := fetchFunc()
				if err != nil {
					fmt.Fprintf(w, "data: Error retrieving metrics: %s\n\n", err.Error())
					return
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
				w.Flush()
			}
		}
	}))

	return nil
}

// DockerHandler streams the collected Docker metrics to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current Docker metrics.
// The event-stream connection is kept open until the client closes it.
func (controller *DockerController) DockerHandler(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.dockerService.GetDockerMetrics)
}

// DockerStatus streams the collected Docker daemon status to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current Docker daemon status.
// The event-stream connection is kept open until the client closes it.
func (controller *DockerController) DockerStatus(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.dockerService.GetDockerStatus)
}
