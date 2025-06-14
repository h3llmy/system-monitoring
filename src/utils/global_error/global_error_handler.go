package global_error

import (
	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/response"
)

func GlobalErrorHandler(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(response.BaseErrorResponse{
		Error:   true,
		Message: err.Error(),
	})
}
