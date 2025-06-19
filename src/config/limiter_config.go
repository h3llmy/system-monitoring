package config

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/h3llmy/system-monitoring/src/response"
)

var (
	LimiterConfig limiter.Config = limiter.Config{
		Max:        500, // Maximum of 500 requests per minute
		Expiration: time.Minute,
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(response.BaseErrorResponse{
				Error:   true,
				Message: "Rate limit exceeded",
			})
		},
	}
)
