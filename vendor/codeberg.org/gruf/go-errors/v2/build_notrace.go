//go:build !errtrace
// +build !errtrace

package errors

import "runtime"

// IncludesStacktrace is a compile-time flag used to indicate
// whether to include stacktraces on error wrap / creation.
const IncludesStacktrace = false

type trace struct{}

// set will set the actual trace value
// only when correct build flag is set.
func (trace) set([]runtime.Frame) {}

// value returns the actual trace value
// only when correct build flag is set.
func (trace) value() Callers { return nil }
