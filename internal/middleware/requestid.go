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
	"encoding/base32"
	"encoding/binary"
	"sync/atomic"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"github.com/gin-gonic/gin"
)

var (
	// request counter.
	count atomic.Uint32

	// server start time in milliseconds.
	start = uint64(time.Now().UnixMilli())

	// shorthand to binary.
	be = binary.BigEndian

	// b32 is a base 32 encoding based on a human-readable character set (no padding).
	b32 = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(-1)
)

// NewRequestID generates a new request ID string.
func NewRequestID() string {
	var buf [12]byte

	// Generate unique request
	// ID from request count and
	// time of server initialization.
	be.PutUint64(buf[0:], start)
	be.PutUint32(buf[8:], count.Add(1))
	return b32.EncodeToString(buf[:])
}

// AddRequestID returns a gin middleware which adds a unique ID to each request (both response header and context).
func AddRequestID(header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(header)
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
