package logger

import (
	stdfmt "fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"codeberg.org/gruf/go-bytes"
)

// DefaultTextFormat is the default TextFormat instance
var DefaultTextFormat = TextFormat{
	Strict:   false,
	Verbose:  false,
	MaxDepth: 10,
	Levels:   DefaultLevels(),
}

// TextFormat is the default LogFormat implementation, with very similar formatting to the
// standard "fmt" package's '%#v' operator. The main difference being that pointers are
// dereferenced as far as possible in order to reach a printable value. It is also *mildly* faster.
type TextFormat struct {
	// Strict defines whether to use strict key-value pair formatting, i.e. should the level
	// timestamp and msg be formatted as key-value pairs (with forced quoting for msg)
	Strict bool

	// Verbose defines whether to increase output verbosity, i.e. include types with nil values
	// and force values implementing .String() / .AppendFormat() to be printed as a struct etc.
	Verbose bool

	// MaxDepth specifies the max depth of fields the formatter will iterate
	MaxDepth uint8

	// Levels defines the map of log LEVELs to level strings
	Levels Levels
}

// fmt returns a new format instance based on receiver TextFormat and given buffer
func (f TextFormat) fmt(buf *bytes.Buffer) format {
	var flags uint8
	if f.Verbose {
		flags |= vboseBit
	}
	return format{
		flags: flags,
		curd:  0,
		maxd:  f.MaxDepth,
		buf:   buf,
	}
}

func (f TextFormat) AppendKey(buf *bytes.Buffer, key string) {
	if len(key) > 0 {
		// only append if key is non-zero length
		appendString(f.fmt(buf).SetIsKey(true), key)
		buf.WriteByte('=')
	}
}

func (f TextFormat) AppendLevel(buf *bytes.Buffer, lvl LEVEL) {
	if f.Strict {
		// Strict format, append level key
		buf.WriteString(`level=`)
		buf.WriteString(f.Levels.Get(lvl))
		return
	}

	// Write level string
	buf.WriteByte('[')
	buf.WriteString(f.Levels.Get(lvl))
	buf.WriteByte(']')
}

func (f TextFormat) AppendTimestamp(buf *bytes.Buffer, now string) {
	if f.Strict {
		// Strict format, use key and quote
		buf.WriteString(`time=`)
		appendString(f.fmt(buf), now)
		return
	}

	// Write time as-is
	buf.WriteString(now)
}

func (f TextFormat) AppendValue(buf *bytes.Buffer, value interface{}) {
	appendIfaceOrRValue(f.fmt(buf).SetIsKey(false), value)
}

func (f TextFormat) AppendByte(buf *bytes.Buffer, value byte) {
	appendByte(f.fmt(buf), value)
}

func (f TextFormat) AppendBytes(buf *bytes.Buffer, value []byte) {
	appendBytes(f.fmt(buf), value)
}

func (f TextFormat) AppendString(buf *bytes.Buffer, value string) {
	appendString(f.fmt(buf), value)
}

