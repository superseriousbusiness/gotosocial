package internal

func ErrWithKey(err error, key string) error {
	return &errorWithKey{key: key, err: err}
}

type errorWithKey struct {
	key string
	err error
}

func (err *errorWithKey) Error() string {
	return err.err.Error() + ": " + err.key
}

func (err *errorWithKey) Unwrap() error {
	return err.err
}

func ErrWithMsg(err error, msg string) error {
	return &errorWithMsg{msg: msg, err: err}
}

type errorWithMsg struct {
	msg string
	err error
}

func (err *errorWithMsg) Error() string {
	return err.msg + ": " + err.err.Error()
}

func (err *errorWithMsg) Unwrap() error {
	return err.err
}

func WrapErr(inner, outer error) error {
	return &wrappedError{inner: inner, outer: outer}
}

type wrappedError struct {
	inner error
	outer error
}

func (err *wrappedError) Is(other error) bool {
	return err.inner == other || err.outer == other
}

func (err *wrappedError) Error() string {
	return err.inner.Error() + ": " + err.outer.Error()
}

func (err *wrappedError) Unwrap() []error {
	return []error{err.inner, err.outer}
}
