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

//go:build !noerrcaller

package gtserror

import (
	"errors"
	"fmt"
	"runtime"
	"strings"

	"codeberg.org/gruf/go-caller"
)

// Caller returns whether created errors will prepend calling function name.
const Caller = true

// cerror wraps an error with a string
// prefix of the caller function name.
type cerror struct {
	c string
	e error
}

func (ce *cerror) Error() string {
	msg := ce.e.Error()
	return ce.c + ": " + msg
}

func (ce *cerror) Unwrap() error {
	return ce.e
}

// newAt is the same as New() but allows specifying calldepth.
func newAt(calldepth int, msg string) error {
	return &cerror{
		c: getCaller(calldepth + 1),
		e: errors.New(msg),
	}
}

// newfAt is the same as Newf() but allows specifying calldepth.
func newfAt(calldepth int, msgf string, args ...any) error {
	return &cerror{
		c: getCaller(calldepth + 1),
		e: fmt.Errorf(msgf, args...),
	}
}

// getCaller fetches the calling function name, skipping 'depth'.
func getCaller(depth int) string {
	pcs := make([]uintptr, 1)

	// Fetch calling function at depth.
	_ = runtime.Callers(depth, pcs)

	// Get cached calling func name.
	name := caller.Get(pcs[0])

	// Drop package / everything but function name itself.
	if idx := strings.LastIndexByte(name, '.'); idx >= 0 {
		name = name[idx+1:]
	}

	return name
}
