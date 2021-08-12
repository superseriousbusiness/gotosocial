package vm_escaped

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
	appendString        = encoder.AppendEscapedString
	appendByteSlice     = encoder.AppendByteSlice
	appendNumber        = encoder.AppendNumber
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
	return fmt.Errorf("encoder (escaped): opcode %s has not been implemented", op)
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
	return append(b, ',')
}

func appendColon(b []byte) []byte {
	last := len(b) - 1
	b[last] = ':'
	return b
}

func appendMapKeyValue(_ *encoder.RuntimeContext, _ *encoder.Opcode, b, key, value []byte) []byte {
	b = append(b, key...)
	b[len(b)-1] = ':'
	return append(b, value...)
}

func appendMapEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	b[len(b)-1] = '}'
	b = append(b, ',')
	return b
}

func appendInterface(ctx *encoder.RuntimeContext, codeSet *encoder.OpcodeSet, opt encoder.Option, _ *encoder.Opcode, b []byte, iface *emptyInterface, ptrOffset uintptr) ([]byte, error) {
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

	bb, err := Run(ctx, b, ifaceCodeSet, opt)
	if err != nil {
		return nil, err
	}

	ctx.Ptrs = oldPtrs
	return bb, nil
}

func appendMarshalJSON(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalJSON(ctx, code, b, v, true)
}

func appendMarshalText(code *encoder.Opcode, b []byte, v interface{}) ([]byte, error) {
	return encoder.AppendMarshalText(code, b, v, true)
}

func appendArrayHead(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	return append(b, '[')
}

func appendArrayEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	b[last] = ']'
	return append(b, ',')
}

func appendEmptyArray(b []byte) []byte {
	return append(b, '[', ']', ',')
}

func appendEmptyObject(b []byte) []byte {
	return append(b, '{', '}', ',')
}

func appendObjectEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	b[last] = '}'
	return append(b, ',')
}

func appendStructHead(b []byte) []byte {
	return append(b, '{')
}

func appendStructKey(_ *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	return append(b, code.EscapedKey...)
}

func appendStructEnd(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte {
	return append(b, '}', ',')
}

func appendStructEndSkipLast(ctx *encoder.RuntimeContext, code *encoder.Opcode, b []byte) []byte {
	last := len(b) - 1
	if b[last] == ',' {
		b[last] = '}'
		return appendComma(b)
	}
	return appendStructEnd(ctx, code, b)
}

func restoreIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, _ uintptr)               {}
func storeIndent(_ uintptr, _ *encoder.Opcode, _ uintptr)                                 {}
func appendMapKeyIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte    { return b }
func appendArrayElemIndent(_ *encoder.RuntimeContext, _ *encoder.Opcode, b []byte) []byte { return b }
