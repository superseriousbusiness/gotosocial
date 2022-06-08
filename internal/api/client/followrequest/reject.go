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

package followrequest

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// FollowRequestRejectPOSTHandler swagger:operation POST /api/v1/follow_requests/{account_id}/reject rejectFollowRequest
//
// Reject/deny follow request from the given account ID.
//
// ---
// tags:
// - follow_requests
//
// produces:
// - application/json
//
// parameters:
// - name: account_id
//   type: string
//   description: ID of the account requesting to follow you.
//   in: path
//   required: true
//
// security:
// - OAuth2 Bearer:
//   - write:follows
//
// responses:
//   '200':
//     name: account relationship
//     description: Your relationship to this account.
//     schema:
//       "$ref": "#/definitions/accountRelationship"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '404':
//      description: not found
//   '406':
//      description: not acceptable
//   '500':
//      description: internal server error
func (m *Module) FollowRequestRejectPOSTHandler(c *gin.Context) {
	authed, err := oauth.Authed(c, true, true, true, true)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if _, err := api.NegotiateAccept(c, api.JSONAcceptHeaders...); err != nil {
		api.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	originAccountID := c.Param(IDKey)
	if originAccountID == "" {
		err := errors.New("no account id specified")
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	relationship, errWithCode := m.processor.FollowRequestReject(c.Request.Context(), authed, originAccountID)
	if errWithCode != nil {
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	c.JSON(http.StatusOK, relationship)
}
