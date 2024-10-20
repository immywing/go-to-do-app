package datastores_test

import (
	"testing"

	"go-to-do-app/to-do-lib/datastores"
	todoerrors "go-to-do-app/to-do-lib/errors"
	"go-to-do-app/to-do-lib/models"

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
	expected := models.ToDo{Id: uuid.New(), Title: "test", Priority: "Low", Complete: false, UserId: uuid.New().String()}
	store.AddItem(expected)
	expected.Priority = "High"
	expected.Complete = true
	actual, _ := store.UpdateItem(expected)
	if actual != expected {
		t.Errorf("Expected: %+v, Got: %+v", expected, actual)
	}
}

func TestUpdateNonExistientToDo(t *testing.T) {
	store := datastores.NewInMemDataStore()
	td := models.ToDo{Id: uuid.New(), Title: "test", Priority: "Low", Complete: false, UserId: uuid.New().String()}
	expected := todoerrors.NotFoundError{Message: "ToDo Not Found"}
	_, actual := store.UpdateItem(td)
	_, ok := actual.(*todoerrors.NotFoundError)
	if !ok {
		t.Errorf("Expected: %T, Got: %T", expected, actual)
	}
}

func TestInMemGetToDo(t *testing.T) {
	store := datastores.NewInMemDataStore()
	id := uuid.New()
	userId := uuid.New().String()
	expected := models.ToDo{Id: id, Title: "test", Priority: "Low", Complete: false, UserId: userId}
	store.AddItem(expected)
	actual, err := store.GetItem(userId, id)
	if err != nil {
		t.Errorf("datastore unable to find item that was created with uuid: %s", id)
	}
	if actual != expected {
		t.Errorf("Expected: %+v, Got: %+v", expected, actual)
	}
}
