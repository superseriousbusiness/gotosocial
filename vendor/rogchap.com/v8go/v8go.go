// Copyright 2019 Roger Chapman and the v8go contributors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

/*
Package v8go provides an API to execute JavaScript.
*/
package v8go

// #include "v8go.h"
// #include <stdlib.h>
import "C"
import (
	"strings"
	"unsafe"
)

// Version returns the version of the V8 Engine with the -v8go suffix
func Version() string {
	return C.GoString(C.Version())
}

// SetFlags sets flags for V8. For possible flags: https://github.com/v8/v8/blob/master/src/flags/flag-definitions.h
// Flags are expected to be prefixed with `--`, for example: `--harmony`.
// Flags can be reverted using the `--no` prefix equivalent, for example: `--use_strict` vs `--nouse_strict`.
// Flags will affect all Isolates created, even after creation.
func SetFlags(flags ...string) {
	cflags := C.CString(strings.Join(flags, " "))
	C.SetFlags(cflags)
	C.free(unsafe.Pointer(cflags))
}
