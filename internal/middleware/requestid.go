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

package middleware

import (
	"bufio"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"github.com/gin-gonic/gin"
)

var (
	// crand provides buffered reads of random input.
	crand = bufio.NewReader(rand.Reader)
	mrand sync.Mutex

	// base32enc is a base 32 encoding based on a human-readable character set (no padding).
	base32enc = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(-1)
)

// NewRequestID generates a new request ID string.
func NewRequestID() string {
	// 0:8  = timestamp
	// 8:12 = entropy
	//
	// inspired by ULID.
	b := make([]byte, 12)

	// Get current time in milliseconds.
	ms := uint64(time.Now().UnixMilli()) // #nosec G115 -- Pre-1970 clock?

	// Store binary time data in byte buffer.
	binary.LittleEndian.PutUint64(b[0:8], ms)

	mrand.Lock()
	// Read random bits into buffer end.
	_, _ = io.ReadFull(crand, b[8:12])
	mrand.Unlock()

	// Encode the binary time+entropy ID.
	return base32enc.EncodeToString(b)
}

// AddRequestID returns a gin middleware which adds a unique ID to each request (both response header and context).
func AddRequestID(header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(header)
		// Have we found anything?
		if id == "" {
			// Generate new ID.
			id = NewRequestID()

			// Set the request ID in the req header in case
			// we pass the request along to another service.
			c.Request.Header.Set(header, id)
		}

		// Store request ID in new request context and set on gin ctx.
		ctx := gtscontext.SetRequestID(c.Request.Context(), id)
		c.Request = c.Request.WithContext(ctx)

		// Set the request ID in the rsp header.
		c.Writer.Header().Set(header, id)
	}
}
