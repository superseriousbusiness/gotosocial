// +build !mips,!mips64,!ppc64,!s390x,!amd64,!386,!arm,!arm64,!mipsle,!mips64le,!ppc64le,!riscv64,!wasm

package consts

import "unsafe"

var IsLittleEndian = *(*uint16)(unsafe.Pointer(&[2]byte{0, 1})) != 1
