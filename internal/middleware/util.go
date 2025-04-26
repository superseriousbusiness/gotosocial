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

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"github.com/gin-gonic/gin"
)

// respondBlocked responds to the given gin context with
// status forbidden, and a generic forbidden JSON response,
// finally aborting the gin handler chain.
func respondBlocked(c *gin.Context) {
	apiutil.Data(c,
		http.StatusForbidden,
		apiutil.AppJSON,
		apiutil.StatusForbiddenJSON,
	)
	c.Abort()
}

// respondInternalServerError responds to the given gin context
// with status internal server error, a generic internal server
// error JSON response, sets the given error on the gin context
// for later logging, finally aborting the gin handler chain.
func respondInternalServerError(c *gin.Context, err error) {
	apiutil.Data(c,
		http.StatusInternalServerError,
		apiutil.AppJSON,
		apiutil.StatusInternalServerErrorJSON,
	)
	_ = c.Error(err)
	c.Abort()
}
