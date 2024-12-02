package format

import (
	"reflect"
	"strconv"
	"unicode/utf8"

	"codeberg.org/gruf/go-byteutil"
)

const (
	// Flag bit constants, note they are prioritised in this order.
	IsKeyBit = uint8(1) << 0 // set to indicate key formatting
	VboseBit = uint8(1) << 1 // set to indicate verbose formatting
	IsValBit = uint8(1) << 2 // set to indicate value formatting
	PanicBit = uint8(1) << 3 // set after panic to prevent recursion
)

// format provides formatting of values into a Buffer.
type format struct {
	// Flags are the currently set value flags.
	Flags uint8

	// Derefs is the current value dereference count.
	Derefs uint8

	// CurDepth is the current Format iterator depth.
	CurDepth uint8

	// VType is the current value type.
	VType string

	// Config is the set Formatter config (MUST NOT be nil).
	Config *Formatter

	// Buffer is the currently set output buffer.
	Buffer *byteutil.Buffer
}

// AtMaxDepth returns whether format is currently at max depth.
func (f format) AtMaxDepth() bool {
	return f.CurDepth > f.Config.MaxDepth
}

// Key returns whether the isKey flag is set.
func (f format) Key() bool {
	return (f.Flags & IsKeyBit) != 0
}

// Value returns whether the isVal flag is set.
func (f format) Value() bool {
	return (f.Flags & IsValBit) != 0
}

// Verbose returns whether the verbose flag is set.
func (f format) Verbose() bool {
	return (f.Flags & VboseBit) != 0
}

// Panic returns whether the panic flag is set.
func (f format) Panic() bool {
	return (f.Flags & PanicBit) != 0
}

// SetKey returns format instance with the IsKey bit set to true,
// note this resets the dereference count.
func (f format) SetKey() format {
	flags := f.Flags | IsKeyBit
	flags &= ^IsValBit
	return format{
		Flags:    flags,
		CurDepth: f.CurDepth,
		Config:   f.Config,
		Buffer:   f.Buffer,
	}
}

// SetValue returns format instance with the IsVal bit set to true,
// note this resets the dereference count.
func (f format) SetValue() format {
	flags := f.Flags | IsValBit
	flags &= ^IsKeyBit
	return format{
		Flags:    flags,
		CurDepth: f.CurDepth,
		Config:   f.Config,
		Buffer:   f.Buffer,
	}
}

// SetVerbose returns format instance with the Vbose bit set to true,
// note this resets the dereference count.
func (f format) SetVerbose() format {
	return format{
		Flags:    f.Flags | VboseBit,
		CurDepth: f.CurDepth,
		Config:   f.Config,
		Buffer:   f.Buffer,
	}
}

// SetPanic returns format instance with the panic bit set to true,
// note this resets the dereference count and sets IsVal (unsetting IsKey) bit.
func (f format) SetPanic() format {
	flags := f.Flags | PanicBit
	flags |= IsValBit
	flags &= ^IsKeyBit
	return format{
		Flags:    flags,
		CurDepth: f.CurDepth,
		Config:   f.Config,
		Buffer:   f.Buffer,
	}
}

// IncrDepth returns format instance with depth incremented and derefs reset.
func (f format) IncrDepth() format {
	return format{
		Flags:    f.Flags,
		Derefs:   f.Derefs,
		CurDepth: f.CurDepth + 1,
		Config:   f.Config,
		Buffer:   f.Buffer,
	}
}

// IncrDerefs returns format instance with dereference count incremented.
func (f format) IncrDerefs() format {
	return format{
		Flags:    f.Flags,
		Derefs:   f.Derefs + 1,
		CurDepth: f.CurDepth,
		Config:   f.Config,
		Buffer:   f.Buffer,
	}
}

func (f format) AppendType() {
	const derefs = `********************************` +
		`********************************` +
		`********************************` +
		`********************************` +
		`********************************` +
		`********************************` +
		`********************************` +
		`********************************`
	f.Buffer.B = append(f.Buffer.B, derefs[:f.Derefs]...)
	f.Buffer.B = append(f.Buffer.B, f.VType...)
}

