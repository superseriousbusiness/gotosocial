//go:build go1.24 && !go1.26

package xunsafe

import (
	"reflect"
	"unsafe"
)

func init() {
	// TypeOf(reflect.Type{}) == *struct{ abi.Type{} }
	t := reflect.TypeOf(reflect.TypeOf(0)).Elem()
	if t.Size() != unsafe.Sizeof(Abi_Type{}) {
		panic("Abi_Type{} not in sync with abi.Type{}")
	}
}

const (
	// see: go/src/internal/abi/type.go
	Abi_KindDirectIface uint8 = 1 << 5
	Abi_KindMask        uint8 = (1 << 5) - 1
)

// Abi_Type is a copy of the memory layout of abi.Type{}.
//
// see: go/src/internal/abi/type.go
type Abi_Type struct {
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

// Abi_EmptyInterface is a copy of the memory layout of abi.EmptyInterface{},
// which is to say also the memory layout of any method-less interface.
//
// see: go/src/internal/abi/iface.go
type Abi_EmptyInterface struct {
	Type *Abi_Type
	Data unsafe.Pointer
}

// Abi_NonEmptyInterface is a copy of the memory layout of abi.NonEmptyInterface{},
// which is to say also the memory layout of any interface containing method(s).
//
// see: go/src/internal/abi/iface.go on 1.25+
// see: go/src/reflect/value.go on 1.24
type Abi_NonEmptyInterface struct {
	ITab uintptr
	Data unsafe.Pointer
}

// see: go/src/internal/abi/type.go Type.Kind()
func Abi_Type_Kind(t reflect.Type) uint8 {
	iface := (*Abi_NonEmptyInterface)(unsafe.Pointer(&t))
	atype := (*Abi_Type)(unsafe.Pointer(iface.Data))
	return atype.Kind_ & Abi_KindMask
}

// see: go/src/internal/abi/type.go Type.IfaceIndir()
func Abi_Type_IfaceIndir(t reflect.Type) bool {
	iface := (*Abi_NonEmptyInterface)(unsafe.Pointer(&t))
	atype := (*Abi_Type)(unsafe.Pointer(iface.Data))
	return atype.Kind_&Abi_KindDirectIface == 0
}

// PackIface packs a new reflect.nonEmptyInterface{} using shielded
// itab and data pointer, returning a pointer for caller casting.
func PackIface(itab uintptr, word unsafe.Pointer) unsafe.Pointer {
	return unsafe.Pointer(&Abi_NonEmptyInterface{
		ITab: itab,
		Data: word,
	})
}

// GetIfaceITab generates a new value of given type,
// casts it to the generic param interface type, and
// returns the .itab portion of the abi.NonEmptyInterface{}.
// this is useful for later calls to PackIface for known type.
func GetIfaceITab[I any](t reflect.Type) uintptr {
	s := reflect.New(t).Elem().Interface().(I)
	i := (*Abi_NonEmptyInterface)(unsafe.Pointer(&s))
	return i.ITab
}

// UnpackEface returns the .Data portion of an abi.EmptyInterface{}.
func UnpackEface(a any) unsafe.Pointer {
	return (*Abi_EmptyInterface)(unsafe.Pointer((&a))).Data
}

// see: go/src/internal/unsafeheader/unsafeheader.go
type Unsafeheader_Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// see: go/src/internal/unsafeheader/unsafeheader.go
type Unsafeheader_String struct {
	Data unsafe.Pointer
	Len  int
}
