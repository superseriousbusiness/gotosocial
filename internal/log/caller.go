/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package log

import (
	"runtime"
	"strings"
	"sync"
)

var (
	// fnCache is a cache of PCs to their calculated function names.
	fnCache = map[uintptr]string{}

	// strCache is a cache of strings to the originally allocated version
	// of that string contents. so we don't have hundreds of the same instances
	// of string floating around in memory.
	strCache = map[string]string{}

	// cacheMu protects fnCache and strCache.
	cacheMu sync.Mutex
)

// Caller fetches the calling function name, skipping 'depth'. Results are cached per PC.
func Caller(depth int) string {
	var rpc [1]uintptr

	// Fetch pcs of callers
	n := runtime.Callers(depth+1, rpc[:])

	if n > 0 {
		// Look for value in cache
		cacheMu.Lock()
		fn, ok := fnCache[rpc[0]]
		cacheMu.Unlock()

		if ok {
			return fn
		}

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

			// Cache this func name
			cacheMu.Lock()
			fn, ok := strCache[name]
			if !ok {
				// Cache ptr to this allocated str
				strCache[name] = name
				fn = name
			}
			fnCache[rpc[0]] = fn
			cacheMu.Unlock()

			return fn
		}
	}

	return "???"
}
