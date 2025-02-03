package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/controller"
)

// JellyfinRouter adds routes for accessing jellyfin metrics.
//
// The /jellyfin route is added to the provided version group.
func JellyfinRouter(version fiber.Router, controller *controller.JellyfinController){
	jellyfinRoute := version.Group("/jellyfin")

	jellyfinRoute.Get("/Count", controller.GetJellyfinCount)
}