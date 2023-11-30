//go:build !errcaller
// +build !errcaller

package errors

// IncludesCaller is a compile-time flag used to indicate whether
// to include calling function prefix on error wrap / creation.
const IncludesCaller = false

type caller struct{}

// set will set the actual caller value
// only when correct build flag is set.
func (caller) set(string) {}

// value returns the actual caller value
// only when correct build flag is set.
func (caller) value() string { return "" }
