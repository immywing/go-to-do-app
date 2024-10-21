package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go-to-do-app/to-do-lib/datastores"
	"go-to-do-app/to-do-lib/logging"
	"go-to-do-app/to-do-server/server"
)

var (
	mode         = flag.String("mode", "", "set the mode the application should run in (in-mem, json-store, pgdb)")
	jsonPath     = flag.String("json", "", "filepath of json file to use as datastore")
	shutdownChan = make(chan bool)
)

func listenForShutdownCommand(server server.ToDoServer) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Todo server running\n!Q to close the server")
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		if strings.Trim(strings.Fields(text)[0], " \n") == "!Q" {
			server.Shutdown()
			break
		}
	}
}

func run() {
	flag.Parse()

	var store datastores.DataStore
	if *mode == "" {
		logging.LogWithTrace(
			context.Background(),
			map[string]interface{}{},
			"no valid mode provided to start server with datastore",
		)
		os.Exit(1)
	}
	if *mode == "pgdb" {
		fmt.Fprintf(os.Stderr, "Error: the mode '%s' is not yet implemented\n", *mode)
		os.Exit(1)
	}
	if *mode == "in-mem" {
		store = datastores.NewInMemDataStore()
	}
	if *mode == "json-store" {
		if filepath.Ext(*jsonPath) != ".json" {
			logging.LogWithTrace(
				context.Background(),
				map[string]interface{}{"path": *jsonPath},
				"no valid path to json file provided",
			)
			os.Exit(1)
		}
		store = datastores.NewJsonDatastore(*jsonPath)
	}

	srv := server.NewToDoServer(":8081", shutdownChan, store)
	go srv.Start()
	listenForShutdownCommand(srv)
	srv.AwaitShutdown()
}

func main() {
	run()
}
