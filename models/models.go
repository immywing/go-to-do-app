package models

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	todoerrors "to-do-app/errors"

	"github.com/google/uuid"
)

type priority = string

const (
	PriorityLow    priority = "Low"
	PriorityMedium priority = "Medium"
	PriorityHigh   priority = "High"
)

func ParsePriority(p string) (priority, error) {
	p = strings.ToUpper(string(p[0])) + strings.ToLower(p[1:])
	switch priority(p) {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return priority(p), nil
	default:
		return "", fmt.Errorf("invalid priority: %s. Valid options are: %s, %s, %s", p, PriorityLow, PriorityMedium, PriorityHigh)
	}
}

type ToDo struct {
	Id       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Priority priority  `json:"priority"`
	Complete bool      `json:"complete"`
}

func ToDoFromCLI(id *string, title *string, priority *string, complete *bool) (ToDo, error) {
	uuid, err := uuid.Parse(*id)
	if err != nil {
		return ToDo{}, err
	}
	if *title == "" {
		return ToDo{}, &todoerrors.ValidationError{Field: *title, Err: errors.New("title is required")}
	}
	p, err := ParsePriority(*priority)
	if err != nil {
		return ToDo{}, &todoerrors.ValidationError{Field: *priority, Err: err}
	}
	return ToDo{Id: uuid, Title: *title, Priority: p, Complete: *complete}, nil
}

type CreateListRequest struct {
	Title string `json:"title"`
}

type DataStore interface {
	AddItem(item ToDo) (ToDo, error)
	GetItem(itemId uuid.UUID) (ToDo, error)
	UpdateItem(item ToDo) (ToDo, error)
}

type inMemDatastore struct {
	Items map[uuid.UUID]ToDo
	mut   sync.RWMutex
}

func (ds *inMemDatastore) AddItem(item ToDo) (ToDo, error) {
	if item.Title == "" {
		return ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := ParsePriority(item.Priority)
	if err != nil {
		return ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	ds.Items[item.Id] = item
	return item, nil
}

func (ds *inMemDatastore) GetItem(itemId uuid.UUID) (ToDo, error) {
	// itemUUID, err := uuid.Parse(itemId)
	// if err != nil {
	// 	return ToDo{}, &todoerrors.ValidationError{Field: itemId, Err: err}
	// }
	if item, exists := ds.Items[itemId]; exists {
		return item, nil
	}
	return ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func (ds *inMemDatastore) UpdateItem(item ToDo) (ToDo, error) {
	if item.Title == "" {
		return ToDo{}, &todoerrors.ValidationError{Field: item.Title, Err: errors.New("invalid title")}
	}
	p, err := ParsePriority(item.Priority)
	if err != nil {
		return ToDo{}, &todoerrors.ValidationError{Field: item.Priority, Err: err}
	}
	item.Priority = p
	if item, exists := ds.Items[item.Id]; exists {
		ds.Items[item.Id] = item
		return ds.Items[item.Id], nil
	}
	return ToDo{}, &todoerrors.NotFoundError{Message: "ToDo Not Found"}
}

func NewInMemDataStore() DataStore {
	return &inMemDatastore{Items: make(map[uuid.UUID]ToDo), mut: sync.RWMutex{}}
}
