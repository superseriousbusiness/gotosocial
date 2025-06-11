// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package util

import (
	"fmt"
	"os"
	"runtime"

	"codeberg.org/gruf/go-errors/v2"
)

// Must executes 'fn' repeatedly until
// it successfully returns without panic.
func Must(fn func()) {
	if fn == nil {
		panic("nil func")
	}
	for !func() (done bool) {
		defer Recover()
		fn()
		done = true
		return
	}() { //nolint
	}
}

// Recover wraps runtime.recover() to dump the current
// stack to stderr on panic and return the panic value.
func Recover() any {
	if r := recover(); r != nil {
		// Gather calling func frames.
		pcs := make([]uintptr, 10)
		n := runtime.Callers(3, pcs)
		i := runtime.CallersFrames(pcs[:n])
		c := gatherFrames(i, n)
		fmt.Fprintf(os.Stderr, "recovered panic: %v\n\n%s\n", r, c.String())
		return r
	}
	return nil
}

// gatherFrames collates runtime frames from a frame iterator.
func gatherFrames(iter *runtime.Frames, n int) errors.Callers {
	if iter == nil {
		return nil
	}
	frames := make([]runtime.Frame, 0, n)
	for {
		f, ok := iter.Next()
		if !ok {
			break
		}
		frames = append(frames, f)
	}
	return frames
}
