package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/immywing/go-to-do-app/server/wiring"
	"github.com/immywing/go-to-do-app/to-do-lib/datastores"
	"github.com/immywing/go-to-do-app/to-do-lib/logging"
)

var (
	mode     = flag.String("mode", "", "set the mode the application should run in (in-mem, json-store, pgdb)")
	jsonPath = flag.String("json", "", "filepath of json file to use as datastore")
)

func run() {
	flag.Parse()

	var store datastores.DataStore
	if *mode == "pgdb" {
		fmt.Fprintf(os.Stderr, "Error: the mode '%s' is not yet implemented\n", *mode)
		os.Exit(1)
	} else if *mode == "in-mem" {
		store = datastores.NewInMemDataStore()
	} else if *mode == "json-store" {
		if filepath.Ext(*jsonPath) != ".json" {
			logging.LogWithTrace(
				context.Background(),
				map[string]interface{}{"path": *jsonPath},
				"no valid path to json file provided",
			)
			os.Exit(1)
		}
		store = datastores.NewJsonDatastore(*jsonPath)
	} else {
		logging.LogWithTrace(
			context.Background(),
			map[string]interface{}{},
			"no valid mode provided to start server with datastore",
		)
		os.Exit(1)
	}

	// Create a channel to listen for shutdown signals
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	wiring.WireEndpoints()
	go wiring.Start(&store, shutdownChan) // Start server in a goroutine

	// Wait for a shutdown signal to ensure a graceful exit
	<-shutdownChan
}

func main() {
	run()
}
