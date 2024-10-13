package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"to-do-app/apiclient"
	"to-do-app/logging"
	"to-do-app/models"
	"to-do-app/server"
)

var (
	startServer = flag.Bool("start-server", false, "Start the To-Do server")
	mode        = flag.String("mode", "", "set the mode the application should run in (in-mem, json-store, pgdb)")
	post        = flag.Bool("post", false, "Add new Todo")
	put         = flag.Bool("put", false, "updateTodo")
	get         = flag.Bool("get", false, "Get existing Todo")
	id          = flag.String("id", "", "UUID of ToDo item")
	title       = flag.String("title", "", "Title of ToDo item")
	priority    = flag.String("priority", "", "Priority of ToDo item")
	complete    = flag.Bool("complete", false, "Completion status of ToDo item")
)

func run() {
	flag.Parse()
	todoflags := map[string]interface{}{"id": *id, "title": *title, "priority": *priority, "complete": *complete}
	if *startServer {
		var store models.DataStore
		if *mode == "pgdb" || *mode == "json-store" {
			fmt.Fprintf(os.Stderr, "Error: the mode '%s' is not yet implemented\n", *mode)
			os.Exit(1)
		} else if *mode == "in-mem" {
			store = models.NewInMemDataStore()
		} else {
			logging.LogWithTrace(
				context.Background(),
				map[string]interface{}{},
				"no valid mode provided to start server with datastore",
			)
			os.Exit(1)
		}
		server.Start(&store)
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
