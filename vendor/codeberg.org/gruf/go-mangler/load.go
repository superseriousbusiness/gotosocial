package mangler

import (
	"reflect"
)

// loadMangler is the top-most Mangler load function. It guarantees that a Mangler
// function will be returned for given value interface{} and reflected type. Else panics.
func loadMangler(t reflect.Type) Mangler {
	ctx := typecontext{rtype: t}
	ctx.direct = true

	// Load mangler fn
	mng := load(ctx)
	if mng != nil {
		return mng
	}

	// No mangler function could be determined
	panic("cannot mangle type: " + t.String())
}

// load will load a Mangler or reflect Mangler for given type and iface 'a'.
// Note: allocates new interface value if nil provided, i.e. if coming via reflection.
func load(ctx typecontext) Mangler {
	if ctx.rtype == nil {
		// There is no reflect type to search by
		panic("cannot mangle nil interface{} type")
	}

	// Search by reflection.
	mng := loadReflect(ctx)
	if mng != nil {
		return mng
	}

	return nil
}

// loadReflect will load a Mangler (or rMangler) function for the given reflected type info.
// NOTE: this is used as the top level load function for nested reflective searches.
func loadReflect(ctx typecontext) Mangler {
	switch ctx.rtype.Kind() {
	case reflect.Pointer:
		return loadReflectPtr(ctx)

	case reflect.String:
		return mangle_string

	case reflect.Struct:
		return loadReflectStruct(ctx)

	case reflect.Array:
		return loadReflectArray(ctx)

	case reflect.Slice:
		return loadReflectSlice(ctx)

	case reflect.Bool:
		return mangle_bool

	case reflect.Int,
		reflect.Uint,
		reflect.Uintptr:
		return mangle_int

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
func loadReflectPtr(ctx typecontext) Mangler {
	var n uint

	// Iteratively dereference ptrs
	for ctx.rtype.Kind() == reflect.Pointer {
		ctx.rtype = ctx.rtype.Elem()
		n++
	}

	// Set ptr type.
	ctx.isptr = true

	// Search for elemn type mangler.
	if mng := load(ctx); mng != nil {
		return deref_ptr_mangler(ctx, mng, n)
	}

	return nil
}

// loadReflectKnownSlice loads a Mangler function for a
// known slice-of-element type (in this case, primtives).
func loadReflectKnownSlice(ctx typecontext) Mangler {
	switch ctx.rtype.Kind() {
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

// loadReflectSlice ...
func loadReflectSlice(ctx typecontext) Mangler {

	// Get nested element type.
	elem := ctx.rtype.Elem()

	// Set this as nested type.
	ctx.set_nested(false)
	ctx.rtype = elem

	// Preferably look for known slice mangler func
	if mng := loadReflectKnownSlice(ctx); mng != nil {
		return mng
	}

	// Use nested mangler iteration.
	if mng := load(ctx); mng != nil {
		return iter_slice_mangler(ctx, mng)
	}

	return nil
}

// loadReflectArray ...
func loadReflectArray(ctx typecontext) Mangler {

	// Get nested element type.
	elem := ctx.rtype.Elem()

	// Set this as a nested value type.
	direct := ctx.rtype.Len() <= 1
	ctx.set_nested(direct)
	ctx.rtype = elem

	// Use manglers for nested iteration.
	if mng := load(ctx); mng != nil {
		return iter_array_mangler(ctx, mng)
	}

	return nil
}

// loadReflectStruct ...
func loadReflectStruct(ctx typecontext) Mangler {
	var mngs []Mangler

	// Set this as a nested value type.
	direct := ctx.rtype.NumField() <= 1
	ctx.set_nested(direct)

	// Gather manglers for all fields.
	for i := 0; i < ctx.ntype.NumField(); i++ {

		// Update context with field at index.
		ctx.rtype = ctx.ntype.Field(i).Type

		// Load mangler.
		mng := load(ctx)
		if mng == nil {
			return nil
		}

		// Append next to map.
		mngs = append(mngs, mng)
	}

	// Use manglers for nested iteration.
	return iter_struct_mangler(ctx, mngs)
}
