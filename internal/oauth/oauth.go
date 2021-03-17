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
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/go-session/session"
	"github.com/gotosocial/gotosocial/internal/api"
	"github.com/gotosocial/gotosocial/internal/gtsmodel"
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/errors"
	"github.com/gotosocial/oauth2/v4/manage"
	"github.com/gotosocial/oauth2/v4/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const methodAny = "ANY"

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

	api := &API{
		manager: manager,
		server:  srv,
		conn:    conn,
		log:     log,
	}

	api.server.SetPasswordAuthorizationHandler(api.PasswordAuthorizationHandler)
	api.server.SetUserAuthorizationHandler(api.UserAuthorizationHandler)
	api.server.SetClientInfoHandler(server.ClientFormHandler)
	return api
}

func (a *API) AddRoutes(s api.Server) error {
	s.AttachHandler(methodAny, "/auth/sign_in", a.SignInHandler)
	s.AttachHandler(methodAny, "/oauth/token", gin.WrapF(a.TokenHandler))
	s.AttachHandler(methodAny, "/oauth/authorize", gin.WrapF(a.AuthorizeHandler))
	s.AttachHandler(methodAny, "/auth", gin.WrapF(a.AuthHandler))
	return nil
}

func incorrectPassword() (string, error) {
	return "", errors.New("password/email combination was incorrect")
}

/*
	MAIN HANDLERS -- serve these through a server/router
*/

// SignInHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The handler will then redirect to the auth handler served at /auth
func (a *API) SignInHandler(c *gin.Context) {
	s := sessions.Default(c)
	if r.Method == "POST" {
		if r.Form == nil {
			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		s.Set("username", r.Form.Get("username"))
		s.Save()

		w.Header().Set("Location", "/auth")
		w.WriteHeader(http.StatusFound)
		return
	}
	http.ServeContent(w, r, "sign_in.html", time.Unix(0, 0), bytes.NewReader([]byte(signInHTML)))
}

// TokenHandler should be served at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
// See https://docs.joinmastodon.org/methods/apps/oauth/#obtain-a-token
func (a *API) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := a.server.HandleTokenRequest(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// AuthorizeHandler should be served at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept. See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (a *API) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(nil, w, r)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, ok := store.Get("username"); !ok {
		w.Header().Set("Location", "/auth/sign_in")
		w.WriteHeader(http.StatusFound)
		return
	}

	http.ServeContent(w, r, "authorize.html", time.Unix(0, 0), bytes.NewReader([]byte(authorizeHTML)))
}

// AuthHandler should be served at https://example.org/auth
func (a *API) AuthHandler(w http.ResponseWriter, r *http.Request) {
	store, err := session.Start(r.Context(), w, r)
	if err != nil {
		a.log.Errorf("error creating session in authhandler: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
	}
	r.Form = form

	store.Delete("ReturnUri")
	store.Save()

	if err := a.server.HandleAuthorizeRequest(w, r); err != nil {
		a.log.Errorf("error in authhandler during handleauthorizerequest: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

/*
	SUB-HANDLERS -- don't serve these directly, they should be attached to the oauth2 server
*/

// PasswordAuthorizationHandler takes a username (in this case, we use an email address)
// and a password. The goal is to authenticate the password against the one for that email
// address stored in the database. If OK, we return the userid (a uuid) for that user,
// so that it can be used in further Oauth flows to generate a token/retreieve an oauth client from the db.
func (a *API) PasswordAuthorizationHandler(email string, password string) (userid string, err error) {
	a.log.Debugf("entering password authorization handler with email: %s and password: %s", email, password)

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

// UserAuthorizationHandler gets the user's email address from the session key 'username'
// or redirects to the /auth/sign_in page, if this key is not present.
func (a *API) UserAuthorizationHandler(w http.ResponseWriter, r *http.Request) (string, error) {

	a.log.Errorf("entering userauthorizationhandler")

	sessionStore, err := session.Start(r.Context(), w, r)
	if err != nil {
		a.log.Errorf("error starting session: %s", err)
		return "", err
	}

	v, ok := sessionStore.Get("username")
	if !ok {
		if err := r.ParseForm(); err != nil {
			a.log.Errorf("error parsing form: %s", err)
			return "", err
		}

		sessionStore.Set("ReturnUri", r.Form)
		sessionStore.Save()

		w.Header().Set("Location", "/auth/sign_in")
		w.WriteHeader(http.StatusFound)
		return v.(string), nil
	}

	sessionStore.Delete("username")
	sessionStore.Save()
	return v.(string), nil
}
