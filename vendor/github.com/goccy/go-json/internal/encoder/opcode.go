package encoder

import (
	"fmt"
	"math"
	"strings"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

const uintptrSize = 4 << (^uintptr(0) >> 63)

type Opcode struct {
	Op               OpType        // operation type
	Type             *runtime.Type // go type
	DisplayIdx       int           // opcode index
	Key              []byte        // struct field key
	EscapedKey       []byte        // struct field key ( HTML escaped )
	PtrNum           int           // pointer number: e.g. double pointer is 2.
	DisplayKey       string        // key text to display
	IsTaggedKey      bool          // whether tagged key
	AnonymousKey     bool          // whether anonymous key
	AnonymousHead    bool          // whether anonymous head or not
	Indirect         bool          // whether indirect or not
	Nilcheck         bool          // whether needs to nilcheck or not
	AddrForMarshaler bool          // whether needs to addr for marshaler or not
	IsNextOpPtrType  bool          // whether next operation is ptr type or not
	IsNilableType    bool          // whether type is nilable or not
	RshiftNum        uint8         // use to take bit for judging whether negative integer or not
	Mask             uint64        // mask for number
	Indent           int           // indent number

	Idx     uintptr // offset to access ptr
	HeadIdx uintptr // offset to access slice/struct head
	ElemIdx uintptr // offset to access array/slice/map elem
	Length  uintptr // offset to access slice/map length or array length
	MapIter uintptr // offset to access map iterator
	MapPos  uintptr // offset to access position list for sorted map
	Offset  uintptr // offset size from struct header
	Size    uintptr // array/slice elem size

	MapKey    *Opcode       // map key
	MapValue  *Opcode       // map value
	Elem      *Opcode       // array/slice elem
	End       *Opcode       // array/slice/struct/map end
	PrevField *Opcode       // prev struct field
	NextField *Opcode       // next struct field
	Next      *Opcode       // next opcode
	Jmp       *CompiledCode // for recursive call
}

func rshitNum(bitSize uint8) uint8 {
	return bitSize - 1
}

func (c *Opcode) setMaskAndRshiftNum(bitSize uint8) {
	switch bitSize {
	case 8:
		c.Mask = math.MaxUint8
	case 16:
		c.Mask = math.MaxUint16
	case 32:
		c.Mask = math.MaxUint32
	case 64:
		c.Mask = math.MaxUint64
	}
	c.RshiftNum = rshitNum(bitSize)
}

func (c *Opcode) ToHeaderType(isString bool) OpType {
	switch c.Op {
	case OpInt:
		if isString {
			return OpStructHeadIntString
		}
		return OpStructHeadInt
	case OpIntPtr:
		if isString {
			return OpStructHeadIntPtrString
		}
		return OpStructHeadIntPtr
	case OpUint:
		if isString {
			return OpStructHeadUintString
		}
		return OpStructHeadUint
	case OpUintPtr:
		if isString {
			return OpStructHeadUintPtrString
		}
		return OpStructHeadUintPtr
	case OpFloat32:
		if isString {
			return OpStructHeadFloat32String
		}
		return OpStructHeadFloat32
	case OpFloat32Ptr:
		if isString {
			return OpStructHeadFloat32PtrString
		}
		return OpStructHeadFloat32Ptr
	case OpFloat64:
		if isString {
			return OpStructHeadFloat64String
		}
		return OpStructHeadFloat64
	case OpFloat64Ptr:
		if isString {
			return OpStructHeadFloat64PtrString
		}
		return OpStructHeadFloat64Ptr
	case OpString:
		if isString {
			return OpStructHeadStringString
		}
		return OpStructHeadString
	case OpStringPtr:
		if isString {
			return OpStructHeadStringPtrString
		}
		return OpStructHeadStringPtr
	case OpNumber:
		if isString {
			return OpStructHeadNumberString
		}
		return OpStructHeadNumber
	case OpNumberPtr:
		if isString {
			return OpStructHeadNumberPtrString
		}
		return OpStructHeadNumberPtr
	case OpBool:
		if isString {
			return OpStructHeadBoolString
		}
		return OpStructHeadBool
	case OpBoolPtr:
		if isString {
			return OpStructHeadBoolPtrString
		}
		return OpStructHeadBoolPtr
	case OpBytes:
		return OpStructHeadBytes
	case OpBytesPtr:
		return OpStructHeadBytesPtr
	case OpMap:
		return OpStructHeadMap
	case OpMapPtr:
		c.Op = OpMap
		return OpStructHeadMapPtr
	case OpArray:
		return OpStructHeadArray
	case OpArrayPtr:
		c.Op = OpArray
		return OpStructHeadArrayPtr
	case OpSlice:
		return OpStructHeadSlice
	case OpSlicePtr:
		c.Op = OpSlice
		return OpStructHeadSlicePtr
	case OpMarshalJSON:
		return OpStructHeadMarshalJSON
	case OpMarshalJSONPtr:
		return OpStructHeadMarshalJSONPtr
	case OpMarshalText:
		return OpStructHeadMarshalText
	case OpMarshalTextPtr:
		return OpStructHeadMarshalTextPtr
	}
	return OpStructHead
}

func (c *Opcode) ToFieldType(isString bool) OpType {
	switch c.Op {
	case OpInt:
		if isString {
			return OpStructFieldIntString
		}
		return OpStructFieldInt
	case OpIntPtr:
		if isString {
			return OpStructFieldIntPtrString
		}
		return OpStructFieldIntPtr
	case OpUint:
		if isString {
			return OpStructFieldUintString
		}
		return OpStructFieldUint
	case OpUintPtr:
		if isString {
			return OpStructFieldUintPtrString
		}
		return OpStructFieldUintPtr
	case OpFloat32:
		if isString {
			return OpStructFieldFloat32String
		}
		return OpStructFieldFloat32
	case OpFloat32Ptr:
		if isString {
			return OpStructFieldFloat32PtrString
		}
		return OpStructFieldFloat32Ptr
	case OpFloat64:
		if isString {
			return OpStructFieldFloat64String
		}
		return OpStructFieldFloat64
	case OpFloat64Ptr:
		if isString {
			return OpStructFieldFloat64PtrString
		}
		return OpStructFieldFloat64Ptr
	case OpString:
		if isString {
			return OpStructFieldStringString
		}
		return OpStructFieldString
	case OpStringPtr:
		if isString {
			return OpStructFieldStringPtrString
		}
		return OpStructFieldStringPtr
	case OpNumber:
		if isString {
			return OpStructFieldNumberString
		}
		return OpStructFieldNumber
	case OpNumberPtr:
		if isString {
			return OpStructFieldNumberPtrString
		}
		return OpStructFieldNumberPtr
	case OpBool:
		if isString {
			return OpStructFieldBoolString
		}
		return OpStructFieldBool
	case OpBoolPtr:
		if isString {
			return OpStructFieldBoolPtrString
		}
		return OpStructFieldBoolPtr
	case OpBytes:
		return OpStructFieldBytes
	case OpBytesPtr:
		return OpStructFieldBytesPtr
	case OpMap:
		return OpStructFieldMap
	case OpMapPtr:
		c.Op = OpMap
		return OpStructFieldMapPtr
	case OpArray:
		return OpStructFieldArray
	case OpArrayPtr:
		c.Op = OpArray
		return OpStructFieldArrayPtr
	case OpSlice:
		return OpStructFieldSlice
	case OpSlicePtr:
		c.Op = OpSlice
		return OpStructFieldSlicePtr
	case OpMarshalJSON:
		return OpStructFieldMarshalJSON
	case OpMarshalJSONPtr:
		return OpStructFieldMarshalJSONPtr
	case OpMarshalText:
		return OpStructFieldMarshalText
	case OpMarshalTextPtr:
		return OpStructFieldMarshalTextPtr
	}
	return OpStructField
}

func newOpCode(ctx *compileContext, op OpType) *Opcode {
	return newOpCodeWithNext(ctx, op, newEndOp(ctx))
}

func opcodeOffset(idx int) uintptr {
	return uintptr(idx) * uintptrSize
}

func copyOpcode(code *Opcode) *Opcode {
	codeMap := map[uintptr]*Opcode{}
	return code.copy(codeMap)
}

func newOpCodeWithNext(ctx *compileContext, op OpType, next *Opcode) *Opcode {
	return &Opcode{
		Op:         op,
		Type:       ctx.typ,
		DisplayIdx: ctx.opcodeIndex,
		Indent:     ctx.indent,
		Idx:        opcodeOffset(ctx.ptrIndex),
		Next:       next,
	}
}

func newEndOp(ctx *compileContext) *Opcode {
	return newOpCodeWithNext(ctx, OpEnd, nil)
}

func (c *Opcode) copy(codeMap map[uintptr]*Opcode) *Opcode {
	if c == nil {
		return nil
	}
	addr := uintptr(unsafe.Pointer(c))
	if code, exists := codeMap[addr]; exists {
		return code
	}
	copied := &Opcode{
		Op:               c.Op,
		Type:             c.Type,
		DisplayIdx:       c.DisplayIdx,
		Key:              c.Key,
		EscapedKey:       c.EscapedKey,
		DisplayKey:       c.DisplayKey,
		PtrNum:           c.PtrNum,
		Mask:             c.Mask,
		RshiftNum:        c.RshiftNum,
		IsTaggedKey:      c.IsTaggedKey,
		AnonymousKey:     c.AnonymousKey,
		AnonymousHead:    c.AnonymousHead,
		Indirect:         c.Indirect,
		Nilcheck:         c.Nilcheck,
		AddrForMarshaler: c.AddrForMarshaler,
		IsNextOpPtrType:  c.IsNextOpPtrType,
		IsNilableType:    c.IsNilableType,
		Indent:           c.Indent,
		Idx:              c.Idx,
		HeadIdx:          c.HeadIdx,
		ElemIdx:          c.ElemIdx,
		Length:           c.Length,
		MapIter:          c.MapIter,
		MapPos:           c.MapPos,
		Offset:           c.Offset,
		Size:             c.Size,
	}
	codeMap[addr] = copied
	copied.MapKey = c.MapKey.copy(codeMap)
	copied.MapValue = c.MapValue.copy(codeMap)
	copied.Elem = c.Elem.copy(codeMap)
	copied.End = c.End.copy(codeMap)
	copied.PrevField = c.PrevField.copy(codeMap)
	copied.NextField = c.NextField.copy(codeMap)
	copied.Next = c.Next.copy(codeMap)
	copied.Jmp = c.Jmp
	return copied
}

func (c *Opcode) BeforeLastCode() *Opcode {
	code := c
	for {
		var nextCode *Opcode
		switch code.Op.CodeType() {
		case CodeArrayElem, CodeSliceElem, CodeMapKey:
			nextCode = code.End
		default:
			nextCode = code.Next
		}
		if nextCode.Op == OpEnd {
			return code
		}
		code = nextCode
	}
}

func (c *Opcode) TotalLength() int {
	var idx int
	for code := c; code.Op != OpEnd; {
		idx = int(code.Idx / uintptrSize)
		if code.Op == OpRecursiveEnd {
			break
		}
		switch code.Op.CodeType() {
		case CodeArrayElem, CodeSliceElem, CodeMapKey:
			code = code.End
		default:
			code = code.Next
		}
	}
	return idx + 2 // opEnd + 1
}

func (c *Opcode) decOpcodeIndex() {
	for code := c; code.Op != OpEnd; {
		code.DisplayIdx--
		code.Idx -= uintptrSize
		if code.HeadIdx > 0 {
			code.HeadIdx -= uintptrSize
		}
		if code.ElemIdx > 0 {
			code.ElemIdx -= uintptrSize
		}
		if code.MapIter > 0 {
			code.MapIter -= uintptrSize
		}
		if code.Length > 0 && code.Op.CodeType() != CodeArrayHead && code.Op.CodeType() != CodeArrayElem {
			code.Length -= uintptrSize
		}
		switch code.Op.CodeType() {
		case CodeArrayElem, CodeSliceElem, CodeMapKey:
			code = code.End
		default:
			code = code.Next
		}
	}
}

func (c *Opcode) decIndent() {
	for code := c; code.Op != OpEnd; {
		code.Indent--
		switch code.Op.CodeType() {
		case CodeArrayElem, CodeSliceElem, CodeMapKey:
			code = code.End
		default:
			code = code.Next
		}
	}
}

func (c *Opcode) dumpHead(code *Opcode) string {
	var length uintptr
	if code.Op.CodeType() == CodeArrayHead {
		length = code.Length
	} else {
		length = code.Length / uintptrSize
	}
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][headIdx:%d][elemIdx:%d][length:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.HeadIdx/uintptrSize,
		code.ElemIdx/uintptrSize,
		length,
	)
}

