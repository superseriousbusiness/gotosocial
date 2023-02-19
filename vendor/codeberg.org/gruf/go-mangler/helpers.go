package mangler

import (
	"reflect"
	"unsafe"
)

type (
	// serializing interfacing types.
	stringer        interface{ String() string }
	binarymarshaler interface{ MarshalBinary() ([]byte, error) }
	textmarshaler   interface{ MarshalText() ([]byte, error) }
	jsonmarshaler   interface{ MarshalJSON() ([]byte, error) }
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

func deref_ptr_mangler(mangle Mangler, count int) rMangler {
	return func(buf []byte, v reflect.Value) []byte {
		for i := 0; i < count; i++ {
			// Check for nil
			if v.IsNil() {
				buf = append(buf, '0')
				return buf
			}

			// Further deref ptr
			buf = append(buf, '1')
			v = v.Elem()
		}

		// Mangle fully deref'd ptr
		return mangle(buf, v.Interface())
	}
}

func deref_ptr_rmangler(mangle rMangler, count int) rMangler {
	return func(buf []byte, v reflect.Value) []byte {
		for i := 0; i < count; i++ {
			// Check for nil
			if v.IsNil() {
				buf = append(buf, '0')
				return buf
			}

			// Further deref ptr
			buf = append(buf, '1')
			v = v.Elem()
		}

		// Mangle fully deref'd ptr
		return mangle(buf, v)
	}
}

func array_to_slice_mangler(mangle Mangler) rMangler {
	return func(buf []byte, v reflect.Value) []byte {
		// Get slice of whole array
		v = v.Slice(0, v.Len())

		// Mangle as known slice type
		return mangle(buf, v.Interface())
	}
}

func iter_array_mangler(mangle Mangler) rMangler {
	return func(buf []byte, v reflect.Value) []byte {
		n := v.Len()
		for i := 0; i < n; i++ {
			buf = mangle(buf, v.Index(i).Interface())
			buf = append(buf, ',')
		}
		if n > 0 {
			buf = buf[:len(buf)-1]
		}
		return buf
	}
}

func iter_array_rmangler(mangle rMangler) rMangler {
	return func(buf []byte, v reflect.Value) []byte {
		n := v.Len()
		for i := 0; i < n; i++ {
			buf = mangle(buf, v.Index(i))
			buf = append(buf, ',')
		}
		if n > 0 {
			buf = buf[:len(buf)-1]
		}
		return buf
	}
}

func iter_map_rmangler(kMangle, vMangle rMangler) rMangler {
	return func(buf []byte, v reflect.Value) []byte {
		r := v.MapRange()
		for r.Next() {
			buf = kMangle(buf, r.Key())
			buf = append(buf, ':')
			buf = vMangle(buf, r.Value())
			buf = append(buf, ',')
		}
		if v.Len() > 0 {
			buf = buf[:len(buf)-1]
		}
		return buf
	}
}

// iface_value returns the raw value ptr for input boxed within interface{} type.
func iface_value(a any) unsafe.Pointer {
	type eface struct {
		Type  unsafe.Pointer
		Value unsafe.Pointer
	}
	return (*eface)(unsafe.Pointer(&a)).Value
}
