package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"to-do-app/logging"
	"to-do-app/models"

	"github.com/google/uuid"
)

type APIClient struct {
	BaseURL    string
	httpClient *http.Client
}

func (c *APIClient) Get(ctx context.Context, id string) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		logging.LogWithTrace(ctx, map[string]interface{}{"uuid": uuid}, err.Error())
	}
	url := fmt.Sprintf("%s?id=%s", c.BaseURL, id)
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

func (c *APIClient) Send(ctx context.Context, item models.ToDo, method string) {
	body, err := json.Marshal(item)
	if err != nil {
		//
	}
	req, err := http.NewRequest(method, c.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		logging.LogWithTrace(
			ctx, map[string]interface{}{"header": req.Header, "body": req.Body}, "Failed request")
		return
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		fmt.Println("Error performing POST request:", err)
		return
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		//
	}
	logData := map[string]interface{}{
		"method":       "POST",
		"url":          c.BaseURL,
		"statusCode":   resp.StatusCode,
		"responseBody": string(body),
	}
	logging.LogWithTrace(ctx, logData, fmt.Sprintf("POST request from CLI"))
}

func (c *APIClient) PingServer() (bool, error) {
	c.httpClient.Timeout = 2 * time.Second
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/%s", c.BaseURL, "swagger.yaml"))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func NewAPIClient(baseURL string) APIClient {
	return APIClient{BaseURL: baseURL, httpClient: &http.Client{}}
}