func (c *Opcode) dumpMapHead(code *Opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][headIdx:%d][elemIdx:%d][length:%d][mapIter:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.HeadIdx/uintptrSize,
		code.ElemIdx/uintptrSize,
		code.Length/uintptrSize,
		code.MapIter/uintptrSize,
	)
}

func (c *Opcode) dumpMapEnd(code *Opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][mapPos:%d][length:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.MapPos/uintptrSize,
		code.Length/uintptrSize,
	)
}

func (c *Opcode) dumpElem(code *Opcode) string {
	var length uintptr
	if code.Op.CodeType() == CodeArrayElem {
		length = code.Length
	} else {
		length = code.Length / uintptrSize
	}
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][headIdx:%d][elemIdx:%d][length:%d][size:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.HeadIdx/uintptrSize,
		code.ElemIdx/uintptrSize,
		length,
		code.Size,
	)
}

func (c *Opcode) dumpField(code *Opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][key:%s][offset:%d][headIdx:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.DisplayKey,
		code.Offset,
		code.HeadIdx/uintptrSize,
	)
}

func (c *Opcode) dumpKey(code *Opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][elemIdx:%d][length:%d][mapIter:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.ElemIdx/uintptrSize,
		code.Length/uintptrSize,
		code.MapIter/uintptrSize,
	)
}

