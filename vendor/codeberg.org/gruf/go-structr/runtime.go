package structr

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"codeberg.org/gruf/go-mangler/v2"
	"codeberg.org/gruf/go-xunsafe"
)

// struct_field contains pre-prepared type
// information about a struct's field member,
// including memory offset and hash function.
type struct_field struct {

	// mangle ...
	mangle mangler.Mangler

	// zero value data, used when
	// nil encountered during ptr
	// offset following.
	zero unsafe.Pointer

	// mangled zero value string,
	// to check zero value keys.
	zerostr string

	// offsets defines whereabouts in
	// memory this field is located,
	// and after how many dereferences.
	offsets []next_offset
}

// next_offset defines a next offset location
// in a struct_field, first by the number of
// derefences required, then by offset from
// that final memory location.
type next_offset struct {
	derefs int
	offset uintptr
}

// get_type_iter returns a prepared xunsafe.TypeIter{} for generic parameter type,
// with flagIndir specifically set as we always take a reference to value type.
func get_type_iter[T any]() xunsafe.TypeIter {
	rtype := reflect.TypeOf((*T)(nil)).Elem()
	flags := xunsafe.Reflect_flag(xunsafe.Abi_Type_Kind(rtype))
	flags |= xunsafe.Reflect_flagIndir // always comes from unsafe ptr
	return xunsafe.ToTypeIter(rtype, flags)
}

// find_field will search for a struct field with given set of names,
// where names is a len > 0 slice of names account for struct nesting.
func find_field(t xunsafe.TypeIter, names []string) (sfield struct_field, ftype reflect.Type) {
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
				panic(fmt.Sprintf("field is not exported: %s", name))
			}
			return name
		}

		// field is the iteratively searched
		// struct field value in below loop.
		field reflect.StructField
	)

	for len(names) > 0 {
		// Pop next name.
		name := pop_name()

		var n int
		rtype := t.Type
		flags := t.Flag

		// Iteratively dereference pointer types.
		for rtype.Kind() == reflect.Pointer {

			// If this actual indirect memory,
			// increase dereferences counter.
			if flags&xunsafe.Reflect_flagIndir != 0 {
				n++
			}

			// Get next elem type.
			rtype = rtype.Elem()

			// Get next set of dereferenced element type flags.
			flags = xunsafe.ReflectPointerElemFlags(flags, rtype)

			// Update type iter info.
			t = t.Child(rtype, flags)
		}

		// Check for valid struct type.
		if rtype.Kind() != reflect.Struct {
			panic(fmt.Sprintf("field %s is not struct (or ptr-to): %s", rtype, name))
		}

		// Set offset info.
		var off next_offset
		off.derefs = n

		var ok bool

		// Look for the next field by name.
		field, ok = rtype.FieldByName(name)
		if !ok {
			panic(fmt.Sprintf("unknown field: %s", name))
		}

		// Set next offset value.
		off.offset = field.Offset
		sfield.offsets = append(sfield.offsets, off)

		// Calculate value flags, and set next nested field type.
		flags = xunsafe.ReflectStructFieldFlags(t.Flag, field.Type)
		t = t.Child(field.Type, flags)
	}

	// Set final field type.
	ftype = t.TypeInfo.Type

	// Get mangler from type info.
	sfield.mangle = mangler.Get(t)

	// Get field type as zero interface.
	v := reflect.New(t.Type).Elem()
	vi := v.Interface()

	// Get argument mangler from iface.
	ti := xunsafe.TypeIterFrom(vi)
	mangleArg := mangler.Get(ti)

	// Calculate zero value string.
	zptr := xunsafe.UnpackEface(vi)
	zstr := string(mangleArg(nil, zptr))
	sfield.zerostr = zstr
	sfield.zero = zptr

	return
}

// extract_fields extracts given structfields from the provided value type,
// this is done using predetermined struct field memory offset locations.
func extract_fields(ptr unsafe.Pointer, fields []struct_field) []unsafe.Pointer {

	// Prepare slice of field value pointers.
	ptrs := make([]unsafe.Pointer, len(fields))
	if len(ptrs) != len(fields) {
		panic(assert("BCE"))
	}

	for i, field := range fields {
		// loop scope.
		fptr := ptr

		for _, offset := range field.offsets {
			// Dereference any ptrs to offset.
			fptr = deref(fptr, offset.derefs)
			if fptr == nil {
				break
			}

			// Jump forward by offset to next ptr.
			fptr = unsafe.Pointer(uintptr(fptr) +
				offset.offset)
		}

		if fptr == nil {
			// Use zero value.
			fptr = field.zero
		}

		// Set field ptr.
		ptrs[i] = fptr
	}

	return ptrs
}

// pkey_field contains pre-prepared type
// information about a primary key struct's
// field member, including memory offset.
type pkey_field struct {

	// zero value data, used when
	// nil encountered during ptr
	// offset following.
	zero unsafe.Pointer

	// offsets defines whereabouts in
	// memory this field is located.
	offsets []next_offset
}

// extract_pkey will extract a pointer from 'ptr', to
// the primary key struct field defined by 'field'.
func extract_pkey(ptr unsafe.Pointer, field pkey_field) unsafe.Pointer {
	for _, offset := range field.offsets {

		// Dereference any ptrs to offset.
		ptr = deref(ptr, offset.derefs)
		if ptr == nil {
			break
		}

		// Jump forward by offset to next ptr.
		ptr = unsafe.Pointer(uintptr(ptr) +
			offset.offset)
	}

	if ptr == nil {
		// Use zero value.
		ptr = field.zero
	}

	return ptr
}

// deref will dereference ptr 'n' times (or until nil).
func deref(p unsafe.Pointer, n int) unsafe.Pointer {
	for ; n > 0; n-- {
		if p == nil {
			return nil
		}
		p = *(*unsafe.Pointer)(p)
	}
	return p
}

// assert can be called to indicated a block
// of code should not be able to be reached,
// it returns a BUG report with callsite.
func assert(assert string) string {
	pcs := make([]uintptr, 1)
	_ = runtime.Callers(2, pcs)
	funcname := "go-structr" // by default use just our library name
	if frames := runtime.CallersFrames(pcs); frames != nil {
		frame, _ := frames.Next()
		funcname = frame.Function
		if i := strings.LastIndexByte(funcname, '/'); i != -1 {
			funcname = funcname[i+1:]
		}
	}
	var buf strings.Builder
	buf.Grow(32 + len(assert) + len(funcname))
	buf.WriteString("BUG: assertion \"")
	buf.WriteString(assert)
	buf.WriteString("\" failed in ")
	buf.WriteString(funcname)
	return buf.String()
}
