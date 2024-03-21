package mangler

import (
	"reflect"
)

// loadMangler is the top-most Mangler load function. It guarantees that a Mangler
// function will be returned for given value interface{} and reflected type. Else panics.
func loadMangler(a any, t reflect.Type) Mangler {
	// Load mangler fn
	mng := load(a, t)
	if mng != nil {
		return mng
	}

	// No mangler function could be determined
	panic("cannot mangle type: " + t.String())
}

// load will load a Mangler or reflect Mangler for given type and iface 'a'.
// Note: allocates new interface value if nil provided, i.e. if coming via reflection.
func load(a any, t reflect.Type) Mangler {
	if t == nil {
		// There is no reflect type to search by
		panic("cannot mangle nil interface{} type")
	}

	if a == nil {
		// Alloc new iface instance
		v := reflect.New(t).Elem()
		a = v.Interface()
	}

	// Check for Mangled implementation.
	if _, ok := a.(Mangled); ok {
		return mangle_mangled
	}

	// Search mangler by reflection.
	mng := loadReflect(t)
	if mng != nil {
		return mng
	}

	// Prefer iface mangler.
	mng = loadIface(a)
	if mng != nil {
		return mng
	}

	return nil
}

// loadIface is used as a near-last-resort interface{} type switch
// loader for types implementating other known (slower) functions.
func loadIface(a any) Mangler {
	switch a.(type) {
	case binarymarshaler:
		return mangle_binary
	case byteser:
		return mangle_byteser
	case stringer:
		return mangle_stringer
	case textmarshaler:
		return mangle_text
	case jsonmarshaler:
		return mangle_json
	default:
		return nil
	}
}

// loadReflect will load a Mangler (or rMangler) function for the given reflected type info.
// NOTE: this is used as the top level load function for nested reflective searches.
func loadReflect(t reflect.Type) Mangler {
	switch t.Kind() {
	case reflect.Pointer:
		return loadReflectPtr(t)

	case reflect.String:
		return mangle_string

	case reflect.Struct:
		return loadReflectStruct(t)

	case reflect.Array:
		return loadReflectArray(t)

	case reflect.Slice:
		return loadReflectSlice(t)

	case reflect.Map:
		return loadReflectMap(t)

	case reflect.Bool:
		return mangle_bool

	case reflect.Int,
		reflect.Uint,
		reflect.Uintptr:
		return mangle_platform_int()

	case reflect.Int8, reflect.Uint8:
		return mangle_8bit

	case reflect.Int16, reflect.Uint16:
		return mangle_16bit

	case reflect.Int32, reflect.Uint32:
		return mangle_32bit

	case reflect.Int64, reflect.Uint64:
		return mangle_64bit

	case reflect.Float32:
		return mangle_32bit

	case reflect.Float64:
		return mangle_64bit

	case reflect.Complex64:
		return mangle_64bit

	case reflect.Complex128:
		return mangle_128bit

	default:
		return nil
	}
}

// loadReflectPtr loads a Mangler (or rMangler) function for a ptr's element type.
// This also handles further dereferencing of any further ptr indrections (e.g. ***int).
func loadReflectPtr(t reflect.Type) Mangler {
	var count int

	// Elem
	et := t

	// Iteratively dereference ptrs
	for et.Kind() == reflect.Pointer {
		et = et.Elem()
		count++
	}

	// Search for ptr elemn type mangler.
	if mng := load(nil, et); mng != nil {
		return deref_ptr_mangler(et, mng, count)
	}

	return nil
}

// loadReflectKnownSlice loads a Mangler function for a
// known slice-of-element type (in this case, primtives).
func loadReflectKnownSlice(et reflect.Type) Mangler {
	switch et.Kind() {
	case reflect.String:
		return mangle_string_slice

	case reflect.Bool:
		return mangle_bool_slice

	case reflect.Int,
		reflect.Uint,
		reflect.Uintptr:
		return mangle_platform_int_slice()

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

// loadReflectSlice ...
func loadReflectSlice(t reflect.Type) Mangler {
	// Element type
	et := t.Elem()

	// Preferably look for known slice mangler func
	if mng := loadReflectKnownSlice(et); mng != nil {
		return mng
	}

	// Fallback to nested mangler iteration.
	if mng := load(nil, et); mng != nil {
		return iter_slice_mangler(t, mng)
	}

	return nil
}

// loadReflectArray ...
func loadReflectArray(t reflect.Type) Mangler {
	// Element type.
	et := t.Elem()

	// Use manglers for nested iteration.
	if mng := load(nil, et); mng != nil {
		return iter_array_mangler(t, mng)
	}

	return nil
}

// loadReflectMap ...
func loadReflectMap(t reflect.Type) Mangler {
	// Map types.
	kt := t.Key()
	et := t.Elem()

	// Load manglers.
	kmng := load(nil, kt)
	emng := load(nil, et)

	// Use manglers for nested iteration.
	if kmng != nil && emng != nil {
		return iter_map_mangler(t, kmng, emng)
	}

	return nil
}

// loadReflectStruct ...
func loadReflectStruct(t reflect.Type) Mangler {
	var mngs []Mangler

	// Gather manglers for all fields.
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Load mangler for field type.
		mng := load(nil, field.Type)
		if mng == nil {
			return nil
		}

		// Append next to map.
		mngs = append(mngs, mng)
	}

	// Use manglers for nested iteration.
	return iter_struct_mangler(t, mngs)
}