func (f TextFormat) AppendStrings(buf *bytes.Buffer, value []string) {
	appendStringSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendBool(buf *bytes.Buffer, value bool) {
	appendBool(f.fmt(buf), value)
}

func (f TextFormat) AppendBools(buf *bytes.Buffer, value []bool) {
	appendBoolSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendInt(buf *bytes.Buffer, value int) {
	appendInt(f.fmt(buf), int64(value))
}

func (f TextFormat) AppendInts(buf *bytes.Buffer, value []int) {
	appendIntSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendUint(buf *bytes.Buffer, value uint) {
	appendUint(f.fmt(buf), uint64(value))
}

func (f TextFormat) AppendUints(buf *bytes.Buffer, value []uint) {
	appendUintSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendFloat(buf *bytes.Buffer, value float64) {
	appendFloat(f.fmt(buf), value)
}

func (f TextFormat) AppendFloats(buf *bytes.Buffer, value []float64) {
	appendFloatSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendTime(buf *bytes.Buffer, value time.Time) {
	appendTime(f.fmt(buf), value)
}

func (f TextFormat) AppendTimes(buf *bytes.Buffer, value []time.Time) {
	appendTimeSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendDuration(buf *bytes.Buffer, value time.Duration) {
	appendDuration(f.fmt(buf), value)
}

func (f TextFormat) AppendDurations(buf *bytes.Buffer, value []time.Duration) {
	appendDurationSlice(f.fmt(buf), value)
}

func (f TextFormat) AppendMsg(buf *bytes.Buffer, a ...interface{}) {
	if f.Strict {
		// Strict format, use key and quote
		buf.WriteString(`msg=`)
		buf.B = strconv.AppendQuote(buf.B, stdfmt.Sprint(a...))
		return
	}

	// Write message as-is
	stdfmt.Fprint(buf, a...)
}

func (f TextFormat) AppendMsgf(buf *bytes.Buffer, s string, a ...interface{}) {
	if f.Strict {
		// Strict format, use key and quote
		buf.WriteString(`msg=`)
		buf.B = strconv.AppendQuote(buf.B, stdfmt.Sprintf(s, a...))
		return
	}

	// Write message as-is
	stdfmt.Fprintf(buf, s, a...)
}

// format is the object passed among the append___ formatting functions
type format struct {
	flags uint8         // 'isKey' and 'verbose' flags
	drefs uint8         // current value deref count
	curd  uint8         // current depth
	maxd  uint8         // maximum depth
	buf   *bytes.Buffer // out buffer
}

const (
	// flag bit constants
	isKeyBit = uint8(1) << 0
	vboseBit = uint8(1) << 1
)

// AtMaxDepth returns whether format is currently at max depth.
func (f format) AtMaxDepth() bool {
	return f.curd >= f.maxd
}

// Derefs returns no. times current value has been dereferenced.
func (f format) Derefs() uint8 {
	return f.drefs
}

// IsKey returns whether the isKey flag is set.
func (f format) IsKey() bool {
	return (f.flags & isKeyBit) != 0
}

// Verbose returns whether the verbose flag is set.
func (f format) Verbose() bool {
	return (f.flags & vboseBit) != 0
}

// SetIsKey returns format instance with the isKey bit set to value.
func (f format) SetIsKey(is bool) format {
	flags := f.flags
	if is {
		flags |= isKeyBit
	} else {
		flags &= ^isKeyBit
	}
	return format{
		flags: flags,
		drefs: f.drefs,
		curd:  f.curd,
		maxd:  f.maxd,
		buf:   f.buf,
	}
}

// IncrDepth returns format instance with depth incremented.
func (f format) IncrDepth() format {
	return format{
		flags: f.flags,
		drefs: f.drefs,
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
		fmt.buf.WriteByte('*')
	}
	fmt.buf.WriteString(t)
}

// appendNilType writes nil to buf, type included if verbose.
func appendNilType(fmt format, t string) {
	if fmt.Verbose() {
		fmt.buf.WriteByte('(')
		appendType(fmt, t)
		fmt.buf.WriteString(`)(nil)`)
	} else {
		fmt.buf.WriteString(`nil`)
	}
}

// appendNilFace writes nil to buf, type included if verbose.
func appendNilIface(fmt format, i interface{}) {
	if fmt.Verbose() {
		fmt.buf.WriteByte('(')
		appendType(fmt, reflect.TypeOf(i).String())
		fmt.buf.WriteString(`)(nil)`)
	} else {
		fmt.buf.WriteString(`nil`)
	}
}

// appendNilRValue writes nil to buf, type included if verbose.
func appendNilRValue(fmt format, v reflect.Value) {
	if fmt.Verbose() {
		fmt.buf.WriteByte('(')
		appendType(fmt, v.Type().String())
		fmt.buf.WriteString(`)(nil)`)
	} else {
		fmt.buf.WriteString(`nil`)
	}
}

// appendByte writes a single byte to buf
func appendByte(fmt format, b byte) {
	fmt.buf.WriteByte(b)
}

// appendBytes writes a quoted byte slice to buf
func appendBytes(fmt format, b []byte) {
	if !fmt.IsKey() && b == nil {
		// Values CAN be nil formatted
		appendNilType(fmt, `[]byte`)
	} else {
		// unsafe cast as string to prevent reallocation
		appendString(fmt, *(*string)(unsafe.Pointer(&b)))
	}
}

// appendString writes an escaped, double-quoted string to buf
func appendString(fmt format, s string) {
	if !fmt.IsKey() || !strconv.CanBackquote(s) {
		// All non-keys and multiline keys get quoted + escaped
		fmt.buf.B = strconv.AppendQuote(fmt.buf.B, s)
		return
	} else if containsSpaceOrTab(s) {
		// Key containing spaces/tabs, quote this
		fmt.buf.WriteByte('"')
		fmt.buf.WriteString(s)
		fmt.buf.WriteByte('"')
		return
	}

	// Safe to leave unquoted
	fmt.buf.WriteString(s)
}

// appendStringSlice writes a slice of strings to buf
func appendStringSlice(fmt format, s []string) {
	// Check for nil slice
	if s == nil {
		appendNilType(fmt, `[]string`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, s := range s {
		appendString(fmt.SetIsKey(false), s)
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(s) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendBool writes a formatted bool to buf
func appendBool(fmt format, b bool) {
	fmt.buf.B = strconv.AppendBool(fmt.buf.B, b)
}

// appendBool writes a slice of formatted bools to buf
func appendBoolSlice(fmt format, b []bool) {
	// Check for nil slice
	if b == nil {
		appendNilType(fmt, `[]bool`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, b := range b {
		appendBool(fmt, b)
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(b) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendInt writes a formatted int to buf
func appendInt(fmt format, i int64) {
	fmt.buf.B = strconv.AppendInt(fmt.buf.B, i, 10)
}

// appendIntSlice writes a slice of formatted int to buf
func appendIntSlice(fmt format, i []int) {
	// Check for nil slice
	if i == nil {
		appendNilType(fmt, `[]int`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, i := range i {
		appendInt(fmt, int64(i))
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(i) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendUint writes a formatted uint to buf
func appendUint(fmt format, u uint64) {
	fmt.buf.B = strconv.AppendUint(fmt.buf.B, u, 10)
}

// appendUintSlice writes a slice of formatted uint to buf
func appendUintSlice(fmt format, u []uint) {
	// Check for nil slice
	if u == nil {
		appendNilType(fmt, `[]uint`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, u := range u {
		appendUint(fmt, uint64(u))
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(u) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendFloat writes a formatted float to buf
func appendFloat(fmt format, f float64) {
	fmt.buf.B = strconv.AppendFloat(fmt.buf.B, f, 'G', -1, 64)
}

// appendFloatSlice writes a slice formatted floats to buf
func appendFloatSlice(fmt format, f []float64) {
	// Check for nil slice
	if f == nil {
		appendNilType(fmt, `[]float64`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, f := range f {
		appendFloat(fmt, f)
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(f) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendTime writes a formatted, quoted time string to buf
func appendTime(fmt format, t time.Time) {
	appendString(fmt.SetIsKey(true), t.Format(time.RFC1123))
}

// appendTimeSlice writes a slice of formatted time strings to buf
func appendTimeSlice(fmt format, t []time.Time) {
	// Check for nil slice
	if t == nil {
		appendNilType(fmt, `[]time.Time`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, t := range t {
		appendString(fmt.SetIsKey(true), t.Format(time.RFC1123))
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(t) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendDuration writes a formatted, quoted duration string to buf
func appendDuration(fmt format, d time.Duration) {
	appendString(fmt.SetIsKey(true), d.String())
}

// appendDurationSlice writes a slice of formatted, quoted duration strings to buf
func appendDurationSlice(fmt format, d []time.Duration) {
	// Check for nil slice
	if d == nil {
		appendNilType(fmt, `[]time.Duration`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, d := range d {
		appendString(fmt.SetIsKey(true), d.String())
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(d) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendComplex writes a formatted complex128 to buf
func appendComplex(fmt format, c complex128) {
	appendFloat(fmt, real(c))
	fmt.buf.WriteByte('+')
	appendFloat(fmt, imag(c))
	fmt.buf.WriteByte('i')
}

// appendComplexSlice writes a slice of formatted complex128s to buf
func appendComplexSlice(fmt format, c []complex128) {
	// Check for nil slice
	if c == nil {
		appendNilType(fmt, `[]complex128`)
		return
	}

	fmt.buf.WriteByte('[')

	// Write elements
	for _, c := range c {
		appendComplex(fmt, c)
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if len(c) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// notNil will safely check if 'v' is nil without dealing with weird Go interface nil bullshit.
func notNil(i interface{}) bool {
	// cast to get fat pointer
	e := *(*struct {
		typeOf  unsafe.Pointer // ignored
		valueOf unsafe.Pointer
	})(unsafe.Pointer(&i))

	// check if value part is nil
	return (e.valueOf != nil)
}

// appendIfaceOrRValueNext performs appendIfaceOrRValue checking + incr depth
func appendIfaceOrRValueNext(fmt format, i interface{}) {
	// Check we haven't hit max
	if fmt.AtMaxDepth() {
		fmt.buf.WriteString("...")
		return
	}

	// Incr the depth
	fmt = fmt.IncrDepth()

	// Make actual call
	appendIfaceOrRValue(fmt, i)
}

// appendIfaceOrReflectValue will attempt to append as interface, falling back to reflection
func appendIfaceOrRValue(fmt format, i interface{}) {
	if !appendIface(fmt, i) {
		appendRValue(fmt, reflect.ValueOf(i))
	}
}

// appendValueOrIfaceNext performs appendRValueOrIface checking + incr depth
func appendRValueOrIfaceNext(fmt format, v reflect.Value) {
	// Check we haven't hit max
	if fmt.AtMaxDepth() {
		fmt.buf.WriteString("...")
		return
	}

	// Incr the depth
	fmt = fmt.IncrDepth()

	// Make actual call
	appendRValueOrIface(fmt, v)
}

// appendRValueOrIface will attempt to interface the reflect.Value, falling back to using this directly
func appendRValueOrIface(fmt format, v reflect.Value) {
	if !v.CanInterface() || !appendIface(fmt, v.Interface()) {
		appendRValue(fmt, v)
	}
}

// appendIface parses and writes a formatted interface value to buf
func appendIface(fmt format, i interface{}) bool {
	switch i := i.(type) {
	case nil:
		fmt.buf.WriteString(`nil`)
	case byte:
		appendByte(fmt, i)
	case []byte:
		appendBytes(fmt, i)
	case string:
		appendString(fmt, i)
	case []string:
		appendStringSlice(fmt, i)
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
	case []int:
		appendIntSlice(fmt, i)
	case uint:
		appendUint(fmt, uint64(i))
	case uint16:
		appendUint(fmt, uint64(i))
	case uint32:
		appendUint(fmt, uint64(i))
	case uint64:
		appendUint(fmt, i)
	case []uint:
		appendUintSlice(fmt, i)
	case float32:
		appendFloat(fmt, float64(i))
	case float64:
		appendFloat(fmt, i)
	case []float64:
		appendFloatSlice(fmt, i)
	case bool:
		appendBool(fmt, i)
	case []bool:
		appendBoolSlice(fmt, i)
	case time.Time:
		appendTime(fmt, i)
	case []time.Time:
		appendTimeSlice(fmt, i)
	case time.Duration:
		appendDuration(fmt, i)
	case []time.Duration:
		appendDurationSlice(fmt, i)
	case complex64:
		appendComplex(fmt, complex128(i))
	case complex128:
		appendComplex(fmt, i)
	case []complex128:
		appendComplexSlice(fmt, i)
	case map[string]interface{}:
		appendIfaceMap(fmt, i)
	case error:
		if notNil(i) /* use safer nil check */ {
			appendString(fmt, i.Error())
		} else {
			appendNilIface(fmt, i)
		}
	case Formattable:
		switch {
		// catch nil case first
		case !notNil(i):
			appendNilIface(fmt, i)

		// not permitted
		case fmt.Verbose():
			return false

		// use func
		default:
			fmt.buf.B = i.AppendFormat(fmt.buf.B)
		}
	case stdfmt.Stringer:
		switch {
		// catch nil case first
		case !notNil(i):
			appendNilIface(fmt, i)

		// not permitted
		case fmt.Verbose():
			return false

		// use func
		default:
			appendString(fmt, i.String())
		}
	default:
		return false // could not handle
	}

	return true
}

// appendReflectValue will safely append a reflected value
func appendRValue(fmt format, v reflect.Value) {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		appendFloat(fmt, v.Float())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		appendInt(fmt, v.Int())
	case reflect.Uint8:
		appendByte(fmt, uint8(v.Uint()))
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		appendUint(fmt, v.Uint())
	case reflect.Bool:
		appendBool(fmt, v.Bool())
	case reflect.Array:
		appendArrayType(fmt, v)
	case reflect.Slice:
		appendSliceType(fmt, v)
	case reflect.Map:
		appendMapType(fmt, v)
	case reflect.Struct:
		appendStructType(fmt, v)
	case reflect.Ptr:
		if v.IsNil() {
			appendNilRValue(fmt, v)
		} else {
			appendRValue(fmt.IncrDerefs(), v.Elem())
		}
	case reflect.UnsafePointer:
		fmt.buf.WriteString("(unsafe.Pointer)")
		fmt.buf.WriteByte('(')
		if u := v.Pointer(); u != 0 {
			fmt.buf.WriteString("0x")
			fmt.buf.B = strconv.AppendUint(fmt.buf.B, uint64(u), 16)
		} else {
			fmt.buf.WriteString(`nil`)
		}
		fmt.buf.WriteByte(')')
	case reflect.Uintptr:
		fmt.buf.WriteString("(uintptr)")
		fmt.buf.WriteByte('(')
		if u := v.Uint(); u != 0 {
			fmt.buf.WriteString("0x")
			fmt.buf.B = strconv.AppendUint(fmt.buf.B, u, 16)
		} else {
			fmt.buf.WriteString(`nil`)
		}
		fmt.buf.WriteByte(')')
	case reflect.String:
		appendString(fmt, v.String())
	case reflect.Complex64, reflect.Complex128:
		appendComplex(fmt, v.Complex())
	case reflect.Func, reflect.Chan, reflect.Interface:
		if v.IsNil() {
			appendNilRValue(fmt, v)
		} else {
			fmt.buf.WriteString(v.String())
		}
	default:
		fmt.buf.WriteString(v.String())
	}
}

// appendIfaceMap writes a map of key-value pairs (as a set of fields) to buf
func appendIfaceMap(fmt format, v map[string]interface{}) {
	// Catch nil map
	if v == nil {
		appendNilType(fmt, `map[string]interface{}`)
		return
	}

	fmt.buf.WriteByte('{')

	// Write map pairs!
	for key, value := range v {
		appendString(fmt.SetIsKey(true), key)
		fmt.buf.WriteByte('=')
		appendIfaceOrRValueNext(fmt.SetIsKey(false), value)
		fmt.buf.WriteByte(' ')
	}

	// Drop last space
	if len(v) > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte('}')
}

// appendArrayType writes an array of unknown type (parsed by reflection) to buf, unlike appendSliceType does NOT catch nil slice
func appendArrayType(fmt format, v reflect.Value) {
	// get no. elements
	n := v.Len()

	fmt.buf.WriteByte('[')

	// Write values
	for i := 0; i < n; i++ {
		appendRValueOrIfaceNext(fmt.SetIsKey(false), v.Index(i))
		fmt.buf.WriteByte(',')
	}

	// Drop last comma
	if n > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte(']')
}

// appendSliceType writes a slice of unknown type (parsed by reflection) to buf
func appendSliceType(fmt format, v reflect.Value) {
	if v.IsNil() {
		appendNilRValue(fmt, v)
	} else {
		appendArrayType(fmt, v)
	}
}

// appendMapType writes a map of unknown types (parsed by reflection) to buf
func appendMapType(fmt format, v reflect.Value) {
	// Catch nil map
	if v.IsNil() {
		appendNilRValue(fmt, v)
		return
	}

	// Get a map iterator
	r := v.MapRange()
	n := v.Len()

	fmt.buf.WriteByte('{')

	// Iterate pairs
	for r.Next() {
		appendRValueOrIfaceNext(fmt.SetIsKey(true), r.Key())
		fmt.buf.WriteByte('=')
		appendRValueOrIfaceNext(fmt.SetIsKey(false), r.Value())
		fmt.buf.WriteByte(' ')
	}

	// Drop last space
	if n > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte('}')
}

// appendStructType writes a struct (as a set of key-value fields) to buf
func appendStructType(fmt format, v reflect.Value) {
	// Get value type & no. fields
	t := v.Type()
	n := v.NumField()
	w := 0

	// If verbose, append the type

	fmt.buf.WriteByte('{')

	// Iterate fields
	for i := 0; i < n; i++ {
		vfield := v.Field(i)
		name := t.Field(i).Name

		// Append field name
		appendString(fmt.SetIsKey(true), name)
		fmt.buf.WriteByte('=')

		if !vfield.CanInterface() {
			// This is an unexported field
			appendRValue(fmt.SetIsKey(false), vfield)
		} else {
			// This is an exported field!
			appendRValueOrIfaceNext(fmt.SetIsKey(false), vfield)
		}

		// Iter written count
		fmt.buf.WriteByte(' ')
		w++
	}

	// Drop last space
	if w > 0 {
		fmt.buf.Truncate(1)
	}

	fmt.buf.WriteByte('}')
}

// containsSpaceOrTab checks if "s" contains space or tabs
func containsSpaceOrTab(s string) bool {
	for _, r := range s {
		if r == ' ' || r == '\t' {
			return true
		}
	}
	return false
}
