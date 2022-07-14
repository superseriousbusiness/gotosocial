package entry

import (
	"context"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

// Entry represents a single entry in a log. It provides methods
// for writing formatted data to a byte buffer before making a
// singular write to an io.Writer.
type Entry struct {
	ctx context.Context
	lvl level.LEVEL
	buf byteutil.Buffer
	fmt Formatter
}

// NewEntry returns a new Entry initialized with buffer.
func NewEntry(b []byte, fmt Formatter) *Entry {
	e := &Entry{buf: byteutil.Buffer{B: b}}
	e.SetFormat(fmt)
	e.Reset()
	return e
}

// SetFormat will update the entry's formatter to provided.
func (e *Entry) SetFormat(fmt Formatter) {
	if fmt == nil {
		// use default if nil
		fmt = DefaultFormatter
	}
	e.fmt = fmt
}

// Context returns the context associated with this entry.
func (e *Entry) Context() context.Context {
	return e.ctx
}

// WithContext updates the entry to use the provided context.
func (e *Entry) WithContext(ctx context.Context) {
	e.ctx = ctx
}

// Level returns the currently set entry level.
func (e *Entry) Level() level.LEVEL {
	return e.lvl
}

// WithLevel updates the entry to provided level and writes formatted level to the entry.
func (e *Entry) WithLevel(lvl level.LEVEL) {
	e.fmt.Level(&e.buf, lvl)
	e.buf.B = append(e.buf.B, ' ')
	e.lvl = lvl
}

// Timestamp will write formatted timestamp to the entry.
func (e *Entry) Timestamp() {
	e.fmt.Timestamp(&e.buf)
	e.buf.B = append(e.buf.B, ' ')
}

// Caller will write formatted caller information to the entry.
func (e *Entry) Caller(calldepth int) {
	e.fmt.Caller(&e.buf, calldepth+1)
	e.buf.B = append(e.buf.B, ' ')
}

// Field will write formatted key-value pair as field to the entry.
func (e *Entry) Field(key string, value interface{}) {
	e.Fields(kv.Field{K: key, V: value})
}

// Fields will write formatted key-value fields to the entry.
func (e *Entry) Fields(fields ...kv.Field) {
	e.fmt.Fields(&e.buf, fields)
	e.buf.B = append(e.buf.B, ' ')
}

// Msg will write given arguments as message to the entry.
func (e *Entry) Msg(a ...interface{}) {
	e.fmt.Msg(&e.buf, a...)
	e.buf.B = append(e.buf.B, ' ')
}

// Msgf will write given format string and arguments as message to the entry.
func (e *Entry) Msgf(s string, a ...interface{}) {
	e.fmt.Msgf(&e.buf, s, a...)
	e.buf.B = append(e.buf.B, ' ')
}

// Buffer returns the underlying format buffer.
func (e *Entry) Buffer() *byteutil.Buffer {
	return &e.buf
}

// Bytes returns the currently accumulated entry buffer bytes.
func (e *Entry) Bytes() []byte {
	return e.buf.B
}

// String returns the currently accumulated entry buffer bytes as
// string, please note this points to the underlying byte slice.
func (e *Entry) String() string {
	return byteutil.B2S(e.buf.B)
}

// Reset will reset the entry's current context, level and buffer.
func (e *Entry) Reset() {
	e.ctx = context.Background()
	e.lvl = level.UNSET
	e.buf.Reset()
}
