package format

import (
	"reflect"
	"strconv"
	"unsafe"
)

// Formattable defines a type capable of being formatted and appended to a byte buffer.
type Formattable interface {
	AppendFormat([]byte) []byte
}

// format is the object passed among the append___ formatting functions.
type format struct {
	flags uint8   // 'isKey' and 'verbose' flags
	drefs uint8   // current value deref count
	curd  uint8   // current depth
	maxd  uint8   // maximum depth
	buf   *Buffer // out buffer
}

const (
	// flag bit constants.
	isKeyBit = uint8(1) << 0
	isValBit = uint8(1) << 1
	vboseBit = uint8(1) << 2
	panicBit = uint8(1) << 3
)

// AtMaxDepth returns whether format is currently at max depth.
func (f format) AtMaxDepth() bool {
	return f.curd > f.maxd
}

// Derefs returns no. times current value has been dereferenced.
func (f format) Derefs() uint8 {
	return f.drefs
}

// IsKey returns whether the isKey flag is set.
func (f format) IsKey() bool {
	return (f.flags & isKeyBit) != 0
}

// IsValue returns whether the isVal flag is set.
func (f format) IsValue() bool {
	return (f.flags & isValBit) != 0
}

// Verbose returns whether the verbose flag is set.
func (f format) Verbose() bool {
	return (f.flags & vboseBit) != 0
}

// Panic returns whether the panic flag is set.
func (f format) Panic() bool {
	return (f.flags & panicBit) != 0
}

// SetIsKey returns format instance with the isKey bit set to value.
func (f format) SetIsKey() format {
	return format{
		flags: f.flags & ^isValBit | isKeyBit,
		curd:  f.curd,
		maxd:  f.maxd,
		buf:   f.buf,
	}
}

// SetIsValue returns format instance with the isVal bit set to value.
func (f format) SetIsValue() format {
	return format{
		flags: f.flags & ^isKeyBit | isValBit,
		curd:  f.curd,
		maxd:  f.maxd,
		buf:   f.buf,
	}
}

// SetPanic returns format instance with the panic bit set to value.
func (f format) SetPanic() format {
	return format{
		flags: f.flags | panicBit /* handle panic as value */ | isValBit & ^isKeyBit,
		curd:  f.curd,
		maxd:  f.maxd,
		buf:   f.buf,
	}
}

// IncrDepth returns format instance with depth incremented.
func (f format) IncrDepth() format {
	return format{
		flags: f.flags,
		curd:  f.curd + 1,
		maxd:  f.maxd,
		buf:   f.buf,
	}
}

// IncrDerefs returns format instance with dereference count incremented.
func (f format) IncrDerefs() format {
	return format{
		flags: f.flags,
		drefs: f.drefs + 1,
		curd:  f.curd,
		maxd:  f.maxd,
		buf:   f.buf,
	}
}

// appendType appends a type using supplied type str.
func appendType(fmt format, t string) {
	for i := uint8(0); i < fmt.Derefs(); i++ {
		fmt.buf.AppendByte('*')
	}
	fmt.buf.AppendString(t)
}

// appendNilType Appends nil to buf, type included if verbose.
func appendNilType(fmt format, t string) {
	if fmt.Verbose() {
		fmt.buf.AppendByte('(')
		appendType(fmt, t)
		fmt.buf.AppendString(`)(nil)`)
	} else {
		fmt.buf.AppendString(`nil`)
	}
}

// appendByte Appends a single byte to buf.
func appendByte(fmt format, b byte) {
	if fmt.IsValue() || fmt.Verbose() {
		fmt.buf.AppendString(`'` + string(b) + `'`)
	} else {
		fmt.buf.AppendByte(b)
	}
}

// appendBytes Appends a quoted byte slice to buf.
func appendBytes(fmt format, b []byte) {
	if b == nil {
		// Bytes CAN be nil formatted
		appendNilType(fmt, `[]byte`)
	} else {
		// Append bytes as slice
		fmt.buf.AppendByte('[')
		for _, b := range b {
			fmt.buf.AppendByte(b)
			fmt.buf.AppendByte(',')
		}
		if len(b) > 0 {
			fmt.buf.Truncate(1)
		}
		fmt.buf.AppendByte(']')
	}
}

// appendString Appends an escaped, double-quoted string to buf.
func appendString(fmt format, s string) {
	switch {
	// Key in a key-value pair
	case fmt.IsKey():
		if !strconv.CanBackquote(s) {
			// Requires quoting AND escaping
			fmt.buf.B = strconv.AppendQuote(fmt.buf.B, s)
		} else if containsSpaceOrTab(s) {
			// Contains space, needs quotes
			fmt.buf.AppendString(`"` + s + `"`)
		} else {
			// All else write as-is
			fmt.buf.AppendString(s)
		}

	// Value in a key-value pair (always escape+quote)
	case fmt.IsValue():
		fmt.buf.B = strconv.AppendQuote(fmt.buf.B, s)

	// Verbose but neither key nor value (always quote)
	case fmt.Verbose():
		fmt.buf.AppendString(`"` + s + `"`)

	// All else
	default:
		fmt.buf.AppendString(s)
	}
}

