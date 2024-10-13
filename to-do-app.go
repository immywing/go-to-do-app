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
	todoerrors "to-do-app/errors"
	inmem "to-do-app/inmem"
	"to-do-app/logging"
	"to-do-app/models"

	"github.com/google/uuid"
)

var mode = flag.String("mode", "", "set the mode the application should run in (in-mem, json-store, pgdb)")

var datastore models.DataStore

func writeJSONResponse(w http.ResponseWriter, statusCode int, data []byte) {
	ctx := logging.AddTraceID(context.Background())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
	logData := map[string]interface{}{
		"statusCode":   statusCode,
		"responseBody": data,
	}
	logging.LogWithTrace(ctx, logData, "Json response Written")
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, []byte(fmt.Sprintf(`{"error": %s}`, message)))
}

func handleDataStoreError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *todoerrors.NotFoundError:
		writeErrorResponse(w, http.StatusNotFound, e.Message)
	case *todoerrors.ValidationError:
		writeErrorResponse(w, http.StatusBadRequest, e.Error())
	default:
		writeErrorResponse(w, http.StatusInternalServerError, "Internal server error")
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
		writeErrorResponse(w, http.StatusBadRequest, "Expected Json in request body")
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println(err, body)
		writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON format: %s", err.Error()))
		return
	}
	item, err = f(item)
	if err != nil {
		handleDataStoreError(w, err)
		return
	}
	resp, err := json.Marshal(item)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
	}
	writeJSONResponse(w, http.StatusCreated, resp)
}

func getToDo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErrorResponse(w, http.StatusBadRequest, "missing 'id' query paramater")
	}
	uuid, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, http.StatusBadGateway, fmt.Sprintf("error parsing uuid: %s", err.Error()))
	}
	item, err := datastore.GetItem(uuid)
	if err != nil {
		handleDataStoreError(w, err)
	}
	resp, err := json.Marshal(item)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
	}
	writeJSONResponse(w, http.StatusCreated, resp)
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
	if *mode == "" {
		fmt.Fprintln(os.Stderr, "Error: application must be provided one of the following modes using --mode=<in-mem|json-store|pgdb>")
		flag.Usage()
		os.Exit(1)
	}
	if *mode == "pgdb" || *mode == "json-store" {
		fmt.Fprintf(os.Stderr, "Error: the mode '%s' is not yet implemented\n", *mode)
		os.Exit(1)
	}
	if *mode == "in-mem" {
		inmem.Run(&datastore)
	}
}

func main() {
	go run()
	http.HandleFunc("/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "to-do-app-api-v1.yaml")
	})
	http.HandleFunc("/swagger-ui", swaggerUI)
	http.HandleFunc("/v1/todo", ToDoHandler)
	http.ListenAndServe(":8081", nil)
}
