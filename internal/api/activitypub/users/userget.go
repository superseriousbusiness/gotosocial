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

package users

import (
	"errors"
	"net/http"
	"strings"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/gin-gonic/gin"
)

// UsersGETHandler should be served at https://example.org/users/:username.
//
// The goal here is to return the activitypub representation of an account
// in the form of a vocab.ActivityStreamsPerson. This should only be served
// to REMOTE SERVERS that present a valid signature on the GET request, on
// behalf of a user, otherwise we risk leaking information about users publicly.
//
// And of course, the request should be refused if the account or server making the
// request is blocked.
func (m *Module) UsersGETHandler(c *gin.Context) {
	// usernames on our instance are always lowercase
	requestedUsername := strings.ToLower(c.Param(UsernameKey))
	if requestedUsername == "" {
		err := errors.New("no username specified in request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	contentType, err := apiutil.NegotiateAccept(c, apiutil.ActivityPubOrHTMLHeaders...)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if contentType == string(apiutil.TextHTML) {
		// redirect to the user's profile
		c.Redirect(http.StatusSeeOther, "/@"+requestedUsername)
		return
	}

	resp, errWithCode := m.processor.Fedi().UserGet(c.Request.Context(), requestedUsername, c.Request.URL)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	apiutil.JSONType(c, http.StatusOK, contentType, resp)
}
