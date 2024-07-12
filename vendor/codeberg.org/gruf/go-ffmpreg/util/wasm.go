package util

import (
	"bytes"
	"context"

	"github.com/tetratelabs/wazero/api"
)

// NOTE:
// the below functions are not very well optimized
// for repeated calls. this is relying on the fact
// that the only place they get used (tempnam), is
// not called very often, should only be once per run
// so calls to ExportedFunction() and Call() instead
// of caching api.Function and using CallWithStack()
// will work out the same (if only called once).

// maxaddr is the maximum
// wasm32 memory address.
const maxaddr = ^uint32(0)

func malloc(ctx context.Context, mod api.Module, sz uint32) uint32 {
	stack, err := mod.ExportedFunction("malloc").Call(ctx, uint64(sz))
	if err != nil {
		panic(err)
	}
	ptr := api.DecodeU32(stack[0])
	if ptr == 0 {
		panic("out of memory")
	}
	return ptr
}

func free(ctx context.Context, mod api.Module, ptr uint32) {
	if ptr != 0 {
		mod.ExportedFunction("free").Call(ctx, uint64(ptr))
	}
}

func view(ctx context.Context, mod api.Module, ptr uint32, n uint32) []byte {
	if n == 0 {
		n = maxaddr - ptr
	}
	mem := mod.Memory()
	b, ok := mem.Read(ptr, n)
	if !ok {
		panic("out of range")
	}
	return b
}

func read(ctx context.Context, mod api.Module, ptr, n uint32) []byte {
	return bytes.Clone(view(ctx, mod, ptr, n))
}

func readString(ctx context.Context, mod api.Module, ptr, n uint32) string {
	return string(view(ctx, mod, ptr, n))
}

func write(ctx context.Context, mod api.Module, b []byte) uint32 {
	mem := mod.Memory()
	len := uint32(len(b))
	ptr := malloc(ctx, mod, len)
	ok := mem.Write(ptr, b)
	if !ok {
		panic("out of range")
	}
	return ptr
}

func writeString(ctx context.Context, mod api.Module, str string) uint32 {
	mem := mod.Memory()
	len := uint32(len(str) + 1)
	ptr := malloc(ctx, mod, len)
	ok := mem.WriteString(ptr, str)
	if !ok {
		panic("out of range")
	}
	return ptr
}
