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

package announcements

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// AnnouncementsGETHandler swagger:operation GET /api/v1/announcements announcementsGet
//
// Get an array of currently active announcements.
//
// THIS ENDPOINT IS CURRENTLY NOT FULLY IMPLEMENTED: it will always return an empty array.
//
//	---
//	tags:
//	- announcements
//
//	produces:
//	- application/json
//
//	security:
//	- OAuth2 Bearer: []
//
//	responses:
//		'200':
//			schema:
//				type: array
//				items:
//					type: object
//				maxItems: 0
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) AnnouncementsGETHandler(c *gin.Context) {
	_, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, apiutil.EmptyJSONArray)
}
