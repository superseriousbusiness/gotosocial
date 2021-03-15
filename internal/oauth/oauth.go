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

package oauth

import (
	"github.com/go-pg/pg/v10"
	"github.com/gotosocial/gotosocial/internal/api"
	"github.com/gotosocial/gotosocial/internal/gtsmodel"
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/errors"
	"github.com/gotosocial/oauth2/v4/manage"
	"github.com/gotosocial/oauth2/v4/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type API struct {
	manager *manage.Manager
	server  *server.Server
	conn    *pg.DB
	log     *logrus.Logger
}

func New(ts oauth2.TokenStore, cs oauth2.ClientStore, conn *pg.DB, log *logrus.Logger) *API {
	manager := manage.NewDefaultManager()
	manager.MapTokenStorage(ts)
	manager.MapClientStorage(cs)

	srv := server.NewDefaultServer(manager)
	srv.SetInternalErrorHandler(func(err error) *errors.Response {
		log.Errorf("internal oauth error: %s", err)
		return nil
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Errorf("internal response error: %s", re.Error)
	})

	return &API{
		manager: manager,
		server:  srv,
		conn:    conn,
		log:     log,
	}
}

func (a *API) AddRoutes(s api.Server) error {
	return nil
}

func incorrectPassword() (string, error) {
	return "", errors.New("password/email combination was incorrect")
}

func (a *API) PasswordAuthorizationHandler(email string, password string) (userid string, err error) {
	// first we select the user from the database based on email address, bail if no user found for that email
	gtsUser := &gtsmodel.User{}
	if err := a.conn.Model(gtsUser).Where("email = ?", email).Select(); err != nil {
		a.log.Debugf("user %s was not retrievable from db during oauth authorization attempt: %s", email, err)
		return incorrectPassword()
	}

	// make sure a password is actually set and bail if not
	if gtsUser.EncryptedPassword == "" {
		a.log.Warnf("encrypted password for user %s was empty for some reason", gtsUser.Email)
		return incorrectPassword()
	}

	// compare the provided password with the encrypted one from the db, bail if they don't match
	if err := bcrypt.CompareHashAndPassword([]byte(gtsUser.EncryptedPassword), []byte(password)); err != nil {
		a.log.Debugf("password hash didn't match for user %s during login attempt: %s", gtsUser.Email, err)
		return incorrectPassword()
	}

	// If we've made it this far the email/password is correct so we need the oauth client-id of the user
	// This is, conveniently, the same as the user ID, so we can just return it.
	userid = gtsUser.ID
	return
}
