package mangler

import (
	"reflect"
	"unsafe"
)

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
			buf = append(buf, '.')
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
