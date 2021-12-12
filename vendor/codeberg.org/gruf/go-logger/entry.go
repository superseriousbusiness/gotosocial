package logger

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/gruf/go-bytes"
)

// Entry defines an entry in the log, it is NOT safe for concurrent use
type Entry struct {
	ctx context.Context
	lvl LEVEL
	buf *bytes.Buffer
	log *Logger
}

// Context returns the current set Entry context.Context
func (e *Entry) Context() context.Context {
	return e.ctx
}

// WithContext updates Entry context value to the supplied
func (e *Entry) WithContext(ctx context.Context) *Entry {
	e.ctx = ctx
	return e
}

// Level appends the supplied level to the log entry, and sets the entry level.
// Please note this CAN be called and append log levels multiple times
func (e *Entry) Level(lvl LEVEL) *Entry {
	e.log.Format.AppendLevel(e.buf, lvl)
	e.buf.WriteByte(' ')
	e.lvl = lvl
	return e
}

// Timestamp appends the current timestamp to the log entry. Please note this
// CAN be called and append the timestamp multiple times
func (e *Entry) Timestamp() *Entry {
	e.log.Format.AppendTimestamp(e.buf, clock.NowFormat())
	e.buf.WriteByte(' ')
	return e
}

// TimestampIf performs Entry.Timestamp() only IF timestamping is enabled for the Logger.
// Please note this CAN be called multiple times
func (e *Entry) TimestampIf() *Entry {
	if e.log.Timestamp {
		e.Timestamp()
	}
	return e
}

// Hooks applies currently set Hooks to the Entry. Please note this CAN be
// called and perform the Hooks multiple times
func (e *Entry) Hooks() *Entry {
	for _, hook := range e.log.Hooks {
		hook.Do(e)
	}
	return e
}

