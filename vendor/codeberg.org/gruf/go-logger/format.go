package logger

import (
	"time"

	"codeberg.org/gruf/go-bytes"
)

// Check our types impl LogFormat
var _ LogFormat = &TextFormat{}

// LogFormat defines a method of formatting log entries
type LogFormat interface {
	// AppendLevel appends given log level to the log buffer
	AppendLevel(buf *bytes.Buffer, lvl LEVEL)

	// AppendTimestamp appends given time format string to the log buffer
	AppendTimestamp(buf *bytes.Buffer, fmtNow string)

	// AppendField appends given key-value pair to the log buffer
	AppendField(buf *bytes.Buffer, key string, value interface{})

	// AppendFields appends given key-values pairs to the log buffer
	AppendFields(buf *bytes.Buffer, fields map[string]interface{})

	// AppendValue appends given interface formatted as value to the log buffer
	AppendValue(buf *bytes.Buffer, value interface{})

	// AppendValues appends given interfaces formatted as values to the log buffer
	AppendValues(buf *bytes.Buffer, slice []interface{})

	// AppendArgs appends given interfaces raw to the log buffer
	AppendArgs(buf *bytes.Buffer, args []interface{})

	// AppendByteField appends given byte value as key-value pair to the log buffer
	AppendByteField(buf *bytes.Buffer, key string, value byte)

	// AppendBytesField appends given byte slice value as key-value pair to the log buffer
	AppendBytesField(buf *bytes.Buffer, key string, value []byte)

	// AppendStringField appends given string value as key-value pair to the log buffer
	AppendStringField(buf *bytes.Buffer, key string, value string)

	// AppendStringsField appends given string slice value as key-value pair to the log buffer
	AppendStringsField(buf *bytes.Buffer, key string, value []string)

	// AppendBoolField appends given bool value as key-value pair to the log buffer
	AppendBoolField(buf *bytes.Buffer, key string, value bool)

	// AppendBoolsField appends given bool slice value as key-value pair to the log buffer
	AppendBoolsField(buf *bytes.Buffer, key string, value []bool)

	// AppendIntField appends given int value as key-value pair to the log buffer
	AppendIntField(buf *bytes.Buffer, key string, value int)

	// AppendIntsField appends given int slice value as key-value pair to the log buffer
	AppendIntsField(buf *bytes.Buffer, key string, value []int)

	// AppendUintField appends given uint value as key-value pair to the log buffer
	AppendUintField(buf *bytes.Buffer, key string, value uint)

	// AppendUintsField appends given uint slice value as key-value pair to the log buffer
	AppendUintsField(buf *bytes.Buffer, key string, value []uint)

	// AppendFloatField appends given float value as key-value pair to the log buffer
	AppendFloatField(buf *bytes.Buffer, key string, value float64)

	// AppendFloatsField appends given float slice value as key-value pair to the log buffer
	AppendFloatsField(buf *bytes.Buffer, key string, value []float64)

	// AppendTimeField appends given time value as key-value pair to the log buffer
	AppendTimeField(buf *bytes.Buffer, key string, value time.Time)

	// AppendTimesField appends given time slice value as key-value pair to the log buffer
	AppendTimesField(buf *bytes.Buffer, key string, value []time.Time)

	// AppendDurationField appends given duration value as key-value pair to the log buffer
	AppendDurationField(buf *bytes.Buffer, key string, value time.Duration)

	// AppendDurationsField appends given duration slice value as key-value pair to the log buffer
	AppendDurationsField(buf *bytes.Buffer, key string, value []time.Duration)

	// AppendMsg appends given msg as key-value pair to the log buffer using fmt.Sprint(...) formatting
	AppendMsg(buf *bytes.Buffer, a ...interface{})

	// AppendMsgf appends given msg format string as key-value pair to the log buffer using fmt.Sprintf(...) formatting
	AppendMsgf(buf *bytes.Buffer, s string, a ...interface{})
}
