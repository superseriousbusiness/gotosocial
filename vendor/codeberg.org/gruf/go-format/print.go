package format

import (
	"io"
	"os"
	"sync"
)

// pool is the global printer buffer pool.
var pool = sync.Pool{
	New: func() interface{} {
		return &Buffer{}
	},
}

// getBuf fetches a buffer from pool.
func getBuf() *Buffer {
	return pool.Get().(*Buffer)
}

// putBuf places a Buffer back in pool.
func putBuf(buf *Buffer) {
	if buf.Cap() > 64<<10 {
		return // drop large
	}
	buf.Reset()
	pool.Put(buf)
}

// Sprint will format supplied values, returning this string.
func Sprint(v ...interface{}) string {
	buf := Buffer{}
	Append(&buf, v...)
	return buf.String()
}

// Sprintf will format supplied format string and args, returning this string.
// See Formatter.Appendf() for more documentation.
func Sprintf(s string, a ...interface{}) string {
	buf := Buffer{}
	Appendf(&buf, s, a...)
	return buf.String()
}

// Print will format supplied values, print this to os.Stdout.
func Print(v ...interface{}) {
	Fprint(os.Stdout, v...) //nolint
}

// Printf will format supplied format string and args, printing this to os.Stdout.
// See Formatter.Appendf() for more documentation.
func Printf(s string, a ...interface{}) {
	Fprintf(os.Stdout, s, a...) //nolint
}

// Println will format supplied values, append a trailing newline and print this to os.Stdout.
func Println(v ...interface{}) {
	Fprintln(os.Stdout, v...) //nolint
}

// Fprint will format supplied values, writing this to an io.Writer.
func Fprint(w io.Writer, v ...interface{}) (int, error) {
	buf := getBuf()
	Append(buf, v...)
	n, err := w.Write(buf.B)
	putBuf(buf)
	return n, err
}

// Fprintf will format supplied format string and args, writing this to an io.Writer.
// See Formatter.Appendf() for more documentation.
func Fprintf(w io.Writer, s string, a ...interface{}) (int, error) {
	buf := getBuf()
	Appendf(buf, s, a...)
	n, err := w.Write(buf.B)
	putBuf(buf)
	return n, err
}

// Println will format supplied values, append a trailing newline and writer this to an io.Writer.
func Fprintln(w io.Writer, v ...interface{}) (int, error) {
	buf := getBuf()
	Append(buf, v...)
	buf.AppendByte('\n')
	n, err := w.Write(buf.B)
	putBuf(buf)
	return n, err
}
