package format

import (
	"reflect"
	"unsafe"
)

// derefPointerType returns a FormatFunc capable of dereferencing
// and formatting the given pointer type currently in typenode{}.
// note this will fetch a sub-FormatFunc for resulting value type.
func (fmt *Formatter) derefPointerType(t typenode) FormatFunc {
	var n int
	rtype := t.rtype
	flags := t.flags

	// Iteratively dereference pointer types.
	for rtype.Kind() == reflect.Pointer {

		// If this is actual indirect
		// memory, increase dereferences.
		if flags&reflect_flagIndir != 0 {
			n++
		}

		// Get next elem type.
		rtype = rtype.Elem()

		// Get next set of dereferenced elem type flags.
		flags = reflect_pointer_elem_flags(flags, rtype)
	}

	// Wrap value as typenode.
	vt := t.next(rtype, flags)

	// Get value format func.
	fn := fmt.loadOrGet(vt)
	if fn == nil {
		panic("unreachable")
	}

	if !t.needs_typestr() {
		if n <= 0 {
			// No derefs are needed.
			return func(s *State) {
				if s.P == nil {
					// Final check.
					appendNil(s)
					return
				}

				// Format
				// final
				// value.
				fn(s)
			}
		}

		return func(s *State) {
			// Deref n number times.
			for i := n; i > 0; i-- {

				if s.P == nil {
					// Nil check.
					appendNil(s)
					return
				}

				// Further deref pointer value.
				s.P = *(*unsafe.Pointer)(s.P)
			}

			if s.P == nil {
				// Final check.
				appendNil(s)
				return
			}

			// Format
			// final
			// value.
			fn(s)
		}
	}

	// Final type string with ptrs.
	typestr := t.typestr_with_ptrs()

	if n <= 0 {
		// No derefs are needed.
		return func(s *State) {
			if s.P == nil {
				// Final nil value check.
				appendNilType(s, typestr)
				return
			}

			// Format
			// final
			// value.
			fn(s)
		}
	}

	return func(s *State) {
		// Deref n number times.
		for i := n; i > 0; i-- {
			if s.P == nil {
				// Check for nil value.
				appendNilType(s, typestr)
				return
			}

			// Further deref pointer value.
			s.P = *(*unsafe.Pointer)(s.P)
		}

		if s.P == nil {
			// Final nil value check.
			appendNilType(s, typestr)
			return
		}

		// Format
		// final
		// value.
		fn(s)
	}
}
