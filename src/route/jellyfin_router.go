package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/controller"
	"go.uber.org/dig"
)

// JellyfinRouter adds routes for accessing jellyfin metrics.
//
// The /jellyfin route is added to the provided version group.
func JellyfinRouter(version fiber.Router, container *dig.Container) {
	jellyfinRoute := version.Group("/jellyfin")

	container.Invoke(func(
		controller *controller.JellyfinController,
	) {
		jellyfinRoute.Get("/count", controller.GetJellyfinCount)
	})
}
