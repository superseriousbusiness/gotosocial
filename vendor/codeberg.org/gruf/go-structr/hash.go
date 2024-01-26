package structr

import (
	"reflect"
	"unsafe"

	"github.com/zeebo/xxh3"
)

func hasher(t reflect.Type) func(*xxh3.Hasher, any) bool {
	switch t.Kind() {
	case reflect.Int,
		reflect.Uint,
		reflect.Uintptr:
		switch unsafe.Sizeof(int(0)) {
		case 4:
			return hash32bit
		case 8:
			return hash64bit
		default:
			panic("unexpected platform int size")
		}

	case reflect.Int8,
		reflect.Uint8:
		return hash8bit

	case reflect.Int16,
		reflect.Uint16:
		return hash16bit

	case reflect.Int32,
		reflect.Uint32,
		reflect.Float32:
		return hash32bit

	case reflect.Int64,
		reflect.Uint64,
		reflect.Float64,
		reflect.Complex64:
		return hash64bit

	case reflect.String:
		return hashstring

	case reflect.Pointer:
		switch t.Elem().Kind() {
		case reflect.Int,
			reflect.Uint,
			reflect.Uintptr:
			switch unsafe.Sizeof(int(0)) {
			case 4:
				return hash32bitptr
			case 8:
				return hash64bitptr
			default:
				panic("unexpected platform int size")
			}

		case reflect.Int8,
			reflect.Uint8:
			return hash8bitptr

		case reflect.Int16,
			reflect.Uint16:
			return hash16bitptr

		case reflect.Int32,
			reflect.Uint32,
			reflect.Float32:
			return hash32bitptr

		case reflect.Int64,
			reflect.Uint64,
			reflect.Float64,
			reflect.Complex64:
			return hash64bitptr

		case reflect.String:
			return hashstringptr
		}

	case reflect.Slice:
		switch t.Elem().Kind() {
		case reflect.Int,
			reflect.Uint,
			reflect.Uintptr:
			switch unsafe.Sizeof(int(0)) {
			case 4:
				return hash32bitslice
			case 8:
				return hash64bitslice
			default:
				panic("unexpected platform int size")
			}

		case reflect.Int8,
			reflect.Uint8:
			return hash8bitslice

		case reflect.Int16,
			reflect.Uint16:
			return hash16bitslice

		case reflect.Int32,
			reflect.Uint32,
			reflect.Float32:
			return hash32bitslice

		case reflect.Int64,
			reflect.Uint64,
			reflect.Float64,
			reflect.Complex64:
			return hash64bitslice

		case reflect.String:
			return hashstringslice
		}
	}
	switch {
	case t.Implements(reflect.TypeOf((*interface{ MarshalBinary() ([]byte, error) })(nil)).Elem()):
		return hashbinarymarshaler

	case t.Implements(reflect.TypeOf((*interface{ Bytes() []byte })(nil)).Elem()):
		return hashbytesmethod

	case t.Implements(reflect.TypeOf((*interface{ String() string })(nil)).Elem()):
		return hashstringmethod

	case t.Implements(reflect.TypeOf((*interface{ MarshalText() ([]byte, error) })(nil)).Elem()):
		return hashtextmarshaler

	case t.Implements(reflect.TypeOf((*interface{ MarshalJSON() ([]byte, error) })(nil)).Elem()):
		return hashjsonmarshaler
	}
	panic("unhashable type")
}

func hash8bit(h *xxh3.Hasher, a any) bool {
	u := *(*uint8)(iface_value(a))
	_, _ = h.Write([]byte{u})
	return u == 0
}

func hash8bitptr(h *xxh3.Hasher, a any) bool {
	u := (*uint8)(iface_value(a))
	if u == nil {
		_, _ = h.Write([]byte{
			0,
		})
		return true
	} else {
		_, _ = h.Write([]byte{
			1,
			byte(*u),
		})
		return false
	}
}

func hash8bitslice(h *xxh3.Hasher, a any) bool {
	b := *(*[]byte)(iface_value(a))
	_, _ = h.Write(b)
	return b == nil
}

func hash16bit(h *xxh3.Hasher, a any) bool {
	u := *(*uint16)(iface_value(a))
	_, _ = h.Write([]byte{
		byte(u),
		byte(u >> 8),
	})
	return u == 0
}

func hash16bitptr(h *xxh3.Hasher, a any) bool {
	u := (*uint16)(iface_value(a))
	if u == nil {
		_, _ = h.Write([]byte{
			0,
		})
		return true
	} else {
		_, _ = h.Write([]byte{
			1,
			byte(*u),
			byte(*u >> 8),
		})
		return false
	}
}

