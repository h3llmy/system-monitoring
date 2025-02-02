package router

import (
	"github.com/h3llmy/system-monitoring/src/controller"
	"github.com/h3llmy/system-monitoring/src/service"
	"github.com/h3llmy/system-monitoring/src/utils"

	"github.com/gofiber/fiber/v2"
)

// Routes sets up all routes for the application.
func Routes(app *fiber.App) {
	monitoringService := service.NewSystemMonitorService()
	monitoringController := controller.NewMonitoringController(monitoringService)

	jfService := service.NewJellyfinService(utils.NewHttpClient())
	jellyfinController := controller.NewJellyfinController(*jfService)

	v1 := app.Group("/api/v1")

	MonitoringRoutes(v1, monitoringController)
	JellyfinRouter(v1, jellyfinController)
}