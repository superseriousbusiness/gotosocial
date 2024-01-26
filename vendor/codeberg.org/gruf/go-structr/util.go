package structr

import (
	"fmt"
	"reflect"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/zeebo/xxh3"
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

	// Get final type hash func.
	sfield.hasher = hasher(t)

	return
}

// panicf provides a panic with string formatting.
func panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}

// hashPool provides a memory pool of xxh3
// hasher objects used indexing field vals.
var hashPool sync.Pool

// gethashbuf fetches hasher from memory pool.
func getHasher() *xxh3.Hasher {
	v := hashPool.Get()
	if v == nil {
		v = new(xxh3.Hasher)
	}
	return v.(*xxh3.Hasher)
}

// putHasher replaces hasher in memory pool.
func putHasher(h *xxh3.Hasher) {
	h.Reset()
	hashPool.Put(h)
}
