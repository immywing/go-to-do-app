package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-to-do-app/to-do-lib/logging"
	"go-to-do-app/to-do-lib/models"

	"github.com/google/uuid"
)

type APIClient struct {
	BaseURL    string
	httpClient *http.Client
}

func (c *APIClient) Get(ctx context.Context, id string, user_id string, version string) {
	uuid, err := uuid.Parse(id)
	if err != nil {
		logging.LogWithTrace(ctx, map[string]interface{}{"uuid": uuid}, err.Error())
	}
	url := fmt.Sprintf("%s%s/todo?id=%s&user_id=%s", c.BaseURL, version, id, user_id)
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
	body, _ := io.ReadAll(resp.Body)
	logData := map[string]interface{}{
		"method":       "GET",
		"url":          url,
		"statusCode":   resp.StatusCode,
		"responseBody": string(body),
		"error":        err,
	}
	logging.LogWithTrace(ctx, logData, "GET request from CLI")
}

func (c *APIClient) Send(ctx context.Context, item models.ToDo, method string, version string) {
	body, _ := json.Marshal(item)
	temp := "/todo"
	fmt.Println(fmt.Sprintf("%s%s%s", c.BaseURL, version, temp))
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s%s", c.BaseURL, version, temp), bytes.NewBuffer(body))
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
	body, _ = io.ReadAll(resp.Body)
	logData := map[string]interface{}{
		"method":       "POST",
		"url":          c.BaseURL,
		"statusCode":   resp.StatusCode,
		"responseBody": string(body),
	}
	logging.LogWithTrace(ctx, logData, "POST request from CLI")
}

func (c *APIClient) PingServer() (bool, error) {
	c.httpClient.Timeout = 2 * time.Second
	resp, err := c.httpClient.Get(c.BaseURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK, nil
}

func NewAPIClient(baseURL string) APIClient {
	return APIClient{BaseURL: baseURL, httpClient: &http.Client{}}
}
