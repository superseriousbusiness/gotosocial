package structr

import (
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-mangler"
)

// findField will search for a struct field with given set of names, where names is a len > 0 slice of names account for nesting.
func findField(t reflect.Type, names []string, allowZero bool) (sfield structfield, ok bool) {
	var (
		// isExported returns whether name is exported
		// from a package; can be func or struct field.
		isExported = func(name string) bool {
			r, _ := utf8.DecodeRuneInString(name)
			return unicode.IsUpper(r)
		}

		// popName pops the next name from
		// the provided slice of field names.
		popName = func() string {
			// Pop next name.
			name := names[0]
			names = names[1:]

			// Ensure valid name.
			if !isExported(name) {
				panicf("field is not exported: %s", name)
			}

			return name
		}

		// field is the iteratively searched-for
		// struct field value in below loop.
		field reflect.StructField
	)

	for len(names) > 0 {
		// Pop next name.
		name := popName()

		// Follow any ptrs leading to field.
		for t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		if t.Kind() != reflect.Struct {
			// The end type after following ptrs must be struct.
			panicf("field %s is not struct (ptr): %s", t, name)
		}

		// Look for next field by name.
		field, ok = t.FieldByName(name)
		if !ok {
			return
		}

		// Append next set of indices required to reach field.
		sfield.index = append(sfield.index, field.Index...)

		// Set the next type.
		t = field.Type
	}

	// Get final type mangler func.
	sfield.mangler = mangler.Get(t)

	if allowZero {
		var buf []byte

		// Allocate field instance.
		v := reflect.New(field.Type)
		v = v.Elem()

		// Serialize this zero value into buf.
		buf = sfield.mangler(buf, v.Interface())

		// Set zero value str.
		sfield.zero = string(buf)
	}

	return
}

// panicf provides a panic with string formatting.
func panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

// bufpool provides a memory pool of byte
// buffers used when encoding key types.
var bufPool sync.Pool

// getBuf fetches buffer from memory pool.
func getBuf() *byteutil.Buffer {
	v := bufPool.Get()
	if v == nil {
		buf := new(byteutil.Buffer)
		buf.B = make([]byte, 0, 512)
		v = buf
	}
	return v.(*byteutil.Buffer)
}

// putBuf replaces buffer in memory pool.
func putBuf(buf *byteutil.Buffer) {
	if buf.Cap() > int(^uint16(0)) {
		return // drop large bufs
	}
	buf.Reset()
	bufPool.Put(buf)
}
