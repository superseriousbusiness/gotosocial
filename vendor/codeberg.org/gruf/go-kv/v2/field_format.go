//go:build kvformat
// +build kvformat

package kv

import (
	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv/v2/format"
)

var argsDefault = format.DefaultArgs()

var argsVerbose = func() format.Args {
	args := format.DefaultArgs()
	args.SetWithType()
	args.SetNoMethod()
	return args
}()

// AppendFormat will append formatted format of Field to 'buf'. See .String() for details.
func (f Field) AppendFormat(buf *byteutil.Buffer, vbose bool) {
	var args format.Args
	if vbose {
		args = argsVerbose
	} else {
		args = argsDefault
	}
	AppendQuoteString(buf, f.K)
	buf.WriteByte('=')
	buf.B = format.Global.Append(buf.B, f.V, args)
}

// Value returns the formatted value string of this Field.
func (f Field) Value(vbose bool) string {
	var args format.Args
	if vbose {
		args = argsVerbose
	} else {
		args = argsDefault
	}
	buf := make([]byte, 0, bufsize/2)
	buf = format.Global.Append(buf, f.V, args)
	return byteutil.B2S(buf)
}
