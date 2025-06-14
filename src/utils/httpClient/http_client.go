package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	Headers    map[string]string
}

// NewClient returns a new Client with the provided timeout set on the underlying HTTP client.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

// SetBaseURL sets the base URL that the client will use for making HTTP requests.
// All requests will have this URL prepended to the provided request URL.
func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}

// SetHeaders sets the headers that the client will use for making HTTP requests.
// All requests will have these headers set.
func (c *Client) SetHeaders(headers map[string]string) {
	c.Headers = headers
}

// Request makes an HTTP request to the provided URL with the provided method and body.
// The request is constructed by prepending the client's BaseURL to the provided URL.
// The request headers are set from the client's Headers map.
// The request body is constructed by marshaling the provided body to JSON.
// The response body is read and returned as a byte array.
// If an error occurs while making the request, it is returned instead.
func (c *Client) Request(method, url string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	fullURL := c.BaseURL + url
	req, err := http.NewRequest(method, fullURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	for key, value := range c.Headers {
		req.Header.Set(key, value)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

// Get makes a GET request to the provided URL and returns the response body as a byte array.
// If an error occurs while making the request, it is returned instead.
func (c *Client) Get(url string) ([]byte, error) {
	return c.Request(http.MethodGet, url, nil)
}

// Post makes a POST request to the provided URL with the given body and returns the response body as a byte array.
// The body is marshaled to JSON before being sent.
// If an error occurs while making the request, it is returned instead.
func (c *Client) Post(url string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPost, url, body)
}

// Put makes a PUT request to the provided URL with the given body and returns the response body as a byte array.
// The body is marshaled to JSON before being sent.
// If an error occurs while making the request, it is returned instead.
func (c *Client) Put(url string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPut, url, body)
}

// Delete makes a DELETE request to the provided URL and returns the response body as a byte array.
// If an error occurs while making the request, it is returned instead.
func (c *Client) Delete(url string) ([]byte, error) {
	return c.Request(http.MethodDelete, url, nil)
}
