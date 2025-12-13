package testutil

import (
	"errors"
)

const DatabaseError = "database error occurred"

type RepoGetData[T any] struct {
	Data T
	Err  error
}

// Return a generic typed error return for a Database call.
func GetMockRepoError[T any]() *RepoGetData[T] {
	return &RepoGetData[T]{
		Data: *new(T),
		Err:  errors.New(DatabaseError),
	}
}

// Wrap a generic Data into a RepoGetData struct.
func ToRepoGetData[T any](Data T) *RepoGetData[T] {
	return &RepoGetData[T]{
		Data: Data,
		Err:  nil,
	}
}
