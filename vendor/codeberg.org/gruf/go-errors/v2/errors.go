package errors

import (
	"fmt"
)

// New returns a new error created from message.
func New(msg string) error {
	return create(msg, nil)
}

// Newf returns a new error created from message format and args.
func Newf(msgf string, args ...interface{}) error {
	return create(fmt.Sprintf(msgf, args...), nil)
}

// Wrap will wrap supplied error within a new error created from message.
func Wrap(err error, msg string) error {
	return create(msg, err)
}

// Wrapf will wrap supplied error within a new error created from message format and args.
func Wrapf(err error, msgf string, args ...interface{}) error {
	return create(fmt.Sprintf(msgf, args...), err)
}

// Stacktrace fetches first stored stacktrace of callers from error chain.
func Stacktrace(err error) Callers {
	var e interface {
		Stacktrace() Callers
	}
	if !As(err, &e) {
		return nil
	}
	return e.Stacktrace()
}
