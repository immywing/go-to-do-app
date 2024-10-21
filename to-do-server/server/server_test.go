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
	statuses := []bool{true, false}
	priorities := []string{models.PriorityLow, models.PriorityMedium, models.PriorityHigh}
	itmev1 := models.ToDo{Id: uuid.Max, Title: "test", Priority: "High", Complete: false, UserId: ""}
	itemv2 := models.ToDo{Id: uuid.Max, Title: "test", Priority: "High", Complete: false, UserId: fmt.Sprintf("TestToDoUser")}
	versions := make(map[string]models.ToDo)
	for _, datastore := range stores {
		expectedV1, _ := datastore.AddItem(itmev1)
		expectedV2, _ := datastore.AddItem(itemv2)
		versions["v1"] = expectedV1
		versions["v2"] = expectedV2
		shutdownChan := make(chan bool)
		srv := server.NewToDoServer(":8081", shutdownChan, datastore)
		go srv.Start()

		var wg sync.WaitGroup
		numRequests := 50000
		for version, expectedItem := range versions {
			for i := 1; i < numRequests; i++ {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					var actual models.ToDo
					expected := expectedItem
					expected.Priority = priorities[rand.IntN(len(statuses))]
					expected.Complete = statuses[rand.IntN(len(statuses))]
					body, err := json.Marshal(expected)
					if err != nil {
						// error marshalling struct
					}
					client := http.Client{}
					endpoint := fmt.Sprintf("http://localhost:8081/%s/todo", version)
					req, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(body))
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
		}

		wg.Wait()
		srv.Shutdown()
	}
}
