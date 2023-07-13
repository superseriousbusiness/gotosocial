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
	"strings"

	"github.com/gin-gonic/gin"
)

type CacheControlConfig struct {
	// Slice of Cache-Control directives, which will be
	// joined comma-separated and served as the value of
	// the Cache-Control header.
	//
	// If no directives are set, the Cache-Control header
	// will not be sent in the response at all.
	//
	// For possible Cache-Control directive values, see:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Cache-Control
	Directives []string

	// Slice of Vary header values, which will be joined
	// comma-separated and served as the value of the Vary
	// header in the response.
	//
	// If no Vary header values are supplied, then the
	// Vary header will be omitted in the response.
	//
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Vary
	Vary []string
}

// CacheControl returns a new gin middleware which allows
// routes to control cache settings on response headers.
func CacheControl(config CacheControlConfig) gin.HandlerFunc {
	if len(config.Directives) == 0 {
		// No Cache-Control directives provided,
		// return empty/stub function.
		return func(c *gin.Context) {}
	}

	// Cache control is usually done on hot paths so
	// parse vars outside of the returned function.
	var (
		ccHeader   = strings.Join(config.Directives, ", ")
		varyHeader = strings.Join(config.Vary, ", ")
	)

	if varyHeader == "" {
		return func(c *gin.Context) {
			c.Header("Cache-Control", ccHeader)
		}
	}

	return func(c *gin.Context) {
		c.Header("Cache-Control", ccHeader)
		c.Header("Vary", varyHeader)
	}
}
