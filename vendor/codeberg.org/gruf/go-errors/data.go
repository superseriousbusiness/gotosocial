package errors

import (
	"sync"

	"codeberg.org/gruf/go-bytes"
	"codeberg.org/gruf/go-logger"
)

// global logfmt data formatter
var logfmt = logger.TextFormat{Strict: false}

// KV is a structure for setting key-value pairs in ErrorData
type KV struct {
	Key   string
	Value interface{}
}

// ErrorData defines a way to set and access contextual error data.
// The default implementation of this is thread-safe
type ErrorData interface {
	// Value will attempt to fetch value for given key in ErrorData
	Value(string) (interface{}, bool)

	// Append adds the supplied key-values to ErrorData, similar keys DO overwrite
	Append(...KV)

	// String returns a string representation of the ErrorData
	String() string
}

// NewData returns a new ErrorData implementation
func NewData() ErrorData {
	return &errorData{
		data: make(map[string]interface{}, 10),
	}
}

// errorData is our ErrorData implementation, this is essentially
// just a thread-safe string-interface map implementation
type errorData struct {
	data map[string]interface{}
	buf  bytes.Buffer
	mu   sync.Mutex
}

func (d *errorData) Value(key string) (interface{}, bool) {
	d.mu.Lock()
	v, ok := d.data[key]
	d.mu.Unlock()
	return v, ok
}

func (d *errorData) Append(kvs ...KV) {
	d.mu.Lock()
	for i := range kvs {
		k := kvs[i].Key
		v := kvs[i].Value
		d.data[k] = v
	}
	d.mu.Unlock()
}

func (d *errorData) String() string {
	d.mu.Lock()

	d.buf.Reset()
	d.buf.B = append(d.buf.B, '{')
	logfmt.AppendFields(&d.buf, d.data)
	d.buf.B = append(d.buf.B, '}')

	d.mu.Unlock()
	return d.buf.StringPtr()
}
