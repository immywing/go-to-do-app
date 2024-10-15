package models_test

import (
	"testing"
	"to-do-app/models"

	"github.com/google/uuid"
)

func TestParsePriorityWithValidStrings(t *testing.T) {
	inputs := []string{"Low", "Medium", "High"}
	for _, s := range inputs {
		_, err := models.ParsePriority(s)
		if err != nil {
			t.Errorf("parser failed with %s error", err)
		}
	}
}

func TestParsePriorityWithValidLowercaseStrings(t *testing.T) {
	inputs := []string{"low", "medium", "high"}
	for _, s := range inputs {
		_, err := models.ParsePriority(s)
		if err != nil {
			t.Errorf("parser failed with %s error", err)
		}
	}
}

func TestParsePriorityErrorsWithInvalidString(t *testing.T) {
	input := "Critical"
	ret, err := models.ParsePriority(input)
	if err == nil {
		t.Errorf("Expected parser to fail given an input of %s, but returned %s", input, ret)
	}
}

func TestNewInMemDataStore(t *testing.T) {
	store := models.NewInMemDataStore()
	if store == nil {
		t.Fatal("Expected a non-nil DataStore, but got nil")
	}
	_, ok := store.(models.DataStore)
	if !ok {
		t.Fatalf("Expected the returned value to implement the DataStore interface, but got %T", store)
	}
}

func TestUpdateToDo(t *testing.T) {
	store := models.NewInMemDataStore()
	expected := models.ToDo{Id: uuid.New(), Title: "test", Priority: "Low", Complete: false}
	store.AddItem(expected)
	expected.Priority = "High"
	expected.Complete = true
	actual, _ := store.UpdateItem(expected)
	if actual != expected {
		t.Errorf("Expected: %+v, Got: %+v", expected, actual)
	}
}

func TestGetToDo(t *testing.T) {
	store := models.NewInMemDataStore()
	id := uuid.New()
	expected := models.ToDo{Id: id, Title: "test", Priority: "Low", Complete: false}
	store.AddItem(expected)
	actual, err := store.GetItem(id)
	if err != nil {
		t.Errorf("datastore unable to find item that was created with uuid: %s", id)
	}
	if actual != expected {
		t.Errorf("Expected: %+v, Got: %+v", expected, actual)
	}
}