func (f format) AppendNil() {
	if !f.Verbose() {
		f.Buffer.B = append(f.Buffer.B, `nil`...)
		return
	}

	// Append nil with type
	f.Buffer.B = append(f.Buffer.B, '(')
	f.AppendType()
	f.Buffer.B = append(f.Buffer.B, `)(nil`...)
	f.Buffer.B = append(f.Buffer.B, ')')
}

func (f format) AppendByte(b byte) {
	switch {
	// Always quoted
	case f.Key():
		f.Buffer.B = append(f.Buffer.B, '\'')
		f.Buffer.B = append(f.Buffer.B, Byte2Str(b)...)
		f.Buffer.B = append(f.Buffer.B, '\'')

	// Always quoted ASCII with type
	case f.Verbose():
		f._AppendPrimitiveTyped(func(f format) {
			f.Buffer.B = append(f.Buffer.B, '\'')
			f.Buffer.B = append(f.Buffer.B, Byte2Str(b)...)
			f.Buffer.B = append(f.Buffer.B, '\'')
		})

	// Always quoted
	case f.Value():
		f.Buffer.B = append(f.Buffer.B, '\'')
		f.Buffer.B = append(f.Buffer.B, Byte2Str(b)...)
		f.Buffer.B = append(f.Buffer.B, '\'')

	// Append as raw byte
	default:
		f.Buffer.B = append(f.Buffer.B, b)
	}
}

func (f format) AppendBytes(b []byte) {
	switch {
	// Bytes CAN be nil formatted
	case b == nil:
		f.AppendNil()

	// Quoted only if spaces/requires escaping
	case f.Key():
		s := byteutil.B2S(b)
		f.AppendStringSafe(s)

	// Append as separate ASCII quoted bytes in slice
	case f.Verbose():
		f._AppendArrayTyped(func(f format) {
			for i := 0; i < len(b); i++ {
				f.Buffer.B = append(f.Buffer.B, '\'')
				f.Buffer.B = append(f.Buffer.B, Byte2Str(b[i])...)
				f.Buffer.B = append(f.Buffer.B, `',`...)
			}
			if len(b) > 0 {
				f.Buffer.Truncate(1)
			}
		})

	// Quoted only if spaces/requires escaping
	case f.Value():
		s := byteutil.B2S(b)
		f.AppendStringSafe(s)

	// Append as raw bytes
	default:
		f.Buffer.B = append(f.Buffer.B, b...)
	}
}

func (f format) AppendRune(r rune) {
	switch {
	// Quoted only if spaces/requires escaping
	case f.Key():
		f.AppendRuneKey(r)

	// Always quoted ASCII with type
	case f.Verbose():
		f._AppendPrimitiveTyped(func(f format) {
			f.Buffer.B = strconv.AppendQuoteRuneToASCII(f.Buffer.B, r)
		})

	// Always quoted value
	case f.Value():
		f.Buffer.B = strconv.AppendQuoteRune(f.Buffer.B, r)

	// Append as raw rune
	default:
		f.Buffer.WriteRune(r)
	}
}

func (f format) AppendRuneKey(r rune) {
	if utf8.RuneLen(r) > 1 && (r < ' ' && r != '\t') || r == '`' || r == '\u007F' {
		// Quote and escape this rune
		f.Buffer.B = strconv.AppendQuoteRuneToASCII(f.Buffer.B, r)
	} else {
		// Simply append rune
		f.Buffer.WriteRune(r)
	}
}

func (f format) AppendRunes(r []rune) {
	switch {
	// Runes CAN be nil formatted
	case r == nil:
		f.AppendNil()

	// Quoted only if spaces/requires escaping
	case f.Key():
		f.AppendStringSafe(string(r))

	// Append as separate ASCII quoted bytes in slice
	case f.Verbose():
		f._AppendArrayTyped(func(f format) {
			for i := 0; i < len(r); i++ {
				f.Buffer.B = strconv.AppendQuoteRuneToASCII(f.Buffer.B, r[i])
				f.Buffer.B = append(f.Buffer.B, ',')
			}
			if len(r) > 0 {
				f.Buffer.Truncate(1)
			}
		})

	// Quoted only if spaces/requires escaping
	case f.Value():
		f.AppendStringSafe(string(r))

	// Append as raw bytes
	default:
		for i := 0; i < len(r); i++ {
			f.Buffer.WriteRune(r[i])
		}
	}
}

