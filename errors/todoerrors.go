package todoerrors

import "fmt"

type NotFoundError struct {
	Message string
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ValidationError struct {
	Field string
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("Validation error on field %s: %v", e.Field, e.Err)
}
