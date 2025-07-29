//go:build go1.24 && !go1.25

package format

import (
	"reflect"
	"unsafe"
)

const (
	// see: go/src/internal/abi/type.go
	abi_KindDirectIface uint8 = 1 << 5
	abi_KindMask        uint8 = (1 << 5) - 1
)

// abi_Type is a copy of the memory layout of abi.Type{}.
//
// see: go/src/internal/abi/type.go
type abi_Type struct {
	_        uintptr
	PtrBytes uintptr
	_        uint32
	_        uint8
	_        uint8
	_        uint8
	Kind_    uint8
	_        func(unsafe.Pointer, unsafe.Pointer) bool
	_        *byte
	_        int32
	_        int32
}

// abi_EmptyInterface is a copy of the memory layout of abi.EmptyInterface{},
// which is to say also the memory layout of any method-less interface.
//
// see: go/src/internal/abi/iface.go
type abi_EmptyInterface struct {
	Type *abi_Type
	Data unsafe.Pointer
}

// see: go/src/internal/abi/type.go Type.Kind()
func abi_Type_Kind(t reflect.Type) uint8 {
	iface := (*reflect_nonEmptyInterface)(unsafe.Pointer(&t))
	atype := (*abi_Type)(unsafe.Pointer(iface.word))
	return atype.Kind_ & abi_KindMask
}

// see: go/src/internal/abi/type.go Type.IfaceIndir()
func abi_Type_IfaceIndir(t reflect.Type) bool {
	iface := (*reflect_nonEmptyInterface)(unsafe.Pointer(&t))
	atype := (*abi_Type)(unsafe.Pointer(iface.word))
	return atype.Kind_&abi_KindDirectIface == 0
}

// pack_iface packs a new reflect.nonEmptyInterface{} using shielded itab
// pointer and data (word) pointer, returning a pointer for caller casting.
func pack_iface(itab uintptr, word unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(&reflect_nonEmptyInterface{
		itab: itab,
		word: word,
	})
}

// get_iface_ITab generates a new value of given type,
// casts it to the generic param interface type, and
// returns the .itab portion of the reflect.nonEmptyInterface{}.
// this is useful for later calls to pack_iface for known type.
func get_iface_ITab[I any](t reflect.Type) uintptr {
	s := reflect.New(t).Elem().Interface().(I)
	i := (*reflect_nonEmptyInterface)(unsafe.Pointer(&s))
	return i.itab
}

// unpack_eface returns the .Data portion of an abi.EmptyInterface{}.
func unpack_eface(a any) unsafe.Pointer {
	return (*abi_EmptyInterface)(unsafe.Pointer((&a))).Data
}

// add returns the ptr addition of starting ptr and a delta.
func add(ptr unsafe.Pointer, delta uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(ptr) + delta)
}

// typeof is short-hand for reflect.TypeFor[T]().
func typeof[T any]() reflect.Type {
	return reflect.TypeFor[T]()
}

// see: go/src/reflect/value.go
type reflect_flag uintptr

const (
	// see: go/src/reflect/value.go
	reflect_flagKindWidth                = 5 // there are 27 kinds
	reflect_flagKindMask    reflect_flag = 1<<reflect_flagKindWidth - 1
	reflect_flagStickyRO    reflect_flag = 1 << 5
	reflect_flagEmbedRO     reflect_flag = 1 << 6
	reflect_flagIndir       reflect_flag = 1 << 7
	reflect_flagAddr        reflect_flag = 1 << 8
	reflect_flagMethod      reflect_flag = 1 << 9
	reflect_flagMethodShift              = 10
	reflect_flagRO          reflect_flag = reflect_flagStickyRO | reflect_flagEmbedRO

	// custom flag to indicate key types.
	flagKeyType = 1 << 10
)

// reflect_iface_elem_flags returns the reflect_flag expected of an unboxed interface element of type.
//
// see: go/src/reflect/value.go unpackElem()
func reflect_iface_elem_flags(elemType reflect.Type) reflect_flag {
	if elemType == nil {
		return 0
	}
	flags := reflect_flag(abi_Type_Kind(elemType))
	if abi_Type_IfaceIndir(elemType) {
		flags |= reflect_flagIndir
	}
	return flags
}

// reflect_pointer_elem_flags returns the reflect_flag expected of a dereferenced pointer element of type.
//
// see: go/src/reflect/value.go Value.Elem()
func reflect_pointer_elem_flags(ptrFlags reflect_flag, elemType reflect.Type) reflect_flag {
	return ptrFlags | reflect_flagIndir | reflect_flagAddr | reflect_flag(abi_Type_Kind(elemType))
}

