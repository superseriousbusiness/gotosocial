package errors

import (
	"errors"
	"reflect"
	_ "unsafe"

	"codeberg.org/gruf/go-bitutil"
)

// Is reports whether any error in err's chain matches any of targets
// (up to a max of 64 targets).
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error is considered to match a target if it is equal to that target or if
// it implements a method Is(error) bool such that Is(target) returns true.
func Is(err error, targets ...error) bool {
	var flags bitutil.Flags64

	// Flags only has 64 bit-slots
	if len(targets) > 64 {
		panic("too many targets")
	}

	// Check if error is nil so we can catch
	// the fast-case where a target is nil
	isNil := (err == nil)

	for i := 0; i < len(targets); {
		// Drop nil targets
		if targets[i] == nil {
			if isNil /* match! */ {
				return true
			}
			targets = append(targets[:i], targets[i+1:]...)
			continue
		}

		// Check if this error is directly comparable
		if reflect.TypeOf(targets[i]).Comparable() {
			flags = flags.Set(uint8(i))
		}

		i++
	}

	for err != nil {
		// Check if this layer supports .Is interface
		is, ok := err.(interface{ Is(error) bool })

		if !ok {
			// Error does not support interface
			//
			// Only try perform direct compare
			for i := 0; i < len(targets); i++ {
				// Try directly compare errors
				if flags.Get(uint8(i)) &&
					err == targets[i] {
					return true
				}
			}
		} else {
			// Error supports the .Is interface
			//
			// Perform direct compare AND .Is()
			for i := 0; i < len(targets); i++ {
				if (flags.Get(uint8(i)) &&
					err == targets[i]) ||
					is.Is(targets[i]) {
					return true
				}
			}
		}

		// Unwrap to next layer
		err = errors.Unwrap(err)
	}

	return false
}

// As finds the first error in err's chain that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
//
//go:linkname As errors.As
func As(err error, target interface{}) bool

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
//
//go:linkname Unwrap errors.Unwrap
func Unwrap(err error) error
