package main

import (
	"context"
	"flag"
	"fmt"
	"strconv"

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
	version  = flag.String("version", "", "version of the api to use")
)

func cliParse() {
	flag.Parse()
	fmt.Println(*post)
	todoflags := map[string]string{
		"user-id":  *userId,
		"id":       *id,
		"title":    *title,
		"priority": *priority,
		"complete": strconv.FormatBool(*complete),
		"version":  *version,
	}
	fmt.Println(todoflags)
	var item models.ToDo
	ctx := logging.AddTraceID(context.Background())
	client := apiclient.NewAPIClient("http://localhost:8081/")
	// if serverup, err := client.PingServer(); !serverup || err != nil {
	// 	// logging.LogWithTrace(ctx, todoflags, "failed to ping server. check server is alive.")
	// }
	if *post || *put {
		var err error
		item, err = models.NewToDo(userId, id, title, priority, complete)
		if err != nil {
			// logging.LogWithTrace(ctx, todoflags, err.Error())
		}
		err = item.Validate(*version)
		if err != nil {
			// logging.LogWithTrace(ctx, todoflags, err.Error())
		}
	}
	if *post {
		client.Req(ctx, "POST", item, todoflags)
	}
	if *put {
		client.Req(ctx, "PUT", item, todoflags)
	}
	if *get {
		client.Req(ctx, "GET", item, todoflags)
	}
}

func main() {
	cliParse()
}
