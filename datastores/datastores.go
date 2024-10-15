package datastores

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	todoerrors "to-do-app/errors"
	"to-do-app/logging"
	"to-do-app/models"

	"github.com/google/uuid"
)

type DataStore interface {
	AddItem(item models.ToDo) (models.ToDo, error)
	GetItem(itemId uuid.UUID) (models.ToDo, error)
	UpdateItem(item models.ToDo) (models.ToDo, error)
}

type inMemDatastore struct {
	Items map[uuid.UUID]models.ToDo
	mut   sync.Mutex
}

func (ds *inMemDatastore) AddItem(item models.ToDo) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if item.Title == "" {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := models.ParsePriority(item.Priority)
	if err != nil {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	ds.Items[item.Id] = item
	return ds.Items[item.Id], nil
}

func (ds *inMemDatastore) GetItem(itemId uuid.UUID) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if item, exists := ds.Items[itemId]; exists {
		return item, nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *inMemDatastore) UpdateItem(item models.ToDo) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	if item.Title == "" {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := models.ParsePriority(item.Priority)
	if err != nil {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	if _, exists := ds.Items[item.Id]; exists {
		ds.Items[item.Id] = item
		return ds.Items[item.Id], nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func SaveToJson(todos map[uuid.UUID]models.ToDo, fpath string) {
	items := make([]models.ToDo, 0)
	for _, item := range todos {
		items = append(items, item)
	}
	bytes, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// Write the JSON to the file
	err = os.WriteFile(fpath, bytes, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Data written successfully!")
}

func NewInMemDataStore() DataStore {
	return &inMemDatastore{Items: make(map[uuid.UUID]models.ToDo), mut: sync.Mutex{}}
}

func LoadJsonStore(fpath string) map[uuid.UUID]models.ToDo {
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
	// store := NewInMemDataStore()
	items := make(map[uuid.UUID]models.ToDo)
	for _, item := range todos {
		items[item.Id] = item
		if err != nil {
			// fmt.Errorf("error with todo: %+v", item)
		}
	}
	return items
}

type JsonDatastore struct {
	fpath string
	mut   sync.Mutex
}

func (ds *JsonDatastore) AddItem(item models.ToDo) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	items := LoadJsonStore(ds.fpath)
	if item.Title == "" {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := models.ParsePriority(item.Priority)
	if err != nil {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	items[item.Id] = item
	SaveToJson(items, ds.fpath)
	return items[item.Id], nil
}

func (ds *JsonDatastore) GetItem(itemId uuid.UUID) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	items := LoadJsonStore(ds.fpath)
	if item, exists := items[itemId]; exists {
		return item, nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *JsonDatastore) UpdateItem(item models.ToDo) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	items := LoadJsonStore(ds.fpath)
	if item.Title == "" {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := models.ParsePriority(item.Priority)
	if err != nil {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	if _, exists := items[item.Id]; exists {
		items[item.Id] = item
		SaveToJson(items, ds.fpath)
		return items[item.Id], nil
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func NewJsonDatastore(path string) DataStore {
	return &JsonDatastore{fpath: path, mut: sync.Mutex{}}
}
