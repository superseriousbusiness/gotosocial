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

// Stacktrace fetches a stored stacktrace of callers from an error, or returns nil.
func Stacktrace(err error) Callers {
	var callers Callers
	if err, ok := err.(interface { //nolint
		Stacktrace() Callers
	}); ok {
		callers = err.Stacktrace()
	}
	return callers
}
