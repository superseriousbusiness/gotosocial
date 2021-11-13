package logger

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"codeberg.org/gruf/go-bytes"
)

// Check our types impl LogFormat
var _ LogFormat = &TextFormat{}

// LogFormat defines a method of formatting log entries
type LogFormat interface {
	AppendLevel(buf *bytes.Buffer, lvl LEVEL)
	AppendTimestamp(buf *bytes.Buffer, fmtNow string)
	AppendField(buf *bytes.Buffer, key string, value interface{})
	AppendFields(buf *bytes.Buffer, fields map[string]interface{})
	AppendByteField(buf *bytes.Buffer, key string, value byte)
	AppendBytesField(buf *bytes.Buffer, key string, value []byte)
	AppendStringField(buf *bytes.Buffer, key string, value string)
	AppendStringsField(buf *bytes.Buffer, key string, value []string)
	AppendBoolField(buf *bytes.Buffer, key string, value bool)
	AppendBoolsField(buf *bytes.Buffer, key string, value []bool)
	AppendIntField(buf *bytes.Buffer, key string, value int)
	AppendIntsField(buf *bytes.Buffer, key string, value []int)
	AppendUintField(buf *bytes.Buffer, key string, value uint)
	AppendUintsField(buf *bytes.Buffer, key string, value []uint)
	AppendFloatField(buf *bytes.Buffer, key string, value float64)
	AppendFloatsField(buf *bytes.Buffer, key string, value []float64)
	AppendTimeField(buf *bytes.Buffer, key string, value time.Time)
	AppendTimesField(buf *bytes.Buffer, key string, value []time.Time)
	AppendDurationField(buf *bytes.Buffer, key string, value time.Duration)
	AppendDurationsField(buf *bytes.Buffer, key string, value []time.Duration)
	AppendMsg(buf *bytes.Buffer, a ...interface{})
	AppendMsgf(buf *bytes.Buffer, s string, a ...interface{})
}

// TextFormat is the default LogFormat implementation, with very similar formatting to logfmt
type TextFormat struct {
	// Strict defines whether to use strict key-value pair formatting,
	// i.e. should the level, timestamp and msg be formatted as key-value pairs
	// or simply be printed as-is
	Strict bool

	// Levels defines the map of log LEVELs to level strings this LogFormat will use
	Levels Levels
}

// NewLogFmt returns a newly set LogFmt object, with DefaultLevels() set
func NewLogFmt(strict bool) *TextFormat {
	return &TextFormat{
		Strict: strict,
		Levels: DefaultLevels(),
	}
}

// appendReflectValue will safely append a reflected value
func appendReflectValue(buf *bytes.Buffer, v reflect.Value, isKey bool) {
	switch v.Kind() {
	case reflect.Slice:
		appendSliceType(buf, v)
	case reflect.Map:
		appendMapType(buf, v)
	case reflect.Struct:
		appendStructType(buf, v)
	case reflect.Ptr:
		if v.IsNil() {
			appendNil(buf)
		} else {
			appendIface(buf, v.Elem().Interface(), isKey)
		}
	default:
		// Just print reflect string
		appendString(buf, v.String())
	}
}

// appendKey should only be used in the case of directly setting key-value pairs,
// not in the case of appendMapType, appendStructType
func appendKey(buf *bytes.Buffer, key string) {
	if len(key) > 0 {
		// Only write key if here
		appendStringUnquoted(buf, key)
		buf.WriteByte('=')
	}
}

// appendSlice performs provided fn and writes square brackets [] around it
func appendSlice(buf *bytes.Buffer, fn func()) {
	buf.WriteByte('[')
	fn()
	buf.WriteByte(']')
}

// appendMap performs provided fn and writes curly braces {} around it
func appendMap(buf *bytes.Buffer, fn func()) {
	buf.WriteByte('{')
	fn()
	buf.WriteByte('}')
}

// appendStruct performs provided fn and writes curly braces {} around it
func appendStruct(buf *bytes.Buffer, fn func()) {
	buf.WriteByte('{')
	fn()
	buf.WriteByte('}')
}

// appendNil writes a nil value placeholder to buf
func appendNil(buf *bytes.Buffer) {
	buf.WriteString(`<nil>`)
}

// appendByte writes a single byte to buf
func appendByte(buf *bytes.Buffer, b byte) {
	buf.WriteByte(b)
}

// appendBytes writes a quoted byte slice to buf
func appendBytes(buf *bytes.Buffer, b []byte) {
	buf.WriteByte('"')
	buf.Write(b)
	buf.WriteByte('"')
}