// Byte appends a byte value to the log entry
func (e *Entry) Byte(value byte) *Entry {
	e.log.Format.AppendByte(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// ByteField appends a byte value as key-value pair to the log entry
func (e *Entry) ByteField(key string, value byte) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendByte(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Bytes appends a byte slice value as to the log entry
func (e *Entry) Bytes(value []byte) *Entry {
	e.log.Format.AppendBytes(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// BytesField appends a byte slice value as key-value pair to the log entry
func (e *Entry) BytesField(key string, value []byte) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendBytes(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Str appends a string value to the log entry
func (e *Entry) Str(value string) *Entry {
	e.log.Format.AppendString(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// StrField appends a string value as key-value pair to the log entry
func (e *Entry) StrField(key string, value string) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendString(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Strs appends a string slice value to the log entry
func (e *Entry) Strs(value []string) *Entry {
	e.log.Format.AppendStrings(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// StrsField appends a string slice value as key-value pair to the log entry
func (e *Entry) StrsField(key string, value []string) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendStrings(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Int appends an int value to the log entry
func (e *Entry) Int(value int) *Entry {
	e.log.Format.AppendInt(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// IntField appends an int value as key-value pair to the log entry
func (e *Entry) IntField(key string, value int) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendInt(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Ints appends an int slice value to the log entry
func (e *Entry) Ints(value []int) *Entry {
	e.log.Format.AppendInts(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// IntsField appends an int slice value as key-value pair to the log entry
func (e *Entry) IntsField(key string, value []int) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendInts(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Uint appends a uint value to the log entry
func (e *Entry) Uint(value uint) *Entry {
	e.log.Format.AppendUint(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// UintField appends a uint value as key-value pair to the log entry
func (e *Entry) UintField(key string, value uint) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendUint(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Uints appends a uint slice value to the log entry
func (e *Entry) Uints(value []uint) *Entry {
	e.log.Format.AppendUints(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// UintsField appends a uint slice value as key-value pair to the log entry
func (e *Entry) UintsField(key string, value []uint) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendUints(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Float appends a float value to the log entry
func (e *Entry) Float(value float64) *Entry {
	e.log.Format.AppendFloat(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// FloatField appends a float value as key-value pair to the log entry
func (e *Entry) FloatField(key string, value float64) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendFloat(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Floats appends a float slice value to the log entry
func (e *Entry) Floats(value []float64) *Entry {
	e.log.Format.AppendFloats(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// FloatsField appends a float slice value as key-value pair to the log entry
func (e *Entry) FloatsField(key string, value []float64) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendFloats(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Bool appends a bool value to the log entry
func (e *Entry) Bool(value bool) *Entry {
	e.log.Format.AppendBool(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// BoolField appends a bool value as key-value pair to the log entry
func (e *Entry) BoolField(key string, value bool) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendBool(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Bools appends a bool slice value to the log entry
func (e *Entry) Bools(value []bool) *Entry {
	e.log.Format.AppendBools(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// BoolsField appends a bool slice value as key-value pair to the log entry
func (e *Entry) BoolsField(key string, value []bool) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendBools(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Time appends a time.Time value to the log entry
func (e *Entry) Time(value time.Time) *Entry {
	e.log.Format.AppendTime(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// TimeField appends a time.Time value as key-value pair to the log entry
func (e *Entry) TimeField(key string, value time.Time) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendTime(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Times appends a time.Time slice value to the log entry
func (e *Entry) Times(value []time.Time) *Entry {
	e.log.Format.AppendTimes(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// TimesField appends a time.Time slice value as key-value pair to the log entry
func (e *Entry) TimesField(key string, value []time.Time) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendTimes(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// DurationField appends a time.Duration value to the log entry
func (e *Entry) Duration(value time.Duration) *Entry {
	e.log.Format.AppendDuration(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// DurationField appends a time.Duration value as key-value pair to the log entry
func (e *Entry) DurationField(key string, value time.Duration) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendDuration(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Durations appends a time.Duration slice value to the log entry
func (e *Entry) Durations(value []time.Duration) *Entry {
	e.log.Format.AppendDurations(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// DurationsField appends a time.Duration slice value as key-value pair to the log entry
func (e *Entry) DurationsField(key string, value []time.Duration) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendDurations(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Field appends an interface value as key-value pair to the log entry
func (e *Entry) Field(key string, value interface{}) *Entry {
	e.log.Format.AppendKey(e.buf, key)
	e.log.Format.AppendValue(e.buf, value)
	e.buf.WriteByte(' ')
	return e
}

// Fields appends a map of key-value pairs to the log entry
func (e *Entry) Fields(fields map[string]interface{}) *Entry {
	for key, value := range fields {
		e.Field(key, value)
	}
	return e
}

// Values appends the given values to the log entry formatted as values, without a key.
func (e *Entry) Values(values ...interface{}) *Entry {
	for _, value := range values {
		e.log.Format.AppendValue(e.buf, value)
		e.buf.WriteByte(' ')
	}
	return e
}

// Append will append the given args formatted using fmt.Sprint(a...) to the Entry.
func (e *Entry) Append(a ...interface{}) *Entry {
	fmt.Fprint(e.buf, a...)
	e.buf.WriteByte(' ')
	return e
}

// Appendf will append the given format string and args using fmt.Sprintf(s, a...) to the Entry.
func (e *Entry) Appendf(s string, a ...interface{}) *Entry {
	fmt.Fprintf(e.buf, s, a...)
	e.buf.WriteByte(' ')
	return e
}

// Msg appends the fmt.Sprint() formatted final message to the log and calls .Send()
func (e *Entry) Msg(a ...interface{}) {
	e.log.Format.AppendMsg(e.buf, a...)
	e.Send()
}

// Msgf appends the fmt.Sprintf() formatted final message to the log and calls .Send()
func (e *Entry) Msgf(s string, a ...interface{}) {
	e.log.Format.AppendMsgf(e.buf, s, a...)
	e.Send()
}

// Send triggers write of the log entry, skipping if the entry's log LEVEL
// is below the currently set Logger level, and releases the Entry back to
// the Logger's Entry pool. So it is NOT safe to continue using this Entry
// object after calling .Send(), .Msg() or .Msgf()
func (e *Entry) Send() {
	// If nothing to do, return
	if e.lvl < e.log.Level || e.buf.Len() < 1 {
		e.reset()
		return
	}

	// Ensure a final new line
	if e.buf.B[e.buf.Len()-1] != '\n' {
		e.buf.WriteByte('\n')
	}

	// Write, reset and release
	e.log.Output.Write(e.buf.B)
	e.reset()
}

func (e *Entry) reset() {
	// Reset all
	e.ctx = nil
	e.buf.Reset()
	e.lvl = unset

	// Release to pool
	e.log.pool.Put(e)
}
