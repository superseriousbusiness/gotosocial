package vm_indent

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"github.com/goccy/go-json/internal/encoder"
	"github.com/goccy/go-json/internal/runtime"
)

const uintptrSize = 4 << (^uintptr(0) >> 63)

var (
	appendInt           = encoder.AppendInt
	appendUint          = encoder.AppendUint
	appendFloat32       = encoder.AppendFloat32
	appendFloat64       = encoder.AppendFloat64
	appendString        = encoder.AppendString
	appendByteSlice     = encoder.AppendByteSlice
	appendNumber        = encoder.AppendNumber
	appendStructEnd     = encoder.AppendStructEndIndent
	appendIndent        = encoder.AppendIndent
	errUnsupportedValue = encoder.ErrUnsupportedValue
	errUnsupportedFloat = encoder.ErrUnsupportedFloat
	mapiterinit         = encoder.MapIterInit
	mapiterkey          = encoder.MapIterKey
	mapitervalue        = encoder.MapIterValue
	mapiternext         = encoder.MapIterNext
	maplen              = encoder.MapLen
)

type emptyInterface struct {
	typ *runtime.Type
	ptr unsafe.Pointer
}

func errUnimplementedOp(op encoder.OpType) error {
	return fmt.Errorf("encoder (indent): opcode %s has not been implemented", op)
}

func load(base uintptr, idx uintptr) uintptr {
	addr := base + idx
	return **(**uintptr)(unsafe.Pointer(&addr))
}

func store(base uintptr, idx uintptr, p uintptr) {
	addr := base + idx
	**(**uintptr)(unsafe.Pointer(&addr)) = p
}

func loadNPtr(base uintptr, idx uintptr, ptrNum int) uintptr {
	addr := base + idx
	p := **(**uintptr)(unsafe.Pointer(&addr))
	for i := 0; i < ptrNum; i++ {
		if p == 0 {
			return 0
		}
		p = ptrToPtr(p)
	}
	return p
}

func ptrToUint64(p uintptr) uint64              { return **(**uint64)(unsafe.Pointer(&p)) }
func ptrToFloat32(p uintptr) float32            { return **(**float32)(unsafe.Pointer(&p)) }
func ptrToFloat64(p uintptr) float64            { return **(**float64)(unsafe.Pointer(&p)) }
func ptrToBool(p uintptr) bool                  { return **(**bool)(unsafe.Pointer(&p)) }
func ptrToBytes(p uintptr) []byte               { return **(**[]byte)(unsafe.Pointer(&p)) }
func ptrToNumber(p uintptr) json.Number         { return **(**json.Number)(unsafe.Pointer(&p)) }
func ptrToString(p uintptr) string              { return **(**string)(unsafe.Pointer(&p)) }
func ptrToSlice(p uintptr) *runtime.SliceHeader { return *(**runtime.SliceHeader)(unsafe.Pointer(&p)) }
func ptrToPtr(p uintptr) uintptr {
	return uintptr(**(**unsafe.Pointer)(unsafe.Pointer(&p)))
}
func ptrToNPtr(p uintptr, ptrNum int) uintptr {
	for i := 0; i < ptrNum; i++ {
		if p == 0 {
			return 0
		}
		p = ptrToPtr(p)
	}
	return p
}

func ptrToUnsafePtr(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}
func ptrToInterface(code *encoder.Opcode, p uintptr) interface{} {
	return *(*interface{})(unsafe.Pointer(&emptyInterface{
		typ: code.Type,
		ptr: *(*unsafe.Pointer)(unsafe.Pointer(&p)),
	}))
}

func appendBool(b []byte, v bool) []byte {
	if v {
		return append(b, "true"...)
	}
	return append(b, "false"...)
}

func appendNull(b []byte) []byte {
	return append(b, "null"...)
}

func appendComma(b []byte) []byte {
	return append(b, ',', '\n')
}

func appendColon(b []byte) []byte {
	return append(b, ':', ' ')
}

