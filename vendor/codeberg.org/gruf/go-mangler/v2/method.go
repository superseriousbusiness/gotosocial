package mangler

import (
	"reflect"
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

var (
	// mangleable type for implement checks.
	mangleableType = reflect.TypeFor[Mangleable]()
)

type Mangleable interface {
	Mangle(dst []byte) []byte
}

// getMethodType returns a *possible* Mangler to handle case
// of a type that implements any known interface{} types, else nil.
func getMethodType(t xunsafe.TypeIter) Mangler {
	switch {
	case t.Type.Implements(mangleableType):
		switch t.Type.Kind() {
		case reflect.Interface:
			return getInterfaceMangleableType(t)
		default:
			return getConcreteMangleableType(t)
		}
	default:
		return nil
	}
}

// getInterfaceMangleableType returns a Mangler to handle case of an interface{}
// type that implements Mangleable{}, i.e. Mangleable{} itself and any superset of.
func getInterfaceMangleableType(t xunsafe.TypeIter) Mangler {
	switch t.Indirect() && !t.IfaceIndir() {
	case true:
		return func(buf []byte, ptr unsafe.Pointer) []byte {
			ptr = *(*unsafe.Pointer)(ptr)
			if ptr == nil || (*xunsafe.Abi_NonEmptyInterface)(ptr).Data == nil {
				buf = append(buf, '0')
				return buf
			}
			v := *(*Mangleable)(ptr)
			buf = append(buf, '1')
			buf = v.Mangle(buf)
			return buf
		}
	case false:
		return func(buf []byte, ptr unsafe.Pointer) []byte {
			if ptr == nil || (*xunsafe.Abi_NonEmptyInterface)(ptr).Data == nil {
				buf = append(buf, '0')
				return buf
			}
			v := *(*Mangleable)(ptr)
			buf = append(buf, '1')
			buf = v.Mangle(buf)
			return buf
		}
	default:
		panic("unreachable")
	}
}

// getConcreteMangleableType returns a Manlger to handle case of concrete
// (i.e. non-interface{}) type that has a Mangleable{} method receiver.
func getConcreteMangleableType(t xunsafe.TypeIter) Mangler {
	itab := xunsafe.GetIfaceITab[Mangleable](t.Type)
	switch {
	case t.Indirect() && !t.IfaceIndir():
		return func(buf []byte, ptr unsafe.Pointer) []byte {
			ptr = *(*unsafe.Pointer)(ptr)
			if ptr == nil {
				buf = append(buf, '0')
				return buf
			}
			v := *(*Mangleable)(xunsafe.PackIface(itab, ptr))
			buf = append(buf, '1')
			buf = v.Mangle(buf)
			return buf
		}
	case t.Type.Kind() == reflect.Pointer && t.Type.Implements(mangleableType):
		// if the interface implementation is received by
		// value type, the pointer type will also support
		// it but it requires an extra dereference check.
		return func(buf []byte, ptr unsafe.Pointer) []byte {
			if ptr == nil {
				buf = append(buf, '0')
				return buf
			}
			v := *(*Mangleable)(xunsafe.PackIface(itab, ptr))
			buf = append(buf, '1')
			buf = v.Mangle(buf)
			return buf
		}
	default:
		return func(buf []byte, ptr unsafe.Pointer) []byte {
			v := *(*Mangleable)(xunsafe.PackIface(itab, ptr))
			buf = v.Mangle(buf)
			return buf
		}
	}
}
