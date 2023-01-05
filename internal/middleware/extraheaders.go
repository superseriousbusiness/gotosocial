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

import "github.com/gin-gonic/gin"

// ExtraHeaders returns a new gin middleware which adds various extra headers to the response.
func ExtraHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Inform all callers which server implementation this is.
		c.Header("Server", "gotosocial")
		// Prevent google chrome cohort tracking. Originally this was referred
		// to as FlocBlock. Floc was replaced by Topics in 2022 and the spec says
		// that interest-cohort will also block Topics (as of 2022-Nov).
		//
		// See: https://smartframe.io/blog/google-topics-api-everything-you-need-to-know
		//
		// See: https://github.com/patcg-individual-drafts/topics
		c.Header("Permissions-Policy", "browsing-topics=()")
	}
}
