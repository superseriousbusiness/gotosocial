package mangler

import (
	"reflect"
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

// iterSliceType returns a Mangler capable of iterating
// and mangling the given slice type currently in TypeIter{}.
// note this will fetch sub-Mangler for slice element type.
func iterSliceType(t xunsafe.TypeIter) Mangler {

	// Get nested element type.
	elem := t.Type.Elem()
	esz := elem.Size()

	// Get nested elem TypeIter{} with flags.
	flags := xunsafe.ReflectSliceElemFlags(elem)
	et := t.Child(elem, flags)

	// Prefer to use a known slice mangler func.
	if fn := mangleKnownSlice(et); fn != nil {
		return fn
	}

	// Get elem mangler.
	fn := loadOrGet(et)
	if fn == nil {
		return nil
	}

	return func(buf []byte, ptr unsafe.Pointer) []byte {
		// Get data as unsafe slice header.
		hdr := (*xunsafe.Unsafeheader_Slice)(ptr)
		if hdr == nil || hdr.Data == nil {

			// Append nil indicator.
			buf = append(buf, '0')
			return buf
		}

		// Append not-nil flag.
		buf = append(buf, '1')

		for i := 0; i < hdr.Len; i++ {
			// Mangle at array index.
			offset := esz * uintptr(i)
			ptr = add(hdr.Data, offset)
			buf = fn(buf, ptr)
			buf = append(buf, ',')
		}

		if hdr.Len > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

// mangleKnownSlice loads a Mangler function for a
// known slice-of-element type (in this case, primtives).
func mangleKnownSlice(t xunsafe.TypeIter) Mangler {
	switch t.Type.Kind() {
	case reflect.String:
		return mangle_string_slice
	case reflect.Bool:
		return mangle_bool_slice
	case reflect.Int,
		reflect.Uint,
		reflect.Uintptr:
		return mangle_int_slice
	case reflect.Int8, reflect.Uint8:
		return mangle_8bit_slice
	case reflect.Int16, reflect.Uint16:
		return mangle_16bit_slice
	case reflect.Int32, reflect.Uint32:
		return mangle_32bit_slice
	case reflect.Int64, reflect.Uint64:
		return mangle_64bit_slice
	case reflect.Float32:
		return mangle_32bit_slice
	case reflect.Float64:
		return mangle_64bit_slice
	case reflect.Complex64:
		return mangle_64bit_slice
	case reflect.Complex128:
		return mangle_128bit_slice
	default:
		return nil
	}
}
