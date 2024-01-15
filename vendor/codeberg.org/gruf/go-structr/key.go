package structr

import (
	"reflect"
	"strings"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-mangler"
)

// KeyGen is the underlying index key generator
// used within Index, and therefore Cache itself.
type KeyGen[StructType any] struct {

	// fields contains our representation of
	// the struct fields contained in the
	// creation of keys by this generator.
	fields []structfield

	// zero specifies whether zero
	// value fields are permitted.
	zero bool
}

// NewKeyGen returns a new initialized KeyGen for the receiving generic
// parameter type, comprising of the given field strings, and whether to
// allow zero values to be included within generated output strings.
func NewKeyGen[T any](fields []string, allowZero bool) KeyGen[T] {
	var kgen KeyGen[T]

	// Preallocate expected struct field slice.
	kgen.fields = make([]structfield, len(fields))

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
		kgen.fields[i] = sfield
	}

	// Set config flags.
	kgen.zero = allowZero

	return kgen
}

// FromParts generates key string from individual key parts.
func (kgen *KeyGen[T]) FromParts(parts ...any) (key string, ok bool) {
	buf := getBuf()
	if ok = kgen.AppendFromParts(buf, parts...); ok {
		key = string(buf.B)
	}
	putBuf(buf)
	return
}

// FromValue generates key string from a value, via reflection.
func (kgen *KeyGen[T]) FromValue(value T) (key string, ok bool) {
	buf := getBuf()
	rvalue := reflect.ValueOf(value)
	if ok = kgen.appendFromRValue(buf, rvalue); ok {
		key = string(buf.B)
	}
	putBuf(buf)
	return
}

// AppendFromParts generates key string into provided buffer, from individual key parts.
func (kgen *KeyGen[T]) AppendFromParts(buf *byteutil.Buffer, parts ...any) bool {
	if len(parts) != len(kgen.fields) {
		// User must provide correct number of parts for key.
		panicf("incorrect number key parts: want=%d received=%d",
			len(parts),
			len(kgen.fields),
		)
	}

	if kgen.zero {
		// Zero values are permitted,
		// mangle all values and ignore
		// zero value return booleans.
		for i, part := range parts {

			// Mangle this value into buffer.
			_ = kgen.fields[i].Mangle(buf, part)

			// Append part separator.
			buf.B = append(buf.B, '.')
		}
	} else {
		// Zero values are NOT permitted.
		for i, part := range parts {

			// Mangle this value into buffer.
			z := kgen.fields[i].Mangle(buf, part)

			if z {
				// The value was zero for
				// this type, return early.
				return false
			}

			// Append part separator.
			buf.B = append(buf.B, '.')
		}
	}

	// Drop the last separator.
	buf.B = buf.B[:len(buf.B)-1]

	return true
}

// AppendFromValue generates key string into provided buffer, from a value via reflection.
func (kgen *KeyGen[T]) AppendFromValue(buf *byteutil.Buffer, value T) bool {
	return kgen.appendFromRValue(buf, reflect.ValueOf(value))
}

// appendFromRValue is the underlying generator function for the exported ___FromValue() functions,
// accepting a reflected input. We do not expose this as the reflected value is EXPECTED to be right.
func (kgen *KeyGen[T]) appendFromRValue(buf *byteutil.Buffer, rvalue reflect.Value) bool {
	// Follow any ptrs leading to value.
	for rvalue.Kind() == reflect.Pointer {
		rvalue = rvalue.Elem()
	}

	if kgen.zero {
		// Zero values are permitted,
		// mangle all values and ignore
		// zero value return booleans.
		for i := range kgen.fields {

			// Get the reflect value's field at idx.
			fv := rvalue.FieldByIndex(kgen.fields[i].index)
			fi := fv.Interface()

			// Mangle this value into buffer.
			_ = kgen.fields[i].Mangle(buf, fi)

			// Append part separator.
			buf.B = append(buf.B, '.')
		}
	} else {
		// Zero values are NOT permitted.
		for i := range kgen.fields {

			// Get the reflect value's field at idx.
			fv := rvalue.FieldByIndex(kgen.fields[i].index)
			fi := fv.Interface()

			// Mangle this value into buffer.
			z := kgen.fields[i].Mangle(buf, fi)

			if z {
				// The value was zero for
				// this type, return early.
				return false
			}

			// Append part separator.
			buf.B = append(buf.B, '.')
		}
	}

	// Drop the last separator.
	buf.B = buf.B[:len(buf.B)-1]

	return true
}

type structfield struct {
	// index is the reflected index
	// of this field (this takes into
	// account struct nesting).
	index []int

	// zero is the possible mangled
	// zero value for this field.
	zero string

	// mangler is the mangler function for
	// serializing values of this field.
	mangler mangler.Mangler
}

// Mangle mangles the given value, using the determined type-appropriate
// field's type. The returned boolean indicates whether this is a zero value.
func (f *structfield) Mangle(buf *byteutil.Buffer, value any) (isZero bool) {
	s := len(buf.B) // start pos.
	buf.B = f.mangler(buf.B, value)
	e := len(buf.B) // end pos.
	isZero = (f.zero == string(buf.B[s:e]))
	return
}
