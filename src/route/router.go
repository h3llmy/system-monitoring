package router

import (
	"github.com/h3llmy/system-monitoring/src/controller"
	"github.com/h3llmy/system-monitoring/src/service"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	monitoringService := service.NewSystemMonitorService()
	monitoringController := controller.NewMonitoringController(monitoringService)

	v1 := app.Group("/api/v1")

	MonitoringRoutes(v1, monitoringController)
}