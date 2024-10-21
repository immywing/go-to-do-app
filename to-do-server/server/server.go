package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"go-to-do-app/to-do-lib/apiclient"
	"go-to-do-app/to-do-lib/datastores"
	todoerrors "go-to-do-app/to-do-lib/errors"
	"go-to-do-app/to-do-lib/logging"
	"go-to-do-app/to-do-lib/models"

	"github.com/google/uuid"
)

type ToDoServer struct {
	server       *http.Server
	shutdownChan chan bool
}

func NewToDoServer(address string, shutdownChannel chan bool, datastore datastores.DataStore) ToDoServer {
	return ToDoServer{
		server:       &http.Server{Addr: address, Handler: wiredMux(datastore)},
		shutdownChan: shutdownChannel,
	}
}

func (s *ToDoServer) Shutdown() {
	s.shutdownChan <- true
}

func (s *ToDoServer) AwaitShutdown() {
	<-s.shutdownChan
}

func wiredMux(datastore datastores.DataStore) *http.ServeMux {
	routes := map[string]http.HandlerFunc{
		"/":                serveTemplate("./templates/home.html", nil),
		"/styles.css":      serveFile("./templates/styles.css"),
		"/v1/swagger.yaml": serveFile("./api-specs/to-do-app-api-v1.yaml"),
		"/v2/swagger.yaml": serveFile("./api-specs/to-do-app-api-v2.yaml"),
		"/v1/swagger-ui":   serveTemplate("./templates/swagger-ui-template.html", "v1"),
		"/v2/swagger-ui":   serveTemplate("./templates/swagger-ui-template.html", "v2"),
		"/v1/todo":         toDoHTTPHandler(datastore),
		"/v2/todo":         toDoHTTPHandler(datastore),
		"/search":          serveTemplate("./templates/todoform.html", "GET"),
		"/update":          serveTemplate("./templates/todoform.html", "PUT"),
		"/add":             serveTemplate("./templates/todoform.html", "POST"),
		"/item":            handleWebForm,
	}

	mux := http.NewServeMux()
	for route, handler := range routes {
		mux.HandleFunc(route, handler)
	}
	return mux
}

func (s *ToDoServer) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
	}()
	<-s.shutdownChan
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown error: %v\n", err)
	} else {
		fmt.Println("Server shut down gracefully")
	}
	s.shutdownChan <- true
}

func serveFile(filePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filePath)
	}
}

func toDoHTTPHandler(datastore datastores.DataStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		toDoHandler(datastore, w, r)
	}
}

func serveTemplate(path string, data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

func handleWebForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	method := r.FormValue("form_method")
	fmt.Println(method)
	args := map[string]string{
		"user-id":  r.FormValue("user_id"),
		"id":       r.FormValue("id"),
		"version":  r.FormValue("api_version"),
		"title":    r.FormValue("title"),
		"priority": r.FormValue("priority"),
		"complete": r.FormValue("complete"),
	}
	var itemIn models.ToDo
	ctx := logging.AddTraceID(r.Context())
	client := apiclient.NewAPIClient("http://localhost:8081/")
	if item, err := client.Req(ctx, method, itemIn, args); err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
	} else {
		temp := serveTemplate("./templates/todoitem.html", item)
		temp(w, r)
	}
}

func WriteJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data []byte) {
	ctx := logging.AddTraceID(r.Context())
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
	WriteJSONResponse(w, r, statusCode, []byte(fmt.Sprintf(`{"error": "%s"}`, message)))
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

func PostputToDo(w http.ResponseWriter, r *http.Request, f func(item models.ToDo) (models.ToDo, error)) {
	w.Header().Set("Content-Type", "application/json")
	defer r.Body.Close()
	var item models.ToDo
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, err.Error())
		return
	}
	pathparts := strings.Split(r.URL.Path, "/")
	err := item.Validate(pathparts[1])
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid body: %s", err.Error()))
		return
	}
	item, err = f(item)
	if err != nil {
		handleDataStoreError(w, r, err)
		return
	}
	MarshalAndWrite(w, r, item)
}

func getToDo(datastore datastores.DataStore, w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userId := r.URL.Query().Get("user_id")
	ver := strings.Split(r.URL.Path, "/")[1]
	uuid, err := uuid.Parse(id)
	if id == "" || (userId == "" && ver == "v2") || err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, "missing 'id' query paramater")
		return
	}
	var item models.ToDo
	if item, err = datastore.GetItem(userId, uuid); err != nil {
		handleDataStoreError(w, r, err)
		return
	}
	MarshalAndWrite(w, r, item)
}

func MarshalAndWrite(w http.ResponseWriter, r *http.Request, b interface{}) {
	resp, err := json.Marshal(b)
	if err != nil {
		writeErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	WriteJSONResponse(w, r, http.StatusOK, resp)
}

func toDoHandler(datastore datastores.DataStore, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getToDo(datastore, w, r)
	case http.MethodPost:
		PostputToDo(w, r, datastore.AddItem)
	case http.MethodPut:
		PostputToDo(w, r, datastore.UpdateItem)
	}
}