func appendInterface(ctx *encoder.RuntimeContext, codeSet *encoder.OpcodeSet, opt encoder.Option, code *encoder.Opcode, b []byte, iface *emptyInterface, ptrOffset uintptr) ([]byte, error) {
	ctx.KeepRefs = append(ctx.KeepRefs, unsafe.Pointer(iface))
	ifaceCodeSet, err := encoder.CompileToGetCodeSet(uintptr(unsafe.Pointer(iface.typ)))
	if err != nil {
		return nil, err
	}

	totalLength := uintptr(codeSet.CodeLength)
	nextTotalLength := uintptr(ifaceCodeSet.CodeLength)

	curlen := uintptr(len(ctx.Ptrs))
	offsetNum := ptrOffset / uintptrSize

	newLen := offsetNum + totalLength + nextTotalLength
	if curlen < newLen {
		ctx.Ptrs = append(ctx.Ptrs, make([]uintptr, newLen-curlen)...)
	}
	oldPtrs := ctx.Ptrs

	newPtrs := ctx.Ptrs[(ptrOffset+totalLength*uintptrSize)/uintptrSize:]
	newPtrs[0] = uintptr(iface.ptr)

	ctx.Ptrs = newPtrs

	oldBaseIndent := ctx.BaseIndent
	ctx.BaseIndent = code.Indent
	bb, err := Run(ctx, b, ifaceCodeSet, opt)
	if err != nil {
		return nil, err
	}
	ctx.BaseIndent = oldBaseIndent

	ctx.Ptrs = oldPtrs

	return bb, nil
}

func appendMapKeyValue(ctx *encoder.RuntimeContext, code *encoder.Opcode, b, key, value []byte) []byte {
	b = appendIndent(ctx, b, code.Indent+1)
	b = append(b, key...)
	b[len(b)-2] = ':'
	b[len(b)-1] = ' '
	return append(b, value...)
}

func appendMapEnd(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = b[:len(b)-2]
	b = append(b, '\n')
	b = appendIndent(ctx, b, code.Indent)
	return append(b, '}', ',', '\n')
}

func appendArrayHead(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = append(b, '[', '\n')
	return appendIndent(ctx, b, code.Indent+1)
}

func appendArrayEnd(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = b[:len(b)-2]
	b = append(b, '\n')
	b = appendIndent(ctx, b, code.Indent)
	return append(b, ']', ',', '\n')
}

func appendEmptyArray(b []byte) []byte {
	return append(b, '[', ']', ',', '\n')
}

func appendEmptyObject(b []byte) []byte {
	return append(b, '{', '}', ',', '\n')
}

func appendObjectEnd(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	b[last] = '\n'
	b = appendIndent(ctx, b, code.Indent-1)
	return append(b, '}', ',', '\n')
}

func appendMarshalJSON(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalJSONIndent(ctx, code, b, v, false)
}

func appendMarshalText(code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalTextIndent(code, b, v, false)
}

func appendStructHead(b []byte) []byte {
	return append(b, '{', '\n')
}

func appendStructKey(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	b = appendIndent(ctx, b, code.Indent)
	b = append(b, code.Key...)
	return append(b, ' ')
}

func appendStructEndSkipLast(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	if b[last-1] == '{' {
		b[last] = '}'
	} else {
		if b[last] == '\n' {
			// to remove ',' and '\n' characters
			b = b[:len(b)-2]
		}
		b = append(b, '\n')
		b = appendIndent(ctx, b, code.Indent-1)
		b = append(b, '}')
	}
	return appendComma(b)
}

func restoreIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, ctxptr uintptr) {
	ctx.BaseIndent = int(load(ctxptr, code.Length))
}

func storeIndent(ctxptr uintptr, code *encoder.Opcode, indent uintptr) {
	store(ctxptr, code.End.Next.Length, indent)
}

func appendArrayElemIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return appendIndent(ctx, b, code.Indent+1)
}

func appendMapKeyIndent(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return appendIndent(ctx, b, code.Indent)
}
