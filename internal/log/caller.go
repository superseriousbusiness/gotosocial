package log

import (
	"runtime"
	"strings"
	"sync"
)

var (
	// fnCache is a cache of PCs to their calculated function names.
	fnCache = map[uintptr]string{}

	// strCache is a cache of strings to the originally allocated version
	// of that string contents. so we don't have hundreds of the same instances
	// of string floating around in memory.
	strCache = map[string]string{}

	// cacheMu protects fnCache and strCache.
	cacheMu sync.Mutex
)

// caller fetches the calling function name, skipping 'depth'.
func caller(depth int) string {
	var rpc [1]uintptr

	// Fetch pcs of callers
	n := runtime.Callers(depth+1, rpc[:])

	if n > 0 {
		// Look for value in cache
		cacheMu.Lock()
		fn, ok := fnCache[rpc[0]]
		cacheMu.Unlock()

		if ok {
			return fn
		}

		// Fetch frame info for caller pc
		frame, _ := runtime.CallersFrames(rpc[:]).Next()

		if frame.PC != 0 {
			name := frame.Function

			// Drop all but the package name and function name, no mod path
			if idx := strings.LastIndex(name, "/"); idx >= 0 {
				name = name[idx+1:]
			}

			// Drop any generic type parameter markers
			if idx := strings.Index(name, "[...]"); idx >= 0 {
				name = name[:idx] + name[idx+5:]
			}

			// Cache this func name
			cacheMu.Lock()
			fn, ok := strCache[name]
			if !ok {
				// Cache ptr to this allocated str
				strCache[name] = name
				fn = name
			}
			fnCache[rpc[0]] = fn
			cacheMu.Unlock()

			return fn
		}
	}

	return "???"
}
