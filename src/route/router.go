package router

import (
	"system-monitoring/src/controller"
	"system-monitoring/src/service"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	monitoringService := service.NewSystemMonitorService()
	monitoringController := controller.NewMonitoringController(monitoringService)

	v1 := app.Group("/api/v1")

	MonitoringRoutes(v1, monitoringController)
}