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

package markers

import (
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/gin-gonic/gin"
)

// MarkersPOSTHandler swagger:operation POST /api/v1/markers markersPost
//
// Update timeline markers by name
//
//	---
//	tags:
//	- markers
//
//	consumes:
//	- multipart/form-data
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: home[last_read_id]
//		type: string
//		description: Last status ID read on the home timeline.
//		in: formData
//	-
//		name: notifications[last_read_id]
//		type: string
//		description: Last notification ID read on the notifications timeline.
//		in: formData
//
//	security:
//	- OAuth2 Bearer:
//		- write:statuses
//
//	responses:
//		'200':
//			description: Requested markers
//			schema:
//				"$ref": "#/definitions/markers"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'409':
//			description: conflict (when two clients try to update the same timeline at the same time)
//		'500':
//			description: internal server error
func (m *Module) MarkersPOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.MarkerPostRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	markers := make([]*gtsmodel.Marker, 0, apimodel.MarkerNameNumValues)
	if homeLastReadID := form.HomeLastReadID(); homeLastReadID != "" {
		markers = append(markers, &gtsmodel.Marker{
			AccountID:  authed.Account.ID,
			Name:       gtsmodel.MarkerNameHome,
			LastReadID: homeLastReadID,
		})
	}
	if notificationsLastReadID := form.NotificationsLastReadID(); notificationsLastReadID != "" {
		markers = append(markers, &gtsmodel.Marker{
			AccountID:  authed.Account.ID,
			Name:       gtsmodel.MarkerNameNotifications,
			LastReadID: notificationsLastReadID,
		})
	}

	marker, errWithCode := m.processor.Markers().Update(c.Request.Context(), markers)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, marker)
}
