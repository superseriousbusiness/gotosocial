package errors

import (
	"errors"
)

// See: errors.Is().
func Is(err error, target error) bool { return errors.Is(err, target) }

// IsV2 calls Is(err, target) for each target within targets.
func IsV2(err error, targets ...error) bool {
	for _, target := range targets {
		if Is(err, target) {
			return true
		}
	}
	return false
}

// See: errors.As().
func As(err error, target any) bool { return errors.As(err, target) }

// AsV2 is functionally similar to As(), instead
// leveraging generics to handle allocation and
// returning of a concrete generic parameter type.
func AsV2[Type any](err error) Type {
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

// See: errors.Unwrap().
func Unwrap(err error) error { return errors.Unwrap(err) }

// UnwrapV2 is functionally similar to Unwrap(), except that
// it also handles the case of interface{ Unwrap() []error }.
func UnwrapV2(err error) []error {
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		if e := u.Unwrap(); err != nil {
			return []error{e}
		}
	case interface{ Unwrap() []error }:
		return u.Unwrap()
	}
	return nil
}

// See: errors.Join().
func Join(errs ...error) error { return errors.Join(errs...) }
