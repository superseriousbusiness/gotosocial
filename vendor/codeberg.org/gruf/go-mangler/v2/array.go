package mangler

import (
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

// iterArrayType returns a Mangler capable of iterating
// and mangling the given array type currently in TypeIter{}.
// note this will fetch sub-Mangler for array element type.
func iterArrayType(t xunsafe.TypeIter) Mangler {

	// Array element type.
	elem := t.Type.Elem()

	// Get nested elem TypeIter with appropriate flags.
	flags := xunsafe.ReflectArrayElemFlags(t.Flag, elem)
	et := t.Child(elem, flags)

	// Get elem mangler.
	fn := loadOrGet(et)
	if fn == nil {
		return nil
	}

	// Array element in-memory size.
	esz := t.Type.Elem().Size()

	// No of elements.
	n := t.Type.Len()
	switch n {
	case 0:
		return empty_mangler
	case 1:
		return fn
	default:
		return func(buf []byte, ptr unsafe.Pointer) []byte {
			for i := 0; i < n; i++ {
				// Mangle data at array index.
				offset := esz * uintptr(i)
				eptr := add(ptr, offset)
				buf = fn(buf, eptr)
				buf = append(buf, ',')
			}

			if n > 0 {
				// Drop final comma.
				buf = buf[:len(buf)-1]
			}

			return buf
		}
	}
}
