package json

import (
	"io"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/encoder/vm"
	"github.com/goccy/go-json/internal/encoder/vm_escaped"
	"github.com/goccy/go-json/internal/encoder/vm_escaped_indent"
	"github.com/goccy/go-json/internal/encoder/vm_indent"
)

// An Encoder writes JSON values to an output stream.
type Encoder struct {
	w                 io.Writer
	enabledIndent     bool
	enabledHTMLEscape bool
	prefix            string
	indentStr         string
}

type EncodeOption int

const (
	EncodeOptionHTMLEscape EncodeOption = 1 << iota
	EncodeOptionIndent
	EncodeOptionUnorderedMap
	EncodeOptionDebug
)

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, enabledHTMLEscape: true}
}

// Encode writes the JSON encoding of v to the stream, followed by a newline character.
//
// See the documentation for Marshal for details about the conversion of Go values to JSON.
func (e *Encoder) Encode(v interface{}) error {
	return e.EncodeWithOption(v)
}

// EncodeWithOption call Encode with EncodeOption.
func (e *Encoder) EncodeWithOption(v interface{}, optFuncs ...EncodeOptionFunc) error {
	ctx := encoder.TakeRuntimeContext()

	err := e.encodeWithOption(ctx, v, optFuncs...)

	encoder.ReleaseRuntimeContext(ctx)
	return err
}

func (e *Encoder) encodeWithOption(ctx *encoder.RuntimeContext, v interface{}, optFuncs ...EncodeOptionFunc) error {
	var opt EncodeOption
	if e.enabledHTMLEscape {
		opt |= EncodeOptionHTMLEscape
	}
	for _, optFunc := range optFuncs {
		opt = optFunc(opt)
	}
	var (
		buf []byte
		err error
	)
	if e.enabledIndent {
		buf, err = encodeIndent(ctx, v, e.prefix, e.indentStr, opt)
	} else {
		buf, err = encode(ctx, v, opt)
	}
	if err != nil {
		return err
	}
	if e.enabledIndent {
		buf = buf[:len(buf)-2]
	} else {
		buf = buf[:len(buf)-1]
	}
	buf = append(buf, '\n')
	if _, err := e.w.Write(buf); err != nil {
		return err
	}
	return nil
}

// SetEscapeHTML specifies whether problematic HTML characters should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability of the output, SetEscapeHTML(false) disables this behavior.
func (e *Encoder) SetEscapeHTML(on bool) {
	e.enabledHTMLEscape = on
}

// SetIndent instructs the encoder to format each subsequent encoded value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (e *Encoder) SetIndent(prefix, indent string) {
	if prefix == "" && indent == "" {
		e.enabledIndent = false
		return
	}
	e.prefix = prefix
	e.indentStr = indent
	e.enabledIndent = true
}

func marshal(v interface{}, opt EncodeOption) ([]byte, error) {
	ctx := encoder.TakeRuntimeContext()

	buf, err := encode(ctx, v, opt|EncodeOptionHTMLEscape)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}

	// this line exists to escape call of `runtime.makeslicecopy` .
	// if use `make([]byte, len(buf)-1)` and `copy(copied, buf)`,
	// dst buffer size and src buffer size are differrent.
	// in this case, compiler uses `runtime.makeslicecopy`, but it is slow.
	buf = buf[:len(buf)-1]
	copied := make([]byte, len(buf))
	copy(copied, buf)

	encoder.ReleaseRuntimeContext(ctx)
	return copied, nil
}

func marshalNoEscape(v interface{}, opt EncodeOption) ([]byte, error) {
	ctx := encoder.TakeRuntimeContext()

	buf, err := encodeNoEscape(ctx, v, opt|EncodeOptionHTMLEscape)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}

	// this line exists to escape call of `runtime.makeslicecopy` .
	// if use `make([]byte, len(buf)-1)` and `copy(copied, buf)`,
	// dst buffer size and src buffer size are differrent.
	// in this case, compiler uses `runtime.makeslicecopy`, but it is slow.
	buf = buf[:len(buf)-1]
	copied := make([]byte, len(buf))
	copy(copied, buf)

	encoder.ReleaseRuntimeContext(ctx)
	return copied, nil
}