func (f format) AppendString(s string) {
	switch {
	// Quoted only if spaces/requires escaping
	case f.Key():
		f.AppendStringSafe(s)

	// Always quoted with type
	case f.Verbose():
		f._AppendPrimitiveTyped(func(f format) {
			f.AppendStringQuoted(s)
		})

	// Quoted only if spaces/requires escaping
	case f.Value():
		f.AppendStringSafe(s)

	// All else
	default:
		f.Buffer.B = append(f.Buffer.B, s...)
	}
}

func (f format) AppendStringSafe(s string) {
	if len(s) > SingleTermLine || !IsSafeASCII(s) {
		// Requires quoting AND escaping
		f.Buffer.B = strconv.AppendQuote(f.Buffer.B, s)
	} else if ContainsDoubleQuote(s) {
		// Contains double quotes, needs escaping
		f.Buffer.B = append(f.Buffer.B, '"')
		f.Buffer.B = AppendEscape(f.Buffer.B, s)
		f.Buffer.B = append(f.Buffer.B, '"')
	} else if len(s) == 0 || ContainsSpaceOrTab(s) {
		// Contains space / empty, needs quotes
		f.Buffer.B = append(f.Buffer.B, '"')
		f.Buffer.B = append(f.Buffer.B, s...)
		f.Buffer.B = append(f.Buffer.B, '"')
	} else {
		// All else write as-is
		f.Buffer.B = append(f.Buffer.B, s...)
	}
}

func (f format) AppendStringQuoted(s string) {
	if len(s) > SingleTermLine || !IsSafeASCII(s) {
		// Requires quoting AND escaping
		f.Buffer.B = strconv.AppendQuote(f.Buffer.B, s)
	} else if ContainsDoubleQuote(s) {
		// Contains double quotes, needs escaping
		f.Buffer.B = append(f.Buffer.B, '"')
		f.Buffer.B = AppendEscape(f.Buffer.B, s)
		f.Buffer.B = append(f.Buffer.B, '"')
	} else {
		// Simply append with quotes
		f.Buffer.B = append(f.Buffer.B, '"')
		f.Buffer.B = append(f.Buffer.B, s...)
		f.Buffer.B = append(f.Buffer.B, '"')
	}
}

func (f format) AppendBool(b bool) {
	if f.Verbose() {
		// Append as bool with type information
		f._AppendPrimitiveTyped(func(f format) {
			f.Buffer.B = strconv.AppendBool(f.Buffer.B, b)
		})
	} else {
		// Simply append as bool
		f.Buffer.B = strconv.AppendBool(f.Buffer.B, b)
	}
}

func (f format) AppendInt(i int64) {
	f._AppendPrimitiveType(func(f format) {
		f.Buffer.B = strconv.AppendInt(f.Buffer.B, i, 10)
	})
}

func (f format) AppendUint(u uint64) {
	f._AppendPrimitiveType(func(f format) {
		f.Buffer.B = strconv.AppendUint(f.Buffer.B, u, 10)
	})
}

func (f format) AppendFloat(l float64) {
	f._AppendPrimitiveType(func(f format) {
		f.AppendFloatValue(l)
	})
}

func (f format) AppendFloatValue(l float64) {
	f.Buffer.B = strconv.AppendFloat(f.Buffer.B, l, 'f', -1, 64)
}

func (f format) AppendComplex(c complex128) {
	f._AppendPrimitiveType(func(f format) {
		f.AppendFloatValue(real(c))
		f.Buffer.B = append(f.Buffer.B, '+')
		f.AppendFloatValue(imag(c))
		f.Buffer.B = append(f.Buffer.B, 'i')
	})
}

func (f format) AppendPtr(u uint64) {
	f._AppendPtrType(func(f format) {
		if u == 0 {
			// Append as nil
			f.Buffer.B = append(f.Buffer.B, `nil`...)
		} else {
			// Append as hex number
			f.Buffer.B = append(f.Buffer.B, `0x`...)
			f.Buffer.B = strconv.AppendUint(f.Buffer.B, u, 16)
		}
	})
}

func (f format) AppendInterfaceOrReflect(i interface{}) {
	if !f.AppendInterface(i) {
		// Interface append failed, used reflected value + type
		f.AppendReflectValue(reflect.ValueOf(i), reflect.TypeOf(i))
	}
}

