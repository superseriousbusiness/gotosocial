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
	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
)

// RobotsHeaders adds robots directives to the X-Robots-Tag HTTP header.
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Robots-Tag
//
// If mode == "aiOnly" then only the noai and noimageai values will be set,
// and other headers will be left alone (for route groups / handlers to set).
//
// If mode == "allowSome" then noai, noimageai, and some indexing will be set.
//
// If mode == "" then noai, noimageai, noindex, and nofollow will be set
// (ie., as restrictive as possible).
func RobotsHeaders(mode string) gin.HandlerFunc {
	const (
		key = "X-Robots-Tag"
		// Some AI scrapers respect the following tags
		// to opt-out of their crawling and datasets.
		// We add them regardless of allowSome.
		noai = "noai, noimageai"
	)

	switch mode {

	// Just set ai headers and
	// leave the other headers be.
	case "aiOnly":
		return func(c *gin.Context) {
			c.Writer.Header().Set(key, noai)
		}

	// Allow some limited indexing.
	case "allowSome":
		return func(c *gin.Context) {
			c.Writer.Header().Set(key, apiutil.RobotsDirectivesAllowSome)
			c.Writer.Header().Add(key, noai)
		}

	// Disallow indexing via noindex, nofollow.
	default:
		return func(c *gin.Context) {
			c.Writer.Header().Set(key, apiutil.RobotsDirectivesDisallow)
			c.Writer.Header().Add(key, noai)
		}
	}
}
