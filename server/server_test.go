package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"to-do-app/models"

	"github.com/google/uuid"
)

func TestConcurrentPutRequests(t *testing.T) {
	datastore := models.NewInMemDataStore()
	uuid := uuid.New()
	title := "test item"
	priority := "High"
	itemA := models.ToDo{Id: uuid, Title: title, Priority: priority, Complete: false}
	itemB := models.ToDo{Id: uuid, Title: title, Priority: priority, Complete: true}
	bodyA, err := json.Marshal(itemA)
	if err != nil {
		t.Errorf("Failed to marshal item: %v", err)
		return
	}
	bodyB, err := json.Marshal(itemB)
	if err != nil {
		t.Errorf("Failed to marshal item: %v", err)
		return
	}
	datastore.AddItem(itemA)

	var wg sync.WaitGroup
	numRequests := 1000
	var actual models.ToDo
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var expected models.ToDo
			var body []byte
			if i%2 == 0 {
				expected = itemA
				body = bodyA
			} else {
				expected = itemB
				body = bodyB
			}
			client := http.Client{}
			req, err := http.NewRequest("POST", "http://localhost:8081/v1/todo", bytes.NewBuffer(body))
			if err != nil {
				return
			}
			req.Header.Set("Accept", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Error performing POST request:", err)
				return
			}
			defer resp.Body.Close()
			respbody, err := io.ReadAll(resp.Body)
			err = json.Unmarshal(respbody, &actual)
			if err != nil {
				t.Errorf("Error unmarshalling response from server")
			}
			if actual != expected {
				t.Errorf("Expected : %+v, Got: %+v", expected, actual)
			}
		}(i)
	}
	wg.Wait()
}
