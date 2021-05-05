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
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	l := m.log.WithFields(logrus.Fields{
		"func": "UsersGETHandler",
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

	// get the account the request is referring to
	requestedAccount := &gtsmodel.Account{}
	if err := m.db.GetLocalAccountByUsername(requestedUsername, requestedAccount); err != nil {
		l.Errorf("database error getting account with username %s: %s", requestedUsername, err)
		// we'll just return not authorized here to avoid giving anything away
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	// and create a transport for it
	transport, err := m.federator.TransportController().NewTransport(requestedAccount.PublicKeyURI, requestedAccount.PrivateKey)
	if err != nil {
		l.Errorf("error creating transport for username %s: %s", requestedUsername, err)
		// we'll just return not authorized here to avoid giving anything away
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	// authenticate the request
	authentication, err := federation.AuthenticateFederatedRequest(transport, c.Request)
	if err != nil {
		l.Errorf("error authenticating GET user request: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	if !authentication.Authenticated {
		l.Debug("request not authorized")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	requestingAccount := &gtsmodel.Account{}
	if authentication.RequestingPublicKeyID != nil {
		if err := m.db.GetWhere("public_key_uri", authentication.RequestingPublicKeyID.String(), requestingAccount); err != nil {

		}
	}

	authorization, err := federation.AuthorizeFederatedRequest

	person, err := m.tc.AccountToAS(requestedAccount)
	if err != nil {
		l.Errorf("error converting account to ap person: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	data, err := person.Serialize()
	if err != nil {
		l.Errorf("error serializing user: %s", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authorized"})
		return
	}

	c.JSON(http.StatusOK, data)
}
