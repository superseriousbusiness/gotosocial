package mangler

import (
	"math/bits"
	_ "unsafe"
)

// Notes:
//   the use of unsafe conversion from the direct interface values to
//   the chosen types in each of the below functions allows us to convert
//   not only those types directly, but anything type-aliased to those
//   types. e.g. `time.Duration` directly as int64.

func mangle_string(buf []byte, a any) []byte {
	return append(buf, *(*string)(iface_value(a))...)
}

func mangle_string_ptr(buf []byte, a any) []byte {
	if ptr := (*string)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		return append(buf, *ptr...)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_string_slice(buf []byte, a any) []byte {
	s := *(*[]string)(iface_value(a))
	for _, s := range s {
		buf = append(buf, s...)
		buf = append(buf, ',')
	}
	if len(s) > 0 {
		buf = buf[:len(buf)-1]
	}
	return buf
}

func mangle_bool(buf []byte, a any) []byte {
	if *(*bool)(iface_value(a)) {
		return append(buf, '1')
	}
	return append(buf, '0')
}

func mangle_bool_ptr(buf []byte, a any) []byte {
	if ptr := (*bool)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		if *ptr {
			return append(buf, '1')
		}
		return append(buf, '0')
	}
	buf = append(buf, '0')
	return buf
}

func mangle_bool_slice(buf []byte, a any) []byte {
	for _, b := range *(*[]bool)(iface_value(a)) {
		if b {
			buf = append(buf, '1')
		} else {
			buf = append(buf, '0')
		}
	}
	return buf
}

func mangle_8bit(buf []byte, a any) []byte {
	return append(buf, *(*uint8)(iface_value(a)))
}

func mangle_8bit_ptr(buf []byte, a any) []byte {
	if ptr := (*uint8)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		return append(buf, *ptr)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_8bit_slice(buf []byte, a any) []byte {
	return append(buf, *(*[]uint8)(iface_value(a))...)
}

func mangle_16bit(buf []byte, a any) []byte {
	return append_uint16(buf, *(*uint16)(iface_value(a)))
}

func mangle_16bit_ptr(buf []byte, a any) []byte {
	if ptr := (*uint16)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		return append_uint16(buf, *ptr)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_16bit_slice(buf []byte, a any) []byte {
	for _, u := range *(*[]uint16)(iface_value(a)) {
		buf = append_uint16(buf, u)
	}
	return buf
}

func mangle_32bit(buf []byte, a any) []byte {
	return append_uint32(buf, *(*uint32)(iface_value(a)))
}

func mangle_32bit_ptr(buf []byte, a any) []byte {
	if ptr := (*uint32)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		return append_uint32(buf, *ptr)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_32bit_slice(buf []byte, a any) []byte {
	for _, u := range *(*[]uint32)(iface_value(a)) {
		buf = append_uint32(buf, u)
	}
	return buf
}

func mangle_64bit(buf []byte, a any) []byte {
	return append_uint64(buf, *(*uint64)(iface_value(a)))
}

func mangle_64bit_ptr(buf []byte, a any) []byte {
	if ptr := (*uint64)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		return append_uint64(buf, *ptr)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_64bit_slice(buf []byte, a any) []byte {
	for _, u := range *(*[]uint64)(iface_value(a)) {
		buf = append_uint64(buf, u)
	}
	return buf
}

// mangle_platform_int contains the correct iface mangler on runtime for platform int size.
var mangle_platform_int = func() Mangler {
	switch bits.UintSize {
	case 32:
		return mangle_32bit
	case 64:
		return mangle_64bit
	default:
		panic("unexpected platform int size")
	}
}()

// mangle_platform_int_ptr contains the correct iface mangler on runtime for platform int size.
var mangle_platform_int_ptr = func() Mangler {
	switch bits.UintSize {
	case 32:
		return mangle_32bit_ptr
	case 64:
		return mangle_64bit_ptr
	default:
		panic("unexpected platform int size")
	}
}()

// mangle_platform_int_slice contains the correct iface mangler on runtime for platform int size.
var mangle_platform_int_slice = func() Mangler {
	switch bits.UintSize {
	case 32:
		return mangle_32bit_slice
	case 64:
		return mangle_64bit_slice
	default:
		panic("unexpected platform int size")
	}
}()

// uint128 provides an easily mangleable data type for 128bit data types to be cast into.
type uint128 [2]uint64

func mangle_128bit(buf []byte, a any) []byte {
	u2 := *(*uint128)(iface_value(a))
	buf = append_uint64(buf, u2[0])
	buf = append_uint64(buf, u2[1])
	return buf
}

func mangle_128bit_ptr(buf []byte, a any) []byte {
	if ptr := (*uint128)(iface_value(a)); ptr != nil {
		buf = append(buf, '1')
		buf = append_uint64(buf, (*ptr)[0])
		buf = append_uint64(buf, (*ptr)[1])
	}
	buf = append(buf, '0')
	return buf
}

func mangle_128bit_slice(buf []byte, a any) []byte {
	for _, u2 := range *(*[]uint128)(iface_value(a)) {
		buf = append_uint64(buf, u2[0])
		buf = append_uint64(buf, u2[1])
	}
	return buf
}

func mangle_mangled(buf []byte, a any) []byte {
	if v := a.(Mangled); v != nil {
		buf = append(buf, '1')
		return v.Mangle(buf)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_binary(buf []byte, a any) []byte {
	if v := a.(binarymarshaler); v != nil {
		b, err := v.MarshalBinary()
		if err != nil {
			panic("mangle_binary: " + err.Error())
		}
		buf = append(buf, '1')
		return append(buf, b...)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_stringer(buf []byte, a any) []byte {
	if v := a.(stringer); v != nil {
		buf = append(buf, '1')
		return append(buf, v.String()...)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_text(buf []byte, a any) []byte {
	if v := a.(textmarshaler); v != nil {
		b, err := v.MarshalText()
		if err != nil {
			panic("mangle_text: " + err.Error())
		}
		buf = append(buf, '1')
		return append(buf, b...)
	}
	buf = append(buf, '0')
	return buf
}

func mangle_json(buf []byte, a any) []byte {
	if v := a.(jsonmarshaler); v != nil {
		b, err := v.MarshalJSON()
		if err != nil {
			panic("mangle_json: " + err.Error())
		}
		buf = append(buf, '1')
		return append(buf, b...)
	}
	buf = append(buf, '0')
	return buf
}
