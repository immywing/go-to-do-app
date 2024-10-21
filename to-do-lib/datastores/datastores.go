package datastores

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	todoerrors "go-to-do-app/to-do-lib/errors"
	"go-to-do-app/to-do-lib/logging"
	"go-to-do-app/to-do-lib/models"

	"github.com/google/uuid"
)

// type record struct {
// 	mut  sync.Mutex
// 	data models.ToDo
// }

// func (t *record) Lock() {
// 	t.mut.Lock()
// }

// func (t *record) Unlock() {
// 	t.mut.Unlock()
// }

type DataStore interface {
	AddItem(item models.ToDo) (models.ToDo, error)
	GetItem(userId string, itemId uuid.UUID) (models.ToDo, error)
	UpdateItem(item models.ToDo) (models.ToDo, error)
	Close()
}

type inMemDatastore struct {
	Items map[string]map[uuid.UUID]models.ToDo
	mut   sync.Mutex
}

func (ds *inMemDatastore) AddItem(item models.ToDo) (models.ToDo, error) {
	item.Id = uuid.New()
	ds.mut.Lock()
	defer ds.mut.Unlock()

	if user, exists := ds.Items[item.UserId]; exists {
		user[item.Id] = item
	} else {
		ds.Items[item.UserId] = map[uuid.UUID]models.ToDo{item.Id: item}
	}
	return ds.Items[item.UserId][item.Id], nil
}

func (ds *inMemDatastore) GetItem(userId string, itemId uuid.UUID) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if item, exists := ds.Items[userId][itemId]; exists {
		return item, nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *inMemDatastore) UpdateItem(item models.ToDo) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()

	if user, exists := ds.Items[item.UserId]; exists {
		if _, iexist := user[item.Id]; iexist {
			user[item.Id] = item
			return ds.Items[item.UserId][item.Id], nil
		}
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *inMemDatastore) Close() {
	//no action for in mem
}

func NewInMemDataStore() DataStore {
	return &inMemDatastore{Items: make(map[string]map[uuid.UUID]models.ToDo), mut: sync.Mutex{}}
}

func LoadJsonStore(fpath string) map[string]map[uuid.UUID]models.ToDo {
	file, err := os.Open(fpath)
	ctx := context.Background()
	logging.AddTraceID(ctx)
	if err != nil {
		logging.LogWithTrace(ctx, map[string]interface{}{}, err.Error())
	}
	defer file.Close()
	var todos []models.ToDo
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&todos)
	if err != nil {
		logging.LogWithTrace(ctx, map[string]interface{}{}, err.Error())
	}
	items := make(map[string]map[uuid.UUID]models.ToDo)
	for _, item := range todos {
		items[item.UserId] = map[uuid.UUID]models.ToDo{item.Id: item}
		if err != nil {
			logging.LogWithTrace(ctx, map[string]interface{}{}, fmt.Sprintf("error with todo: %+v", item))
		}
	}
	return items
}

type JsonDatastore struct {
	fpath string
	mut   sync.Mutex
	items map[string]map[uuid.UUID]models.ToDo
}

func (ds *JsonDatastore) AddItem(item models.ToDo) (models.ToDo, error) {
	item.Id = uuid.New()
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if user, exists := ds.items[item.UserId]; exists {
		user[item.Id] = item
	} else {
		ds.items[item.UserId] = map[uuid.UUID]models.ToDo{item.Id: item}
	}
	ds.Close()
	return ds.items[item.UserId][item.Id], nil
}

func (ds *JsonDatastore) GetItem(userId string, itemId uuid.UUID) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if item, exists := ds.items[userId][itemId]; exists {
		return item, nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *JsonDatastore) UpdateItem(item models.ToDo) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if user, exists := ds.items[item.UserId]; exists {
		if _, iexist := user[item.Id]; iexist {
			ds.items[item.UserId][item.Id] = item
			ds.Close()
			return ds.items[item.UserId][item.Id], nil
		}
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *JsonDatastore) Close() {
	items := make([]models.ToDo, 0)
	for _, user := range ds.items {
		for _, item := range user {
			items = append(items, item)
		}
	}
	bytes, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
	// Write the JSON to the file
	err = os.WriteFile(ds.fpath, bytes, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Data written successfully!!!!!")
}

func NewJsonDatastore(path string) DataStore {
	return &JsonDatastore{fpath: path, items: LoadJsonStore(path), mut: sync.Mutex{}}
}
