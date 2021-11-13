package logger

import (
	"context"
	"time"

	"codeberg.org/gruf/go-bytes"
)

// Entry defines an entry in the log
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

// Byte appends a byte value as key-value pair to the log entry
func (e *Entry) Byte(key string, value byte) *Entry {
	e.log.Format.AppendByteField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Bytes appends a byte slice value as key-value pair to the log entry
func (e *Entry) Bytes(key string, value []byte) *Entry {
	e.log.Format.AppendBytesField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Str appends a string value as key-value pair to the log entry
func (e *Entry) Str(key string, value string) *Entry {
	e.log.Format.AppendStringField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Strs appends a string slice value as key-value pair to the log entry
func (e *Entry) Strs(key string, value []string) *Entry {
	e.log.Format.AppendStringsField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Int appends an int value as key-value pair to the log entry
func (e *Entry) Int(key string, value int) *Entry {
	e.log.Format.AppendIntField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Ints appends an int slice value as key-value pair to the log entry
func (e *Entry) Ints(key string, value []int) *Entry {
	e.log.Format.AppendIntsField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Uint appends a uint value as key-value pair to the log entry
func (e *Entry) Uint(key string, value uint) *Entry {
	e.log.Format.AppendUintField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Uints appends a uint slice value as key-value pair to the log entry
func (e *Entry) Uints(key string, value []uint) *Entry {
	e.log.Format.AppendUintsField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Float appends a float value as key-value pair to the log entry
func (e *Entry) Float(key string, value float64) *Entry {
	e.log.Format.AppendFloatField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Floats appends a float slice value as key-value pair to the log entry
func (e *Entry) Floats(key string, value []float64) *Entry {
	e.log.Format.AppendFloatsField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Bool appends a bool value as key-value pair to the log entry
func (e *Entry) Bool(key string, value bool) *Entry {
	e.log.Format.AppendBoolField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Bools appends a bool slice value as key-value pair to the log entry
func (e *Entry) Bools(key string, value []bool) *Entry {
	e.log.Format.AppendBoolsField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Time appends a time.Time value as key-value pair to the log entry
func (e *Entry) Time(key string, value time.Time) *Entry {
	e.log.Format.AppendTimeField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Times appends a time.Time slice value as key-value pair to the log entry
func (e *Entry) Times(key string, value []time.Time) *Entry {
	e.log.Format.AppendTimesField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Duration appends a time.Duration value as key-value pair to the log entry
func (e *Entry) Duration(key string, value time.Duration) *Entry {
	e.log.Format.AppendDurationField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Durations appends a time.Duration slice value as key-value pair to the log entry
func (e *Entry) Durations(key string, value []time.Duration) *Entry {
	e.log.Format.AppendDurationsField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Field appends an interface value as key-value pair to the log entry
func (e *Entry) Field(key string, value interface{}) *Entry {
	e.log.Format.AppendField(e.buf, key, value)
	e.buf.WriteByte(' ')
	return e
}

// Fields appends a map of key-value pairs to the log entry
func (e *Entry) Fields(fields map[string]interface{}) *Entry {
	e.log.Format.AppendFields(e.buf, fields)
	e.buf.WriteByte(' ')
	return e
}

// Msg appends the formatted final message to the log and calls .Send()
func (e *Entry) Msg(a ...interface{}) {
	e.log.Format.AppendMsg(e.buf, a...)
	e.Send()
}

// Msgf appends the formatted final message to the log and calls .Send()
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

	// Final new line
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
