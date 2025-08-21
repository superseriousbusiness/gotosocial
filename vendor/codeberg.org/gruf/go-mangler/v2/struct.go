package mangler

import (
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

// field stores the minimum necessary
// data for iterating and mangling
// each field in a given struct.
type field struct {
	mangle Mangler
	offset uintptr
}

// iterStructType returns a Mangler capable of iterating
// and mangling the given struct type currently in TypeIter{}.
// note this will fetch sub-Manglers for each struct field.
func iterStructType(t xunsafe.TypeIter) Mangler {

	// Number of struct fields.
	n := t.Type.NumField()

	// Gather mangler functions.
	fields := make([]field, n)
	for i := 0; i < n; i++ {

		// Get struct field at index.
		sfield := t.Type.Field(i)
		rtype := sfield.Type

		// Get nested field TypeIter with appropriate flags.
		flags := xunsafe.ReflectStructFieldFlags(t.Flag, rtype)
		ft := t.Child(sfield.Type, flags)

		// Get field mangler.
		fn := loadOrGet(ft)
		if fn == nil {
			return nil
		}

		// Set field info.
		fields[i] = field{
			mangle: fn,
			offset: sfield.Offset,
		}
	}

	// Handle no. fields.
	switch len(fields) {
	case 0:
		return empty_mangler
	case 1:
		return fields[0].mangle
	default:
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
}
