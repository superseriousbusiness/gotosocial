package storage

import (
	"fmt"
	"syscall"
)

// errorString is our own simple error type
type errorString string

// Error implements error
func (e errorString) Error() string {
	return string(e)
}

// Extend appends extra information to an errorString
func (e errorString) Extend(s string, a ...interface{}) errorString {
	return errorString(string(e) + ": " + fmt.Sprintf(s, a...))
}

var (
	// ErrNotFound is the error returned when a key cannot be found in storage
	ErrNotFound = errorString("store/storage: key not found")

	// ErrAlreadyExist is the error returned when a key already exists in storage
	ErrAlreadyExists = errorString("store/storage: key already exists")

	// ErrInvalidkey is the error returned when an invalid key is passed to storage
	ErrInvalidKey = errorString("store/storage: invalid key")

	// errPathIsFile is returned when a path for a disk config is actually a file
	errPathIsFile = errorString("store/storage: path is file")

	// errNoHashesWritten is returned when no blocks are written for given input value
	errNoHashesWritten = errorString("storage/storage: no hashes written")

	// errInvalidNode is returned when read on an invalid node in the store is attempted
	errInvalidNode = errorString("store/storage: invalid node")

	// errCorruptNodes is returned when nodes with missing blocks are found during a BlockStorage clean
	errCorruptNodes = errorString("store/storage: corrupted nodes")
)

// errSwapNoop performs no error swaps
func errSwapNoop(err error) error {
	return err
}

// ErrSwapNotFound swaps syscall.ENOENT for ErrNotFound
func errSwapNotFound(err error) error {
	if err == syscall.ENOENT {
		return ErrNotFound
	}
	return err
}

// errSwapExist swaps syscall.EEXIST for ErrAlreadyExists
func errSwapExist(err error) error {
	if err == syscall.EEXIST {
		return ErrAlreadyExists
	}
	return err
}
