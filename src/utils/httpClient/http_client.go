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

func NewClient(timeout time.Duration) *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) SetBaseURL(baseURL string) {
	c.BaseURL = baseURL
}

func (c *Client) SetHeaders(headers map[string]string) {
	c.Headers = headers
}

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

func (c *Client) Get(url string) ([]byte, error) {
	return c.Request(http.MethodGet, url, nil)
}

func (c *Client) Post(url string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPost, url, body)
}

func (c *Client) Put(url string, body interface{}) ([]byte, error) {
	return c.Request(http.MethodPut, url, body)
}

func (c *Client) Delete(url string) ([]byte, error) {
	return c.Request(http.MethodDelete, url, nil)
}
