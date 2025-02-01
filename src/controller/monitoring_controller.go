package controller

import (
	"bufio"
	"fmt"
	"system-monitoring/src/service"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type MonitoringController struct {
	monitoringService service.MonitoringService
}

// NewMonitoringController initializes a new MonitoringController with dependency injection.
func NewMonitoringController(monitoringService service.MonitoringService) *MonitoringController {
	go monitoringService.CollectMetrics()

	return &MonitoringController{
		monitoringService: monitoringService,
	}
}

// MonitoringHandler streams system metrics to the client.
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