func (f format) AppendInterfaceOrReflectNext(v reflect.Value, t reflect.Type) {
	// Check we haven't hit max
	if f.AtMaxDepth() {
		f.Buffer.B = append(f.Buffer.B, `...`...)
		return
	}

	// Incr the depth
	f = f.IncrDepth()

	// Make actual call
	f.AppendReflectOrInterface(v, t)
}

func (f format) AppendReflectOrInterface(v reflect.Value, t reflect.Type) {
	if !v.CanInterface() ||
		!f.AppendInterface(v.Interface()) {
		// Interface append failed, use reflect
		f.AppendReflectValue(v, t)
	}
}

func (f format) AppendInterface(i interface{}) bool {
	switch i := i.(type) {
	// Reflect types
	case reflect.Type:
		f.AppendReflectType(i)
	case reflect.Value:
		f.Buffer.B = append(f.Buffer.B, `reflect.Value`...)
		f.Buffer.B = append(f.Buffer.B, '(')
		f.Buffer.B = append(f.Buffer.B, i.String()...)
		f.Buffer.B = append(f.Buffer.B, ')')

	// Bytes, runes and string types
	case rune:
		f.VType = `int32`
		f.AppendRune(i)
	case []rune:
		f.VType = `[]int32`
		f.AppendRunes(i)
	case byte:
		f.VType = `uint8`
		f.AppendByte(i)
	case []byte:
		f.VType = `[]uint8`
		f.AppendBytes(i)
	case string:
		f.VType = `string`
		f.AppendString(i)

	// Int types
	case int:
		f.VType = `int`
		f.AppendInt(int64(i))
	case int8:
		f.VType = `int8`
		f.AppendInt(int64(i))
	case int16:
		f.VType = `int16`
		f.AppendInt(int64(i))
	case int64:
		f.VType = `int64`
		f.AppendInt(int64(i))

	// Uint types
	case uint:
		f.VType = `uint`
		f.AppendUint(uint64(i))
	case uint16:
		f.VType = `uint16`
		f.AppendUint(uint64(i))
	case uint32:
		f.VType = `uint32`
		f.AppendUint(uint64(i))
	case uint64:
		f.VType = `uint64`
		f.AppendUint(uint64(i))

	// Float types
	case float32:
		f.VType = `float32`
		f.AppendFloat(float64(i))
	case float64:
		f.VType = `float64`
		f.AppendFloat(float64(i))

	// Bool type
	case bool:
		f.VType = `bool`
		f.AppendBool(i)

	// Complex types
	case complex64:
		f.VType = `complex64`
		f.AppendComplex(complex128(i))
	case complex128:
		f.VType = `complex128`
		f.AppendComplex(complex128(i))

	// Method types
	case error:
		return f._AppendMethodType(func() string {
			return i.Error()
		}, i)
	case interface{ String() string }:
		return f._AppendMethodType(func() string {
			return i.String()
		}, i)

	// No quick handler
	default:
		return false
	}

	return true
}

func (f format) AppendReflectType(t reflect.Type) {
	switch f.VType = `reflect.Type`; {
	case isNil(t) /* safer nil check */ :
		f.AppendNil()
	case f.Verbose():
		f.AppendType()
		f.Buffer.B = append(f.Buffer.B, '(')
		f.Buffer.B = append(f.Buffer.B, t.String()...)
		f.Buffer.B = append(f.Buffer.B, ')')
	default:
		f.Buffer.B = append(f.Buffer.B, t.String()...)
	}
}

