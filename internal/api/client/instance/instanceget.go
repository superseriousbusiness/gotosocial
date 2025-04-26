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

package instance

import (
	"net/http"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"

	"github.com/gin-gonic/gin"
)

// InstanceInformationV1GETHandlerV1 swagger:operation GET /api/v1/instance instanceGetV1
//
// View instance information.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			description: "Instance information."
//			schema:
//				"$ref": "#/definitions/instanceV1"
//		'406':
//			description: not acceptable
//		'500':
//			description: internal error
func (m *Module) InstanceInformationGETHandlerV1(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	switch config.GetInstanceStatsMode() {

	case config.InstanceStatsModeBaffle:
		// Replace actual stats with cached randomized ones.
		instance.Stats["user_count"] = util.Ptr(int(instance.RandomStats.TotalUsers))
		instance.Stats["status_count"] = util.Ptr(int(instance.RandomStats.Statuses))

	case config.InstanceStatsModeZero:
		// Replace actual stats with zero.
		instance.Stats["user_count"] = new(int)
		instance.Stats["status_count"] = new(int)

	default:
		// serve or default.
		// Leave stats alone.
	}

	apiutil.JSON(c, http.StatusOK, instance)
}

// InstanceInformationGETHandlerV2 swagger:operation GET /api/v2/instance instanceGetV2
//
// View instance information.
//
//	---
//	tags:
//	- instance
//
//	produces:
//	- application/json
//
//	responses:
//		'200':
//			description: "Instance information."
//			schema:
//				"$ref": "#/definitions/instanceV2"
//		'406':
//			description: not acceptable
//		'500':
//			description: internal error
func (m *Module) InstanceInformationGETHandlerV2(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	instance, errWithCode := m.processor.InstanceGetV2(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	switch config.GetInstanceStatsMode() {

	case config.InstanceStatsModeBaffle:
		// Replace actual stats with cached randomized ones.
		instance.Usage.Users.ActiveMonth = int(instance.RandomStats.MonthlyActiveUsers)

	case config.InstanceStatsModeZero:
		// Replace actual stats with zero.
		instance.Usage.Users.ActiveMonth = 0

	default:
		// serve or default.
		// Leave stats alone.
	}

	apiutil.JSON(c, http.StatusOK, instance)
}
