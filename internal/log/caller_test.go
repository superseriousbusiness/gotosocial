package log_test

import (
	"runtime"
	"strings"
	"testing"

	"codeberg.org/gruf/go-atomics"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// noopt exists to prevent certain optimisations during benching.
var noopt = atomics.NewString()

func BenchmarkCaller(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			name := log.Caller(2)
			noopt.Store(name)
		}
	})
}

func BenchmarkCallerNoCache(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var rpc [1]uintptr

			// Fetch pcs of callers
			n := runtime.Callers(2, rpc[:])

			if n > 0 {
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

					noopt.Store(name)
				}
			}
		}
	})
}
