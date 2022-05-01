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

package user

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/api"
)

// StatusGETHandler serves the target status as an activitystreams NOTE so that other AP servers can parse it.
func (m *Module) StatusGETHandler(c *gin.Context) {
	l := logrus.WithFields(logrus.Fields{
		"func": "StatusGETHandler",
		"url":  c.Request.RequestURI,
	})

	// usernames on our instance are always lowercase
	requestedUsername := strings.ToLower(c.Param(UsernameKey))
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	// status IDs on our instance are always uppercase
	requestedStatusID := strings.ToUpper(c.Param(StatusIDKey))
	if requestedStatusID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no status id specified in request"})
		return
	}

	format, err := api.NegotiateAccept(c, api.ActivityPubAcceptHeaders...)
	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("negotiated format: %s", format)

	ctx := transferContext(c)

	status, errWithCode := m.processor.GetFediStatus(ctx, requestedUsername, requestedStatusID, c.Request.URL)
	if errWithCode != nil {
		l.Info(errWithCode.Error())
		c.JSON(errWithCode.Code(), gin.H{"error": errWithCode.Safe()})
		return
	}

	b, mErr := json.Marshal(status)
	if mErr != nil {
		err := fmt.Errorf("could not marshal json: %s", mErr)
		l.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, format, b)
}
