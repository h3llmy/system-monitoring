package config

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/h3llmy/system-monitoring/src/response"
)

var (
	KeyAuthConfig = keyauth.Config{
		KeyLookup:    "header:X-API-Key",
		ErrorHandler: errorHandler,
		Validator:    authenticate,
	}
)

func errorHandler(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusUnauthorized).JSON(response.BaseErrorResponse{
		Error:   true,
		Message: "Unauthorized",
	})
}

func authenticate(ctx *fiber.Ctx, key string) (bool, error) {
	authenticated := os.Getenv("API_KEY") == key
	if !authenticated {
		return false, nil
	}
	return true, nil
}
