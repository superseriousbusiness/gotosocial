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
	"time"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-errors/v2"
	"codeberg.org/gruf/go-kv"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Logger returns a gin middleware which provides request logging and panic recovery.
func Logger(logClientIP bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Initialize the logging fields
		fields := make(kv.Fields, 5, 7)

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
				pcs := make([]uintptr, 10)
				n := runtime.Callers(3, pcs)
				iter := runtime.CallersFrames(pcs[:n])
				callers := errors.Callers(gatherFrames(iter, n))
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
			fields[1] = kv.Field{"userAgent", c.Request.UserAgent()}
			fields[2] = kv.Field{"method", c.Request.Method}
			fields[3] = kv.Field{"statusCode", code}
			fields[4] = kv.Field{"path", path}

			// Set optional request logging fields.
			if logClientIP {
				fields = append(fields, kv.Field{
					"clientIP", c.ClientIP(),
				})
			}

			ctx := c.Request.Context()
			if pubKeyID := gtscontext.HTTPSignaturePubKeyID(ctx); pubKeyID != nil {
				fields = append(fields, kv.Field{
					"pubKeyID", pubKeyID.String(),
				})
			}

			// Create log entry with fields
			l := log.New()
			l = l.WithContext(ctx)
			l = l.WithFields(fields...)

			// Default is info
			lvl := log.INFO

			if code >= 500 {
				// Actual error.
				lvl = log.ERROR
			}

			if len(c.Errors) > 0 {
				// Always attach any found errors.
				l = l.WithField("errors", c.Errors)
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

			// Finally, write log entry with status text + body size.
			l.Logf(lvl, "%s: wrote %s", statusText, size)
		}()

		// Process request
		c.Next()
	}
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