// reflect_array_elem_flags returns the reflect_flag expected of an element of type in an array.
//
// see: go/src/reflect/value.go Value.Index()
func reflect_array_elem_flags(arrayFlags reflect_flag, elemType reflect.Type) reflect_flag {
	return arrayFlags&(reflect_flagIndir|reflect_flagAddr) | reflect_flag(abi_Type_Kind(elemType))
}

// reflect_slice_elem_flags returns the reflect_flag expected of a slice element of type.
//
// see: go/src/reflect/value.go Value.Index()
func reflect_slice_elem_flags(elemType reflect.Type) reflect_flag {
	return reflect_flagAddr | reflect_flagIndir | reflect_flag(abi_Type_Kind(elemType))
}

// reflect_struct_field_flags returns the reflect_flag expected of a struct field of type.
//
// see: go/src/reflect/value.go Value.Field()
func reflect_struct_field_flags(structFlags reflect_flag, fieldType reflect.Type) reflect_flag {
	return structFlags&(reflect_flagIndir|reflect_flagAddr) | reflect_flag(abi_Type_Kind(fieldType))
}

// reflect_map_key_flags returns the reflect_flag expected of a map key of type (with our own key type mask set).
//
// see: go/src/reflect/map_swiss.go MapIter.Key()
func reflect_map_key_flags(keyType reflect.Type) reflect_flag {
	return flagKeyType | reflect_flag(abi_Type_Kind(keyType))
}

// reflect_map_elem_flags returns the reflect_flag expected of a map element of type.
//
// see: go/src/reflect/map_swiss.go MapIter.Value()
func reflect_map_elem_flags(elemType reflect.Type) reflect_flag {
	return reflect_flag(abi_Type_Kind(elemType))
}

// reflect_nonEmptyInterface is a copy of the memory layout of reflect.nonEmptyInterface,
// which is also to say the memory layout of any non-empty (i.e. w/ method) interface.
//
// see: go/src/reflect/value.go
type reflect_nonEmptyInterface struct {
	itab uintptr
	word unsafe.Pointer
}

// reflect_Value is a copy of the memory layout of reflect.Value{}.
//
// see: go/src/reflect/value.go
type reflect_Value struct {
	typ_ unsafe.Pointer
	ptr  unsafe.Pointer
	reflect_flag
}

func init() {
	if unsafe.Sizeof(reflect_Value{}) != unsafe.Sizeof(reflect.Value{}) {
		panic("reflect_Value{} not in sync with reflect.Value{}")
	}
}

// reflect_type_data returns the .word from the reflect.Type{} cast
// as the reflect.nonEmptyInterface{}, which itself will be a pointer
// to the actual abi.Type{} that this reflect.Type{} is wrapping.
func reflect_type_data(t reflect.Type) unsafe.Pointer {
	return (*reflect_nonEmptyInterface)(unsafe.Pointer(&t)).word
}

// build_reflect_value manually builds a reflect.Value{} by setting the internal field members.
func build_reflect_value(rtype reflect.Type, data unsafe.Pointer, flags reflect_flag) reflect.Value {
	return *(*reflect.Value)(unsafe.Pointer(&reflect_Value{reflect_type_data(rtype), data, flags}))
}

// maps_Iter is a copy of the memory layout of maps.Iter{}.
//
// see: go/src/internal/runtime/maps/table.go
type maps_Iter struct {
	key  unsafe.Pointer
	elem unsafe.Pointer
	_    uintptr
	_    uintptr
	_    uint64
	_    uint64
	_    uint64
	_    uint8
	_    int
	_    uintptr
	_    struct{ _ unsafe.Pointer }
	_    uint64
}

// reflect_MapIter is a copy of the memory layout of reflect.MapIter{}.
//
// see: go/src/reflect/map_swiss.go
type reflect_MapIter struct {
	m     reflect.Value
	hiter maps_Iter
}

func init() {
	if unsafe.Sizeof(reflect_MapIter{}) != unsafe.Sizeof(reflect.MapIter{}) {
		panic("reflect_MapIter{} not in sync with reflect.MapIter{}")
	}
}

// map_iter creates a new map iterator from value,
// skipping the initial v.MapRange() type checking.
func map_iter(v reflect.Value) *reflect.MapIter {
	var i reflect_MapIter
	i.m = v
	return (*reflect.MapIter)(unsafe.Pointer(&i))
}

// map_key returns ptr to current map key in iter.
func map_key(i *reflect.MapIter) unsafe.Pointer {
	return (*reflect_MapIter)(unsafe.Pointer(i)).hiter.key
}

// map_elem returns ptr to current map element in iter.
func map_elem(i *reflect.MapIter) unsafe.Pointer {
	return (*reflect_MapIter)(unsafe.Pointer(i)).hiter.elem
}

// see: go/src/internal/unsafeheader/unsafeheader.go
type unsafeheader_Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// see: go/src/internal/unsafeheader/unsafeheader.go
type unsafeheader_String struct {
	Data unsafe.Pointer
	Len  int
}
