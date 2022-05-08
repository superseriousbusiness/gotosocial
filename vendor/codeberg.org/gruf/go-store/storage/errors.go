package storage

import (
	"errors"
	"syscall"
)

var (
	// ErrClosed is returned on operations on a closed storage
	ErrClosed = errors.New("store/storage: closed")

	// ErrNotFound is the error returned when a key cannot be found in storage
	ErrNotFound = errors.New("store/storage: key not found")

	// ErrAlreadyExist is the error returned when a key already exists in storage
	ErrAlreadyExists = errors.New("store/storage: key already exists")

	// ErrInvalidkey is the error returned when an invalid key is passed to storage
	ErrInvalidKey = errors.New("store/storage: invalid key")

	// ErrAlreadyLocked is returned on fail opening a storage lockfile
	ErrAlreadyLocked = errors.New("store/storage: storage lock already open")

	// errPathIsFile is returned when a path for a disk config is actually a file
	errPathIsFile = errors.New("store/storage: path is file")

	// errNoHashesWritten is returned when no blocks are written for given input value
	errNoHashesWritten = errors.New("storage/storage: no hashes written")

	// errInvalidNode is returned when read on an invalid node in the store is attempted
	errInvalidNode = errors.New("store/storage: invalid node")

	// errCorruptNode is returned when a block fails to be opened / read during read of a node.
	errCorruptNode = errors.New("store/storage: corrupted node")
)

// wrappedError allows wrapping together an inner with outer error.
type wrappedError struct {
	inner error
	outer error
}

// wrap will return a new wrapped error from given inner and outer errors.
func wrap(outer, inner error) *wrappedError {
	return &wrappedError{
		inner: inner,
		outer: outer,
	}
}

func (e *wrappedError) Is(target error) bool {
	return e.outer == target || e.inner == target
}

func (e *wrappedError) Error() string {
	return e.outer.Error() + ": " + e.inner.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.inner
}

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

// errSwapUnavailable swaps syscall.EAGAIN for ErrAlreadyLocked
func errSwapUnavailable(err error) error {
	if err == syscall.EAGAIN {
		return ErrAlreadyLocked
	}
	return err
}
