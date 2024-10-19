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

type record struct {
	mut  sync.Mutex
	data models.ToDo
}

func (t *record) Lock() {
	t.mut.Lock()
}

func (t *record) Unlock() {
	t.mut.Unlock()
}

type DataStore interface {
	AddItem(item models.ToDo) (models.ToDo, error)
	GetItem(userId string, itemId uuid.UUID) (models.ToDo, error)
	UpdateItem(item models.ToDo) (models.ToDo, error)
	Close()
}

type inMemDatastore struct {
	Items map[string]map[uuid.UUID]*record
	mut   sync.Mutex
}

func (ds *inMemDatastore) AddItem(item models.ToDo) (models.ToDo, error) {
	if user, exists := ds.Items[item.UserId]; exists {
		rec := user[item.Id]
		rec.Lock()
		defer rec.Unlock()
		rec.data = item
	} else {
		rec := record{mut: sync.Mutex{}, data: item}
		ds.Items[item.UserId] = map[uuid.UUID]*record{item.Id: &rec}
	}
	return ds.Items[item.UserId][item.Id].data, nil
}

func (ds *inMemDatastore) GetItem(userId string, itemId uuid.UUID) (models.ToDo, error) {
	if item, exists := ds.Items[userId][itemId]; exists {
		return item.data, nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *inMemDatastore) UpdateItem(item models.ToDo) (models.ToDo, error) {
	if user, exists := ds.Items[item.UserId]; exists {
		if rec, iexist := user[item.Id]; iexist {
			rec.Lock()
			defer rec.Unlock()
			rec.data = item
			return ds.Items[item.UserId][item.Id].data, nil
		}
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *inMemDatastore) Close() {

}

func NewInMemDataStore() DataStore {
	return &inMemDatastore{Items: make(map[string]map[uuid.UUID]*record), mut: sync.Mutex{}}
}

func LoadJsonStore(fpath string) map[string]map[uuid.UUID]*record {
	file, err := os.Open(fpath)
	if err != nil {
		ctx := context.Background()
		logging.AddTraceID(ctx)
		logging.LogWithTrace(ctx, map[string]interface{}{}, err.Error())
	}
	defer file.Close()
	var todos []models.ToDo
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&todos)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		// 	return
	}
	items := make(map[string]map[uuid.UUID]*record)
	for _, item := range todos {
		rec := record{mut: sync.Mutex{}, data: item}
		items[item.UserId] = map[uuid.UUID]*record{item.Id: &rec}
		if err != nil {
			fmt.Errorf("error with todo: %+v", item)
		}
	}
	return items
}

type JsonDatastore struct {
	fpath string
	items map[string]map[uuid.UUID]*record
	mut   sync.Mutex
}

func (ds *JsonDatastore) AddItem(item models.ToDo) (models.ToDo, error) {
	if user, exists := ds.items[item.UserId]; exists {
		rec := user[item.Id]
		rec.Lock()
		defer rec.Unlock()
		rec.data = item
	} else {
		rec := record{mut: sync.Mutex{}, data: item}
		ds.items[item.UserId] = map[uuid.UUID]*record{item.Id: &rec}
		ds.Close()
	}
	return ds.items[item.UserId][item.Id].data, nil
}

func (ds *JsonDatastore) GetItem(userId string, itemId uuid.UUID) (models.ToDo, error) {
	if rec, exists := ds.items[userId][itemId]; exists {
		return rec.data, nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *JsonDatastore) UpdateItem(item models.ToDo) (models.ToDo, error) {
	if user, exists := ds.items[item.UserId]; exists {
		if rec, iexist := user[item.Id]; iexist {
			rec.Lock()
			defer rec.Unlock()
			rec.data = item
			ds.Close()
			return ds.items[item.UserId][item.Id].data, nil
		}
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *JsonDatastore) Close() {
	items := make([]models.ToDo, 0)
	for _, user := range ds.items {
		for _, item := range user {
			items = append(items, item.data)
		}
	}
	bytes, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}
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
