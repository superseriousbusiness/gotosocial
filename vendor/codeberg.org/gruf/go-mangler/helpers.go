//go:build go1.19 && !go1.25

package mangler

import (
	"reflect"
	"unsafe"
)

func append_uint16(b []byte, u uint16) []byte {
	return append(b, // LE
		byte(u),
		byte(u>>8),
	)
}

func append_uint32(b []byte, u uint32) []byte {
	return append(b, // LE
		byte(u),
		byte(u>>8),
		byte(u>>16),
		byte(u>>24),
	)
}

func append_uint64(b []byte, u uint64) []byte {
	return append(b, // LE
		byte(u),
		byte(u>>8),
		byte(u>>16),
		byte(u>>24),
		byte(u>>32),
		byte(u>>40),
		byte(u>>48),
		byte(u>>56),
	)
}

type typecontext struct {
	isptr  bool
	direct bool
	ntype  reflect.Type
	rtype  reflect.Type
}

func (ctx *typecontext) set_nested(direct bool) {
	ctx.direct = ctx.direct && direct && !ctx.isptr
	ctx.ntype = ctx.rtype
	ctx.rtype = nil
	ctx.isptr = false
}

func deref_ptr_mangler(ctx typecontext, mangle Mangler, n uint) Mangler {
	if mangle == nil || n == 0 {
		panic("bad input")
	}

	// If this is a direct value type, i.e. non-nested primitive,
	// or part of a single-field struct / single element array
	// then it can be treated as a direct ptr with 1 less deref.
	if ctx.direct {
		n--
	}

	return func(buf []byte, ptr unsafe.Pointer) []byte {
		// Deref n number times.
		for i := n; i > 0; i-- {

			if ptr == nil {
				// Check for nil values
				buf = append(buf, '0')
				return buf
			}

			// Further deref ptr
			buf = append(buf, '1')
			ptr = *(*unsafe.Pointer)(ptr)
		}

		if ptr == nil {
			// Final nil val check
			buf = append(buf, '0')
			return buf
		}

		// Mangle fully deref'd
		buf = append(buf, '1')
		buf = mangle(buf, ptr)
		return buf
	}
}

func iter_slice_mangler(ctx typecontext, mangle Mangler) Mangler {
	if ctx.rtype == nil || mangle == nil {
		panic("bad input")
	}

	// memory size of elem.
	esz := ctx.rtype.Size()

	return func(buf []byte, ptr unsafe.Pointer) []byte {
		// Get data as slice hdr.
		hdr := (*slice_header)(ptr)

		for i := 0; i < hdr.len; i++ {
			// Mangle data at slice index.
			eptr := array_at(hdr.data, esz, i)
			buf = mangle(buf, eptr)
			buf = append(buf, ',')
		}

		if hdr.len > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func iter_array_mangler(ctx typecontext, mangle Mangler) Mangler {
	if ctx.rtype == nil || mangle == nil {
		panic("bad input")
	}

	// no. array elements.
	n := ctx.ntype.Len()

	// Optimize
	// easy cases.
	switch n {
	case 0:
		return empty_mangler
	case 1:
		return mangle
	}

	// memory size of elem.
	esz := ctx.rtype.Size()

	return func(buf []byte, ptr unsafe.Pointer) []byte {
		for i := 0; i < n; i++ {
			// Mangle data at array index.
			offset := esz * uintptr(i)
			eptr := add(ptr, offset)
			buf = mangle(buf, eptr)
			buf = append(buf, ',')
		}

		if n > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func iter_struct_mangler(ctx typecontext, manglers []Mangler) Mangler {
	if ctx.rtype == nil || len(manglers) != ctx.ntype.NumField() {
		panic("bad input")
	}

	// Optimized easy cases.
	switch len(manglers) {
	case 0:
		return empty_mangler
	case 1:
		return manglers[0]
	}

	type field struct {
		mangle Mangler
		offset uintptr
	}

	// Bundle together the fields and manglers.
	fields := make([]field, ctx.ntype.NumField())
	for i := range fields {
		rfield := ctx.ntype.Field(i)
		fields[i].offset = rfield.Offset
		fields[i].mangle = manglers[i]
		if fields[i].mangle == nil {
			panic("bad input")
		}
	}

	return func(buf []byte, ptr unsafe.Pointer) []byte {
		for i := range fields {
			// Get struct field ptr via offset.
			fptr := add(ptr, fields[i].offset)

			// Mangle the struct field data.
			buf = fields[i].mangle(buf, fptr)
			buf = append(buf, ',')
		}

		if len(fields) > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func empty_mangler(buf []byte, _ unsafe.Pointer) []byte {
	return buf
}

// array_at returns ptr to index in array at ptr, given element size.
func array_at(ptr unsafe.Pointer, esz uintptr, i int) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + esz*uintptr(i))
}

// add returns the ptr addition of starting ptr and a delta.
func add(ptr unsafe.Pointer, delta uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + delta)
}

type slice_header struct {
	data unsafe.Pointer
	len  int
	cap  int
}

func eface_data(a any) unsafe.Pointer {
	type eface struct{ _, data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&a)).data
}