// appendBool Appends a formatted bool to buf.
func appendBool(fmt format, b bool) {
	fmt.buf.B = strconv.AppendBool(fmt.buf.B, b)
}

// appendInt Appends a formatted int to buf.
func appendInt(fmt format, i int64) {
	fmt.buf.B = strconv.AppendInt(fmt.buf.B, i, 10)
}

// appendUint Appends a formatted uint to buf.
func appendUint(fmt format, u uint64) {
	fmt.buf.B = strconv.AppendUint(fmt.buf.B, u, 10)
}

// appendFloat Appends a formatted float to buf.
func appendFloat(fmt format, f float64) {
	fmt.buf.B = strconv.AppendFloat(fmt.buf.B, f, 'G', -1, 64)
}

// appendComplex Appends a formatted complex128 to buf.
func appendComplex(fmt format, c complex128) {
	appendFloat(fmt, real(c))
	fmt.buf.AppendByte('+')
	appendFloat(fmt, imag(c))
	fmt.buf.AppendByte('i')
}

// isNil will safely check if 'v' is nil without dealing with weird Go interface nil bullshit.
func isNil(i interface{}) bool {
	e := *(*struct {
		_ unsafe.Pointer // type
		v unsafe.Pointer // value
	})(unsafe.Pointer(&i))
	return (e.v == nil)
}

// appendIfaceOrReflectValue will attempt to append as interface, falling back to reflection.
func appendIfaceOrRValue(fmt format, i interface{}) {
	if !appendIface(fmt, i) {
		appendRValue(fmt, reflect.ValueOf(i))
	}
}

// appendValueNext checks for interface methods before performing appendRValue, checking + incr depth.
func appendRValueOrIfaceNext(fmt format, v reflect.Value) {
	// Check we haven't hit max
	if fmt.AtMaxDepth() {
		fmt.buf.AppendString("...")
		return
	}

	// Incr the depth
	fmt = fmt.IncrDepth()

	// Make actual call
	if !v.CanInterface() || !appendIface(fmt, v.Interface()) {
		appendRValue(fmt, v)
	}
}

// appendIface parses and Appends a formatted interface value to buf.
func appendIface(fmt format, i interface{}) (ok bool) {
	ok = true // default
	catchPanic := func() {
		if r := recover(); r != nil {
			// DON'T recurse catchPanic()
			if fmt.Panic() {
				panic(r)
			}

			// Attempt to decode panic into buf
			fmt.buf.AppendString(`!{PANIC=`)
			appendIfaceOrRValue(fmt.SetPanic(), r)
			fmt.buf.AppendByte('}')

			// Ensure return
			ok = true
		}
	}

	switch i := i.(type) {
	// Nil type
	case nil:
		fmt.buf.AppendString(`nil`)

	// Reflect types
	case reflect.Type:
		if isNil(i) /* safer nil check */ {
			appendNilType(fmt, `reflect.Type`)
		} else {
			appendType(fmt, `reflect.Type`)
			fmt.buf.AppendString(`(` + i.String() + `)`)
		}
	case reflect.Value:
		appendType(fmt, `reflect.Value`)
		fmt.buf.AppendByte('(')
		fmt.flags |= vboseBit
		appendRValue(fmt, i)
		fmt.buf.AppendByte(')')

	// Bytes and string types
	case byte:
		appendByte(fmt, i)
	case []byte:
		appendBytes(fmt, i)
	case string:
		appendString(fmt, i)

	// Int types
	case int:
		appendInt(fmt, int64(i))
	case int8:
		appendInt(fmt, int64(i))
	case int16:
		appendInt(fmt, int64(i))
	case int32:
		appendInt(fmt, int64(i))
	case int64:
		appendInt(fmt, i)

	// Uint types
	case uint:
		appendUint(fmt, uint64(i))
	// case uint8 :: this is 'byte'
	case uint16:
		appendUint(fmt, uint64(i))
	case uint32:
		appendUint(fmt, uint64(i))
	case uint64:
		appendUint(fmt, i)

	// Float types
	case float32:
		appendFloat(fmt, float64(i))
	case float64:
		appendFloat(fmt, i)

	// Bool type
	case bool:
		appendBool(fmt, i)

	// Complex types
	case complex64:
		appendComplex(fmt, complex128(i))
	case complex128:
		appendComplex(fmt, i)

	// Method types
	case error:
		switch {
		case fmt.Verbose():
			ok = false
		case isNil(i) /* use safer nil check */ :
			appendNilType(fmt, reflect.TypeOf(i).String())
		default:
			defer catchPanic()
			appendString(fmt, i.Error())
		}
	case Formattable:
		switch {
		case fmt.Verbose():
			ok = false
		case isNil(i) /* use safer nil check */ :
			appendNilType(fmt, reflect.TypeOf(i).String())
		default:
			defer catchPanic()
			fmt.buf.B = i.AppendFormat(fmt.buf.B)
		}
	case interface{ String() string }:
		switch {
		case fmt.Verbose():
			ok = false
		case isNil(i) /* use safer nil check */ :
			appendNilType(fmt, reflect.TypeOf(i).String())
		default:
			defer catchPanic()
			appendString(fmt, i.String())
		}

	// No quick handler
	default:
		ok = false
	}

	return ok
}

