package storage

import (
	"errors"
	"strings"
	"syscall"

	"github.com/minio/minio-go/v7"
)

var (
	// ErrClosed is returned on operations on a closed storage
	ErrClosed = new_error("closed")

	// ErrNotFound is the error returned when a key cannot be found in storage
	ErrNotFound = new_error("key not found")

	// ErrAlreadyExist is the error returned when a key already exists in storage
	ErrAlreadyExists = new_error("key already exists")

	// ErrInvalidkey is the error returned when an invalid key is passed to storage
	ErrInvalidKey = new_error("invalid key")

	// ErrAlreadyLocked is returned on fail opening a storage lockfile
	ErrAlreadyLocked = new_error("storage lock already open")
)

// new_error returns a new error instance prefixed by package prefix.
func new_error(msg string) error {
	return errors.New("store/storage: " + msg)
}

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

// transformS3Error transforms an error returned from S3Storage underlying
// minio.Core client, by wrapping where necessary with our own error types.
func transformS3Error(err error) error {
	// Cast this to a minio error response
	ersp, ok := err.(minio.ErrorResponse)
	if ok {
		switch ersp.Code {
		case "NoSuchKey":
			return wrap(ErrNotFound, err)
		case "Conflict":
			return wrap(ErrAlreadyExists, err)
		default:
			return err
		}
	}

	// Check if error has an invalid object name prefix
	if strings.HasPrefix(err.Error(), "Object name ") {
		return wrap(ErrInvalidKey, err)
	}

	return err
}
