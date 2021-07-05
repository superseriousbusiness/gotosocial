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
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// FollowingGETHandler returns a collection of URIs for accounts that the target user follows, formatted so that other AP servers can understand it.
func (m *Module) FollowingGETHandler(c *gin.Context) {
	l := m.log.WithFields(logrus.Fields{
		"func": "FollowingGETHandler",
		"url":  c.Request.RequestURI,
	})

	requestedUsername := c.Param(UsernameKey)
	if requestedUsername == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no username specified in request"})
		return
	}

	// make sure this actually an AP request
	format := c.NegotiateFormat(ActivityPubAcceptHeaders...)
	if format == "" {
		c.JSON(http.StatusNotAcceptable, gin.H{"error": "could not negotiate format with given Accept header(s)"})
		return
	}
	l.Tracef("negotiated format: %s", format)

	// transfer the signature verifier from the gin context to the request context
	ctx := c.Request.Context()
	verifier, signed := c.Get(string(util.APRequestingPublicKeyVerifier))
	if signed {
		ctx = context.WithValue(ctx, util.APRequestingPublicKeyVerifier, verifier)
	}

	user, err := m.processor.GetFediFollowing(ctx, requestedUsername, c.Request.URL) // handles auth as well
	if err != nil {
		l.Info(err.Error())
		c.JSON(err.Code(), gin.H{"error": err.Safe()})
		return
	}

	c.JSON(http.StatusOK, user)
}
