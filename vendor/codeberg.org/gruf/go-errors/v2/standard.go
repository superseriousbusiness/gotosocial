package errors

import (
	_ "unsafe"
)

// Is reports whether any error in err's tree matches target.
//
// The tree consists of err itself, followed by the errors obtained by repeatedly
// calling Unwrap. When err wraps multiple errors, Is examines err followed by a
// depth-first traversal of its children.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
//
// An error type might provide an Is method so it can be treated as equivalent
// to an existing error. For example, if MyError defines
//
//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
//
// then Is(MyError{}, fs.ErrExist) returns true. See [syscall.Errno.Is] for
// an example in the standard library. An Is method should only shallowly
// compare err and the target and not call Unwrap on either.
//
//go:linkname Is errors.Is
func Is(err error, target error) bool

// As finds the first error in err's tree that matches generic parameter type.
//
// The tree consists of err itself, followed by the errors obtained by repeatedly
// calling Unwrap. When err wraps multiple errors, As examines err followed by a
// depth-first traversal of its children.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
func As[Type any](err error) Type {
	var t Type
	var ok bool
	errs := []error{err}
	for len(errs) > 0 {
		// Pop next error to check.
		err := errs[len(errs)-1]
		errs = errs[:len(errs)-1]

		// Check direct type.
		t, ok = err.(Type)
		if ok {
			return t
		}

		// Look for .As() support.
		as, ok := err.(interface {
			As(target any) bool
		})

		if ok {
			// Attempt .As().
			if as.As(&t) {
				return t
			}
		}

		// Try unwrap errors.
		switch u := err.(type) {
		case interface{ Unwrap() error }:
			errs = append(errs, u.Unwrap())
		case interface{ Unwrap() []error }:
			errs = append(errs, u.Unwrap()...)
		}
	}
	return t
}

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error.
// Otherwise, Unwrap returns nil.
//
// Unwrap only calls a method of the form "Unwrap() error".
// In particular Unwrap does not unwrap errors returned by [Join].
//
//go:linkname Unwrap errors.Unwrap
func Unwrap(err error) error

// Join returns an error that wraps the given errors.
// Any nil error values are discarded.
// Join returns nil if every value in errs is nil.
// The error formats as the concatenation of the strings obtained
// by calling the Error method of each element of errs, with a newline
// between each string.
//
// A non-nil error returned by Join implements the Unwrap() []error method.
//
//go:linkname Join errors.Join
func Join(errs ...error) error
