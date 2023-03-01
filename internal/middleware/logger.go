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

package middleware

import (
	"fmt"
	"net/http"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-kv"
	"codeberg.org/gruf/go-logger/v2/level"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Logger returns a gin middleware which provides request logging and panic recovery.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize the logging fields
		fields := make(kv.Fields, 6, 7)

		// Determine pre-handler time
		before := time.Now()

		// defer so that we log *after the request has completed*
		defer func() {
			code := c.Writer.Status()
			path := c.Request.URL.Path

			if r := recover(); r != nil {
				if c.Writer.Status() == 0 {
					// No response was written, send a generic Internal Error
					c.Writer.WriteHeader(http.StatusInternalServerError)
				}

				// Append panic information to the request ctx
				err := fmt.Errorf("recovered panic: %v", r)
				_ = c.Error(err)

				// Dump a stacktrace to error log
				callers := errors.GetCallers(3, 10)
				log.WithContext(c.Request.Context()).
					WithField("stacktrace", callers).Error(err)
			}

			// NOTE:
			// It is very important here that we are ONLY logging
			// the request path, and none of the query parameters.
			// Query parameters can contain sensitive information
			// and could lead to storing plaintext API keys in logs

			// Set request logging fields
			fields[0] = kv.Field{"latency", time.Since(before)}
			fields[1] = kv.Field{"clientIP", c.ClientIP()}
			fields[2] = kv.Field{"userAgent", c.Request.UserAgent()}
			fields[3] = kv.Field{"method", c.Request.Method}
			fields[4] = kv.Field{"statusCode", code}
			fields[5] = kv.Field{"path", path}

			// Create log entry with fields
			l := log.WithContext(c.Request.Context()).
				WithFields(fields...)

			// Default is info
			lvl := level.INFO

			if code >= 500 {
				// This is a server error
				lvl = level.ERROR
				l = l.WithField("error", c.Errors)
			}

			// Generate a nicer looking bytecount
			size := bytesize.Size(c.Writer.Size())

			// Finally, write log entry with status text body size
			l.Logf(lvl, "%s: wrote %s", http.StatusText(code), size)
		}()

		// Process request
		c.Next()
	}
}
