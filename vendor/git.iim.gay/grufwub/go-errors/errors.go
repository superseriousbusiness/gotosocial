package errors

import (
	"errors"
	"fmt"
)

var (
	_ Definition = definition("")
	_ Error      = &derivedError{}
)

// BaseError defines a simple error implementation
type BaseError interface {
	// Error returns the error string
	Error() string

	// Is checks whether an error is equal to this one
	Is(error) bool

	// Unwrap attempts to unwrap any contained errors
	Unwrap() error
}

// Definition describes an error implementation that allows creating
// errors derived from this. e.g. global errors defined at runtime
// that are called with `.New()` or `.Wrap()` to derive new errors with
// extra contextual information when needed
type Definition interface {
	// New returns a new Error based on Definition using
	// supplied string as contextual information
	New(a ...interface{}) Error

	// Newf returns a new Error based on Definition using
	// supplied format string as contextual information
	Newf(string, ...interface{}) Error

	// Wrap returns a new Error, wrapping supplied error with
	// a wrapper with definition as the outer error
	Wrap(error) Error

	// must implement BaseError
	BaseError
}

// Error defines an error implementation that supports wrapping errors, easily
// accessing inner / outer errors in the wrapping structure, and setting extra
// contextual information related to this error
type Error interface {
	// Outer returns the outermost error
	Outer() error

	// Extra allows you to set extra contextual information. Please note
	// that multiple calls to .Extra() will overwrite previously set information
	Extra(...interface{}) Error

	// Extraf allows you to set extra contextual information using a format string.
	// Please note that multiple calls to .Extraf() will overwrite previously set
	// information
	Extraf(string, ...interface{}) Error

	// must implement BaseError
	BaseError
}

// New returns a simple error implementation. This exists so that `go-errors` can
// be a drop-in replacement for the standard "errors" library
func New(msg string) error {
	return definition(msg)
}

// Define returns a new error Definition
func Define(msg string) Definition {
	return definition(msg)
}

// Wrap wraps the supplied inner error within a struct of the outer error
func Wrap(outer, inner error) Error {
	// If this is a wrapped error but inner is nil, use this
	if derived, ok := outer.(*derivedError); ok && derived.inner == nil {
		derived.inner = inner
		return derived
	}

	// Create new derived error
	return &derivedError{
		msg:   "",
		extra: "",
		outer: outer,
		inner: inner,
	}
}

type definition string

func (e definition) New(a ...interface{}) Error {
	return &derivedError{
		msg:   fmt.Sprint(a...),
		extra: "",
		inner: nil,
		outer: e,
	}
}

func (e definition) Newf(msg string, a ...interface{}) Error {
	return &derivedError{
		msg:   fmt.Sprintf(msg, a...),
		extra: "",
		inner: nil,
		outer: e,
	}
}

func (e definition) Wrap(err error) Error {
	return &derivedError{
		msg:   "",
		extra: "",
		inner: err,
		outer: e,
	}
}

func (e definition) Error() string {
	return string(e)
}

func (e definition) Is(err error) bool {
	switch err := err.(type) {
	case definition:
		return e == err
	case *derivedError:
		return err.Is(e)
	default:
		return false
	}
}

func (e definition) Unwrap() error {
	return nil
}

type derivedError struct {
	msg   string // msg provides the set message for this derived error
	extra string // extra provides any extra set contextual information
	inner error  // inner is the error being wrapped
	outer error  // outer is the outmost error in this wrapper
}

func (e *derivedError) Error() string {
	// Error starts with outer error
	s := e.outer.Error() + ` (`

	// Add any message
	if e.msg != "" {
		s += `msg="` + e.msg + `" `
	}

	// Add any wrapped error
	if e.inner != nil {
		s += `wrapped="` + e.inner.Error() + `" `
	}

	// Add any extra information
	if e.extra != "" {
		s += `extra="` + e.extra + `" `
	}

	// Return error string
	return s[:len(s)-1] + `)`
}

func (e *derivedError) Is(err error) bool {
	return errors.Is(e.outer, err) || errors.Is(e.inner, err)
}

func (e *derivedError) Outer() error {
	return e.outer
}

func (e *derivedError) Unwrap() error {
	return e.inner
}

func (e *derivedError) Extra(a ...interface{}) Error {
	e.extra = fmt.Sprint(a...)
	return e
}

func (e *derivedError) Extraf(s string, a ...interface{}) Error {
	e.extra = fmt.Sprintf(s, a...)
	return e
}
