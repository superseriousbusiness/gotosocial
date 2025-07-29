package format

import (
	"reflect"
	"unsafe"
)

type Stringer interface{ String() string }

var (
	// stringer type for implement checks.
	stringerType = typeof[Stringer]()

	// error type for implement checks.
	errorType = typeof[error]()
)

// getMethodType returns a *possible* FormatFunc to handle case
// of a type that implements any known interface{} types, else nil.
func getMethodType(t typenode) FormatFunc {
	switch {
	case t.rtype.Implements(stringerType):
		switch t.rtype.Kind() {
		case reflect.Interface:
			return getInterfaceStringerType(t)
		default:
			return getConcreteStringerType(t)
		}
	case t.rtype.Implements(errorType):
		switch t.rtype.Kind() {
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
func getInterfaceStringerType(t typenode) FormatFunc {
	switch t.indirect() && !t.iface_indir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil || (*reflect_nonEmptyInterface)(s.P).word == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(s.P)
			appendString(s, v.String())
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil || (*reflect_nonEmptyInterface)(s.P).word == nil {
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
func getConcreteStringerType(t typenode) FormatFunc {
	itab := get_iface_ITab[Stringer](t.rtype)
	switch t.indirect() && !t.iface_indir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(pack_iface(itab, s.P))
			appendString(s, v.String())
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*Stringer)(pack_iface(itab, s.P))
			appendString(s, v.String())
		})
	default:
		panic("unreachable")
	}
}

// getInterfaceErrorType returns a FormatFunc to handle case of an interface{}
// type that implements error{}, i.e. error{} itself and any superset of.
func getInterfaceErrorType(t typenode) FormatFunc {
	switch t.indirect() && !t.iface_indir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil || (*reflect_nonEmptyInterface)(s.P).word == nil {
				appendNil(s)
				return
			}
			v := *(*error)(s.P)
			appendString(s, v.Error())
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil || (*reflect_nonEmptyInterface)(s.P).word == nil {
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
func getConcreteErrorType(t typenode) FormatFunc {
	itab := get_iface_ITab[error](t.rtype)
	switch t.indirect() && !t.iface_indir() {
	case true:
		return with_typestr_ptrs(t, func(s *State) {
			s.P = *(*unsafe.Pointer)(s.P)
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*error)(pack_iface(itab, s.P))
			appendString(s, v.Error())
		})
	case false:
		return with_typestr_ptrs(t, func(s *State) {
			if s.P == nil {
				appendNil(s)
				return
			}
			v := *(*error)(pack_iface(itab, s.P))
			appendString(s, v.Error())
		})
	default:
		panic("unreachable")
	}
}
