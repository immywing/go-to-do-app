package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"
	"testing"

	"go-to-do-app/to-do-lib/datastores"
	"go-to-do-app/to-do-lib/models"
	"go-to-do-app/to-do-server/server"

	"github.com/google/uuid"
)

func TestConcurrentPutRequests(t *testing.T) {
	stores := []datastores.DataStore{
		datastores.NewInMemDataStore(),
		datastores.NewJsonDatastore("store.json"),
	}
	items := make([]models.ToDo, 0)
	statuses := []bool{true, false}
	priorities := []string{models.PriorityLow, models.PriorityMedium, models.PriorityHigh}
	for i := 0; i < 10; i++ {
		items = append(items, models.ToDo{Id: uuid.New(), Title: "test", Priority: "High", Complete: false, UserId: uuid.New().String()})
	}
	items[0].Validate("v1")
	for _, datastore := range stores {
		for _, item := range items {
			datastore.AddItem(item)
		}
		shutdownChan := make(chan bool)
		srv := server.NewToDoServer(":8081", shutdownChan, datastore)
		go srv.Start()

		var wg sync.WaitGroup
		numRequests := 100000
		for i := 1; i < numRequests; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				var actual models.ToDo
				expected := items[rand.IntN(len(items))]
				expected.Priority = priorities[rand.IntN(len(statuses))]
				expected.Complete = statuses[rand.IntN(len(statuses))]
				body, err := json.Marshal(expected)
				if err != nil {
					// error marshalling struct
				}
				client := http.Client{}
				req, err := http.NewRequest("PUT", "http://localhost:8081/v2/todo", bytes.NewBuffer(body))
				if err != nil {
					fmt.Println("Error performing PUT request:", err)
					return
				}
				req.Header.Set("Accept", "application/json")
				resp, err := client.Do(req)
				if err != nil {
					// fmt.Println("Error performing PUT request:", err)
					return
				}
				defer resp.Body.Close()
				respbody, err := io.ReadAll(resp.Body)
				if err != nil {
					//Todo
				}
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
		srv.Shutdown()
	}
}
