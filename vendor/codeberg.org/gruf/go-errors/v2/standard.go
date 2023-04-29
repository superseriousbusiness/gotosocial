package errors

import (
	"errors"
	"reflect"
	_ "unsafe"

	"codeberg.org/gruf/go-bitutil"
)

// errtype is a ptr to the error interface type.
var errtype = reflect.TypeOf((*error)(nil)).Elem()

// Comparable is functionally equivalent to calling errors.Is() on multiple errors (up to a max of 64).
func Comparable(err error, targets ...error) bool {
	var flags bitutil.Flags64

	// Flags only has 64 bit-slots
	if len(targets) > 64 {
		panic("too many targets")
	}

	for i := 0; i < len(targets); {
		if targets[i] == nil {
			if err == nil {
				return true
			}

			// Drop nil targets from slice.
			copy(targets[i:], targets[i+1:])
			targets = targets[:len(targets)-1]
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

// Assignable is functionally equivalent to calling errors.As() on multiple errors,
// except that it only checks assignability as opposed to setting the target.
func Assignable(err error, targets ...error) bool {
	if err == nil {
		// Fastest case.
		return false
	}

	for i := 0; i < len(targets); {
		if targets[i] == nil {
			// Drop nil targets from slice.
			copy(targets[i:], targets[i+1:])
			targets = targets[:len(targets)-1]
			continue
		}
		i++
	}

	for err != nil {
		// Check if this layer supports .As interface
		as, ok := err.(interface{ As(any) bool })

		// Get reflected err type.
		te := reflect.TypeOf(err)

		if !ok {
			// Error does not support interface.
			//
			// Check assignability using reflection.
			for i := 0; i < len(targets); i++ {
				tt := reflect.TypeOf(targets[i])
				if te.AssignableTo(tt) {
					return true
				}
			}
		} else {
			// Error supports the .As interface.
			//
			// Check using .As() and reflection.
			for i := 0; i < len(targets); i++ {
				if as.As(targets[i]) {
					return true
				} else if tt := reflect.TypeOf(targets[i]); // nocollapse
				te.AssignableTo(tt) {
					return true
				}
			}
		}

		// Unwrap to next layer.
		err = errors.Unwrap(err)
	}

	return false
}

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
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
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
//
//go:linkname As errors.As
func As(err error, target any) bool

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
//
//go:linkname Unwrap errors.Unwrap
func Unwrap(err error) error
