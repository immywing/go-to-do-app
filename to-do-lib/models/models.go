package models

import (
	"errors"
	"fmt"
	"strings"

	todoerrors "go-to-do-app/to-do-lib/errors"

	"github.com/google/uuid"
)

type priority = string

const (
	PriorityLow    priority = "Low"
	PriorityMedium priority = "Medium"
	PriorityHigh   priority = "High"
)

var (
	V1 = "v1"
	V2 = "v2"
)

func ParsePriority(p string) (priority, error) {
	if len(p) < 1 {
		return "", fmt.Errorf("invalid priority: %s. Valid options are: %s, %s, %s", p, PriorityLow, PriorityMedium, PriorityHigh)
	}
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
	UserId   string    `json:"user_id,omitempty"`
}

func (t *ToDo) Validate(ver string) error {
	if t.Title == "" {
		return &todoerrors.ValidationError{Field: t.Title, Err: errors.New("invalid title")}
	}
	p, err := ParsePriority(t.Priority)
	if err != nil {
		return &todoerrors.ValidationError{Field: t.Priority, Err: err}
	}
	t.Priority = p
	switch ver {
	case V1:
		if t.UserId != "" {
			return &todoerrors.ValidationError{Field: fmt.Sprintf("user_id: %s", t.UserId), Err: errors.New("v1 todo api does not allow user_id")}
		}
	case V2:
		if t.UserId == "" {
			return &todoerrors.ValidationError{Field: fmt.Sprintf("user_id: %s", t.UserId), Err: errors.New("invalid user_id")}
		}
	default:
		return &todoerrors.NotFoundError{Message: fmt.Sprintf("%d not a valid version", t.Id.Version())}
	}
	return nil
}

func NewToDo(userId *string, id *string, title *string, priority *string, complete *bool) (ToDo, error) {
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
