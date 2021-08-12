// +build race

package encoder

import (
	"sync"
	"unsafe"

	"github.com/goccy/go-json/internal/runtime"
)

var setsMu sync.RWMutex

func CompileToGetCodeSet(typeptr uintptr) (*OpcodeSet, error) {
	if typeptr > typeAddr.MaxTypeAddr {
		return compileToGetCodeSetSlowPath(typeptr)
	}
	index := (typeptr - typeAddr.BaseTypeAddr) >> typeAddr.AddrShift
	setsMu.RLock()
	if codeSet := cachedOpcodeSets[index]; codeSet != nil {
		setsMu.RUnlock()
		return codeSet, nil
	}
	setsMu.RUnlock()

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
	setsMu.Lock()
	cachedOpcodeSets[index] = codeSet
	setsMu.Unlock()
	return codeSet, nil
}
