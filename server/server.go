package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	datastores "to-do-app/datastores"
	todoerrors "to-do-app/errors"
	"to-do-app/logging"
	"to-do-app/models"

	"github.com/google/uuid"
)

var (
	datastore datastores.DataStore
	// postputResultChan = make(chan models.ToDo)
	// postputErrorChan  = make(chan error)
)

func Start(store *datastores.DataStore) {
	datastore = *store
	http.HandleFunc("/v1/todo/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "to-do-app-api-v1.yaml")
	})
	http.HandleFunc("/swagger-ui", swaggerUI)
	http.HandleFunc("/v1/todo", ToDoHandler)
	http.ListenAndServe(":8081", nil)
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data []byte) {
	ctx := r.Context()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
	logData := map[string]interface{}{
		"statusCode":   statusCode,
		"responseBody": data,
	}
	logging.LogWithTrace(ctx, logData, "Json response Written")
}

func writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	writeJSONResponse(w, r, statusCode, []byte(fmt.Sprintf(`{"error": %s}`, message)))
}

func handleDataStoreError(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case *todoerrors.NotFoundError:
		writeErrorResponse(w, r, http.StatusNotFound, e.Message)
	case *todoerrors.ValidationError:
		writeErrorResponse(w, r, http.StatusBadRequest, e.Error())
	default:
		writeErrorResponse(w, r, http.StatusInternalServerError, "Internal server error")
	}
}

func swaggerUI(w http.ResponseWriter, r *http.Request) {
	tmpl, _ := template.ParseFiles("swagger-ui-template.html")
	tmpl.Execute(w, nil)
}

func PostputToDo(w http.ResponseWriter, r *http.Request, f func(item models.ToDo) (models.ToDo, error)) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	var item models.ToDo
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, "Expected Json in request body")
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format: %s", err.Error()))
		return
	}
	postputResultChan := make(chan models.ToDo)
	postputErrorChan := make(chan error)
	go func() {
		item, err := f(item)
		if err != nil {
			postputErrorChan <- err
			return
		}
		postputResultChan <- item
	}()
	select {
	case item := <-postputResultChan:
		resp, err := json.Marshal(item)
		if err != nil {
			writeErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
		}
		writeJSONResponse(w, r, http.StatusCreated, resp)
	case err := <-postputErrorChan:
		handleDataStoreError(w, r, err)
		// case <-time.After(time.Second * 10):
		// 	writeErrorResponse(w, r, http.StatusGatewayTimeout, "Request timed out")
	}
}

func getToDo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErrorResponse(w, r, http.StatusBadRequest, "missing 'id' query paramater")
	}
	uuid, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadGateway, fmt.Sprintf("error parsing uuid: %s", err.Error()))
	}
	item, err := datastore.GetItem(uuid)
	if err != nil {
		handleDataStoreError(w, r, err)
	}
	resp, err := json.Marshal(item)
	if err != nil {
		writeErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
	}
	writeJSONResponse(w, r, http.StatusCreated, resp)
}

func ToDoHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getToDo(w, r)
	case http.MethodPost:
		PostputToDo(w, r, datastore.AddItem)
	case http.MethodPut:
		PostputToDo(w, r, datastore.UpdateItem)
	}
}
