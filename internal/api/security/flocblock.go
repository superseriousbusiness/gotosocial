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

package security

import "github.com/gin-gonic/gin"

// FlocBlock is a middleware that prevents google chrome cohort tracking by
// writing the Permissions-Policy header after all other parts of the request
// have been completed. Floc was replaced by Topics in 2022 and the spec says
// that interest-cohort will also block Topics (as of 2022-Nov).
// See: https://smartframe.io/blog/google-topics-api-everything-you-need-to-know
// See: https://github.com/patcg-individual-drafts/topics
func (m *Module) FlocBlock(c *gin.Context) {
	c.Header("Permissions-Policy", "browsing-topics=()")
}