// appendReflectValue will safely append a reflected value.
func appendRValue(fmt format, v reflect.Value) {
	switch v.Kind() {
	// String and byte types
	case reflect.Uint8:
		appendByte(fmt, byte(v.Uint()))
	case reflect.String:
		appendString(fmt, v.String())

	// Float tpyes
	case reflect.Float32, reflect.Float64:
		appendFloat(fmt, v.Float())

	// Int types
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		appendInt(fmt, v.Int())

	// Uint types
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		appendUint(fmt, v.Uint())

	// Complex types
	case reflect.Complex64, reflect.Complex128:
		appendComplex(fmt, v.Complex())

	// Bool type
	case reflect.Bool:
		appendBool(fmt, v.Bool())

	// Slice and array types
	case reflect.Array:
		appendArrayType(fmt, v)
	case reflect.Slice:
		if v.IsNil() {
			appendNilType(fmt, v.Type().String())
		} else {
			appendArrayType(fmt, v)
		}

	// Map types
	case reflect.Map:
		if v.IsNil() {
			appendNilType(fmt, v.Type().String())
		} else {
			appendMapType(fmt, v)
		}

	// Struct types
	case reflect.Struct:
		appendStructType(fmt, v)

	// Deref'able ptr types
	case reflect.Ptr, reflect.Interface:
		if v.IsNil() {
			appendNilType(fmt, v.Type().String())
		} else {
			appendRValue(fmt.IncrDerefs(), v.Elem())
		}

	// 'raw' pointer types
	case reflect.UnsafePointer:
		appendType(fmt, `unsafe.Pointer`)
		fmt.buf.AppendByte('(')
		if u := v.Pointer(); u != 0 {
			fmt.buf.AppendString("0x")
			fmt.buf.B = strconv.AppendUint(fmt.buf.B, uint64(u), 16)
		} else {
			fmt.buf.AppendString(`nil`)
		}
		fmt.buf.AppendByte(')')
	case reflect.Uintptr:
		appendType(fmt, `uintptr`)
		fmt.buf.AppendByte('(')
		if u := v.Uint(); u != 0 {
			fmt.buf.AppendString("0x")
			fmt.buf.B = strconv.AppendUint(fmt.buf.B, u, 16)
		} else {
			fmt.buf.AppendString(`nil`)
		}
		fmt.buf.AppendByte(')')

	// Generic types we don't *exactly* handle
	case reflect.Func, reflect.Chan:
		if v.IsNil() {
			appendNilType(fmt, v.Type().String())
		} else {
			fmt.buf.AppendString(v.String())
		}

	// Unhandled kind
	default:
		fmt.buf.AppendString(v.String())
	}
}

// appendArrayType Appends an array of unknown type (parsed by reflection) to buf, unlike appendSliceType does NOT catch nil slice.
func appendArrayType(fmt format, v reflect.Value) {
	// get no. elements
	n := v.Len()

	fmt.buf.AppendByte('[')

	// Append values
	for i := 0; i < n; i++ {
		appendRValueOrIfaceNext(fmt.SetIsValue(), v.Index(i))
		fmt.buf.AppendByte(',')
	}

	// Drop last comma
	if n > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.AppendByte(']')
}

// appendMapType Appends a map of unknown types (parsed by reflection) to buf.
func appendMapType(fmt format, v reflect.Value) {
	// Prepend type if verbose
	if fmt.Verbose() {
		appendType(fmt, v.Type().String())
	}

	// Get a map iterator
	r := v.MapRange()
	n := v.Len()

	fmt.buf.AppendByte('{')

	// Iterate pairs
	for r.Next() {
		appendRValueOrIfaceNext(fmt.SetIsKey(), r.Key())
		fmt.buf.AppendByte('=')
		appendRValueOrIfaceNext(fmt.SetIsValue(), r.Value())
		fmt.buf.AppendByte(' ')
	}

	// Drop last space
	if n > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.AppendByte('}')
}

// appendStructType Appends a struct (as a set of key-value fields) to buf.
func appendStructType(fmt format, v reflect.Value) {
	// Get value type & no. fields
	t := v.Type()
	n := v.NumField()

	// Prepend type if verbose
	if fmt.Verbose() {
		appendType(fmt, v.Type().String())
	}

	fmt.buf.AppendByte('{')

	// Iterate fields
	for i := 0; i < n; i++ {
		vfield := v.Field(i)
		tfield := t.Field(i)

		// Append field name
		fmt.buf.AppendString(tfield.Name)
		fmt.buf.AppendByte('=')
		appendRValueOrIfaceNext(fmt.SetIsValue(), vfield)

		// Iter written count
		fmt.buf.AppendByte(' ')
	}

	// Drop last space
	if n > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.AppendByte('}')
}

// containsSpaceOrTab checks if "s" contains space or tabs.
func containsSpaceOrTab(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '\t' {
			return true
		}
	}
	return false
}
