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

package accounts

import (
	"errors"
	"net/http"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/gin-gonic/gin"
)

// AccountMutePOSTHandler swagger:operation POST /api/v1/accounts/{id}/mute accountMute
//
// Mute account by ID.
//
// If account was already muted, succeeds anyway. This can be used to update the details of a mute.
//
//	---
//	tags:
//	- accounts
//
//	produces:
//	- application/json
//
//	parameters:
//	-
//		name: id
//		type: string
//		description: The ID of the account to block.
//		in: path
//		required: true
//	-
//		name: notifications
//		type: boolean
//		description: Mute notifications as well as posts.
//		in: formData
//		required: false
//		default: false
//	-
//		name: duration
//		type: number
//		description: How long the mute should last, in seconds. If 0 or not provided, mute lasts indefinitely.
//		in: formData
//		required: false
//		default: 0
//
//	security:
//	- OAuth2 Bearer:
//		- write:mutes
//
//	responses:
//		'200':
//			description: Your relationship to the account.
//			schema:
//				"$ref": "#/definitions/accountRelationship"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden to moved accounts
//		'404':
//			description: not found
//		'406':
//			description: not acceptable
//		'500':
//			description: internal server error
func (m *Module) AccountMutePOSTHandler(c *gin.Context) {
	authed, errWithCode := apiutil.TokenAuth(c,
		true, true, true, true,
		apiutil.ScopeWriteMutes,
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

	targetAcctID := c.Param(IDKey)
	if targetAcctID == "" {
		err := errors.New("no account id specified")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	form := &apimodel.UserMuteCreateUpdateRequest{}
	if err := c.ShouldBind(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if err := normalizeCreateUpdateMute(form); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnprocessableEntity(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	relationship, errWithCode := m.processor.Account().MuteCreate(
		c.Request.Context(),
		authed.Account,
		targetAcctID,
		form,
	)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSON(c, http.StatusOK, relationship)
}

func normalizeCreateUpdateMute(form *apimodel.UserMuteCreateUpdateRequest) error {
	// Apply defaults for missing fields.
	form.Notifications = util.Ptr(util.PtrOrValue(form.Notifications, false))

	// Normalize duration if necessary.
	if form.DurationI != nil {
		// If we parsed this as JSON, duration
		// may be either a float64 or a string.
		duration, err := apiutil.ParseDuration(form.DurationI, "duration")
		if err != nil {
			return err
		}
		form.Duration = duration
	}

	// Interpret zero as indefinite duration.
	if form.Duration != nil && *form.Duration == 0 {
		form.Duration = nil
	}

	return nil
}
