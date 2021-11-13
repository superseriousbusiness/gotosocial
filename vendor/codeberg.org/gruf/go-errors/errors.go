package errors

import "fmt"

// ErrorContext defines a wrappable error with the ability to hold extra contextual information
type ErrorContext interface {
	// implement base error interface
	error

	// Is identifies whether the receiver contains / is the target
	Is(error) bool

	// Unwrap reveals the underlying wrapped error (if any!)
	Unwrap() error

	// Value attempts to fetch contextual data for given key from this ErrorContext
	Value(string) (interface{}, bool)

	// Append allows adding contextual data to this ErrorContext
	Append(...KV) ErrorContext

	// Data returns the contextual data structure associated with this ErrorContext
	Data() ErrorData
}

// New returns a new ErrorContext created from string
func New(msg string) ErrorContext {
	return stringError(msg)
}

// Newf returns a new ErrorContext created from format string
func Newf(s string, a ...interface{}) ErrorContext {
	return stringError(fmt.Sprintf(s, a...))
}

// Wrap ensures supplied error is an ErrorContext, wrapping if necessary
func Wrap(err error) ErrorContext {
	// Nil error, do nothing
	if err == nil {
		return nil
	}

	// Check if this is already wrapped somewhere in stack
	if xerr, ok := err.(*errorContext); ok {
		return xerr
	} else if As(err, &xerr) {
		// This is really not an ideal situation,
		// but we try to make do by salvaging the
		// contextual error data from earlier in
		// stack, setting current error to the top
		// and setting the unwrapped error to inner
		return &errorContext{
			data: xerr.data,
			innr: Unwrap(err),
			err:  err,
		}
	}

	// Return new Error type
	return &errorContext{
		data: NewData(),
		innr: nil,
		err:  err,
	}
}

// WrapMsg wraps supplied error as inner, returning an ErrorContext
// with a new outer error made from the supplied message string
func WrapMsg(err error, msg string) ErrorContext {
	// Nil error, do nothing
	if err == nil {
		return nil
	}

	// Check if this is already wrapped
	var xerr *errorContext
	if As(err, &xerr) {
		return &errorContext{
			data: xerr.data,
			innr: err,
			err:  New(msg),
		}
	}

	// Return new wrapped error
	return &errorContext{
		data: NewData(),
		innr: err,
		err:  stringError(msg),
	}
}

// WrapMsgf wraps supplied error as inner, returning an ErrorContext with
// a new outer error made from the supplied message format string
func WrapMsgf(err error, msg string, a ...interface{}) ErrorContext {
	return WrapMsg(err, fmt.Sprintf(msg, a...))
}

// ErrorData attempts fetch ErrorData from supplied error, returns nil otherwise
func Data(err error) ErrorData {
	x, ok := err.(ErrorContext)
	if ok {
		return x.Data()
	}
	return nil
}

// stringError is the simplest ErrorContext implementation
type stringError string

func (e stringError) Error() string {
	return string(e)
}

func (e stringError) Is(err error) bool {
	se, ok := err.(stringError)
	return ok && e == se
}

func (e stringError) Unwrap() error {
	return nil
}

func (e stringError) Value(key string) (interface{}, bool) {
	return nil, false
}

func (e stringError) Append(kvs ...KV) ErrorContext {
	data := NewData()
	data.Append(kvs...)
	return &errorContext{
		data: data,
		innr: nil,
		err:  e,
	}
}

func (e stringError) Data() ErrorData {
	return nil
}

// errorContext is the *actual* ErrorContext implementation
type errorContext struct {
	// data contains any appended context data, there will only ever be one
	// instance of data within an ErrorContext stack
	data ErrorData

	// innr is the inner wrapped error in this structure, it is only accessible
	// via .Unwrap() or via .Is()
	innr error

	// err is the top-level error in this wrapping structure, we identify
	// as this error type via .Is() and return its error message
	err error
}

func (e *errorContext) Error() string {
	return e.err.Error()
}

func (e *errorContext) Is(err error) bool {
	return Is(e.err, err) || Is(e.innr, err)
}

func (e *errorContext) Unwrap() error {
	return e.innr
}

func (e *errorContext) Value(key string) (interface{}, bool) {
	return e.data.Value(key)
}

func (e *errorContext) Append(kvs ...KV) ErrorContext {
	e.data.Append(kvs...)
	return e
}

func (e *errorContext) Data() ErrorData {
	return e.data
}
