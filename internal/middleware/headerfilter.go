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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/state"

	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
)

func RequestHeaderFilter(state *state.State) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		hdr := c.Request.Header

		var blocked bool
		var err error

		// First pass through any positive filters, only
		// requests that *match* here may pass through.
		blocked, err = state.DB.HeaderMatchPositive(ctx, hdr)
		if err != nil {
			respondInternalServerError(c, err)
			return
		}

		if blocked {
			// Request was blocked,
			// respond and abort here.
			respondBlocked(c)
			return
		}

		// Secondly pass through any negative filters,
		// *any* requests that match here will be denied.
		blocked, err = state.DB.HeaderMatchNegative(ctx, hdr)
		if err != nil {
			respondInternalServerError(c, err)
			return
		}

		if blocked {
			// Request was blocked,
			// respond and abort here.
			respondBlocked(c)
			return
		}

		// Pass to
		// next.
		c.Next()
	}
}

func respondBlocked(c *gin.Context) {
	apiutil.Data(c,
		http.StatusForbidden,
		apiutil.AppJSON,
		apiutil.StatusForbiddenJSON,
	)
	c.Abort()
}

func respondInternalServerError(c *gin.Context, err error) {
	apiutil.Data(c,
		http.StatusInternalServerError,
		apiutil.AppJSON,
		apiutil.StatusInternalServerErrorJSON,
	)
	c.Error(err)
	c.Abort()
}