func (c *Opcode) dumpValue(code *Opcode) string {
	return fmt.Sprintf(
		`[%d]%s%s ([idx:%d][mapIter:%d])`,
		code.DisplayIdx,
		strings.Repeat("-", code.Indent),
		code.Op,
		code.Idx/uintptrSize,
		code.MapIter/uintptrSize,
	)
}

func (c *Opcode) Dump() string {
	codes := []string{}
	for code := c; code.Op != OpEnd; {
		switch code.Op.CodeType() {
		case CodeSliceHead:
			codes = append(codes, c.dumpHead(code))
			code = code.Next
		case CodeMapHead:
			codes = append(codes, c.dumpMapHead(code))
			code = code.Next
		case CodeArrayElem, CodeSliceElem:
			codes = append(codes, c.dumpElem(code))
			code = code.End
		case CodeMapKey:
			codes = append(codes, c.dumpKey(code))
			code = code.End
		case CodeMapValue:
			codes = append(codes, c.dumpValue(code))
			code = code.Next
		case CodeMapEnd:
			codes = append(codes, c.dumpMapEnd(code))
			code = code.Next
		case CodeStructField:
			codes = append(codes, c.dumpField(code))
			code = code.Next
		case CodeStructEnd:
			codes = append(codes, c.dumpField(code))
			code = code.Next
		default:
			codes = append(codes, fmt.Sprintf(
				"[%d]%s%s ([idx:%d])",
				code.DisplayIdx,
				strings.Repeat("-", code.Indent),
				code.Op,
				code.Idx/uintptrSize,
			))
			code = code.Next
		}
	}
	return strings.Join(codes, "\n")
}

