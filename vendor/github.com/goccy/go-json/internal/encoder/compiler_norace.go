// +build !race

package encoder

import (
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

func CompileToGetCodeSet(typeptr uintptr) (*OpcodeSet, error) {
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetCodeSetSlowPath(typeptr)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	if codeSet := cachedOpcodeSets[index]; codeSet != nil {
		return codeSet, nil
	}

	// noescape trick for header.typ ( reflect.*rtype )
	copiedType := *(**runtime.Type)(unsafe.Pointer(&typeptr))

	code, err := compileHead(&compileContext{
		typ:                      copiedType,
		structTypeToCompiledCode: map[uintptr]*CompiledCode{},
	})
	if err != nil {
		return nil, err
	}
	code = copyOpcode(code)
	codeLength := code.TotalLength()
	codeSet := &OpcodeSet{
		Type:       copiedType,
		Code:       code,
		CodeLength: codeLength,
	}
	cachedOpcodeSets[index] = codeSet
	return codeSet, nil
}
