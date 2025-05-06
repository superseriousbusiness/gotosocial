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
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-kv"
	"github.com/gin-gonic/gin"
)

// Logger returns a gin middleware which provides request logging and panic recovery.
func Logger(logClientIP bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Determine pre-handler time
		before := time.Now()

		// defer so that we log *after
		// the request has completed*
		defer func() {

			// Get response status code.
			code := c.Writer.Status()

			// Get request context.
			ctx := c.Request.Context()

			if r := recover(); r != nil {
				if code == 0 {
					// No response was written, send a generic Internal Error
					c.Writer.WriteHeader(http.StatusInternalServerError)
				}

				// Append panic information to the request ctx
				err := fmt.Errorf("recovered panic: %v", r)
				_ = c.Error(err)

				// Dump a stacktrace to error log
				pcs := make([]uintptr, 10)
				n := runtime.Callers(3, pcs)
				iter := runtime.CallersFrames(pcs[:n])
				callers := errors.Callers(gatherFrames(iter, n))
				log.WithContext(c.Request.Context()).
					WithField("stacktrace", callers).Error(err)
			}

			// Initialize the logging fields
			fields := make(kv.Fields, 5, 8)

			// Set request logging fields
			fields[0] = kv.Field{"latency", time.Since(before)}
			fields[1] = kv.Field{"userAgent", c.Request.UserAgent()}
			fields[2] = kv.Field{"method", c.Request.Method}
			fields[3] = kv.Field{"statusCode", code}

			// If the request contains sensitive query
			// data only log path, else log entire URI.
			if sensitiveQuery(c.Request.URL.RawQuery) {
				path := c.Request.URL.Path
				fields[4] = kv.Field{"uri", path}
			} else {
				uri := c.Request.RequestURI
				fields[4] = kv.Field{"uri", uri}
			}

			if logClientIP {
				// Append IP only if configured to.
				fields = append(fields, kv.Field{
					"clientIP", c.ClientIP(),
				})
			}

			if pubKeyID := gtscontext.HTTPSignaturePubKeyID(ctx); pubKeyID != nil {
				// Append public key ID if attached.
				fields = append(fields, kv.Field{
					"pubKeyID", pubKeyID.String(),
				})
			}

			if len(c.Errors) > 0 {
				// Always attach any found errors.
				fields = append(fields, kv.Field{
					"errors", c.Errors,
				})
			}

			// Create entry
			// with fields.
			l := log.New().
				WithContext(ctx).
				WithFields(fields...)

			// Default is info
			lvl := log.INFO

			if code >= 500 {
				// Actual error.
				lvl = log.ERROR
			}

			// Get appropriate text for this code.
			statusText := http.StatusText(code)
			if statusText == "" {
				// Look for custom codes.
				switch code {
				case gtserror.StatusClientClosedRequest:
					statusText = gtserror.StatusTextClientClosedRequest
				default:
					statusText = "Unknown Status"
				}
			}

			// Generate a nicer looking bytecount
			size := bytesize.Size(c.Writer.Size()) // #nosec G115 -- Just logging

			// Write log entry with status text + body size.
			l.Logf(lvl, "%s: wrote %s", statusText, size)
		}()

		// Process
		// request.
		c.Next()
	}
}

// sensitiveQuery checks whether given query string
// contains sensitive data that shouldn't be logged.
func sensitiveQuery(query string) bool {
	return strings.Contains(query, "token")
}

// gatherFrames gathers runtime frames from a frame iterator.
func gatherFrames(iter *runtime.Frames, n int) []runtime.Frame {
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