func prevField(code *Opcode, removedFields map[*Opcode]struct{}) *Opcode {
	if _, exists := removedFields[code]; exists {
		return prevField(code.PrevField, removedFields)
	}
	return code
}

func nextField(code *Opcode, removedFields map[*Opcode]struct{}) *Opcode {
	if _, exists := removedFields[code]; exists {
		return nextField(code.NextField, removedFields)
	}
	return code
}

func linkPrevToNextField(cur *Opcode, removedFields map[*Opcode]struct{}) {
	prev := prevField(cur.PrevField, removedFields)
	prev.NextField = nextField(cur.NextField, removedFields)
	code := prev
	fcode := cur
	for {
		var nextCode *Opcode
		switch code.Op.CodeType() {
		case CodeArrayElem, CodeSliceElem, CodeMapKey:
			nextCode = code.End
		default:
			nextCode = code.Next
		}
		if nextCode == fcode {
			code.Next = fcode.Next
			break
		} else if nextCode.Op == OpEnd {
			break
		}
		code = nextCode
	}
}

func newSliceHeaderCode(ctx *compileContext) *Opcode {
	idx := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	elemIdx := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	length := opcodeOffset(ctx.ptrIndex)
	return &Opcode{
		Op:         OpSlice,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        idx,
		HeadIdx:    idx,
		ElemIdx:    elemIdx,
		Length:     length,
		Indent:     ctx.indent,
	}
}

