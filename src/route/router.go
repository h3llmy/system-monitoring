package router

import (
	"time"

	"github.com/h3llmy/system-monitoring/src/controller"
	"github.com/h3llmy/system-monitoring/src/service"
	httpClient "github.com/h3llmy/system-monitoring/src/utils/httpClient"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/dig"
)

// Routes sets up all routes for the application using dependency injection.
func Routes(app *fiber.App) {
	container := setupContainer()

	v1 := app.Group("/api/v1")

	MonitoringRoutes(v1, container)
	JellyfinRouter(v1, container)
}

// SetupContainer initializes the DI container and registers all dependencies.
func setupContainer() *dig.Container {
	container := dig.New()

	// Provide utility
	container.Provide(func() *httpClient.Client {
		return httpClient.NewClient(10 * time.Second)
	})

	// Provide services
	container.Provide(service.NewSystemMonitorService)
	container.Provide(service.NewJellyfinService)

	// Provide controllers
	container.Provide(controller.NewMonitoringController)
	container.Provide(controller.NewJellyfinController)

	return container
}