func marshalIndent(v interface{}, prefix, indent string, opt EncodeOption) ([]byte, error) {
	ctx := encoder.TakeRuntimeContext()

	buf, err := encodeIndent(ctx, v, prefix, indent, opt|EncodeOptionHTMLEscape)
	if err != nil {
		encoder.ReleaseRuntimeContext(ctx)
		return nil, err
	}

	buf = buf[:len(buf)-2]
	copied := make([]byte, len(buf))
	copy(copied, buf)

	encoder.ReleaseRuntimeContext(ctx)
	return copied, nil
}

func encode(ctx *encoder.RuntimeContext, v interface{}, opt EncodeOption) ([]byte, error) {
	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(b)
		b = encoder.AppendComma(b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encoder.CompileToGetCodeSet(typeptr)
	if err != nil {
		return nil, err
	}

	p := uintptr(header.ptr)
	ctx.Init(p, codeSet.CodeLength)
	ctx.KeepRefs = append(ctx.KeepRefs, header.ptr)

	buf, err := encodeRunCode(ctx, b, codeSet, opt)
	if err != nil {
		return nil, err
	}
	ctx.Buf = buf
	return buf, nil
}

func encodeNoEscape(ctx *encoder.RuntimeContext, v interface{}, opt EncodeOption) ([]byte, error) {
	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(b)
		b = encoder.AppendComma(b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encoder.CompileToGetCodeSet(typeptr)
	if err != nil {
		return nil, err
	}

	p := uintptr(header.ptr)
	ctx.Init(p, codeSet.CodeLength)
	buf, err := encodeRunCode(ctx, b, codeSet, opt)
	if err != nil {
		return nil, err
	}

	ctx.Buf = buf
	return buf, nil
}

func encodeIndent(ctx *encoder.RuntimeContext, v interface{}, prefix, indent string, opt EncodeOption) ([]byte, error) {
	b := ctx.Buf[:0]
	if v == nil {
		b = encoder.AppendNull(b)
		b = encoder.AppendCommaIndent(b)
		return b, nil
	}
	header := (*emptyInterface)(unsafe.Pointer(&v))
	typ := header.typ

	typeptr := uintptr(unsafe.Pointer(typ))
	codeSet, err := encoder.CompileToGetCodeSet(typeptr)
	if err != nil {
		return nil, err
	}

	p := uintptr(header.ptr)
	ctx.Init(p, codeSet.CodeLength)
	buf, err := encodeRunIndentCode(ctx, b, codeSet, prefix, indent, opt)

	ctx.KeepRefs = append(ctx.KeepRefs, header.ptr)

	if err != nil {
		return nil, err
	}

	ctx.Buf = buf
	return buf, nil
}

func encodeRunCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, opt EncodeOption) ([]byte, error) {
	if (opt & EncodeOptionDebug) != 0 {
		return encodeDebugRunCode(ctx, b, codeSet, opt)
	}
	if (opt & EncodeOptionHTMLEscape) != 0 {
		return vm_escaped.Run(ctx, b, codeSet, encoder.Option(opt))
	}
	return vm.Run(ctx, b, codeSet, encoder.Option(opt))
}

func encodeDebugRunCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, opt EncodeOption) ([]byte, error) {
	if (opt & EncodeOptionHTMLEscape) != 0 {
		return vm_escaped.DebugRun(ctx, b, codeSet, encoder.Option(opt))
	}
	return vm.DebugRun(ctx, b, codeSet, encoder.Option(opt))
}

func encodeRunIndentCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, prefix, indent string, opt EncodeOption) ([]byte, error) {
	ctx.Prefix = []byte(prefix)
	ctx.IndentStr = []byte(indent)
	if (opt & EncodeOptionDebug) != 0 {
		return encodeDebugRunIndentCode(ctx, b, codeSet, opt)
	}
	if (opt & EncodeOptionHTMLEscape) != 0 {
		return vm_escaped_indent.Run(ctx, b, codeSet, encoder.Option(opt))
	}
	return vm_indent.Run(ctx, b, codeSet, encoder.Option(opt))
}

func encodeDebugRunIndentCode(ctx *encoder.RuntimeContext, b []byte, codeSet *encoder.OpcodeSet, opt EncodeOption) ([]byte, error) {
	if (opt & EncodeOptionHTMLEscape) != 0 {
		return vm_escaped_indent.DebugRun(ctx, b, codeSet, encoder.Option(opt))
	}
	return vm_indent.DebugRun(ctx, b, codeSet, encoder.Option(opt))
}