func newSliceElemCode(ctx *compileContext, head *Opcode, size uintptr) *Opcode {
	return &Opcode{
		Op:         OpSliceElem,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        opcodeOffset(ctx.ptrIndex),
		HeadIdx:    head.Idx,
		ElemIdx:    head.ElemIdx,
		Length:     head.Length,
		Indent:     ctx.indent,
		Size:       size,
	}
}

func newArrayHeaderCode(ctx *compileContext, alen int) *Opcode {
	idx := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	elemIdx := opcodeOffset(ctx.ptrIndex)
	return &Opcode{
		Op:         OpArray,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        idx,
		HeadIdx:    idx,
		ElemIdx:    elemIdx,
		Indent:     ctx.indent,
		Length:     uintptr(alen),
	}
}

func newArrayElemCode(ctx *compileContext, head *Opcode, length int, size uintptr) *Opcode {
	return &Opcode{
		Op:         OpArrayElem,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        opcodeOffset(ctx.ptrIndex),
		ElemIdx:    head.ElemIdx,
		HeadIdx:    head.HeadIdx,
		Length:     uintptr(length),
		Indent:     ctx.indent,
		Size:       size,
	}
}

func newMapHeaderCode(ctx *compileContext) *Opcode {
	idx := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	elemIdx := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	length := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	mapIter := opcodeOffset(ctx.ptrIndex)
	return &Opcode{
		Op:         OpMap,
		Type:       ctx.typ,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        idx,
		ElemIdx:    elemIdx,
		Length:     length,
		MapIter:    mapIter,
		Indent:     ctx.indent,
	}
}

func newMapKeyCode(ctx *compileContext, head *Opcode) *Opcode {
	return &Opcode{
		Op:         OpMapKey,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        opcodeOffset(ctx.ptrIndex),
		ElemIdx:    head.ElemIdx,
		Length:     head.Length,
		MapIter:    head.MapIter,
		Indent:     ctx.indent,
	}
}

func newMapValueCode(ctx *compileContext, head *Opcode) *Opcode {
	return &Opcode{
		Op:         OpMapValue,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        opcodeOffset(ctx.ptrIndex),
		ElemIdx:    head.ElemIdx,
		Length:     head.Length,
		MapIter:    head.MapIter,
		Indent:     ctx.indent,
	}
}

func newMapEndCode(ctx *compileContext, head *Opcode) *Opcode {
	mapPos := opcodeOffset(ctx.ptrIndex)
	ctx.incPtrIndex()
	idx := opcodeOffset(ctx.ptrIndex)
	return &Opcode{
		Op:         OpMapEnd,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        idx,
		Length:     head.Length,
		MapPos:     mapPos,
		Indent:     ctx.indent,
		Next:       newEndOp(ctx),
	}
}

func newInterfaceCode(ctx *compileContext) *Opcode {
	return &Opcode{
		Op:         OpInterface,
		Type:       ctx.typ,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        opcodeOffset(ctx.ptrIndex),
		Indent:     ctx.indent,
		Next:       newEndOp(ctx),
	}
}

func newRecursiveCode(ctx *compileContext, jmp *CompiledCode) *Opcode {
	return &Opcode{
		Op:         OpRecursive,
		Type:       ctx.typ,
		DisplayIdx: ctx.opcodeIndex,
		Idx:        opcodeOffset(ctx.ptrIndex),
		Indent:     ctx.indent,
		Next:       newEndOp(ctx),
		Jmp:        jmp,
	}
}
