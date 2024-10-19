package server

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

	"github.com/immywing/go-to-do-app/to-do-lib"

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

	http.HandleFunc("/search", handleFormGet)
	http.HandleFunc("/update", handleFormPut)
	http.HandleFunc("/add", handleFormPost)
	http.HandleFunc("/item", handleItem)
}

func handleFormGet(w http.ResponseWriter, r *http.Request) {
	serveForm(w, "GET")
}

func handleFormPut(w http.ResponseWriter, r *http.Request) {
	serveForm(w, "PUT")
}

func handleFormPost(w http.ResponseWriter, r *http.Request) {
	serveForm(w, "POST")
}

func serveForm(w http.ResponseWriter, method string) {
	tmpl, err := template.ParseFiles("./templates/todoform.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, method) // Render the form template without data
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func handleItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleWebGet(w, r)
	case http.MethodPost, http.MethodPut:
		handleWebPostPut(w, r, r.Method)
	}
}

func handleWebPostPut(w http.ResponseWriter, r *http.Request, m string) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	// fmt.Println(r.FormValue("api_version"))
	apiVer := r.FormValue("api_version")
	userID := r.FormValue("user_id")
	itemID := r.FormValue("id")
	title := r.FormValue("title")
	priority := r.FormValue("priority")
	complete := r.FormValue("complete") == "true"
	fmt.Println(apiVer, userID, itemID, title, r.FormValue("complete"))
	itemIn, err := models.NewToDo(&userID, &itemID, &title, &priority, &complete)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json, err := json.Marshal(itemIn)
	if err != nil {
		http.Error(w, "failed to construct item json body", http.StatusInternalServerError)
		return
	}
	apiURL := fmt.Sprintf("http://localhost:8081/%s/todo?", apiVer)
	var resp *http.Response
	if m == http.MethodPost {
		resp, err = http.Post(apiURL, "application/json", bytes.NewBuffer(json))
	} else {
		var req *http.Request
		req, err = http.NewRequest(m, apiURL, bytes.NewBuffer(json))
		c := http.Client{}
		resp, err = c.Do(req)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}
	item, err := unpackToDoItem(body)
	if err != nil {
		http.Error(w, "failed unmarshamlling response", http.StatusInternalServerError)
		return
	}
	serveItem(w, item)
}

func unpackToDoItem(body []byte) (models.ToDo, error) {
	var item models.ToDo
	err := json.Unmarshal(body, &item)
	if err != nil {
		return models.ToDo{}, nil
	}
	return item, nil
}

func serveItem(w http.ResponseWriter, item models.ToDo) {
	tmpl, err := template.ParseFiles("./templates/todoitem.html")
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	fmt.Println(item)
	err = tmpl.Execute(w, item)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

func handleWebGet(w http.ResponseWriter, r *http.Request) {
	apiVer := r.URL.Query().Get("api_version")
	userID := r.URL.Query().Get("user_id")
	itemID := r.URL.Query().Get("id")
	apiURL := fmt.Sprintf("http://localhost:8081/%s/todo?user_id=%s&id=%s", apiVer, userID, itemID)
	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Error calling internal API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}
	item, err := unpackToDoItem(body)
	if err != nil {
		http.Error(w, "failed unmarshamlling response", http.StatusInternalServerError)
		return
	}
	serveItem(w, item)
	// // Render the template with the API response as the result
	// tmpl, err := template.ParseFiles("./templates/todoitem.html")
	// if err != nil {
	// 	http.Error(w, "Error parsing template", http.StatusInternalServerError)
	// 	return
	// }
	// fmt.Println(item)
	// err = tmpl.Execute(w, item)
	// if err != nil {
	// 	http.Error(w, "Error rendering template", http.StatusInternalServerError)
	// }
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		fmt.Printf("Server Shutdown error: %v\n", err)
	} else {
		fmt.Println("Server shut down gracefully")
	}
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

func PostputToDo(w http.ResponseWriter, r *http.Request, f func(item models.ToDo) (models.ToDo, error)) {
	w.Header().Set("Content-Type", "application/json")
	// defer r.Body.Close()
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
	err = item.Validate(pathparts[1]) //r.Url.Path describes: path (relative paths may omit leading slash), should consider if this needs handling to avoid index out of bounds
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
	writeJSONResponse(w, r, http.StatusCreated, resp)
	// versionItem(w, r, item)
	// pathparts := strings.Split(r.URL.Path, "/")
	// fmt.Println()
	// versionPath := strings.Split(r.URL.Path, "/")[0]
	// ver := int(versionPath[1] - '0')
	// if err = item.Validate(ver); err != nil {
	// 	//
	// }
	// if ver, exists := versions[strings.Split(r.URL.Path, "/")[0]]; !exists {
	// 	writeErrorResponse(w, r, http.StatusInternalServerError, "Failed to parse API version")
	// } else {
	// 	item.Validate(ver)
	// }

	// fmt.Println(r.URL.Path, pathParts)
	// postputResultChan := make(chan models.ToDo)
	// postputErrorChan := make(chan error)
	// go func() {
	// 	item, err := f(item)
	// 	if err != nil {
	// 		postputErrorChan <- err
	// 		return
	// 	}
	// 	postputResultChan <- item
	// }()
	// select {
	// case item := <-postputResultChan:
	// 	resp, err := json.Marshal(item)
	// 	if err != nil {
	// 		writeErrorResponse(w, r, http.StatusInternalServerError, "Internal Server Error")
	// 	}
	// 	writeJSONResponse(w, r, http.StatusCreated, resp)
	// case err := <-postputErrorChan:
	// 	handleDataStoreError(w, r, err)
	// case <-time.After(time.Second * 10):
	// 	writeErrorResponse(w, r, http.StatusGatewayTimeout, "Request timed out")
	// }

}

func getToDo(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userId := r.URL.Query().Get("user_id")
	ver := strings.Split(r.URL.Path, "/")[1]
	if id == "" {
		writeErrorResponse(w, r, http.StatusBadRequest, "missing 'id' query paramater")
	}
	if userId == "" && ver == "v2" {
		writeErrorResponse(w, r, http.StatusBadRequest, "missing 'user_id' query paramater")
	}
	uuid, err := uuid.Parse(id)
	if err != nil {
		writeErrorResponse(w, r, http.StatusBadGateway, fmt.Sprintf("error parsing uuid: %s", err.Error()))
	}
	item, err := datastore.GetItem(userId, uuid)
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
