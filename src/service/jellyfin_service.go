package service

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/h3llmy/system-monitoring/src/response"
	"github.com/h3llmy/system-monitoring/src/utils"
)

type JellyfinService struct {
	client utils.HttpClient
	baseUrl string
}

// NewJellyfinService initializes a new JellyfinService with dependency injection.
//
// It expects environment variables JELLYFIN_TOKEN and JELLYFIN_BASE_URL to be set.
func NewJellyfinService(client utils.HttpClient) *JellyfinService{
	token := os.Getenv("JELLYFIN_TOKEN")
	baseUrl := os.Getenv("JELLYFIN_BASE_URL")
	headers := map[string]string{"X-Emby-Token": token}
	client.SetHeaders(headers)
	return &JellyfinService{
		client: client,
		baseUrl: baseUrl,
	}
}


// GetLibrariesItemCount retrieves the current item count for movies, series and songs.
//
// The response is unmarshalled into a response.LibrariesItemCountResponse struct.
func (s *JellyfinService) GetLibrariesItemCount() (*response.LibrariesItemCountResponse, error){
	url := fmt.Sprintf("%s/Items/Counts", s.baseUrl)

	resp, err := s.client.Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)	
	}

	defer resp.Body.Close()

	
	body, err := utils.ReadResponseBody(resp)	
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)	
	}

	var itemCount response.LibrariesItemCountResponse
	err = json.Unmarshal(body, &itemCount); if err != nil {
        return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
    }

	return &itemCount, nil
}