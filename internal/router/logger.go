/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var skipPaths = map[string]interface{}{
	"/api/v1/streaming": nil,
}

func loggerWithConfig(log *logrus.Logger) gin.HandlerFunc {
	logHandler := func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skipPaths[path]; !ok {
			latency := time.Since(start)
			clientIP := c.ClientIP()
			userAgent := c.Request.UserAgent()
			method := c.Request.Method
			statusCode := c.Writer.Status()
			errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
			bodySize := c.Writer.Size()
			if raw != "" {
				path = path + "?" + raw
			}

			l := log.WithFields(logrus.Fields{
				"latency":    latency,
				"clientIP":   clientIP,
				"userAgent":  userAgent,
				"method":     method,
				"statusCode": statusCode,
				"path":       path,
			})

			if errorMessage == "" {
				l.Infof("%s: wrote %d bytes in %v", http.StatusText(statusCode), bodySize, latency)
			} else {
				l.Errorf("%s: %s", http.StatusText(statusCode), errorMessage)
			}
		}
	}

	return logHandler
}
