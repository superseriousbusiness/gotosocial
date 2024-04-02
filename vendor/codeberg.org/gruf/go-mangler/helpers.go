package mangler

import (
	"reflect"
	"unsafe"

	"github.com/modern-go/reflect2"
)

type (
	byteser         interface{ Bytes() []byte }
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

func deref_ptr_mangler(rtype reflect.Type, mangle Mangler, count int) Mangler {
	if rtype == nil || mangle == nil || count == 0 {
		panic("bad input")
	}

	// Get reflect2's type for later
	// unsafe interface data repacking,
	type2 := reflect2.Type2(rtype)

	return func(buf []byte, value any) []byte {
		// Get raw value data.
		ptr := eface_data(value)

		// Deref n - 1 number times.
		for i := 0; i < count-1; i++ {

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
			// Final nil value check.
			buf = append(buf, '0')
			return buf
		}

		// Repack and mangle fully deref'd
		value = type2.UnsafeIndirect(ptr)
		buf = append(buf, '1')
		return mangle(buf, value)
	}
}

func iter_slice_mangler(rtype reflect.Type, mangle Mangler) Mangler {
	if rtype == nil || mangle == nil {
		panic("bad input")
	}

	// Get reflect2's type for later
	// unsafe slice data manipulation.
	slice2 := reflect2.Type2(rtype).(*reflect2.UnsafeSliceType)

	return func(buf []byte, value any) []byte {
		// Get raw value data.
		ptr := eface_data(value)

		// Get length of slice value.
		n := slice2.UnsafeLengthOf(ptr)

		for i := 0; i < n; i++ {
			// Mangle data at each slice index.
			e := slice2.UnsafeGetIndex(ptr, i)
			buf = mangle(buf, e)
			buf = append(buf, ',')
		}

		if n > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func iter_array_mangler(rtype reflect.Type, mangle Mangler) Mangler {
	if rtype == nil || mangle == nil {
		panic("bad input")
	}

	// Get reflect2's type for later
	// unsafe slice data manipulation.
	array2 := reflect2.Type2(rtype).(*reflect2.UnsafeArrayType)
	n := array2.Len()

	return func(buf []byte, value any) []byte {
		// Get raw value data.
		ptr := eface_data(value)

		for i := 0; i < n; i++ {
			// Mangle data at each slice index.
			e := array2.UnsafeGetIndex(ptr, i)
			buf = mangle(buf, e)
			buf = append(buf, ',')
		}

		if n > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func iter_map_mangler(rtype reflect.Type, kmangle, emangle Mangler) Mangler {
	if rtype == nil || kmangle == nil || emangle == nil {
		panic("bad input")
	}

	// Get reflect2's type for later
	// unsafe map data manipulation.
	map2 := reflect2.Type2(rtype).(*reflect2.UnsafeMapType)
	key2, elem2 := map2.Key(), map2.Elem()

	return func(buf []byte, value any) []byte {
		// Get raw value data.
		ptr := eface_data(value)
		ptr = indirect_ptr(ptr)

		// Create iterator for map value.
		iter := map2.UnsafeIterate(ptr)

		// Check if empty map.
		empty := !iter.HasNext()

		for iter.HasNext() {
			// Get key + elem data as ifaces.
			kptr, eptr := iter.UnsafeNext()
			key := key2.UnsafeIndirect(kptr)
			elem := elem2.UnsafeIndirect(eptr)

			// Mangle data for key + elem.
			buf = kmangle(buf, key)
			buf = append(buf, ':')
			buf = emangle(buf, elem)
			buf = append(buf, ',')
		}

		if !empty {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func iter_struct_mangler(rtype reflect.Type, manglers []Mangler) Mangler {
	if rtype == nil || len(manglers) != rtype.NumField() {
		panic("bad input")
	}

	type field struct {
		type2  reflect2.Type
		field  *reflect2.UnsafeStructField
		mangle Mangler
	}

	// Get reflect2's type for later
	// unsafe struct field data access.
	struct2 := reflect2.Type2(rtype).(*reflect2.UnsafeStructType)

	// Bundle together the fields and manglers.
	fields := make([]field, rtype.NumField())
	for i := range fields {
		fields[i].field = struct2.Field(i).(*reflect2.UnsafeStructField)
		fields[i].type2 = fields[i].field.Type()
		fields[i].mangle = manglers[i]
		if fields[i].type2 == nil ||
			fields[i].field == nil ||
			fields[i].mangle == nil {
			panic("bad input")
		}
	}

	return func(buf []byte, value any) []byte {
		// Get raw value data.
		ptr := eface_data(value)

		for i := range fields {
			// Get struct field as iface via offset.
			fptr := fields[i].field.UnsafeGet(ptr)
			field := fields[i].type2.UnsafeIndirect(fptr)

			// Mangle the struct field data.
			buf = fields[i].mangle(buf, field)
			buf = append(buf, ',')
		}

		if len(fields) > 0 {
			// Drop final comma.
			buf = buf[:len(buf)-1]
		}

		return buf
	}
}

func indirect_ptr(p unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(&p)
}

func eface_data(a any) unsafe.Pointer {
	type eface struct{ _, data unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&a)).data
}
