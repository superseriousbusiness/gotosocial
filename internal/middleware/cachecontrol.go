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
	"strings"

	"github.com/gin-gonic/gin"
)

// CacheControl returns a new gin middleware which allows callers to control cache settings on response headers.
//
// For directives, see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
func CacheControl(directives ...string) gin.HandlerFunc {
	ccHeader := strings.Join(directives, ", ")
	return func(c *gin.Context) {
		c.Header("Cache-Control", ccHeader)
	}
}
