package main

import (
	"context"
	"flag"

	"go-to-do-app/to-do-lib/apiclient"
	"go-to-do-app/to-do-lib/logging"
	"go-to-do-app/to-do-lib/models"
)

var (
	post     = flag.Bool("post", false, "Add new Todo")
	put      = flag.Bool("put", false, "updateTodo")
	get      = flag.Bool("get", false, "Get existing Todo")
	id       = flag.String("id", "", "UUID of ToDo item")
	userId   = flag.String("user-id", "", "UUID representing user id")
	title    = flag.String("title", "", "Title of ToDo item")
	priority = flag.String("priority", "", "Priority of ToDo item")
	complete = flag.Bool("complete", false, "Completion status of ToDo item")
)

func cliParse() {
	flag.Parse()
	todoflags := map[string]interface{}{"user-id": *userId, "id": *id, "title": *title, "priority": *priority, "complete": *complete}
	var item models.ToDo
	ctx := logging.AddTraceID(context.Background())
	client := apiclient.NewAPIClient("http://localhost:8081/")
	if serverup, err := client.PingServer(); !serverup || err != nil {
		logging.LogWithTrace(ctx, todoflags, "failed to ping server. Use --start-server to run.")
	}
	if *post || *put {
		var err error
		item, err = models.NewToDo(userId, id, title, priority, complete)
		if err != nil {
			logging.LogWithTrace(ctx, todoflags, err.Error())
		}
	}
	if *post {
		client.Send(ctx, item, "POST", "v1")
	}
	if *put {
		client.Send(ctx, item, "POST", "v1")
	}
	if *get {
		client.Get(ctx, *id)
	}
}

func main() {
	cliParse()
}
