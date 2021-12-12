package logger

import (
	"time"

	"codeberg.org/gruf/go-bytes"
)

// Check our types impl LogFormat
var _ LogFormat = &TextFormat{}

// Formattable defines a type capable of writing a string formatted form
// of itself to a supplied byte buffer, and returning the resulting byte
// buffer. Implementing this will greatly speed up formatting of custom
// types passed to LogFormat (assuming they implement checking for this).
type Formattable interface {
	AppendFormat([]byte) []byte
}

// LogFormat defines a method of formatting log entries
type LogFormat interface {
	// AppendKey appends given key to the log buffer
	AppendKey(buf *bytes.Buffer, key string)

	// AppendLevel appends given log level as key-value pair to the log buffer
	AppendLevel(buf *bytes.Buffer, lvl LEVEL)

	// AppendTimestamp appends given timestamp string as key-value pair to the log buffer
	AppendTimestamp(buf *bytes.Buffer, fmtNow string)

	// AppendValue appends given interface formatted as value to the log buffer
	AppendValue(buf *bytes.Buffer, value interface{})

	// AppendByte appends given byte value to the log buffer
	AppendByte(buf *bytes.Buffer, value byte)

	// AppendBytes appends given byte slice value to the log buffer
	AppendBytes(buf *bytes.Buffer, value []byte)

	// AppendString appends given string value to the log buffer
	AppendString(buf *bytes.Buffer, value string)

	// AppendStrings appends given string slice value to the log buffer
	AppendStrings(buf *bytes.Buffer, value []string)

	// AppendBool appends given bool value to the log buffer
	AppendBool(buf *bytes.Buffer, value bool)

	// AppendBools appends given bool slice value to the log buffer
	AppendBools(buf *bytes.Buffer, value []bool)

	// AppendInt appends given int value to the log buffer
	AppendInt(buf *bytes.Buffer, value int)

	// AppendInts appends given int slice value to the log buffer
	AppendInts(buf *bytes.Buffer, value []int)

	// AppendUint appends given uint value to the log buffer
	AppendUint(buf *bytes.Buffer, value uint)

	// AppendUints appends given uint slice value to the log buffer
	AppendUints(buf *bytes.Buffer, value []uint)

	// AppendFloat appends given float value to the log buffer
	AppendFloat(buf *bytes.Buffer, value float64)

	// AppendFloats appends given float slice value to the log buffer
	AppendFloats(buf *bytes.Buffer, value []float64)

	// AppendTime appends given time value to the log buffer
	AppendTime(buf *bytes.Buffer, value time.Time)

	// AppendTimes appends given time slice value to the log buffer
	AppendTimes(buf *bytes.Buffer, value []time.Time)

	// AppendDuration appends given duration value to the log buffer
	AppendDuration(buf *bytes.Buffer, value time.Duration)

	// AppendDurations appends given duration slice value to the log buffer
	AppendDurations(buf *bytes.Buffer, value []time.Duration)

	// AppendMsg appends given msg as key-value pair to the log buffer using fmt.Sprint(...) formatting
	AppendMsg(buf *bytes.Buffer, a ...interface{})

	// AppendMsgf appends given msg format string as key-value pair to the log buffer using fmt.Sprintf(...) formatting
	AppendMsgf(buf *bytes.Buffer, s string, a ...interface{})
}
