package config

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

var (
	AppEnv            string       = os.Getenv("APP_ENV")
	IsProduction      bool         = AppEnv == "production"
	ApplicationConfig fiber.Config = fiber.Config{
		AppName:           "System Monitoring",
		EnablePrintRoutes: !IsProduction,
	}
)
