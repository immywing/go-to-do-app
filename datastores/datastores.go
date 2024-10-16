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
	GetItem(userId string, itemId uuid.UUID) (models.ToDo, error)
	UpdateItem(item models.ToDo) (models.ToDo, error)
}

type inMemDatastore struct {
	Items map[string]map[uuid.UUID]models.ToDo
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
	// ds.Items[item.UserId][item.Id] = item
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
	if item.Title == "" {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := models.ParsePriority(item.Priority)
	if err != nil {
		return models.ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	if user, exists := ds.Items[item.UserId]; exists {
		if _, iexist := user[item.Id]; iexist {
			ds.Items[item.UserId][item.Id] = item

			return ds.Items[item.UserId][item.Id], nil
		}
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func SaveToJson(todos map[string]map[uuid.UUID]models.ToDo, fpath string) {
	items := make([]models.ToDo, 0)
	for _, user := range todos {
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
	err = os.WriteFile(fpath, bytes, 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("Data written successfully!")
}

func NewInMemDataStore() DataStore {
	return &inMemDatastore{Items: make(map[string]map[uuid.UUID]models.ToDo), mut: sync.Mutex{}}
}

func LoadJsonStore(fpath string) map[string]map[uuid.UUID]models.ToDo {
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
	items := make(map[string]map[uuid.UUID]models.ToDo)
	for _, item := range todos {
		items[item.UserId] = map[uuid.UUID]models.ToDo{item.Id: item}
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
	if user, exists := items[item.UserId]; exists {
		user[item.Id] = item
	} else {
		items[item.UserId] = map[uuid.UUID]models.ToDo{item.Id: item}
	}
	SaveToJson(items, ds.fpath)
	return items[item.UserId][item.Id], nil
}

func (ds *JsonDatastore) GetItem(userId string, itemId uuid.UUID) (models.ToDo, error) {
	ds.mut.Lock()
	defer ds.mut.Unlock()
	items := LoadJsonStore(ds.fpath)
	if item, exists := items[userId][itemId]; exists {
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
	if user, exists := items[item.UserId]; exists {
		if _, iexist := user[item.Id]; iexist {
			items[item.UserId][item.Id] = item
			SaveToJson(items, ds.fpath)
			return items[item.UserId][item.Id], nil
		}
	}
	return models.ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func NewJsonDatastore(path string) DataStore {
	return &JsonDatastore{fpath: path, mut: sync.Mutex{}}
}
