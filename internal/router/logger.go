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

package router

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"codeberg.org/gruf/go-bytesize"
	"codeberg.org/gruf/go-errors/v2"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// loggingMiddleware provides a request logging and panic recovery gin handler.
func loggingMiddleware(c *gin.Context) {
	// Initialize the logging fields
	fields := make(logrus.Fields, 7)

	// Determine pre-handler time
	before := time.Now()

	defer func() {
		code := c.Writer.Status()
		path := c.Request.URL.Path

		if r := recover(); r != nil {
			if c.Writer.Status() == 0 {
				// No response was written, send a generic Internal Error
				c.Writer.WriteHeader(http.StatusInternalServerError)
			}

			// Append panic information to the request ctx
			_ = c.Error(fmt.Errorf("recovered panic: %v", r))

			// Dump a stacktrace to stderr
			callers := errors.GetCallers(3, 10)
			fmt.Fprintf(os.Stderr, "recovered panic: %v\n%s", r, callers)
		}

		// Set request logging fields
		fields["latency"] = time.Since(before)
		fields["clientIP"] = c.ClientIP()
		fields["userAgent"] = c.Request.UserAgent()
		fields["method"] = c.Request.Method
		fields["statusCode"] = code
		fields["path"] = path

		// Create a log entry with fields
		l := logrus.WithFields(fields)
		l.Level = logrus.InfoLevel

		if code >= 500 {
			// This is a server error
			l.Level = logrus.ErrorLevel

			if len(c.Errors) > 0 {
				// Add an error string log field
				fields["error"] = c.Errors.String()
			}
		}

		// Generate a nicer looking bytecount
		size := bytesize.Size(c.Writer.Size())

		// Finally, write log entry with status text body size
		l.Logf(l.Level, "%s: wrote %s", http.StatusText(code), size)
	}()

	// Process request
	c.Next()
}
