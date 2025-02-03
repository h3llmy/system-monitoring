package main

import (
	"fmt"
	"os"

	router "github.com/h3llmy/system-monitoring/src/route"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// main sets up the Fiber application and starts the HTTP server.
//
// The function does the following:
// - Loads environment variables from a .env file.
// - Creates a new Fiber app.
// - Sets up CORS middleware.
// - Adds routes for the monitoring app.
// - Starts the HTTP server on port 3000.
func main() {
	loadEnv()
	app := fiber.New()

	port := os.Getenv("PORT")

	app.Use(cors.New())

	router.Routes(app)

	app.Listen(fmt.Sprintf(":%s", port))
}

// loadEnv loads environment variables from a .env file.
// If the file cannot be loaded, the function panics with an error message.
func loadEnv()  {
	err := godotenv.Load()
	if err != nil {
	  panic("Error loading .env file")
	}
}