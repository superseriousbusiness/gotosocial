package errors

// WithValue wraps err to store given key-value pair, accessible via Value() function.
func WithValue(err error, key any, value any) error {
	if err == nil {
		panic("nil error")
	}
	var kvs []kv
	if e := AsV2[*errWithValues](err); e != nil {
		kvs = e.kvs
	}
	return &errWithValues{
		err: err,
		kvs: append(kvs, kv{key, value}),
	}
}

// Value searches for value stored under given key in error chain.
func Value(err error, key any) any {
	if e := AsV2[*errWithValues](err); e != nil {
		return e.Value(key)
	}
	return nil
}

// simple key-value type.
type kv struct{ k, v any }

// errWithValues wraps an error to provide key-value storage.
type errWithValues struct {
	err error
	kvs []kv
}

func (e *errWithValues) Error() string {
	return e.err.Error()
}

func (e *errWithValues) Unwrap() error {
	return e.err
}

func (e *errWithValues) Value(key any) any {
	for i := range e.kvs {
		if e.kvs[i].k == key {
			return e.kvs[i].v
		}
	}
	return nil
}
