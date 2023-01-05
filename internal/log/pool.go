/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"sync"

	"codeberg.org/gruf/go-byteutil"
)

// bufPool provides a memory pool of log buffers.
var bufPool = sync.Pool{
	New: func() any {
		return &byteutil.Buffer{
			B: make([]byte, 0, 512),
		}
	},
}

// getBuf acquires a buffer from memory pool.
func getBuf() *byteutil.Buffer {
	buf, _ := bufPool.Get().(*byteutil.Buffer)
	return buf
}

// putBuf places (after resetting) buffer back in memory pool, dropping if capacity too large.
func putBuf(buf *byteutil.Buffer) {
	if buf.Cap() > int(^uint16(0)) {
		return // drop large buffer
	}
	buf.Reset()
	bufPool.Put(buf)
}