func (f format) AppendReflectValue(v reflect.Value, t reflect.Type) {
	switch v.Kind() {
	// String/byte types
	case reflect.String:
		f.VType = t.String()
		f.AppendString(v.String())
	case reflect.Uint8:
		f.VType = t.String()
		f.AppendByte(byte(v.Uint()))
	case reflect.Int32:
		f.VType = t.String()
		f.AppendRune(rune(v.Int()))

	// Float tpyes
	case reflect.Float32, reflect.Float64:
		f.VType = t.String()
		f.AppendFloat(v.Float())

	// Int types
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int64:
		f.VType = t.String()
		f.AppendInt(v.Int())

	// Uint types
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f.VType = t.String()
		f.AppendUint(v.Uint())

	// Complex types
	case reflect.Complex64, reflect.Complex128:
		f.VType = t.String()
		f.AppendComplex(v.Complex())

	// Bool type
	case reflect.Bool:
		f.VType = t.String()
		f.AppendBool(v.Bool())

	// Slice and array types
	case reflect.Array:
		f.AppendArray(v, t)
	case reflect.Slice:
		f.AppendSlice(v, t)

	// Map types
	case reflect.Map:
		f.AppendMap(v, t)

	// Struct types
	case reflect.Struct:
		f.AppendStruct(v, t)

	// Interface type
	case reflect.Interface:
		if v.IsNil() {
			// Append nil ptr type
			f.VType = t.String()
			f.AppendNil()
		} else {
			// Append interface
			v = v.Elem()
			t = v.Type()
			f.AppendReflectOrInterface(v, t)
		}

	// Deref'able ptr type
	case reflect.Ptr:
		if v.IsNil() {
			// Append nil ptr type
			f.VType = t.String()
			f.AppendNil()
		} else {
			// Deref to next level
			f = f.IncrDerefs()
			v, t = v.Elem(), t.Elem()
			f.AppendReflectOrInterface(v, t)
		}

	// 'raw' pointer types
	case reflect.UnsafePointer, reflect.Func, reflect.Chan:
		f.VType = t.String()
		f.AppendPtr(uint64(v.Pointer()))
	case reflect.Uintptr:
		f.VType = t.String()
		f.AppendPtr(v.Uint())

	// Zero reflect value
	case reflect.Invalid:
		f.Buffer.B = append(f.Buffer.B, `nil`...)

	// All others
	default:
		f.VType = t.String()
		f.AppendType()
	}
}

func (f format) AppendSlice(v reflect.Value, t reflect.Type) {
	// Get slice value type
	f.VType = t.String()

	if t.Elem().Kind() == reflect.Uint8 {
		// This is a byte slice
		f.AppendBytes(v.Bytes())
		return
	}

	if v.IsNil() {
		// Nil slice
		f.AppendNil()
		return
	}

	if f.Verbose() {
		// Append array with type information
		f._AppendArrayTyped(func(f format) {
			f.AppendArrayElems(v, t)
		})
	} else {
		// Simply append array as elems
		f._AppendArray(func(f format) {
			f.AppendArrayElems(v, t)
		})
	}
}

func (f format) AppendArray(v reflect.Value, t reflect.Type) {
	// Get array value type
	f.VType = t.String()

	if f.Verbose() {
		// Append array with type information
		f._AppendArrayTyped(func(f format) {
			f.AppendArrayElems(v, t)
		})
	} else {
		// Simply append array as elems
		f._AppendArray(func(f format) {
			f.AppendArrayElems(v, t)
		})
	}
}

func (f format) AppendArrayElems(v reflect.Value, t reflect.Type) {
	// Get no. elems
	n := v.Len()

	// Get elem type
	et := t.Elem()

	// Append values
	for i := 0; i < n; i++ {
		f.SetValue().AppendInterfaceOrReflectNext(v.Index(i), et)
		f.Buffer.B = append(f.Buffer.B, ',')
	}

	// Drop last comma
	if n > 0 {
		f.Buffer.Truncate(1)
	}
}

func (f format) AppendMap(v reflect.Value, t reflect.Type) {
	// Get value type
	f.VType = t.String()

	if v.IsNil() {
		// Nil map -- no fields
		f.AppendNil()
		return
	}

	// Append field formatted map fields
	f._AppendFieldType(func(f format) {
		f.AppendMapFields(v, t)
	})
}

func (f format) AppendMapFields(v reflect.Value, t reflect.Type) {
	// Get a map iterator
	r := v.MapRange()
	n := v.Len()

	// Get key/val types
	kt := t.Key()
	kv := t.Elem()

	// Iterate pairs
	for r.Next() {
		f.SetKey().AppendInterfaceOrReflectNext(r.Key(), kt)
		f.Buffer.B = append(f.Buffer.B, '=')
		f.SetValue().AppendInterfaceOrReflectNext(r.Value(), kv)
		f.Buffer.B = append(f.Buffer.B, ' ')
	}

	// Drop last space
	if n > 0 {
		f.Buffer.Truncate(1)
	}
}

func (f format) AppendStruct(v reflect.Value, t reflect.Type) {
	// Get value type
	f.VType = t.String()

	// Append field formatted struct fields
	f._AppendFieldType(func(f format) {
		f.AppendStructFields(v, t)
	})
}

