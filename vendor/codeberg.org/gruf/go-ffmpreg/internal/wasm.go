package internal

import (
	"os"

	"github.com/tetratelabs/wazero"
)

func init() {
	var err error

	if dir := os.Getenv("WAZERO_COMPILATION_CACHE"); dir != "" {
		// Use on-filesystem compilation cache given by env.
		Cache, err = wazero.NewCompilationCacheWithDir(dir)
		if err != nil {
			panic(err)
		}
	} else {
		// Use in-memory compilation cache.
		Cache = wazero.NewCompilationCache()
	}
}

// Shared WASM compilation cache.
var Cache wazero.CompilationCache
