package util

import (
	"context"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/tetratelabs/wazero/api"
)

// Wasm_Tempnam wraps Go_Tempnam to fulfill wazero's api.GoModuleFunc,
// the argument definition is (i32, i32) and return definition is (i32).
// NOTE: the calling module MUST have access to exported malloc / free.
func Wasm_Tempnam(ctx context.Context, mod api.Module, stack []uint64) {
	dirptr := api.DecodeU32(stack[0])
	pfxptr := api.DecodeU32(stack[1])
	dir := readString(ctx, mod, dirptr, 0)
	pfx := readString(ctx, mod, pfxptr, 0)
	tmpstr := Go_Tempnam(dir, pfx)
	tmpptr := writeString(ctx, mod, tmpstr)
	stack[0] = api.EncodeU32(tmpptr)
}

// Go_Tempname is functionally similar to C's tempnam.
func Go_Tempnam(dir, prefix string) string {
	now := time.Now().Unix()
	prefix = path.Join(dir, prefix)
	for i := 0; i < 1000; i++ {
		n := murmur2(uint32(now + int64(i)))
		name := prefix + strconv.FormatUint(uint64(n), 10)
		_, err := os.Stat(name)
		if err == nil {
			continue
		} else if os.IsNotExist(err) {
			return name
		} else {
			panic(err)
		}
	}
	panic("too many attempts")
}

// murmur2 is a simple uint32 murmur2 hash
// impl with fixed seed and input size.
func murmur2(k uint32) (h uint32) {
	const (
		//  seed ^ bitlen
		s = uint32(2147483647) ^ 8

		M = 0x5bd1e995
		R = 24
	)
	h = s
	k *= M
	k ^= k >> R
	k *= M
	h *= M
	h ^= k
	h ^= h >> 13
	h *= M
	h ^= h >> 15
	return
}
