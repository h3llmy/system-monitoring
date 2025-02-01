package router

import (
	"github.com/h3llmy/system-monitoring/src/controller"

	"github.com/gofiber/fiber/v2"
)

func MonitoringRoutes(versionRouter fiber.Router, controller *controller.MonitoringController) {
	versionRouter.Get("/monitoring", controller.MonitoringHandler)
}