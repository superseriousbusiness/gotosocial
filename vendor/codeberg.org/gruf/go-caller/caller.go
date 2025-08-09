package caller

import (
	"runtime"
	"strings"
	"sync/atomic"
)

var (
	// callerCache caches PC values to string names.
	// note this may be a little slower than Caller()
	// calls on startup, but after all PCs are cached
	// this should be ~3x faster + less GC overhead.
	//
	// see the following benchmark:
	// goos: linux
	// goarch: amd64
	// pkg: codeberg.org/gruf/go-caller
	// cpu: AMD Ryzen 7 7840U w/ Radeon  780M Graphics
	// BenchmarkCallerCache
	// BenchmarkCallerCache-16         16796982                66.19 ns/op           24 B/op          3 allocs/op
	// BenchmarkNoCallerCache
	// BenchmarkNoCallerCache-16        5486168               219.9 ns/op           744 B/op          6 allocs/op
	callerCache atomic.Pointer[map[uintptr]string]

	// stringCache caches strings to minimise string memory use
	// by ensuring only 1 instance of the same func name string.
	stringCache atomic.Pointer[map[string]string]
)

// Clear will empty the global caller PC -> func names cache.
func Clear() { callerCache.Store(nil); stringCache.Store(nil) }

// Name returns the calling function name for given
// program counter, formatted to be useful for logging.
func Name(pc uintptr) string {

	// Get frame iterator for program counter.
	frames := runtime.CallersFrames([]uintptr{pc})
	if frames == nil {
		return "???"
	}

	// Get func name from frame.
	frame, _ := frames.Next()
	name := frame.Function
	if name == "" {
		return "???"
	}

	// Drop all but package and function name, no path.
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}

	const params = `[...]`

	// Drop any function generic type parameter markers.
	if idx := strings.Index(name, params); idx >= 0 {
		name = name[:idx] + name[idx+len(params):]
	}

	return name
}

// Get will return calling func information for given PC value,
// caching func names by their PC values to reduce calls to Caller().
func Get(pc uintptr) string {
	var cache map[uintptr]string
	for {
		// Load caller cache map.
		ptr := callerCache.Load()

		if ptr != nil {
			// Look for stored name.
			name, ok := (*ptr)[pc]
			if ok {
				return name
			}

			// Make a clone of existing caller cache map.
			cache = make(map[uintptr]string, len(*ptr)+1)
			for key, value := range *ptr {
				cache[key] = value
			}
		} else {
			// Allocate new caller cache map.
			cache = make(map[uintptr]string, 1)
		}

		// Calculate caller
		// name for PC value.
		name := Name(pc)
		name = getString(name)

		// Store in map.
		cache[pc] = name

		// Attempt to update caller cache map pointer.
		if callerCache.CompareAndSwap(ptr, &cache) {
			return name
		}
	}
}

func getString(key string) string {
	var cache map[string]string
	for {
		// Load string cache map.
		ptr := stringCache.Load()

		if ptr != nil {
			// Check for existing string.
			if str, ok := (*ptr)[key]; ok {
				return str
			}

			// Make a clone of existing string cache map.
			cache = make(map[string]string, len(*ptr)+1)
			for key, value := range *ptr {
				cache[key] = value
			}
		} else {
			// Allocate new string cache map.
			cache = make(map[string]string, 1)
		}

		// Store this str.
		cache[key] = key

		// Attempt to update string cache map pointer.
		if stringCache.CompareAndSwap(ptr, &cache) {
			return key
		}
	}
}
