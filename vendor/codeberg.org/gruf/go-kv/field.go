package kv

import (
	"codeberg.org/gruf/go-byteutil"
)

// bufsize is the default buffer size per field to alloc
// when calling .AppendFormat() from within .String().
const bufsize = 64

// Fields is a typedef for a []Field slice to provide
// slightly more performant string formatting for multiples.
type Fields []Field

// Get will return the field with given 'key'.
func (f Fields) Get(key string) (*Field, bool) {
	for i := 0; i < len(f); i++ {
		if f[i].K == key {
			return &f[i], true
		}
	}
	return nil, false
}

// Set will set an existing field with 'key' to 'value', or append new.
func (f *Fields) Set(key string, value interface{}) {
	for i := 0; i < len(*f); i++ {
		// Update existing value
		if (*f)[i].K == key {
			(*f)[i].V = value
			return
		}
	}

	// Append new field
	*f = append(*f, Field{
		K: key,
		V: value,
	})
}

// AppendFormat appends a string representation of receiving Field(s) to 'b'.
func (f Fields) AppendFormat(buf *byteutil.Buffer, vbose bool) {
	for i := 0; i < len(f); i++ {
		f[i].AppendFormat(buf, vbose)
		buf.WriteByte(' ')
	}
	if len(f) > 0 {
		buf.Truncate(1)
	}
}

// String returns a string representation of receiving Field(s).
func (f Fields) String() string {
	b := make([]byte, 0, bufsize*len(f))
	buf := byteutil.Buffer{B: b}
	f.AppendFormat(&buf, false)
	return buf.String()
}

// GoString performs .String() but with type prefix.
func (f Fields) GoString() string {
	b := make([]byte, 0, bufsize*len(f))
	buf := byteutil.Buffer{B: b}
	f.AppendFormat(&buf, true)
	return "kv.Fields{" + buf.String() + "}"
}

// Field represents an individual key-value field.
type Field struct {
	K string      // Field key
	V interface{} // Field value
}

// Key returns the formatted key string of this Field.
func (f Field) Key() string {
	buf := byteutil.Buffer{B: make([]byte, 0, bufsize/2)}
	AppendQuote(&buf, f.K)
	return buf.String()
}

// String will return a string representation of this Field
// of the form `key=value` where `value` is formatted using
// fmt package's `%+v` directive. If the .X = true (verbose),
// then it uses '%#v'. Both key and value are escaped and
// quoted if necessary to fit on single line.
//
// If the `kvformat` build tag is provided, the formatting
// will be performed by the `kv/format` package. In this case
// the value will be formatted using the `{:v}` directive, or
// `{:?}` if .X = true (verbose).
func (f Field) String() string {
	b := make([]byte, 0, bufsize)
	buf := byteutil.Buffer{B: b}
	f.AppendFormat(&buf, false)
	return buf.String()
}

// GoString performs .String() but with verbose always enabled.
func (f Field) GoString() string {
	b := make([]byte, 0, bufsize)
	buf := byteutil.Buffer{B: b}
	f.AppendFormat(&buf, true)
	return buf.String()
}
