package utils

import (
	"fmt"
	"io"
	"net/http"
)

type HttpClient interface {
	SetHeaders(headers map[string]string)
	Get(url string) (*http.Response, error)
}

type httpClient struct{
	headers map[string]string
	client *http.Client
}

func NewHttpClient() HttpClient{
	return &httpClient{
		client: &http.Client{},
	}
}

func (c *httpClient) SetHeaders(headers map[string]string) {
	c.headers = headers
}

func (c *httpClient) addHeaders(req *http.Request) {
    for key, value := range c.headers {
        req.Header.Set(key, value)
    }
}

func (c *httpClient) Get(url string)(*http.Response, error){
	req, err := http.NewRequest(http.MethodGet,url, nil)
	if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

	c.addHeaders(req)
	
	return c.client.Do(req)
}

func ReadResponseBody(resp *http.Response) ([]byte, error){
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}