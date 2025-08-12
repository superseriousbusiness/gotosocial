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

package scheduledstatuses

import (
	"net/http"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"

	"github.com/gin-gonic/gin"
)

// ScheduledStatusPUTHandler swagger:operation PUT /api/v1/scheduled_statuses/{id} updateScheduledStatus
//
// Update a scheduled status's publishing date
//
//	---
//	tags:
//	- scheduled_statuses
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: ID of the status
//		in: path
//		required: true
//	-
//		name: scheduled_at
//		x-go-name: ScheduledAt
//		description: |-
//			ISO 8601 Datetime at which to schedule a status.
//
//			Must be at least 5 minutes in the future.
//		type: string
//		format: date-time
//		in: formData
//
//	security:
//	- OAuth2 Bearer:
//		- write:statuses
//
//	responses:
//		'200':
//			schema:
//				"$ref": "#/definitions/scheduledStatus"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'422':
//			description: unprocessable content
//		'500':
//			description: internal server error
func (m *Module) ScheduledStatusPUTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteStatuses,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	if authed.Account.IsMoving() {
		apiutil.ForbiddenAfterMove(c)
		return
	}

	if _, err := apiutil.NegotiateAccept(c, apiutil.JSONAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	targetScheduledStatusID, errWithCode := apiutil.ParseID(c.Param(apiutil.IDKey))
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.ScheduledStatusUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	now := time.Now()
	if !now.Add(5 * time.Minute).Before(*form.ScheduledAt) {
		const errText = "scheduled_at must be at least 5 minutes in the future"
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(gtserror.New(errText), errText), m.processor.InstanceGetV1)
		return
	}

	scheduledStatus, errWithCode := m.processor.Status().ScheduledStatusesUpdate(
		c.Request.Context(),
		authed.Account,
		targetScheduledStatusID,
		form.ScheduledAt,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, scheduledStatus)
}
