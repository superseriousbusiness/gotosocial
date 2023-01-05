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
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserAgent returns a gin middleware which aborts requests with
// empty user agent strings, returning code 418 - I'm a teapot.
func UserAgent() gin.HandlerFunc {
	// todo: make this configurable
	return func(c *gin.Context) {
		if ua := c.Request.UserAgent(); ua == "" {
			code := http.StatusTeapot
			err := errors.New(http.StatusText(code) + ": no user-agent sent with request")
			c.AbortWithStatusJSON(code, gin.H{"error": err.Error()})
		}
	}
}
