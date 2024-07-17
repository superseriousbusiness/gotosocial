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
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/util"
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
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
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

	// Normalize mute duration if necessary.
	// If we parsed this as JSON, expires_in
	// may be either a float64 or a string.
	if ei := form.DurationI; ei != nil {
		switch e := ei.(type) {
		case float64:
			form.Duration = util.Ptr(int(e))

		case string:
			duration, err := strconv.Atoi(e)
			if err != nil {
				return fmt.Errorf("could not parse duration value %s as integer: %w", e, err)
			}

			form.Duration = &duration

		default:
			return fmt.Errorf("could not parse expires_in type %T as integer", ei)
		}
	}

	// Interpret zero as indefinite duration.
	if form.Duration != nil && *form.Duration == 0 {
		form.Duration = nil
	}

	return nil
}
