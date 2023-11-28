//go:build errtrace
// +build errtrace

package errors

import (
	"runtime"
	_ "unsafe"
)

// IncludesStacktrace is a compile-time flag used to indicate
// whether to include stacktraces on error wrap / creation.
const IncludesStacktrace = true

type trace []runtime.Frame

// set will set the actual trace value
// only when correct build flag is set.
func (t *trace) set(v []runtime.Frame) {
	*t = trace(v)
}

// value returns the actual trace value
// only when correct build flag is set.
func (t trace) value() Callers {
	return Callers(t)
}
