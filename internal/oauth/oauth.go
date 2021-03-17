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
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
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

const methodAny = "ANY"

type API struct {
	manager *manage.Manager
	server  *server.Server
	conn    *pg.DB
	log     *logrus.Logger
}

type login struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

type authorize struct {
	ForceLogin   string `form:"force_login,omitempty"`
	ResponseType string `form:"response_type"`
	ClientID     string `form:"client_id"`
	RedirectURI  string `form:"redirect_uri"`
	Scope        string `form:"scope,omitempty"`
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
	s.AttachHandler(http.MethodGet, "/auth/sign_in", a.SignInGETHandler)
	s.AttachHandler(http.MethodPost, "/auth/sign_in", a.SignInPOSTHandler)
	s.AttachHandler(methodAny, "/oauth/token", a.TokenHandler)
	s.AttachHandler(http.MethodGet, "/oauth/authorize", a.AuthorizeHandler)
	s.AttachHandler(methodAny, "/auth", a.AuthHandler)
	return nil
}

func incorrectPassword() (string, error) {
	return "", errors.New("password/email combination was incorrect")
}

/*
	MAIN HANDLERS -- serve these through a server/router
*/

// SignInGETHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The form will then POST to the sign in page, which will be handled by SignInPOSTHandler
func (a *API) SignInGETHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(signInHTML))
}

// SignInPOSTHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The handler will then redirect to the auth handler served at /auth
func (a *API) SignInPOSTHandler(c *gin.Context) {
	s := sessions.Default(c)
	form := &login{}
	if err := c.ShouldBind(form); err != nil || form.Username == "" || form.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.Set("username", form.Username)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, "/auth")
}

// TokenHandler should be served at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
// See https://docs.joinmastodon.org/methods/apps/oauth/#obtain-a-token
func (a *API) TokenHandler(c *gin.Context) {
	if err := a.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// AuthorizeHandler should be served as GET at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept. See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (a *API) AuthorizeHandler(c *gin.Context) {
	s := sessions.Default(c)
	form := &authorize{}

	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if form.ResponseType == "" || form.ClientID == "" || form.RedirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing one of: response_type, client_id or redirect_uri"})
		return
	}

	s.Set("force_login", form.ForceLogin)
	s.Set("response_type", form.ResponseType)
	s.Set("client_id", form.ClientID)
	s.Set("redirect_uri", form.RedirectURI)
	s.Set("scope", form.Scope)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	v := s.Get("username")
	if username, ok := v.(string); !ok || username == "" {
		c.Redirect(http.StatusFound, "/auth/sign_in")
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(authorizeHTML))
}

// AuthHandler should be served at https://example.org/auth
func (a *API) AuthHandler(c *gin.Context) {
	s := sessions.Default(c)

	values := url.Values{}

	if v, ok := s.Get("force_login").(string); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing force_login"})
		return
	} else {
		values.Add("force_login", v)
	}

	if v, ok := s.Get("response_type").(string); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing response_type"})
		return
	} else {
		values.Add("response_type", v)
	}

	if v, ok := s.Get("client_id").(string); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing client_id"})
		return
	} else {
		values.Add("client_id", v)
	}

	if v, ok := s.Get("redirect_uri").(string); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing redirect_uri"})
		return
	} else {
		values.Add("redirect_uri", v)
	}

	if v, ok := s.Get("scope").(string); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing scope"})
		return
	} else {
		values.Add("scope", v)
	}

	if v, ok := s.Get("username").(string); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing username"})
		return
	} else {
		values.Add("username", v)
	}

	c.Request.Form = values

	if err := s.Save(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.server.HandleAuthorizeRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// UserAuthorizationHandler gets the user's email address from the form key 'username'
// or redirects to the /auth/sign_in page, if this key is not present.
func (a *API) UserAuthorizationHandler(w http.ResponseWriter, r *http.Request) (string, error) {

	username := r.FormValue("username")
	if username == "" {
		http.Redirect(w, r, "/auth/sign_in", http.StatusFound)
		return "", nil
	}
	return username, nil
}
