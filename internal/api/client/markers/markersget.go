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

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// MarkersGETHandler swagger:operation GET /api/v1/markers markersGet
//
// Get timeline markers by name
//
//	---
//	tags:
//	- markers
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: timeline
//		type: array
//		items:
//			type: string
//			enum:
//				- home
//				- notifications
//		description: Timelines to retrieve.
//		in: query
//
//	security:
//	- OAuth2 Bearer:
//		- read:statuses
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
//		'500':
//			description: internal server error
func (m *Module) MarkersGETHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeReadStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	names, errWithCode := parseMarkerNames(c.QueryArray("timeline[]"))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
	}

	marker, errWithCode := m.processor.Markers().Get(c.Request.Context(), authed.Account, names)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, marker)
}

// parseMarkerNames turns a list of strings into a set of valid marker timeline names, or returns an error.
func parseMarkerNames(nameStrings []string) ([]apimodel.MarkerName, gtserror.WithCode) {
	nameSet := make(map[apimodel.MarkerName]struct{}, apimodel.MarkerNameNumValues)
	for _, timelineString := range nameStrings {
		if err := validate.MarkerName(timelineString); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		nameSet[apimodel.MarkerName(timelineString)] = struct{}{}
	}

	i := 0
	names := make([]apimodel.MarkerName, len(nameSet))
	for name := range nameSet {
		names[i] = name
		i++
	}

	return names, nil
}
