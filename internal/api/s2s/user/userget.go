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

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	l := logrus.WithFields(logrus.Fields{
		"func": "UsersGETHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	format, err := negotiateFormat(c)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": fmt.Sprintf("could not negotiate format with given Accept header(s): %s", err)})
		return
	}
	l.Tracef("negotiated format: %s", format)

	ctx := transferContext(c)

	user, errWithCode := m.processor.GetFediUser(ctx, requestedUsername, c.Request.URL) // GetFediUser handles auth as well
	if errWithCode != nil {
		l.Info(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(user)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, format, b)
}
