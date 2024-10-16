package datastores_test

import (
	"testing"
	"to-do-app/datastores"
	"to-do-app/models"

	"github.com/google/uuid"
)

func TestNewInMemDataStore(t *testing.T) {
	store := datastores.NewInMemDataStore()
	if store == nil {
		t.Fatal("Expected a non-nil DataStore, but got nil")
	}
	_, ok := store.(datastores.DataStore)
	if !ok {
		t.Fatalf("Expected the returned value to implement the DataStore interface, but got %T", store)
	}
}

func TestInMemUpdateToDo(t *testing.T) {
	store := datastores.NewInMemDataStore()
	expected := models.ToDo{Id: uuid.New(), Title: "test", Priority: "Low", Complete: false}
	store.AddItem(expected)
	expected.Priority = "High"
	expected.Complete = true
	actual, _ := store.UpdateItem(expected)
	if actual != expected {
		t.Errorf("Expected: %+v, Got: %+v", expected, actual)
	}
}

func TestInMemGetToDo(t *testing.T) {
	store := datastores.NewInMemDataStore()
	id := uuid.New()
	userId := uuid.New()
	expected := models.ToDo{Id: id, Title: "test", Priority: "Low", Complete: false, UserId: userId.String()}
	store.AddItem(expected)
	actual, err := store.GetItem(userId.String(), id)
	if err != nil {
		t.Errorf("datastore unable to find item that was created with uuid: %s", id)
	}
	if actual != expected {
		t.Errorf("Expected: %+v, Got: %+v", expected, actual)
	}
}