func (f format) AppendStructFields(v reflect.Value, t reflect.Type) {
	// Get field no.
	n := v.NumField()

	// Iterate struct fields
	for i := 0; i < n; i++ {
		vfield := v.Field(i)
		tfield := t.Field(i)

		// Append field name
		f.AppendStringSafe(tfield.Name)
		f.Buffer.B = append(f.Buffer.B, '=')
		f.SetValue().AppendInterfaceOrReflectNext(vfield, tfield.Type)

		// Append separator
		f.Buffer.B = append(f.Buffer.B, ' ')
	}

	// Drop last space
	if n > 0 {
		f.Buffer.Truncate(1)
	}
}

func (f format) _AppendMethodType(method func() string, i interface{}) (ok bool) {
	// Verbose -- no methods
	if f.Verbose() {
		return false
	}

	// Catch nil type
	if isNil(i) {
		f.AppendNil()
		return true
	}

	// Catch any panics
	defer func() {
		if r := recover(); r != nil {
			// DON'T recurse catchPanic()
			if f.Panic() {
				panic(r)
			}

			// Attempt to decode panic into buf
			f.Buffer.B = append(f.Buffer.B, `!{PANIC=`...)
			f.SetPanic().AppendInterfaceOrReflect(r)
			f.Buffer.B = append(f.Buffer.B, '}')

			// Ensure no further attempts
			// to format after return
			ok = true
		}
	}()

	// Get method result
	result := method()

	switch {
	// Append as key formatted
	case f.Key():
		f.AppendStringSafe(result)

	// Append as always quoted
	case f.Value():
		f.AppendStringQuoted(result)

	// Append as-is
	default:
		f.Buffer.B = append(f.Buffer.B, result...)
	}

	return true
}

// _AppendPrimitiveType is a helper to append prefix/suffix for primitives (numbers/bools/bytes/runes).
func (f format) _AppendPrimitiveType(appendPrimitive func(format)) {
	if f.Verbose() {
		// Append value with type information
		f._AppendPrimitiveTyped(appendPrimitive)
	} else {
		// Append simply as-is
		appendPrimitive(f)
	}
}

// _AppendPrimitiveTyped is a helper to append prefix/suffix for primitives (numbers/bools/bytes/runes) with their types (if deref'd).
func (f format) _AppendPrimitiveTyped(appendPrimitive func(format)) {
	if f.Derefs > 0 {
		// Is deref'd, append type info
		f.Buffer.B = append(f.Buffer.B, '(')
		f.AppendType()
		f.Buffer.WriteString(`)(`)
		appendPrimitive(f)
		f.Buffer.B = append(f.Buffer.B, ')')
	} else {
		// Simply append value
		appendPrimitive(f)
	}
}

// _AppendPtrType is a helper to append prefix/suffix for ptr types (with type if necessary).
func (f format) _AppendPtrType(appendPtr func(format)) {
	if f.Verbose() {
		// Append value with type information
		f.Buffer.B = append(f.Buffer.B, '(')
		f.AppendType()
		f.Buffer.WriteString(`)(`)
		appendPtr(f)
		f.Buffer.B = append(f.Buffer.B, ')')
	} else {
		// Append simply as-is
		appendPtr(f)
	}
}

// _AppendArray is a helper to append prefix/suffix for array-types.
func (f format) _AppendArray(appendArray func(format)) {
	f.Buffer.B = append(f.Buffer.B, '[')
	appendArray(f)
	f.Buffer.B = append(f.Buffer.B, ']')
}

// _AppendArrayTyped is a helper to append prefix/suffix for array-types with their types.
func (f format) _AppendArrayTyped(appendArray func(format)) {
	f.AppendType()
	f.Buffer.B = append(f.Buffer.B, '{')
	appendArray(f)
	f.Buffer.B = append(f.Buffer.B, '}')
}

// _AppendFields is a helper to append prefix/suffix for field-types (with type if necessary).
func (f format) _AppendFieldType(appendFields func(format)) {
	if f.Verbose() {
		f.AppendType()
	}
	f.Buffer.B = append(f.Buffer.B, '{')
	appendFields(f)
	f.Buffer.B = append(f.Buffer.B, '}')
}
