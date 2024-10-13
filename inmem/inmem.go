package inmem

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"to-do-app/apiclient"
	"to-do-app/logging"
	"to-do-app/models"
)

const inMemCommands = `
Available commands:
    addlist <title: string>
    additem <listId: string> <title: string> <priority: string>
    getlists
    getlist <listId: string>
    getitem <listId: string> <itemId: string>
    updateitem <listId: string> <itemId: string> <title: string> <priority: string> <complete: bool>`

// func NewInMemInstance(store models.DataStore) map[string]func(args ...string) {
// 	// store.AddList("TEST LIST")
// 	// fmt.Println(store.GetLists())
// 	commandCalls := map[string]func(args ...string){
// 		// "addlist": func(args ...string) {
// 		// 	if len(args) < 1 {
// 		// 		fmt.Println("Error: addlist requires <title: string>")
// 		// 		return
// 		// 	}
// 		// 	store.AddList(args[0])
// 		// },
// 		"additem": func(args ...string) {
// 			if len(args) < 3 {
// 				fmt.Println("Error: additem requires <listId: string> <title: string> <priority: string>")
// 				return
// 			}
// 			_, err := store.AddItem(args[0], args[1])
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 		},
// 		// "getlists": func(args ...string) {
// 		// 	fmt.Println(store.GetLists())
// 		// },
// 		// "getlist": func(args ...string) {
// 		// 	if len(args) < 1 {
// 		// 		fmt.Println("Error: getlist requires <listId: string>")
// 		// 		return
// 		// 	}
// 		// 	model, err := store.GetList(args[0])
// 		// 	if err != nil {
// 		// 		fmt.Println(err)
// 		// 	} else {
// 		// 		fmt.Println(model)
// 		// 	}
// 		// },
// 		"getitem": func(args ...string) {
// 			if len(args) < 2 {
// 				fmt.Println("Error: getitem requires <listId: string> <itemId: string>")
// 				return
// 			}
// 			uuid, err := uuid.Parse(args[0])
// 			model, err := store.GetItem(uuid)
// 			if err != nil {
// 				fmt.Println(err)
// 			} else {
// 				fmt.Println(model)
// 			}
// 		},
// 		"updateitem": func(args ...string) {
// 			if len(args) < 5 {
// 				fmt.Println("Error: updateitem requires <listId: string> <itemId: string> <title: string> <priority: string> <complete: bool>")
// 				return
// 			}
// 			complete := false
// 			if args[4] == "true" {
// 				complete = true
// 			} else if args[4] == "false" {
// 				complete = false
// 			} else {
// 				fmt.Println("Error: complete must be 'true' or 'false'")
// 				return
// 			}
// 			_, err := store.UpdateItem(args[0], args[1], args[2], complete)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 		},
// 		"exit": func(args ...string) {
// 			os.Exit(0)
// 		},
// 	}
// 	return commandCalls
// }

// func callFuncWithTracing(ctx context.Context, f func(args ...string), args ...string) {
// 	f(args...)
// 	logWithTrace(ctx, "")
// }

func Run(store *models.DataStore) {
	*store = models.NewInMemDataStore()
	// commandCalls := NewInMemInstance(*store)
	// fmt.Println(commandCalls)
	client := apiclient.NewAPIClient("http://localhost:8081/")
	reader := bufio.NewReader(os.Stdin)
	fmt.Println(inMemCommands)
	for {
		userInput, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error parsing input: ", err)
		} else {
			userInput = strings.TrimSpace(userInput)
			ctx := logging.AddTraceID(context.Background())
			client.Get(ctx, userInput)
			// parts := strings.Fields(userInput)
			// if len(parts) == 0 {
			// 	fmt.Println("No command entered")
			// 	return
			// }
			// command := parts[0]
			// args := parts[1:]
			// ctx := addTraceID(context.Background())
			// callFuncWithTracing(ctx, commandCalls[command], args...)
		}
	}
}
