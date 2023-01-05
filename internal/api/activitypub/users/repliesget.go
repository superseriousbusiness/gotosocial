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
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
)

// StatusRepliesGETHandler swagger:operation GET /users/{username}/statuses/{status}/replies s2sRepliesGet
//
// Get the replies collection for a status.
//
// Note that the response will be a Collection with a page as `first`, as shown below, if `page` is `false`.
//
// If `page` is `true`, then the response will be a single `CollectionPage` without the wrapping `Collection`.
//
// HTTP signature is required on the request.
//
//	---
//	tags:
//	- s2s/federation
//
//	produces:
//	- application/activity+json
//
//	parameters:
//	-
//		name: username
//		type: string
//		description: Username of the account.
//		in: path
//		required: true
//	-
//		name: status
//		type: string
//		description: ID of the status.
//		in: path
//		required: true
//	-
//		name: page
//		type: boolean
//		description: Return response as a CollectionPage.
//		in: query
//		default: false
//	-
//		name: only_other_accounts
//		type: boolean
//		description: Return replies only from accounts other than the status owner.
//		in: query
//		default: false
//	-
//		name: min_id
//		type: string
//		description: Minimum ID of the next status, used for paging.
//		in: query
//
//	responses:
//		'200':
//			in: body
//			schema:
//				"$ref": "#/definitions/swaggerCollection"
//		'400':
//			description: bad request
//		'401':
//			description: unauthorized
//		'403':
//			description: forbidden
//		'404':
//			description: not found
func (m *Module) StatusRepliesGETHandler(c *gin.Context) {
	// usernames on our instance are always lowercase
	requestedUsername := strings.ToLower(c.Param(UsernameKey))
	if requestedUsername == "" {
		err := errors.New("no username specified in request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// status IDs on our instance are always uppercase
	requestedStatusID := strings.ToUpper(c.Param(StatusIDKey))
	if requestedStatusID == "" {
		err := errors.New("no status id specified in request")
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	format, err := apiutil.NegotiateAccept(c, apiutil.HTMLOrActivityPubHeaders...)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if format == string(apiutil.TextHTML) {
		// redirect to the status
		c.Redirect(http.StatusSeeOther, "/@"+requestedUsername+"/statuses/"+requestedStatusID)
		return
	}

	var page bool
	if pageString := c.Query(PageKey); pageString != "" {
		i, err := strconv.ParseBool(pageString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", PageKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		page = i
	}

	onlyOtherAccounts := false
	onlyOtherAccountsString := c.Query(OnlyOtherAccountsKey)
	if onlyOtherAccountsString != "" {
		i, err := strconv.ParseBool(onlyOtherAccountsString)
		if err != nil {
			err := fmt.Errorf("error parsing %s: %s", OnlyOtherAccountsKey, err)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
			return
		}
		onlyOtherAccounts = i
	}

	minID := ""
	minIDString := c.Query(MinIDKey)
	if minIDString != "" {
		minID = minIDString
	}

	resp, errWithCode := m.processor.GetFediStatusReplies(apiutil.TransferSignatureContext(c), requestedUsername, requestedStatusID, page, onlyOtherAccounts, minID, c.Request.URL)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	b, err := json.Marshal(resp)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}

	c.Data(http.StatusOK, format, b)
}
