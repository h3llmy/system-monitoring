package controller

import (
	"bufio"
	"fmt"
	"time"

	"github.com/h3llmy/system-monitoring/src/service"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type MonitoringController struct {
	monitoringService service.MonitoringService
}


// NewMonitoringController returns a new MonitoringController instance.
// It starts the monitoring service metrics collection goroutine.
func NewMonitoringController(monitoringService service.MonitoringService) *MonitoringController {
	go monitoringService.CollectMetrics()

	return &MonitoringController{
		monitoringService: monitoringService,
	}
}


// MonitoringHandler handles HTTP requests to stream system metrics in real-time.
// It sets up the response headers for server-sent events and uses a stream writer
// to periodically send JSON-encoded system metrics data to the client every second.
// The function retrieves the metrics history from the monitoring service and writes
// it to the response body. If an error occurs while retrieving the metrics, an error
// message is sent instead.
func (controller *MonitoringController) MonitoringHandler(c *fiber.Ctx) error  {
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Transfer-Encoding", "chunked")

	c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			<-ticker.C
			data, err := controller.monitoringService.GetHistory()
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
