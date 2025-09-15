package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/controller"
	"go.uber.org/dig"
)

func DockerRouter(version fiber.Router, container *dig.Container) {
	dockerRoute := version.Group("/docker")

	container.Invoke(func(
		controller *controller.DockerController,
	) {
		dockerRoute.Get("/", controller.DockerHandler)
	})
}
