package format

import (
	"reflect"
	"unsafe"

	"codeberg.org/gruf/go-xunsafe"
)

type Stringer interface{ String() string }

type Formattable interface{ Format(*State) }

var (
	// stringer type for implement checks.
	stringerType = typeof[Stringer]()

	// formattableType type for implement checks.
	formattableType = typeof[Formattable]()

	// error type for implement checks.
	errorType = typeof[error]()
)

// getMethodType returns a *possible* FormatFunc to handle case
// of a type that implements any known interface{} types, else nil.
func getMethodType(t xunsafe.TypeIter) FormatFunc {
	switch {
	case t.Type.Implements(stringerType):
		switch t.Type.Kind() {
		case reflect.Interface:
			return getInterfaceStringerType(t)
		default:
			return getConcreteStringerType(t)
		}
	case t.Type.Implements(formattableType):
		switch t.Type.Kind() {
		case reflect.Interface:
			return getInterfaceFormattableType(t)
		default:
			return getConcreteFormattableType(t)
		}
	case t.Type.Implements(errorType):
		switch t.Type.Kind() {
		case reflect.Interface:
			return getInterfaceErrorType(t)
		default:
			return getConcreteErrorType(t)
		}
	default:
		return nil
	}
}

// getInterfaceStringerType returns a FormatFunc to handle case of an interface{}
// type that implements Stringer{}, i.e. Stringer{} itself and any superset of.
func getInterfaceStringerType(t xunsafe.TypeIter) FormatFunc {
	switch t.Indirect() && !t.IfaceIndir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil || (*xunsafe.Abi_NonEmptyInterface)(s.P).Data == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(s.P)
			appendString(s, v.String())
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil || (*xunsafe.Abi_NonEmptyInterface)(s.P).Data == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(s.P)
			appendString(s, v.String())
		})
	default:
		panic("unreachable")
	}
}

// getConcreteStringerType returns a FormatFunc to handle case of concrete
// (i.e. non-interface{}) type that has a Stringer{} method receiver.
func getConcreteStringerType(t xunsafe.TypeIter) FormatFunc {
	itab := xunsafe.GetIfaceITab[Stringer](t.Type)
	switch {
	case t.Indirect() && !t.IfaceIndir():
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(xunsafe.PackIface(itab, s.P))
			appendString(s, v.String())
		})
	case t.Type.Kind() == reflect.Pointer && t.Type.Implements(stringerType):
		// if the interface implementation is received by
		// value type, the pointer type will also support
		// it but it requires an extra dereference check.
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(xunsafe.PackIface(itab, s.P))
			appendString(s, v.String())
		})
	default:
		return with_typestr_ptrs(t, func(s *State) {
			v := *(*Stringer)(xunsafe.PackIface(itab, s.P))
			appendString(s, v.String())
		})
	}
}

// getInterfaceFormattableType returns a FormatFunc to handle case of an interface{}
// type that implements Formattable{}, i.e. Formattable{} itself and any superset of.
func getInterfaceFormattableType(t xunsafe.TypeIter) FormatFunc {
	switch t.Indirect() && !t.IfaceIndir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil || (*xunsafe.Abi_NonEmptyInterface)(s.P).Data == nil {
				appendNil(s)
				return
			}
			v := *(*Formattable)(s.P)
			v.Format(s)
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil || (*xunsafe.Abi_NonEmptyInterface)(s.P).Data == nil {
				appendNil(s)
				return
			}
			v := *(*Formattable)(s.P)
			v.Format(s)
		})
	default:
		panic("unreachable")
	}
}

// getConcreteFormattableType returns a FormatFunc to handle case of concrete
// (i.e. non-interface{}) type that has a Formattable{} method receiver.
func getConcreteFormattableType(t xunsafe.TypeIter) FormatFunc {
	itab := xunsafe.GetIfaceITab[Formattable](t.Type)
	switch {
	case t.Indirect() && !t.IfaceIndir():
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*Formattable)(xunsafe.PackIface(itab, s.P))
			v.Format(s)
		})
	case t.Type.Kind() == reflect.Pointer && t.Type.Implements(formattableType):
		// if the interface implementation is received by
		// value type, the pointer type will also support
		// it but it requires an extra dereference check.
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*Formattable)(xunsafe.PackIface(itab, s.P))
			v.Format(s)
		})
	default:
		return with_typestr_ptrs(t, func(s *State) {
			v := *(*Formattable)(xunsafe.PackIface(itab, s.P))
			v.Format(s)
		})
	}
}

// getInterfaceErrorType returns a FormatFunc to handle case of an interface{}
// type that implements error{}, i.e. error{} itself and any superset of.
func getInterfaceErrorType(t xunsafe.TypeIter) FormatFunc {
	switch t.Indirect() && !t.IfaceIndir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil || (*xunsafe.Abi_NonEmptyInterface)(s.P).Data == nil {
				appendNil(s)
				return
			}
			v := *(*error)(s.P)
			appendString(s, v.Error())
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil || (*xunsafe.Abi_NonEmptyInterface)(s.P).Data == nil {
				appendNil(s)
				return
			}
			v := *(*error)(s.P)
			appendString(s, v.Error())
		})
	default:
		panic("unreachable")
	}
}

// getConcreteErrorType returns a FormatFunc to handle case of concrete
// (i.e. non-interface{}) type that has an error{} method receiver.
func getConcreteErrorType(t xunsafe.TypeIter) FormatFunc {
	itab := xunsafe.GetIfaceITab[error](t.Type)
	switch {
	case t.Indirect() && !t.IfaceIndir():
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*error)(xunsafe.PackIface(itab, s.P))
			appendString(s, v.Error())
		})
	case t.Type.Kind() == reflect.Pointer && t.Type.Implements(errorType):
		// if the interface implementation is received by
		// value type, the pointer type will also support
		// it but it requires an extra dereference check.
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*error)(xunsafe.PackIface(itab, s.P))
			appendString(s, v.Error())
		})
	default:
		return with_typestr_ptrs(t, func(s *State) {
			v := *(*error)(xunsafe.PackIface(itab, s.P))
			appendString(s, v.Error())
		})
	}
}
