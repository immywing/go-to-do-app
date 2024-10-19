package wiring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"go-to-do-app/to-do-lib/datastores"
	todoerrors "go-to-do-app/to-do-lib/errors"
	"go-to-do-app/to-do-lib/logging"
	"go-to-do-app/to-do-lib/models"

	"github.com/google/uuid"
)

var (
	datastore datastores.DataStore
)

func WireEndpoints() {
	http.HandleFunc("/styles.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./templates/styles.css")
	})
	http.HandleFunc("/v1/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api-specs/to-do-app-api-v1.yaml")
	})
	http.HandleFunc("/v2/swagger.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./api-specs/to-do-app-api-v2.yaml")
	})
	http.HandleFunc("/v1/swagger-ui", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("./templates/swagger-ui-template.html")
		data := "v1"
		tmpl.Execute(w, data)
	})
	http.HandleFunc("/v2/swagger-ui", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("./templates/swagger-ui-template.html")
		data := "v2"
		tmpl.Execute(w, data)
	})
	http.HandleFunc("/v1/todo", ToDoHandler)
	http.HandleFunc("/v2/todo", ToDoHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("./templates/home.html")
		tmpl.Execute(w, nil)
	})
	http.HandleFunc("/search", handleFormGet)
	http.HandleFunc("/update", handleFormPut)
	http.HandleFunc("/add", handleFormPost)
	http.HandleFunc("/item", handleWebForm)
}

func Start(store *datastores.DataStore, shutdownChan chan bool) {
	datastore = *store
	srv := &http.Server{
		Addr: ":8081",
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("ListenAndServe error: %v\n", err)
		}
	}()
	<-shutdownChan
	datastore.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown error: %v\n", err)
	} else {
		fmt.Println("Server shut down gracefully")
	}
	shutdownChan <- true
}

func staticTemplate(w http.ResponseWriter, data interface{}, path string) {
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

func serveItem(w http.ResponseWriter, item models.ToDo) {
	staticTemplate(w, item, "./templates/todoitem.html")
}

func itemNotFound(w http.ResponseWriter, message string) {
	staticTemplate(w, message, "./templates/notfound.html")
}

func handleFormGet(w http.ResponseWriter, r *http.Request) {
	staticTemplate(w, "GET", "./templates/todoform.html")
}

func handleFormPut(w http.ResponseWriter, r *http.Request) {
	staticTemplate(w, "PUT", "./templates/todoform.html")
}

func handleFormPost(w http.ResponseWriter, r *http.Request) {
	staticTemplate(w, "POST", "./templates/todoform.html")
}

func handleWebForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	apiVer := r.FormValue("api_version")
	userID := r.FormValue("user_id")
	itemID := r.FormValue("id")
	title := r.FormValue("title")
	priority := r.FormValue("priority")
	complete := r.FormValue("complete") == "true"
	var apiURL string
	var req *http.Request
	var itemIn models.ToDo
	var buffer []byte
	var err error
	if r.Method == http.MethodGet {
		apiURL = fmt.Sprintf("http://localhost:8081/%s/todo?user_id=%s&id=%s", apiVer, userID, itemID)
	}
	fmt.Println(userID, itemID, title, priority, complete)
	if r.Method == http.MethodPut || r.Method == http.MethodPost {
		apiURL = fmt.Sprintf("http://localhost:8081/%s/todo", apiVer)
		itemIn, err = models.NewToDo(&userID, &itemID, &title, &priority, &complete)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		buffer, err = json.Marshal(itemIn)
		if err != nil {
			http.Error(w, "failed to construct item json body", http.StatusInternalServerError)
			return
		}
	}
	req, err = http.NewRequest(r.Method, apiURL, bytes.NewBuffer(buffer))
	c := http.Client{}
	var item models.ToDo
	resp, err := c.Do(req)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &item)
	if err != nil {
		http.Error(w, "failed unmarshamlling response", http.StatusInternalServerError)
		return
	}
	serveItem(w, item)
}

func writeJSONResponse(w http.ResponseWriter, r *http.Request, statusCode int, data []byte) {
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
	writeJSONResponse(w, r, statusCode, []byte(fmt.Sprintf(`{"error": "%s"}`, message)))
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
	pathparts := strings.Split(r.URL.Path, "/")
	err = item.Validate(pathparts[1])
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadRequest, fmt.Sprintf("Invalid body: %s", err.Error()))
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
	if r.Method == http.MethodPost {
		writeJSONResponse(w, r, http.StatusCreated, resp)
	} else {
		writeJSONResponse(w, r, http.StatusOK, resp)
	}
}

func getToDo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userId := r.URL.Query().Get("user_id")
	ver := strings.Split(r.URL.Path, "/")[1]
	if id == "" {
		writeErrorResponse(w, r, http.StatusBadRequest, "missing 'id' query paramater")
		return
	}
	if userId == "" && ver == "v2" {
		writeErrorResponse(w, r, http.StatusBadRequest, "missing 'user_id' query paramater")
		return
	}
	uuid, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadGateway, fmt.Sprintf("error parsing uuid: %s", err.Error()))
		return
	}
	item, err := datastore.GetItem(userId, uuid)
	if err != nil {
		handleDataStoreError(w, r, err)
		return
	}
	resp, err := json.Marshal(item)
	if err != nil {
		writeErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	writeJSONResponse(w, r, http.StatusOK, resp)
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
