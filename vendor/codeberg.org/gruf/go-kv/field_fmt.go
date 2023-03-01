//go:build !kvformat
// +build !kvformat

package kv

import (
	"fmt"
	"sync"

	"codeberg.org/gruf/go-byteutil"
)

// bufPool is a memory pool of byte buffers.
var bufPool = sync.Pool{
	New: func() interface{} {
		return &byteutil.Buffer{B: make([]byte, 0, 512)}
	},
}

// AppendFormat will append formatted format of Field to 'buf'. See .String() for details.
func (f Field) AppendFormat(buf *byteutil.Buffer, vbose bool) {
	var fmtstr string
	if vbose /* verbose */ {
		fmtstr = `%#v`
	} else /* regular */ {
		fmtstr = `%+v`
	}
	AppendQuote(buf, f.K)
	buf.WriteByte('=')
	appendValuef(buf, fmtstr, f.V)
}

// Value returns the formatted value string of this Field.
func (f Field) Value(vbose bool) string {
	var fmtstr string
	if vbose /* verbose */ {
		fmtstr = `%#v`
	} else /* regular */ {
		fmtstr = `%+v`
	}
	buf := byteutil.Buffer{B: make([]byte, 0, bufsize/2)}
	appendValuef(&buf, fmtstr, f.V)
	return buf.String()
}

// appendValuef appends a quoted value string (formatted by fmt.Appendf) to 'buf'.
func appendValuef(buf *byteutil.Buffer, format string, args ...interface{}) {
	// Write format string to a byte buffer
	fmtbuf := bufPool.Get().(*byteutil.Buffer)
	fmtbuf.B = fmt.Appendf(fmtbuf.B, format, args...)

	// Append quoted value to dst buffer
	AppendQuote(buf, fmtbuf.String())

	// Drop overly large capacity buffers
	if fmtbuf.Cap() > int(^uint16(0)) {
		return
	}

	// Replace buffer in pool
	fmtbuf.Reset()
	bufPool.Put(fmtbuf)
}
