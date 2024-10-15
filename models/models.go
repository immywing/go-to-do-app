package models

import (
	"errors"
	"fmt"
	"strings"
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
