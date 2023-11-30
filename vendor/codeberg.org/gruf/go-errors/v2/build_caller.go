//go:build errcaller
// +build errcaller

package errors

import (
	_ "unsafe"
)

// IncludesCaller is a compile-time flag used to indicate whether
// to include calling function prefix on error wrap / creation.
const IncludesCaller = true

type caller string

// set will set the actual caller value
// only when correct build flag is set.
func (c *caller) set(v string) {
	*c = caller(v)
}

// value returns the actual caller value
// only when correct build flag is set
func (c caller) value() string {
	return string(c)
}
