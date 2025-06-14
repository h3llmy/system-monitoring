package router

import (
	"github.com/h3llmy/system-monitoring/src/controller"
	"go.uber.org/dig"

	"github.com/gofiber/fiber/v2"
)

// MonitoringRoutes adds routes for accessing monitoring metrics.
//
// The /monitoring route is added to the provided version group.
func MonitoringRoutes(version fiber.Router, container *dig.Container) {
	monitoringRoute := version.Group("/monitoring")

	container.Invoke(func(
		controller *controller.MonitoringController,
	) {
		monitoringRoute.Get("/", controller.MonitoringHandler)
		monitoringRoute.Get("/cpu", controller.MonitoringCpuHandler)
		monitoringRoute.Get("/memory", controller.MonitoringMemoryHandler)
		monitoringRoute.Get("/disk", controller.MonitoringDiskHandler)
		monitoringRoute.Get("/network", controller.MonitoringNetworkHandler)
	})
}
