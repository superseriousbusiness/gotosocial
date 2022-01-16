package errors

import (
	"fmt"
	"sync"

	"codeberg.org/gruf/go-format"
)

// KV is a structure for setting key-value pairs in ErrorData.
type KV struct {
	Key   string
	Value interface{}
}

// ErrorData defines a way to set and access contextual error data.
// The default implementation of this is thread-safe.
type ErrorData interface {
	// Value will attempt to fetch value for given key in ErrorData
	Value(string) (interface{}, bool)

	// Append adds the supplied key-values to ErrorData, similar keys DO overwrite
	Append(...KV)

	// Implement byte slice representation formatter.
	format.Formattable

	// Implement string representation formatter.
	fmt.Stringer
}

// NewData returns a new ErrorData implementation.
func NewData() ErrorData {
	return &errorData{
		data: make([]KV, 0, 10),
	}
}

// errorData is our ErrorData implementation, this is essentially
// just a thread-safe string-interface map implementation.
type errorData struct {
	data []KV
	mu   sync.Mutex
}

func (d *errorData) set(key string, value interface{}) {
	for i := range d.data {
		if d.data[i].Key == key {
			// Found existing, update!
			d.data[i].Value = value
			return
		}
	}

	// Add new KV entry to slice
	d.data = append(d.data, KV{
		Key:   key,
		Value: value,
	})
}

func (d *errorData) Value(key string) (interface{}, bool) {
	d.mu.Lock()
	for i := range d.data {
		if d.data[i].Key == key {
			v := d.data[i].Value
			d.mu.Unlock()
			return v, true
		}
	}
	d.mu.Unlock()
	return nil, false
}

func (d *errorData) Append(kvs ...KV) {
	d.mu.Lock()
	for i := range kvs {
		d.set(kvs[i].Key, kvs[i].Value)
	}
	d.mu.Unlock()
}

func (d *errorData) AppendFormat(b []byte) []byte {
	buf := format.Buffer{B: b}
	d.mu.Lock()
	buf.B = append(buf.B, '{')

	// Append data as kv pairs
	for i := range d.data {
		key := d.data[i].Key
		val := d.data[i].Value
		format.Appendf(&buf, "{:k}={:v} ", key, val)
	}

	// Drop trailing space
	if len(d.data) > 0 {
		buf.Truncate(1)
	}

	buf.B = append(buf.B, '}')
	d.mu.Unlock()
	return buf.B
}

func (d *errorData) String() string {
	return string(d.AppendFormat(nil))
}
