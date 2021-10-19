/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
// ---
// tags:
// - s2s/federation
//
// produces:
// - application/activity+json
//
// parameters:
// - name: username
//   type: string
//   description: Username of the account.
//   in: path
//   required: true
// - name: status
//   type: string
//   description: ID of the status.
//   in: path
//   required: true
// - name: page
//   type: boolean
//   description: Return response as a CollectionPage.
//   in: query
//   default: false
// - name: only_other_accounts
//   type: boolean
//   description: Return replies only from accounts other than the status owner.
//   in: query
//   default: false
// - name: min_id
//   type: string
//   description: Minimum ID of the next status, used for paging.
//   in: query
//
// responses:
//   '200':
//      in: body
//      schema:
//        "$ref": "#/definitions/swaggerCollection"
//   '400':
//      description: bad request
//   '401':
//      description: unauthorized
//   '403':
//      description: forbidden
//   '404':
//      description: not found
func (m *Module) StatusRepliesGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func": "StatusRepliesGETHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	requestedStatusID := c.Param(StatusIDKey)
	if requestedStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id specified in request"})
		return
	}

	page := false
	pageString := c.Query(PageKey)
	if pageString != "" {
		i, err := strconv.ParseBool(pageString)
		if err != nil {
			l.Debugf("error parsing page string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse page query param"})
			return
		}
		page = i
	}

	onlyOtherAccounts := false
	onlyOtherAccountsString := c.Query(OnlyOtherAccountsKey)
	if onlyOtherAccountsString != "" {
		i, err := strconv.ParseBool(onlyOtherAccountsString)
		if err != nil {
			l.Debugf("error parsing only_other_accounts string: %s", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't parse only_other_accounts query param"})
			return
		}
		onlyOtherAccounts = i
	}

	minID := ""
	minIDString := c.Query(MinIDKey)
	if minIDString != "" {
		minID = minIDString
	}

	format, err := negotiateFormat(c)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": fmt.Sprintf("could not negotiate format with given Accept header(s): %s", err)})
		return
	}
	l.Tracef("negotiated format: %s", format)

	ctx := transferContext(c)

	replies, errWithCode := m.processor.GetFediStatusReplies(ctx, requestedUsername, requestedStatusID, page, onlyOtherAccounts, minID, c.Request.URL)
	if errWithCode != nil {
		l.Info(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(replies)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, format, b)
}
