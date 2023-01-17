package errors

// WithValue wraps err to store given key-value pair, accessible via Value() function.
func WithValue(err error, key any, value any) error {
	if err == nil {
		panic("nil error")
	}
	return &errWithValue{
		err: err,
		key: key,
		val: value,
	}
}

// Value searches for value stored under given key in error chain.
func Value(err error, key any) any {
	var e *errWithValue

	if !As(err, &e) {
		return nil
	}

	return e.Value(key)
}

type errWithValue struct {
	err error
	key any
	val any
}

func (e *errWithValue) Error() string {
	return e.err.Error()
}

func (e *errWithValue) Is(target error) bool {
	return e.err == target
}

func (e *errWithValue) Unwrap() error {
	return Unwrap(e.err)
}

func (e *errWithValue) Value(key any) any {
	for {
		if key == e.key {
			return e.val
		}

		if !As(e.err, &e) {
			return nil
		}
	}
}
