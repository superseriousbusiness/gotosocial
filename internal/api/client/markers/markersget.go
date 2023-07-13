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
	"github.com/superseriousbusiness/gotosocial/internal/validate"
	"net/http"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
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
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	timelines, errWithCode := parseTimelines(c.QueryArray("timeline[]"))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
	}

	marker, errWithCode := m.processor.Markers().Get(c.Request.Context(), authed.Account, timelines)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	c.JSON(http.StatusOK, marker)
}

// parseTimelines turns a list of strings into a set of valid marker timeline names, or returns an error.
func parseTimelines(timelineStrings []string) ([]apimodel.MarkerTimelineName, gtserror.WithCode) {
	timelineSet := make(map[apimodel.MarkerTimelineName]struct{}, apimodel.MarkerTimelineNameNumValues)
	for _, timelineString := range timelineStrings {
		if err := validate.MarkerTimelineName(timelineString); err != nil {
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		timelineSet[apimodel.MarkerTimelineName(timelineString)] = struct{}{}
	}

	i := 0
	timelines := make([]apimodel.MarkerTimelineName, len(timelineSet))
	for timeline := range timelineSet {
		timelines[i] = timeline
		i++
	}

	return timelines, nil
}
