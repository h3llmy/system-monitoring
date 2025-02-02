package router

import (
	"github.com/h3llmy/system-monitoring/src/controller"

	"github.com/gofiber/fiber/v2"
)

// MonitoringRoutes adds routes for accessing monitoring metrics.
//
// The /monitoring route is added to the provided version group.
func MonitoringRoutes(version fiber.Router, controller *controller.MonitoringController) {
	monitoringRoute := version.Group("/monitoring")

	monitoringRoute.Get("/", controller.MonitoringHandler)
}