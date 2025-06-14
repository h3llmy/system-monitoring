package config

import "github.com/gofiber/fiber/v2/middleware/logger"

var (
	LoggerConfig logger.Config = logger.Config{
		Format:        "${pid} ${status} - ${method} ${path}\n",
		TimeFormat:    "02-Jan-2006",
		DisableColors: false,
	}
)
