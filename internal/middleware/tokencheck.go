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

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/oauth2/v4"
)

// TokenCheck returns a new gin middleware for validating oauth tokens in requests.
//
// The middleware checks the request Authorization header for a valid oauth Bearer token.
//
// If no token was set in the Authorization header, or the token was invalid, the handler will return.
//
// If a valid oauth Bearer token was provided, it will be set on the gin context for further use.
//
// Then, it will check which *gtsmodel.User the token belongs to. If the user is not confirmed, not approved,
// or has been disabled, then the middleware will return early. Otherwise, the User will be set on the
// gin context for further processing by other functions.
//
// Next, it will look up the *gtsmodel.Account for the User. If the Account has been suspended, then the
// middleware will return early. Otherwise, it will set the Account on the gin context too.
//
// Finally, it will check the client ID of the token to see if a *gtsmodel.Application can be retrieved
// for that client ID. This will also be set on the gin context.
//
// If an invalid token is presented, or a user/account/application can't be found, then this middleware
// won't abort the request, since the server might want to still allow public requests that don't have a
// Bearer token set (eg., for public instance information and so on).
func TokenCheck(dbConn db.DB, validateBearerToken func(r *http.Request) (oauth2.TokenInfo, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if c.Request.Header.Get("Authorization") == "" {
			// no token set in the header, we can just bail
			return
		}

		ti, err := validateBearerToken(c.Copy().Request)
		if err != nil {
			log.Debugf("token was passed in Authorization header but we could not validate it: %s", err)
			return
		}
		c.Set(oauth.SessionAuthorizedToken, ti)

		// check for user-level token
		if userID := ti.GetUserID(); userID != "" {
			log.Tracef("authenticated user %s with bearer token, scope is %s", userID, ti.GetScope())

			// fetch user for this token
			user, err := dbConn.GetUserByID(ctx, userID)
			if err != nil {
				if err != db.ErrNoEntries {
					log.Errorf("database error looking for user with id %s: %s", userID, err)
					return
				}
				log.Warnf("no user found for userID %s", userID)
				return
			}

			if user.ConfirmedAt.IsZero() {
				log.Warnf("authenticated user %s has never confirmed thier email address", userID)
				return
			}

			if !*user.Approved {
				log.Warnf("authenticated user %s's account was never approved by an admin", userID)
				return
			}

			if *user.Disabled {
				log.Warnf("authenticated user %s's account was disabled'", userID)
				return
			}

			c.Set(oauth.SessionAuthorizedUser, user)

			// fetch account for this token
			if user.Account == nil {
				acct, err := dbConn.GetAccountByID(ctx, user.AccountID)
				if err != nil {
					if err != db.ErrNoEntries {
						log.Errorf("database error looking for account with id %s: %s", user.AccountID, err)
						return
					}
					log.Warnf("no account found for userID %s", userID)
					return
				}
				user.Account = acct
			}

			if !user.Account.SuspendedAt.IsZero() {
				log.Warnf("authenticated user %s's account (accountId=%s) has been suspended", userID, user.AccountID)
				return
			}

			c.Set(oauth.SessionAuthorizedAccount, user.Account)
		}

		// check for application token
		if clientID := ti.GetClientID(); clientID != "" {
			log.Tracef("authenticated client %s with bearer token, scope is %s", clientID, ti.GetScope())

			// fetch app for this token
			app := &gtsmodel.Application{}
			if err := dbConn.GetWhere(ctx, []db.Where{{Key: "client_id", Value: clientID}}, app); err != nil {
				if err != db.ErrNoEntries {
					log.Errorf("database error looking for application with clientID %s: %s", clientID, err)
					return
				}
				log.Warnf("no app found for client %s", clientID)
				return
			}
			c.Set(oauth.SessionAuthorizedApplication, app)
		}
	}
}
