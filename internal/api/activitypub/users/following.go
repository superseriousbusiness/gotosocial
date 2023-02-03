/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// FollowingGETHandler returns a collection of URIs for accounts that the target user follows, formatted so that other AP servers can understand it.
func (m *Module) FollowingGETHandler(c *gin.Context) {
	// usernames on our instance are always lowercase
	requestedUsername := strings.ToLower(c.Param(UsernameKey))
	if requestedUsername == "" {
		err := errors.New("no username specified in request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	format, err := apiutil.NegotiateAccept(c, apiutil.HTMLOrActivityPubHeaders...)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if format == string(apiutil.TextHTML) {
		// redirect to the user's profile
		c.Redirect(http.StatusSeeOther, "/@"+requestedUsername)
		return
	}

	resp, errWithCode := m.processor.GetFediFollowing(apiutil.TransferSignatureContext(c), requestedUsername, c.Request.URL)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	b, err := json.Marshal(resp)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}

	c.Data(http.StatusOK, format, b)
}
