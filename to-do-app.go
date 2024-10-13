package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/template"
	"to-do-app/apiclient"
	todoerrors "to-do-app/errors"
	"to-do-app/logging"
	"to-do-app/models"

	"github.com/google/uuid"
)

var startServer = flag.Bool("start-server", false, "Start the To-Do server")
var mode = flag.String("mode", "", "set the mode the application should run in (in-mem, json-store, pgdb)")
var post = flag.Bool("post", false, "Add new Todo")
var put = flag.Bool("put", false, "updateTodo")
var get = flag.Bool("get", false, "Get existing Todo")
var id = flag.String("id", "", "UUID of ToDo item")
var title = flag.String("title", "", "Title of ToDo item")
var priority = flag.String("priority", "", "Priority of ToDo item")
var complete = flag.Bool("complete", false, "Completion status of ToDo item")
var datastore models.DataStore

func writeJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data []byte) {
	// ctx := logging.AddTraceID(context.Background())
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

func postputToDo(w http.ResponseWriter, r *http.Request, f func(item models.ToDo) (models.ToDo, error)) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	item := models.ToDo{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, "Expected Json in request body")
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println(err, body)
		writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format: %s", err.Error()))
		return
	}
	item, err = f(item)
	if err != nil {
		handleDataStoreError(w, r, err)
		return
	}
	resp, err := json.Marshal(item)
	if err != nil {
		writeErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
	}
	writeJSONResponse(w, r, http.StatusCreated, resp)
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
		postputToDo(w, r, datastore.AddItem)
	case http.MethodPut:
		postputToDo(w, r, datastore.UpdateItem)
	}
}

func run() {
	flag.Parse()
	todoflags := map[string]interface{}{"id": *id, "title": *title, "priority": *priority, "complete": *complete}
	if *mode == "pgdb" || *mode == "json-store" {
		fmt.Fprintf(os.Stderr, "Error: the mode '%s' is not yet implemented\n", *mode)
		os.Exit(1)
	}
	if *mode == "in-mem" {
		datastore = models.NewInMemDataStore()
	}
	if *startServer {
		http.HandleFunc("/v1/todo/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "to-do-app-api-v1.yaml")
		})
		http.HandleFunc("/swagger-ui", swaggerUI)
		http.HandleFunc("/v1/todo", ToDoHandler)
		http.ListenAndServe(":8081", nil)
	}
	var item models.ToDo
	ctx := logging.AddTraceID(context.Background())
	client := apiclient.NewAPIClient("http://localhost:8081/v1/todo")
	if serverup, err := client.PingServer(); !serverup || err != nil {
		logging.LogWithTrace(ctx, todoflags, "failed to ping server. Use --start-server to run.")
	}
	if *post || *put {
		var err error
		item, err = models.ToDoFromCLI(id, title, priority, complete)
		if err != nil {
			logging.LogWithTrace(ctx, todoflags, err.Error())
		}
	}
	if *post {
		client.Send(ctx, item, "POST")
	}
	if *put {
		client.Send(ctx, item, "POST")
	}
	if *get {
		client.Get(ctx, *id)
	}
}

func main() {
	run()
}
