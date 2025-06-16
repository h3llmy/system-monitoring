package main

import (
	"fmt"
	"log"
	"os"

	"github.com/h3llmy/system-monitoring/src/config"
	router "github.com/h3llmy/system-monitoring/src/route"
	env "github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
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
	app := fiber.New(config.ApplicationConfig)

	port := os.Getenv("PORT")

	initMiddlewares(app)

	router.Routes(app)

	log.Fatal(app.Listen(fmt.Sprintf("127.0.0.1:%s", port)))
}

// initMiddlewares configures and registers middleware for the Fiber application.
//
// It adds the following middleware:
// - Compression: Compresses HTTP responses.
// - CORS: Enables Cross-Origin Resource Sharing.
// - Helmet: Helps secure HTTP headers.
// - Request ID: Generates a unique ID for each request.
// - Healthcheck: Provides a liveness endpoint.
func initMiddlewares(app *fiber.App) {
	app.Use(compress.New())
	app.Use(cors.New())
	app.Use(helmet.New())
	app.Use(requestid.New())
	app.Use(healthcheck.New(config.HealthCheckConfig))
	app.Use(logger.New(config.LoggerConfig))
	app.Use(limiter.New(config.LimiterConfig))
	app.Use(recover.New())

	app.Get("/monitor", monitor.New(config.MonitorConfig))
}

// loadEnv loads environment variables from a .env file.
// If the file cannot be loaded, the function panics with an error message.
func loadEnv() {
	err := env.Load()
	if err != nil {
		panic("Error loading .env file")
	}
}