func hash16bitslice(h *xxh3.Hasher, a any) bool {
	u := *(*[]uint16)(iface_value(a))
	for i := range u {
		_, _ = h.Write([]byte{
			byte(u[i]),
			byte(u[i] >> 8),
		})
	}
	return u == nil
}

func hash32bit(h *xxh3.Hasher, a any) bool {
	u := *(*uint32)(iface_value(a))
	_, _ = h.Write([]byte{
		byte(u),
		byte(u >> 8),
		byte(u >> 16),
		byte(u >> 24),
	})
	return u == 0
}

func hash32bitptr(h *xxh3.Hasher, a any) bool {
	u := (*uint32)(iface_value(a))
	if u == nil {
		_, _ = h.Write([]byte{
			0,
		})
		return true
	} else {
		_, _ = h.Write([]byte{
			1,
			byte(*u),
			byte(*u >> 8),
			byte(*u >> 16),
			byte(*u >> 24),
		})
		return false
	}
}

func hash32bitslice(h *xxh3.Hasher, a any) bool {
	u := *(*[]uint32)(iface_value(a))
	for i := range u {
		_, _ = h.Write([]byte{
			byte(u[i]),
			byte(u[i] >> 8),
			byte(u[i] >> 16),
			byte(u[i] >> 24),
		})
	}
	return u == nil
}

func hash64bit(h *xxh3.Hasher, a any) bool {
	u := *(*uint64)(iface_value(a))
	_, _ = h.Write([]byte{
		byte(u),
		byte(u >> 8),
		byte(u >> 16),
		byte(u >> 24),
		byte(u >> 32),
		byte(u >> 40),
		byte(u >> 48),
		byte(u >> 56),
	})
	return u == 0
}

func hash64bitptr(h *xxh3.Hasher, a any) bool {
	u := (*uint64)(iface_value(a))
	if u == nil {
		_, _ = h.Write([]byte{
			0,
		})
		return true
	} else {
		_, _ = h.Write([]byte{
			1,
			byte(*u),
			byte(*u >> 8),
			byte(*u >> 16),
			byte(*u >> 24),
			byte(*u >> 32),
			byte(*u >> 40),
			byte(*u >> 48),
			byte(*u >> 56),
		})
		return false
	}
}

func hash64bitslice(h *xxh3.Hasher, a any) bool {
	u := *(*[]uint64)(iface_value(a))
	for i := range u {
		_, _ = h.Write([]byte{
			byte(u[i]),
			byte(u[i] >> 8),
			byte(u[i] >> 16),
			byte(u[i] >> 24),
			byte(u[i] >> 32),
			byte(u[i] >> 40),
			byte(u[i] >> 48),
			byte(u[i] >> 56),
		})
	}
	return u == nil
}

func hashstring(h *xxh3.Hasher, a any) bool {
	s := *(*string)(iface_value(a))
	_, _ = h.WriteString(s)
	return s == ""
}

func hashstringptr(h *xxh3.Hasher, a any) bool {
	s := (*string)(iface_value(a))
	if s == nil {
		_, _ = h.Write([]byte{
			0,
		})
		return true
	} else {
		_, _ = h.Write([]byte{
			1,
		})
		_, _ = h.WriteString(*s)
		return false
	}
}

func hashstringslice(h *xxh3.Hasher, a any) bool {
	s := *(*[]string)(iface_value(a))
	for i := range s {
		_, _ = h.WriteString(s[i])
	}
	return s == nil
}

func hashbinarymarshaler(h *xxh3.Hasher, a any) bool {
	i := a.(interface{ MarshalBinary() ([]byte, error) })
	b, _ := i.MarshalBinary()
	_, _ = h.Write(b)
	return b == nil
}

func hashbytesmethod(h *xxh3.Hasher, a any) bool {
	i := a.(interface{ Bytes() []byte })
	b := i.Bytes()
	_, _ = h.Write(b)
	return b == nil
}

func hashstringmethod(h *xxh3.Hasher, a any) bool {
	i := a.(interface{ String() string })
	s := i.String()
	_, _ = h.WriteString(s)
	return s == ""
}

func hashtextmarshaler(h *xxh3.Hasher, a any) bool {
	i := a.(interface{ MarshalText() ([]byte, error) })
	b, _ := i.MarshalText()
	_, _ = h.Write(b)
	return b == nil
}

func hashjsonmarshaler(h *xxh3.Hasher, a any) bool {
	i := a.(interface{ MarshalJSON() ([]byte, error) })
	b, _ := i.MarshalJSON()
	_, _ = h.Write(b)
	return b == nil
}

func iface_value(a any) unsafe.Pointer {
	type eface struct{ _, v unsafe.Pointer }
	return (*eface)(unsafe.Pointer(&a)).v
}
