package main

import (
	router "github.com/h3llmy/system-monitoring/src/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	app.Use(cors.New())

	router.Routes(app)

	app.Listen(":3000")
}