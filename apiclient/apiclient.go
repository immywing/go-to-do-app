package apiclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"to-do-app/logging"
)

type APIClient struct {
	BaseURL    string
	httpClient *http.Client
}

func (c *APIClient) Get(ctx context.Context, path string) {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Error performing GET request:", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		//
	}
	logData := map[string]interface{}{
		"method":       "GET",
		"url":          url,
		"statusCode":   resp.StatusCode,
		"responseBody": string(body),
	}
	logging.LogWithTrace(ctx, logData, "GET request from CLI")
}

func NewAPIClient(baseURL string) APIClient {
	return APIClient{BaseURL: baseURL, httpClient: &http.Client{}}
}
