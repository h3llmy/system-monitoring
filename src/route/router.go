package router

import (
	"time"

	"github.com/h3llmy/system-monitoring/src/controller"
	"github.com/h3llmy/system-monitoring/src/service"
	httpClient "github.com/h3llmy/system-monitoring/src/utils/httpClient"

	"github.com/gofiber/fiber/v2"
)

// Routes sets up all routes for the application.
func Routes(app *fiber.App) {
	// utils initialize
	httpClient := httpClient.NewClient(10 * time.Second)

	// services initialize
	monitoringService := service.NewSystemMonitorService()
	jfService := service.NewJellyfinService(httpClient)

	// controllers initialize
	monitoringController := controller.NewMonitoringController(monitoringService)
	jellyfinController := controller.NewJellyfinController(*jfService)

	v1 := app.Group("/api/v1")

	MonitoringRoutes(v1, monitoringController)
	JellyfinRouter(v1, jellyfinController)
}
