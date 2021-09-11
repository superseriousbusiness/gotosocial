package errors

import "errors"

// Is wraps "errors".Is()
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As wraps "errors".As()
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Unwrap wraps "errors".Unwrap()
func Unwrap(err error) error {
	return errors.Unwrap(err)
}
