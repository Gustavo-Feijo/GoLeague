package testutil

import (
	"errors"
)

const DatabaseError = "database error occurred"

type OperationRestult[T any] struct {
	Data T
	Err  error
}

// Return a generic typed error return for a Database call.
func GetMockRepoError[T any]() *OperationRestult[T] {
	return NewErrorResult[T](DatabaseError)
}

func NewErrorResult[T any](err string) *OperationRestult[T] {
	return &OperationRestult[T]{
		Data: *new(T),
		Err:  errors.New(err),
	}
}

// Wrap a generic Data into a OperationRestult struct.
func NewSuccessResult[T any](Data T) *OperationRestult[T] {
	return &OperationRestult[T]{
		Data: Data,
		Err:  nil,
	}
}
