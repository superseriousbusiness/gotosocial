package structr

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/modern-go/reflect2"
	"github.com/zeebo/xxh3"
)

type structfield struct {
	// _type is the runtime type pointer
	// underlying the struct field type.
	// used for repacking our own erfaces.
	_type reflect2.Type

	// offset is the offset in memory
	// of this struct field from the
	// outer-most value ptr location.
	offset uintptr

	// hasher is the relevant function
	// for hashing value of structfield
	// into the supplied hashbuf, where
	// return value indicates if zero.
	hasher func(*xxh3.Hasher, any) bool
}

// find_field will search for a struct field with given set of names,
// where names is a len > 0 slice of names account for struct nesting.
func find_field(t reflect.Type, names []string) (sfield structfield) {
	var (
		// is_exported returns whether name is exported
		// from a package; can be func or struct field.
		is_exported = func(name string) bool {
			r, _ := utf8.DecodeRuneInString(name)
			return unicode.IsUpper(r)
		}

		// pop_name pops the next name from
		// the provided slice of field names.
		pop_name = func() string {
			name := names[0]
			names = names[1:]
			if !is_exported(name) {
				panicf("field is not exported: %s", name)
			}
			return name
		}

		// field is the iteratively searched
		// struct field value in below loop.
		field reflect.StructField
	)

	switch {
	// The only 2 types we support are
	// structs, and ptrs to a struct.
	case t.Kind() == reflect.Struct:
	case t.Kind() == reflect.Pointer &&
		t.Elem().Kind() == reflect.Struct:
		t = t.Elem()
	default:
		panic("index only support struct{} and *struct{}")
	}

	for len(names) > 0 {
		var ok bool

		// Pop next name.
		name := pop_name()

		// Check for valid struct type.
		if t.Kind() != reflect.Struct {
			panicf("field %s is not struct: %s", t, name)
		}

		// Look for next field by name.
		field, ok = t.FieldByName(name)
		if !ok {
			panicf("unknown field: %s", name)
		}

		// Increment total field offset.
		sfield.offset += field.Offset

		// Set the next type.
		t = field.Type
	}

	// Get field type as reflect2.
	sfield._type = reflect2.Type2(t)

	// Find hasher for type.
	sfield.hasher = hasher(t)

	return
}

// extract_fields extracts given structfields from the provided value type,
// this is done using predetermined struct field memory offset locations.
func extract_fields[T any](value T, fields []structfield) []any {
	// Get ptr to raw value data.
	ptr := unsafe.Pointer(&value)

	// If this is a pointer type deref the value ptr.
	if reflect.TypeOf(value).Kind() == reflect.Pointer {
		ptr = *(*unsafe.Pointer)(ptr)
	}

	// Prepare slice of field ifaces.
	ifaces := make([]any, len(fields))

	for i := 0; i < len(fields); i++ {
		// Manually access field at memory offset and pack eface.
		ptr := unsafe.Pointer(uintptr(ptr) + fields[i].offset)
		ifaces[i] = fields[i]._type.UnsafeIndirect(ptr)
	}

	return ifaces
}

// data_ptr returns the runtime data ptr associated with value.
func data_ptr(a any) unsafe.Pointer {
	return (*struct{ t, v unsafe.Pointer })(unsafe.Pointer(&a)).v
}

// panicf provides a panic with string formatting.
func panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}
