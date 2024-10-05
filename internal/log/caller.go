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

package log

import (
	"runtime"
	"strings"
)

// Caller fetches the calling function name, skipping 'depth'.
//
//go:noinline
func Caller(depth int) string {
	pcs := make([]uintptr, 1)

	// Fetch calling func using depth.
	_ = runtime.Callers(depth, pcs)
	fn := runtime.FuncForPC(pcs[0])

	if fn == nil {
		return ""
	}

	// Get func name.
	name := fn.Name()

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
