package service

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/h3llmy/system-monitoring/src/response"
	httpClient "github.com/h3llmy/system-monitoring/src/utils/httpClient"
)

type JellyfinService struct {
	client *httpClient.Client
}

// NewJellyfinService creates a new JellyfinService instance.
//
// It sets the base URL to the environment variable JELLYFIN_BASE_URL and the
// X-Emby-Token header to the environment variable JELLYFIN_TOKEN.
//
// It returns a pointer to a JellyfinService instance.
func NewJellyfinService(client *httpClient.Client) *JellyfinService {
	jellyfinBaseUrl := os.Getenv("JELLYFIN_BASE_URL")

	headers := map[string]string{
		"X-Emby-Token": os.Getenv("JELLYFIN_TOKEN"),
	}

	client.SetBaseURL(jellyfinBaseUrl)
	client.SetHeaders(headers)

	return &JellyfinService{
		client: client,
	}
}

// GetLibrariesItemCount retrieves the current item count for movies, series and songs from the Jellyfin API.
//
// It returns a JSON response containing the item count for each media type. In case of an error, a JSON response with an
// error message is returned instead.
func (s *JellyfinService) GetLibrariesItemCount() (*response.LibrariesItemCountResponse, error) {
	res, err := s.client.Get("/Items/Counts")
	if err != nil {
		return nil, fmt.Errorf("failed to get Jellyfin count: %w", err)
	}

	var itemsCount response.LibrariesItemCountResponse
	if err = json.Unmarshal(res, &itemsCount); err != nil {
		return nil, fmt.Errorf("failed to read Jellyfin count: %w", err)
	}

	return &itemsCount, nil
}
