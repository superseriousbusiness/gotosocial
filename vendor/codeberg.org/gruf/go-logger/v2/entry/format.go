package entry

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
)

// DefaultFormatter is the global default formatter, useful
// when in case of nil Formatter, or simply wanting defaults.
var DefaultFormatter = NewTextFormatter(nil)

// Formatter provides formatting for various standard
// data types being written to an entry's format buffer.
type Formatter interface {
	// Level will append formatted level string to buffer.
	Level(buf *byteutil.Buffer, lvl level.LEVEL)

	// Timestamp will append formatted timestamp string to buffer.
	Timestamp(buf *byteutil.Buffer)

	// Caller will append formatted caller information to buffer.
	Caller(buf *byteutil.Buffer, calldepth int)

	// Fields will append formatted key-value fields to buffer.
	Fields(buf *byteutil.Buffer, fields []kv.Field)

	// Msg will append formatted msg arguments to buffer.
	Msg(buf *byteutil.Buffer, args ...interface{})

	// Msgf will append formatted msg format string with arguments to buffer.
	Msgf(buf *byteutil.Buffer, format string, args ...interface{})
}

// TextFormatter is a Levels typedef to implement the Formatter interface,
// in a manner that provides logging with timestamp, caller info and level
// prefixing log entries in plain printed text.
type TextFormatter struct {
	lvls level.Levels
}

// NewTextFormatter returns a prepared TextFormatter instance, using
// provided level.Levels for printing of level strings, else uses defaults.
func NewTextFormatter(lvls *level.Levels) Formatter {
	if lvls == nil {
		// Use default levels
		l := level.Default()
		lvls = &l
	}
	return &TextFormatter{lvls: *lvls}
}

func (f *TextFormatter) Level(buf *byteutil.Buffer, lvl level.LEVEL) {
	buf.B = append(buf.B, '[')
	buf.B = append(buf.B, f.lvls.Get(lvl)...)
	buf.B = append(buf.B, ']')
}

func (f *TextFormatter) Timestamp(buf *byteutil.Buffer) {
	const fmt = "2006-01-02 15:04:05"
	buf.B = time.Now().AppendFormat(buf.B, fmt)
}

func (f *TextFormatter) Caller(buf *byteutil.Buffer, calldepth int) {
	fn, file, line := caller(calldepth + 1)
	buf.B = append(buf.B, file...)
	buf.B = append(buf.B, '#')
	buf.B = strconv.AppendInt(buf.B, int64(line), 10)
	buf.B = append(buf.B, ':')
	buf.B = append(buf.B, fn...)
	buf.B = append(buf.B, `()`...)
}

func (f *TextFormatter) Fields(buf *byteutil.Buffer, fields []kv.Field) {
	kv.Fields(fields).AppendFormat(buf)
}

func (f *TextFormatter) Msg(buf *byteutil.Buffer, a ...interface{}) {
	fmt.Fprint(buf, a...)
}

func (f *TextFormatter) Msgf(buf *byteutil.Buffer, s string, a ...interface{}) {
	fmt.Fprintf(buf, s, a...)
}

// FieldFormatter is a level.Levels typedef to implement the Formatter
// interface, in a manner that provides logging with timestamp, caller
// info and level prefixing log entries in key-value formatting.
type FieldFormatter struct {
	lvls level.Levels
	pool sync.Pool
}

// NewFieldFormatter returns a prepared FieldFormatter instance, using
// provided level.Levels for printing of level strings, else uses defaults.
func NewFieldFormatter(lvls *level.Levels) Formatter {
	if lvls == nil {
		// Use default levels
		l := level.Default()
		lvls = &l
	}

	fmt := &FieldFormatter{lvls: *lvls}

	// Setup a msg value buffer pool
	fmt.pool.New = func() interface{} {
		return &byteutil.Buffer{B: make([]byte, 0, 512)}
	}

	return fmt
}

// getBuffer returns a new message format buffer from memory pool.
func (f *FieldFormatter) getBuffer() *byteutil.Buffer {
	return f.pool.Get().(*byteutil.Buffer)
}

// putBuffer resets and places a message format buffer back in memory pool.
func (f *FieldFormatter) putBuffer(buf *byteutil.Buffer) {
	if buf.Cap() > int(^uint16(0)) {
		return // drop large buffer
	}
	buf.Reset()
	f.pool.Put(buf)
}

func (f *FieldFormatter) Level(buf *byteutil.Buffer, lvl level.LEVEL) {
	buf.B = append(buf.B, `level=`...)
	buf.B = append(buf.B, f.lvls.Get(lvl)...)
}

func (f *FieldFormatter) Timestamp(buf *byteutil.Buffer) {
	const fmt = "2006-01-02 15:04:05"
	buf.B = append(buf.B, `timestamp="`...)
	buf.B = time.Now().AppendFormat(buf.B, fmt)
	buf.B = append(buf.B, '"')
}

func (f *FieldFormatter) Caller(buf *byteutil.Buffer, calldepth int) {
	fn, file, line := caller(calldepth + 1)
	buf.B = append(buf.B, `file="`...)
	buf.B = append(buf.B, file...)
	buf.B = append(buf.B, ':')
	buf.B = strconv.AppendInt(buf.B, int64(line), 10)
	buf.B = append(buf.B, `" func="`...)
	buf.B = append(buf.B, fn...)
	buf.B = append(buf.B, `()"`...)
}

func (f *FieldFormatter) Fields(buf *byteutil.Buffer, fields []kv.Field) {
	kv.Fields(fields).AppendFormat(buf)
}

func (f *FieldFormatter) Msg(buf *byteutil.Buffer, a ...interface{}) {
	vbuf := f.getBuffer()
	fmt.Fprint(vbuf, a...)
	kv.Field{K: "msg", V: vbuf.String()}.AppendFormat(buf)
	f.putBuffer(vbuf)
}

func (f *FieldFormatter) Msgf(buf *byteutil.Buffer, s string, a ...interface{}) {
	vbuf := f.getBuffer()
	fmt.Fprintf(vbuf, s, a...)
	kv.Field{K: "msg", V: vbuf.String()}.AppendFormat(buf)
	f.putBuffer(vbuf)
}
