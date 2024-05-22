package storage

import (
	"errors"
)

var (
	// ErrNotFound is the error returned when a key cannot be found in storage
	ErrNotFound = errors.New("storage: key not found")

	// ErrAlreadyExist is the error returned when a key already exists in storage
	ErrAlreadyExists = errors.New("storage: key already exists")

	// ErrInvalidkey is the error returned when an invalid key is passed to storage
	ErrInvalidKey = errors.New("storage: invalid key")
)
