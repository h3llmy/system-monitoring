package controller

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/service"
	"github.com/valyala/fasthttp"
)

type MonitoringController struct {
	monitoringService service.MonitoringService
}

// NewMonitoringController creates a new MonitoringController instance that is responsible for handling monitoring endpoints.
// It takes a service.MonitoringService dependency which is used to interact with the monitoring service. The provided service
// is started in a goroutine to collect system metrics in the background.
func NewMonitoringController(monitoringService service.MonitoringService) *MonitoringController {
	go monitoringService.CollectMetrics()
	return &MonitoringController{monitoringService: monitoringService}
}

// streamHandler is a generic handler that streams data to the client by calling the provided fetchFunc every second.
// The data is sent as a series of events with the type "data" and the actual data as the payload.
// If fetchFunc returns an error, an event with the type "data" is sent with a payload of "Error retrieving metrics: <error message>".
// The event-stream connection is kept open until the client closes it.
func (controller *MonitoringController) streamHandler(c *fiber.Ctx, fetchFunc func() ([]byte, error)) error {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			data, err := fetchFunc()
			if err != nil {
				fmt.Fprintf(w, "data: Error retrieving metrics: %s\n\n", err.Error())
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()
		}
	}))

	return nil
}

// MonitoringHandler streams the collected system metrics history to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current metrics.
// The event-stream connection is kept open until the client closes it.
func (controller *MonitoringController) MonitoringHandler(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.monitoringService.GetHistory)
}

// MonitoringCpuHandler streams the collected CPU usage metrics history to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current CPU usage metrics.
// The event-stream connection is kept open until the client closes it.
func (controller *MonitoringController) MonitoringCpuHandler(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.monitoringService.GetCpuHistory)
}

// MonitoringMemoryHandler streams the collected memory usage metrics history to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current memory usage metrics.
// The event-stream connection is kept open until the client closes it.
func (controller *MonitoringController) MonitoringMemoryHandler(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.monitoringService.GetMemoryHistory)
}

// MonitoringDiskHandler streams the collected disk usage metrics history to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current disk usage metrics.
// The event-stream connection is kept open until the client closes it.
func (controller *MonitoringController) MonitoringDiskHandler(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.monitoringService.GetDiskHistory)
}

// MonitoringNetworkHandler streams the collected network usage metrics history to the client in a series of events.
// The client will receive a continuous stream of events, with the type "data" and a payload of the current network usage metrics.
// The event-stream connection is kept open until the client closes it.
func (controller *MonitoringController) MonitoringNetworkHandler(c *fiber.Ctx) error {
	return controller.streamHandler(c, controller.monitoringService.GetNetworkHistory)
}
