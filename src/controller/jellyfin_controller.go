package controller

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/h3llmy/system-monitoring/src/response"
	"github.com/h3llmy/system-monitoring/src/service"
)

type JellyfinController struct {
	jfService service.JellyfinService
}

// NewJellyfinController returns a new JellyfinController instance.
//
// It accepts a service.JellyfinService dependency which is used to interact with the Jellyfin API.
func NewJellyfinController(jfService service.JellyfinService) *JellyfinController {
	return &JellyfinController{
		jfService: jfService,
	}
}

// GetJellyfinCount handles HTTP requests to retrieve the current item count for movies, series and songs from the Jellyfin API.
//
// It returns a JSON response containing the item count for each media type. In case of an error, a JSON response with an error
// message is returned instead.
func (controller *JellyfinController) GetJellyfinCount(c *fiber.Ctx) error {
	itemCount, err := controller.jfService.GetLibrariesItemCount()
	if err != nil {
		c.Status(http.StatusInternalServerError).JSON(response.BaseErrorResponse{
			Error:   true,
			Message: err.Error(),
		})
	}

	return c.JSON(itemCount)
}
