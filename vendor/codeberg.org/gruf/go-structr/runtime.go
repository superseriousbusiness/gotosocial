package structr

import (
	"fmt"
	"reflect"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"codeberg.org/gruf/go-mangler"
	"github.com/modern-go/reflect2"
)

// struct_field contains pre-prepared type
// information about a struct's field member,
// including memory offset and hash function.
type struct_field struct {

	// type2 is the runtime type pointer
	// underlying the struct field type.
	// used for repacking our own erfaces.
	type2 reflect2.Type

	// offset is the offset in memory
	// of this struct field from the
	// outer-most value ptr location.
	offset uintptr

	// struct field type mangling
	// (i.e. fast serializing) fn.
	mangle mangler.Mangler

	// mangled zero value string,
	// if set this indicates zero
	// values of field not allowed
	zero string
}

// find_field will search for a struct field with given set of names,
// where names is a len > 0 slice of names account for struct nesting.
func find_field(t reflect.Type, names []string) (sfield struct_field) {
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
	sfield.type2 = reflect2.Type2(t)

	// Find mangler for field type.
	sfield.mangle = mangler.Get(t)

	return
}

// extract_fields extracts given structfields from the provided value type,
// this is done using predetermined struct field memory offset locations.
func extract_fields[T any](value T, fields []struct_field) []any {
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
		ifaces[i] = fields[i].type2.UnsafeIndirect(ptr)
	}

	return ifaces
}

// panicf provides a panic with string formatting.
func panicf(format string, args ...any) {
	panic(fmt.Sprintf(format, args...))
}
