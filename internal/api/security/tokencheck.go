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

package security

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// TokenCheck checks if the client has presented a valid oauth Bearer token.
// If so, it will check the User that the token belongs to, and set that in the context of
// the request. Then, it will look up the account for that user, and set that in the request too.
// If user or account can't be found, then the handler won't *fail*, in case the server wants to allow
// public requests that don't have a Bearer token set (eg., for public instance information and so on).
func (m *Module) TokenCheck(c *gin.Context) {
	l := logrus.WithField("func", "OauthTokenMiddleware")
	ctx := c.Request.Context()
	defer c.Next()

	if c.Request.Header.Get("Authorization") == "" {
		// no token set in the header, we can just bail
		return
	}

	ti, err := m.server.ValidationBearerToken(c.Copy().Request)
	if err != nil {
		l.Infof("token was passed in Authorization header but we could not validate it: %s", err)
		return
	}
	c.Set(oauth.SessionAuthorizedToken, ti)

	// check for user-level token
	if userID := ti.GetUserID(); userID != "" {
		l.Tracef("authenticated user %s with bearer token, scope is %s", userID, ti.GetScope())

		// fetch user for this token
		user := &gtsmodel.User{}
		if err := m.db.GetByID(ctx, userID, user); err != nil {
			if err != db.ErrNoEntries {
				l.Errorf("database error looking for user with id %s: %s", userID, err)
				return
			}
			l.Warnf("no user found for userID %s", userID)
			return
		}

		if user.ConfirmedAt.IsZero() {
			l.Warnf("authenticated user %s has never confirmed thier email address", userID)
			return
		}

		if !user.Approved {
			l.Warnf("authenticated user %s's account was never approved by an admin", userID)
			return
		}

		if user.Disabled {
			l.Warnf("authenticated user %s's account was disabled'", userID)
			return
		}

		c.Set(oauth.SessionAuthorizedUser, user)

		// fetch account for this token
		acct, err := m.db.GetAccountByID(ctx, user.AccountID)
		if err != nil {
			if err != db.ErrNoEntries {
				l.Errorf("database error looking for account with id %s: %s", user.AccountID, err)
				return
			}
			l.Warnf("no account found for userID %s", userID)
			return
		}

		if !acct.SuspendedAt.IsZero() {
			l.Warnf("authenticated user %s's account (accountId=%s) has been suspended", userID, user.AccountID)
			return
		}

		c.Set(oauth.SessionAuthorizedAccount, acct)
	}

	// check for application token
	if clientID := ti.GetClientID(); clientID != "" {
		l.Tracef("authenticated client %s with bearer token, scope is %s", clientID, ti.GetScope())

		// fetch app for this token
		app := &gtsmodel.Application{}
		if err := m.db.GetWhere(ctx, []db.Where{{Key: "client_id", Value: clientID}}, app); err != nil {
			if err != db.ErrNoEntries {
				l.Errorf("database error looking for application with clientID %s: %s", clientID, err)
				return
			}
			l.Warnf("no app found for client %s", clientID)
			return
		}
		c.Set(oauth.SessionAuthorizedApplication, app)
	}
}
