package structr

import (
	"reflect"
	"strings"

	"github.com/zeebo/xxh3"
)

// Hasher provides hash checksumming for a configured
// index, based on an arbitrary combination of generic
// paramter struct type's fields. This provides hashing
// both by input of the fields separately, or passing
// an instance of the generic paramter struct type.
//
// Supported field types by the hasher include:
// - ~int
// - ~int8
// - ~int16
// - ~int32
// - ~int64
// - ~float32
// - ~float64
// - ~string
// - slices / ptrs of the above
type Hasher[StructType any] struct {

	// fields contains our representation
	// of struct fields contained in the
	// creation of sums by this hasher.
	fields []structfield

	// zero specifies whether zero
	// value fields are permitted.
	zero bool
}

// NewHasher returns a new initialized Hasher for the receiving generic
// parameter type, comprising of the given field strings, and whether to
// allow zero values to be incldued within generated hash checksum values.
func NewHasher[T any](fields []string, allowZero bool) Hasher[T] {
	var h Hasher[T]

	// Preallocate expected struct field slice.
	h.fields = make([]structfield, len(fields))

	// Get the reflected struct ptr type.
	t := reflect.TypeOf((*T)(nil)).Elem()

	for i, fieldName := range fields {
		// Split name to account for nesting.
		names := strings.Split(fieldName, ".")

		// Look for a usable struct field from type.
		sfield, ok := findField(t, names, allowZero)
		if !ok {
			panicf("failed finding field: %s", fieldName)
		}

		// Set parsed struct field.
		h.fields[i] = sfield
	}

	// Set config flags.
	h.zero = allowZero

	return h
}

// FromParts generates hash checksum (used as index key) from individual key parts.
func (h *Hasher[T]) FromParts(parts ...any) (sum uint64, ok bool) {
	hh := getHasher()
	sum, ok = h.fromParts(hh, parts...)
	putHasher(hh)
	return

}

func (h *Hasher[T]) fromParts(hh *xxh3.Hasher, parts ...any) (sum uint64, ok bool) {
	if len(parts) != len(h.fields) {
		// User must provide correct number of parts for key.
		panicf("incorrect number key parts: want=%d received=%d",
			len(parts),
			len(h.fields),
		)
	}

	if h.zero {
		// Zero values are permitted,
		// mangle all values and ignore
		// zero value return booleans.
		for i, part := range parts {

			// Write mangled part to hasher.
			_ = h.fields[i].hasher(hh, part)
		}
	} else {
		// Zero values are NOT permitted.
		for i, part := range parts {

			// Write mangled field to hasher.
			z := h.fields[i].hasher(hh, part)

			if z {
				// The value was zero for
				// this type, return early.
				return 0, false
			}
		}
	}

	return hh.Sum64(), true
}

// FromValue generates hash checksum (used as index key) from a value, via reflection.
func (h *Hasher[T]) FromValue(value T) (sum uint64, ok bool) {
	rvalue := reflect.ValueOf(value)
	hh := getHasher()
	sum, ok = h.fromRValue(hh, rvalue)
	putHasher(hh)
	return
}

func (h *Hasher[T]) fromRValue(hh *xxh3.Hasher, rvalue reflect.Value) (uint64, bool) {
	// Follow any ptrs leading to value.
	for rvalue.Kind() == reflect.Pointer {
		rvalue = rvalue.Elem()
	}

	if h.zero {
		// Zero values are permitted,
		// mangle all values and ignore
		// zero value return booleans.
		for i := range h.fields {

			// Get the reflect value's field at idx.
			fv := rvalue.FieldByIndex(h.fields[i].index)
			fi := fv.Interface()

			// Write mangled field to hasher.
			_ = h.fields[i].hasher(hh, fi)
		}
	} else {
		// Zero values are NOT permitted.
		for i := range h.fields {

			// Get the reflect value's field at idx.
			fv := rvalue.FieldByIndex(h.fields[i].index)
			fi := fv.Interface()

			// Write mangled field to hasher.
			z := h.fields[i].hasher(hh, fi)

			if z {
				// The value was zero for
				// this type, return early.
				return 0, false
			}
		}
	}

	return hh.Sum64(), true
}

type structfield struct {
	// index is the reflected index
	// of this field (this takes into
	// account struct nesting).
	index []int

	// hasher is the relevant function
	// for hashing value of structfield
	// into the supplied hashbuf, where
	// return value indicates if zero.
	hasher func(*xxh3.Hasher, any) bool
}