// appendBytesUnquoted writes a byte slice to buf as-is
func appendBytesUnquoted(buf *bytes.Buffer, b []byte) {
	buf.Write(b)
}

// appendString writes a quoted string to buf
func appendString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	buf.WriteString(s)
	buf.WriteByte('"')
}

// appendStringUnquoted writes a string as-is to buf
func appendStringUnquoted(buf *bytes.Buffer, s string) {
	buf.WriteString(s)
}

// appendStringSlice writes a slice of strings to buf
func appendStringSlice(buf *bytes.Buffer, s []string) {
	appendSlice(buf, func() {
		for _, s := range s {
			appendString(buf, s)
			buf.WriteByte(',')
		}
		if len(s) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendBool writes a formatted bool to buf
func appendBool(buf *bytes.Buffer, b bool) {
	buf.B = strconv.AppendBool(buf.B, b)
}

// appendBool writes a slice of formatted bools to buf
func appendBoolSlice(buf *bytes.Buffer, b []bool) {
	appendSlice(buf, func() {
		// Write elements
		for _, b := range b {
			appendBool(buf, b)
			buf.WriteByte(',')
		}

		// Drop last comma
		if len(b) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendInt writes a formatted int to buf
func appendInt(buf *bytes.Buffer, i int64) {
	buf.B = strconv.AppendInt(buf.B, i, 10)
}

// appendIntSlice writes a slice of formatted int to buf
func appendIntSlice(buf *bytes.Buffer, i []int) {
	appendSlice(buf, func() {
		// Write elements
		for _, i := range i {
			appendInt(buf, int64(i))
			buf.WriteByte(',')
		}

		// Drop last comma
		if len(i) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendUint writes a formatted uint to buf
func appendUint(buf *bytes.Buffer, u uint64) {
	buf.B = strconv.AppendUint(buf.B, u, 10)
}

// appendUintSlice writes a slice of formatted uint to buf
func appendUintSlice(buf *bytes.Buffer, u []uint) {
	appendSlice(buf, func() {
		// Write elements
		for _, u := range u {
			appendUint(buf, uint64(u))
			buf.WriteByte(',')
		}

		// Drop last comma
		if len(u) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendFloat writes a formatted float to buf
func appendFloat(buf *bytes.Buffer, f float64) {
	buf.B = strconv.AppendFloat(buf.B, f, 'G', -1, 64)
}

// appendFloatSlice writes a slice formatted floats to buf
func appendFloatSlice(buf *bytes.Buffer, f []float64) {
	appendSlice(buf, func() {
		// Write elements
		for _, f := range f {
			appendFloat(buf, f)
			buf.WriteByte(',')
		}

		// Drop last comma
		if len(f) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendTime writes a formatted, quoted time string to buf
func appendTime(buf *bytes.Buffer, t time.Time) {
	buf.WriteByte('"')
	buf.B = t.AppendFormat(buf.B, time.RFC1123)
	buf.WriteByte('"')
}

// appendTimeUnquoted writes a formatted time string to buf as-is
func appendTimeUnquoted(buf *bytes.Buffer, t time.Time) {
	buf.B = t.AppendFormat(buf.B, time.RFC1123)
}

// appendTimeSlice writes a slice of formatted time strings to buf
func appendTimeSlice(buf *bytes.Buffer, t []time.Time) {
	appendSlice(buf, func() {
		// Write elements
		for _, t := range t {
			appendTime(buf, t)
			buf.WriteByte(',')
		}

		// Drop last comma
		if len(t) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendDuration writes a formatted, quoted duration string to buf
func appendDuration(buf *bytes.Buffer, d time.Duration) {
	appendString(buf, d.String())
}

// appendDurationUnquoted writes a formatted duration string to buf as-is
func appendDurationUnquoted(buf *bytes.Buffer, d time.Duration) {
	appendStringUnquoted(buf, d.String())
}

// appendDurationSlice writes a slice of formatted, quoted duration strings to buf
func appendDurationSlice(buf *bytes.Buffer, d []time.Duration) {
	appendSlice(buf, func() {
		// Write elements
		for _, d := range d {
			appendString(buf, d.String())
			buf.WriteByte(',')
		}

		// Drop last comma
		if len(d) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendIface parses and writes a formatted interface value to buf
func appendIface(buf *bytes.Buffer, i interface{}, isKey bool) {
	switch i := i.(type) {
	case nil:
		appendNil(buf)
	case byte:
		appendByte(buf, i)
	case []byte:
		if isKey {
			// Keys don't get quoted
			appendBytesUnquoted(buf, i)
		} else {
			appendBytes(buf, i)
		}
	case string:
		if isKey {
			// Keys don't get quoted
			appendStringUnquoted(buf, i)
		} else {
			appendString(buf, i)
		}
	case []string:
		appendStringSlice(buf, i)
	case int:
		appendInt(buf, int64(i))
	case int8:
		appendInt(buf, int64(i))
	case int16:
		appendInt(buf, int64(i))
	case int32:
		appendInt(buf, int64(i))
	case int64:
		appendInt(buf, i)
	case []int:
		appendIntSlice(buf, i)
	case uint:
		appendUint(buf, uint64(i))
	case uint16:
		appendUint(buf, uint64(i))
	case uint32:
		appendUint(buf, uint64(i))
	case uint64:
		appendUint(buf, i)
	case []uint:
		appendUintSlice(buf, i)
	case float32:
		appendFloat(buf, float64(i))
	case float64:
		appendFloat(buf, i)
	case []float64:
		appendFloatSlice(buf, i)
	case bool:
		appendBool(buf, i)
	case []bool:
		appendBoolSlice(buf, i)
	case time.Time:
		if isKey {
			// Keys don't get quoted
			appendTimeUnquoted(buf, i)
		} else {
			appendTime(buf, i)
		}
	case *time.Time:
		if isKey {
			// Keys don't get quoted
			appendTimeUnquoted(buf, *i)
		} else {
			appendTime(buf, *i)
		}
	case []time.Time:
		appendTimeSlice(buf, i)
	case time.Duration:
		if isKey {
			// Keys dont get quoted
			appendDurationUnquoted(buf, i)
		} else {
			appendDuration(buf, i)
		}
	case []time.Duration:
		appendDurationSlice(buf, i)
	case map[string]interface{}:
		appendIfaceMap(buf, i)
	case error:
		if isKey {
			// Keys don't get quoted
			appendStringUnquoted(buf, i.Error())
		} else {
			appendString(buf, i.Error())
		}
	case fmt.Stringer:
		if isKey {
			// Keys don't get quoted
			appendStringUnquoted(buf, i.String())
		} else {
			appendString(buf, i.String())
		}
	default:
		// If we reached here, we need reflection.
		appendReflectValue(buf, reflect.ValueOf(i), isKey)
	}
}

// appendIfaceMap writes a map of key-value pairs (as a set of fields) to buf
func appendIfaceMap(buf *bytes.Buffer, v map[string]interface{}) {
	appendMap(buf, func() {
		// Write map pairs!
		for key, value := range v {
			appendStringUnquoted(buf, key)
			buf.WriteByte('=')
			appendIface(buf, value, false)
			buf.WriteByte(' ')
		}

		// Drop last space
		if len(v) > 0 {
			buf.Truncate(1)
		}
	})
}

// appendSliceType writes a slice of unknown type (parsed by reflection) to buf
func appendSliceType(buf *bytes.Buffer, v reflect.Value) {
	n := v.Len()
	appendSlice(buf, func() {
		for i := 0; i < n; i++ {
			appendIface(buf, v.Index(i).Interface(), false)
			buf.WriteByte(',')
		}

		// Drop last comma
		if n > 0 {
			buf.Truncate(1)
		}
	})
}

// appendMapType writes a map of unknown types (parsed by reflection) to buf
func appendMapType(buf *bytes.Buffer, v reflect.Value) {
	// Get a map iterator
	r := v.MapRange()
	n := v.Len()

	// Now begin creating the map!
	appendMap(buf, func() {
		// Iterate pairs
		for r.Next() {
			appendIface(buf, r.Key().Interface(), true)
			buf.WriteByte('=')
			appendIface(buf, r.Value().Interface(), false)
			buf.WriteByte(' ')
		}

		// Drop last space
		if n > 0 {
			buf.Truncate(1)
		}
	})
}

// appendStructType writes a struct (as a set of key-value fields) to buf
func appendStructType(buf *bytes.Buffer, v reflect.Value) {
	// Get value type & no. fields
	t := v.Type()
	n := v.NumField()
	w := 0

	// Iter and write struct fields
	appendStruct(buf, func() {
		for i := 0; i < n; i++ {
			vfield := v.Field(i)

			// Check for unexported fields
			if !vfield.CanInterface() {
				continue
			}

			// Append the struct member as field key-value
			appendStringUnquoted(buf, t.Field(i).Name)
			buf.WriteByte('=')
			appendIface(buf, vfield.Interface(), false)
			buf.WriteByte(' ')

			// Iter written count
			w++
		}

		// Drop last space
		if w > 0 {
			buf.Truncate(1)
		}
	})
}

func (f *TextFormat) AppendLevel(buf *bytes.Buffer, lvl LEVEL) {
	if f.Strict {
		// Strict format, append level key
		buf.WriteString(`level=`)
		buf.WriteString(f.Levels.LevelString(lvl))
		return
	}

	// Write level string
	buf.WriteByte('[')
	buf.WriteString(f.Levels.LevelString(lvl))
	buf.WriteByte(']')
}

func (f *TextFormat) AppendTimestamp(buf *bytes.Buffer, now string) {
	if f.Strict {
		// Strict format, use key and quote
		appendStringUnquoted(buf, `time`)
		buf.WriteByte('=')
		appendString(buf, now)
		return
	}

	// Write time as-is
	appendStringUnquoted(buf, now)
}

func (f *TextFormat) AppendField(buf *bytes.Buffer, key string, value interface{}) {
	appendKey(buf, key)
	appendIface(buf, value, false)
}

func (f *TextFormat) AppendFields(buf *bytes.Buffer, fields map[string]interface{}) {
	// Append individual fields
	for key, value := range fields {
		appendKey(buf, key)
		appendIface(buf, value, false)
		buf.WriteByte(' ')
	}

	// Drop last space
	if len(fields) > 0 {
		buf.Truncate(1)
	}
}

func (f *TextFormat) AppendByteField(buf *bytes.Buffer, key string, value byte) {
	appendKey(buf, key)
	appendByte(buf, value)
}

func (f *TextFormat) AppendBytesField(buf *bytes.Buffer, key string, value []byte) {
	appendKey(buf, key)
	appendBytes(buf, value)
}

func (f *TextFormat) AppendStringField(buf *bytes.Buffer, key string, value string) {
	appendKey(buf, key)
	appendString(buf, value)
}

func (f *TextFormat) AppendStringsField(buf *bytes.Buffer, key string, value []string) {
	appendKey(buf, key)
	appendStringSlice(buf, value)
}

func (f *TextFormat) AppendBoolField(buf *bytes.Buffer, key string, value bool) {
	appendKey(buf, key)
	appendBool(buf, value)
}

func (f *TextFormat) AppendBoolsField(buf *bytes.Buffer, key string, value []bool) {
	appendKey(buf, key)
	appendBoolSlice(buf, value)
}

func (f *TextFormat) AppendIntField(buf *bytes.Buffer, key string, value int) {
	appendKey(buf, key)
	appendInt(buf, int64(value))
}

func (f *TextFormat) AppendIntsField(buf *bytes.Buffer, key string, value []int) {
	appendKey(buf, key)
	appendIntSlice(buf, value)
}

func (f *TextFormat) AppendUintField(buf *bytes.Buffer, key string, value uint) {
	appendKey(buf, key)
	appendUint(buf, uint64(value))
}

func (f *TextFormat) AppendUintsField(buf *bytes.Buffer, key string, value []uint) {
	appendKey(buf, key)
	appendUintSlice(buf, value)
}

func (f *TextFormat) AppendFloatField(buf *bytes.Buffer, key string, value float64) {
	appendKey(buf, key)
	appendFloat(buf, value)
}

func (f *TextFormat) AppendFloatsField(buf *bytes.Buffer, key string, value []float64) {
	appendKey(buf, key)
	appendFloatSlice(buf, value)
}

func (f *TextFormat) AppendTimeField(buf *bytes.Buffer, key string, value time.Time) {
	appendKey(buf, key)
	appendTime(buf, value)
}

func (f *TextFormat) AppendTimesField(buf *bytes.Buffer, key string, value []time.Time) {
	appendKey(buf, key)
	appendTimeSlice(buf, value)
}

func (f *TextFormat) AppendDurationField(buf *bytes.Buffer, key string, value time.Duration) {
	appendKey(buf, key)
	appendDuration(buf, value)
}

func (f *TextFormat) AppendDurationsField(buf *bytes.Buffer, key string, value []time.Duration) {
	appendKey(buf, key)
	appendDurationSlice(buf, value)
}

func (f *TextFormat) AppendMsg(buf *bytes.Buffer, a ...interface{}) {
	if f.Strict {
		// Strict format, use key and quote
		buf.WriteString(`msg="`)
		fmt.Fprint(buf, a...)
		buf.WriteByte('"')
		return
	}

	// Write message as-is
	fmt.Fprint(buf, a...)
}

func (f *TextFormat) AppendMsgf(buf *bytes.Buffer, s string, a ...interface{}) {
	if f.Strict {
		// Strict format, use key and quote
		buf.WriteString(`msg="`)
		fmt.Fprintf(buf, s, a...)
		buf.WriteByte('"')
		return
	}

	// Write message as-is
	fmt.Fprintf(buf, s, a...)
}
