package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go-to-do-app/to-do-lib/models"
)

type APIClient struct {
	BaseURL    string
	httpClient *http.Client
}

func (c *APIClient) Req(
	ctx context.Context,
	m string,
	initem models.ToDo, args map[string]string) (models.ToDo, error) {

	var apiURL string
	var req *http.Request
	var itemIn models.ToDo
	var buffer []byte
	var err error
	userid := args["user-id"]
	itemid := args["id"]
	version := args["version"]
	title := args["title"]
	priority := args["priority"]
	complete := args["complete"] == "true"
	fmt.Println(complete, m)
	if m == http.MethodGet {
		apiURL = fmt.Sprintf("http://localhost:8081/%s/todo?user_id=%s&id=%s",
			version, userid, itemid)
	}
	if m == http.MethodPut || m == http.MethodPost {
		apiURL = fmt.Sprintf("http://localhost:8081/%s/todo", args["version"])
		itemIn, err = models.NewToDo(&userid, &itemid, &title, &priority, &complete)
		fmt.Println(itemIn)
		if err != nil {
			return models.ToDo{}, err
		}
		buffer, err = json.Marshal(itemIn)
		if err != nil {
			return models.ToDo{}, err
		}
	}
	req, err = http.NewRequest(m, apiURL, bytes.NewBuffer(buffer))
	var item models.ToDo
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return models.ToDo{}, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return models.ToDo{}, err
	}
	return item, nil
}

func NewAPIClient(baseURL string) APIClient {
	return APIClient{BaseURL: baseURL, httpClient: &http.Client{}}
}
