//go:build !kvformat
// +build !kvformat

package kv

import (
	"fmt"

	"codeberg.org/gruf/go-byteutil"
)

// AppendFormat will append formatted format of Field to 'buf'. See .String() for details.
func (f Field) AppendFormat(buf *byteutil.Buffer) {
	var fmtstr string
	if f.x.Get0() /* verbose */ {
		fmtstr = `%#v`
	} else /* regular */ {
		fmtstr = `%+v`
	}
	appendQuoteKey(buf, f.K)
	buf.WriteByte('=')
	appendQuoteValue(buf, fmt.Sprintf(fmtstr, f.V))
}

// Value returns the formatted value string of this Field.
func (f Field) Value() string {
	var fmtstr string
	if f.x.Get0() /* verbose */ {
		fmtstr = `%#v`
	} else /* regular */ {
		fmtstr = `%+v`
	}
	buf := byteutil.Buffer{B: make([]byte, 0, bufsize/2)}
	appendQuoteValue(&buf, fmt.Sprintf(fmtstr, f.V))
	return buf.String()
}
