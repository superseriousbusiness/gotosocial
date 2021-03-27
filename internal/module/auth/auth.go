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

// Package auth is a module that provides oauth functionality to a router.
// It adds the following paths:
//    /auth/sign_in
//    /oauth/token
//    /oauth/authorize
// It also includes the oauthTokenMiddleware, which can be attached to a router to authenticate every request by Bearer token.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/module"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/router"
	"github.com/superseriousbusiness/gotosocial/pkg/mastotypes"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	authSignInPath     = "/auth/sign_in"
	oauthTokenPath     = "/oauth/token"
	oauthAuthorizePath = "/oauth/authorize"
)

type authModule struct {
	server oauth.Server
	db     db.DB
	log    *logrus.Logger
}

type login struct {
	Email    string `form:"username"`
	Password string `form:"password"`
}

// New returns a new auth module
func New(srv oauth.Server, db db.DB, log *logrus.Logger) module.ClientAPIModule {
	return &authModule{
		server: srv,
		db:     db,
		log:    log,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *authModule) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, authSignInPath, m.signInGETHandler)
	s.AttachHandler(http.MethodPost, authSignInPath, m.signInPOSTHandler)

	s.AttachHandler(http.MethodPost, oauthTokenPath, m.tokenPOSTHandler)

	s.AttachHandler(http.MethodGet, oauthAuthorizePath, m.authorizeGETHandler)
	s.AttachHandler(http.MethodPost, oauthAuthorizePath, m.authorizePOSTHandler)

	s.AttachMiddleware(m.oauthTokenMiddleware)
	return nil
}

/*
	MAIN HANDLERS -- serve these through a server/router
*/

// signInGETHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The form will then POST to the sign in page, which will be handled by SignInPOSTHandler
func (m *authModule) signInGETHandler(c *gin.Context) {
	m.log.WithField("func", "SignInGETHandler").Trace("serving sign in html")
	c.HTML(http.StatusOK, "sign-in.tmpl", gin.H{})
}

// signInPOSTHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The handler will then redirect to the auth handler served at /auth
func (m *authModule) signInPOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "SignInPOSTHandler")
	s := sessions.Default(c)
	form := &login{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("parsed form: %+v", form)

	userid, err := m.validatePassword(form.Email, form.Password)
	if err != nil {
		c.String(http.StatusForbidden, err.Error())
		return
	}

	s.Set("userid", userid)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Trace("redirecting to auth page")
	c.Redirect(http.StatusFound, oauthAuthorizePath)
}

// tokenPOSTHandler should be served as a POST at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
// See https://docs.joinmastodon.org/methods/apps/oauth/#obtain-a-token
func (m *authModule) tokenPOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "TokenPOSTHandler")
	l.Trace("entered TokenPOSTHandler")
	if err := m.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// authorizeGETHandler should be served as GET at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept. See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (m *authModule) authorizeGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "AuthorizeGETHandler")
	s := sessions.Default(c)

	// UserID will be set in the session by AuthorizePOSTHandler if the caller has already gone through the authentication flow
	// If it's not set, then we don't know yet who the user is, so we need to redirect them to the sign in page.
	userID, ok := s.Get("userid").(string)
	if !ok || userID == "" {
		l.Trace("userid was empty, parsing form then redirecting to sign in page")
		if err := parseAuthForm(c, l); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.Redirect(http.StatusFound, authSignInPath)
		}
		return
	}

	// We can use the client_id on the session to retrieve info about the app associated with the client_id
	clientID, ok := s.Get("client_id").(string)
	if !ok || clientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no client_id found in session"})
		return
	}
	app := &model.Application{
		ClientID: clientID,
	}
	if err := m.db.GetWhere("client_id", app.ClientID, app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("no application found for client id %s", clientID)})
		return
	}

	// we can also use the userid of the user to fetch their username from the db to greet them nicely <3
	user := &model.User{
		ID: userID,
	}
	if err := m.db.GetByID(user.ID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acct := &model.Account{
		ID: user.AccountID,
	}

	if err := m.db.GetByID(acct.ID, acct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Finally we should also get the redirect and scope of this particular request, as stored in the session.
	redirect, ok := s.Get("redirect_uri").(string)
	if !ok || redirect == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no redirect_uri found in session"})
		return
	}
	scope, ok := s.Get("scope").(string)
	if !ok || scope == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no scope found in session"})
		return
	}

	// the authorize template will display a form to the user where they can get some information
	// about the app that's trying to authorize, and the scope of the request.
	// They can then approve it if it looks OK to them, which will POST to the AuthorizePOSTHandler
	l.Trace("serving authorize html")
	c.HTML(http.StatusOK, "authorize.tmpl", gin.H{
		"appname":    app.Name,
		"appwebsite": app.Website,
		"redirect":   redirect,
		"scope":      scope,
		"user":       acct.Username,
	})
}

// authorizePOSTHandler should be served as POST at https://example.org/oauth/authorize
// At this point we assume that the user has A) logged in and B) accepted that the app should act for them,
// so we should proceed with the authentication flow and generate an oauth token for them if we can.
// See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (m *authModule) authorizePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "AuthorizePOSTHandler")
	s := sessions.Default(c)

	// At this point we know the user has said 'yes' to allowing the application and oauth client
	// work for them, so we can set the

	// We need to retrieve the original form submitted to the authorizeGEThandler, and
	// recreate it on the request so that it can be used further by the oauth2 library.
	// So first fetch all the values from the session.
	forceLogin, ok := s.Get("force_login").(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing force_login"})
		return
	}
	responseType, ok := s.Get("response_type").(string)
	if !ok || responseType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing response_type"})
		return
	}
	clientID, ok := s.Get("client_id").(string)
	if !ok || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing client_id"})
		return
	}
	redirectURI, ok := s.Get("redirect_uri").(string)
	if !ok || redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing redirect_uri"})
		return
	}
	scope, ok := s.Get("scope").(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing scope"})
		return
	}
	userID, ok := s.Get("userid").(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing userid"})
		return
	}
	// we're done with the session so we can clear it now
	s.Clear()

	// now set the values on the request
	values := url.Values{}
	values.Set("force_login", forceLogin)
	values.Set("response_type", responseType)
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	values.Set("scope", scope)
	values.Set("userid", userID)
	c.Request.Form = values
	l.Tracef("values on request set to %+v", c.Request.Form)

	// and proceed with authorization using the oauth2 library
	if err := m.server.HandleAuthorizeRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

/*
	MIDDLEWARE
*/

// oauthTokenMiddleware checks if the client has presented a valid oauth Bearer token.
// If so, it will check the User that the token belongs to, and set that in the context of
// the request. Then, it will look up the account for that user, and set that in the request too.
// If user or account can't be found, then the handler won't *fail*, in case the server wants to allow
// public requests that don't have a Bearer token set (eg., for public instance information and so on).
func (m *authModule) oauthTokenMiddleware(c *gin.Context) {
	l := m.log.WithField("func", "ValidatePassword")
	l.Trace("entering OauthTokenMiddleware")

	ti, err := m.server.ValidationBearerToken(c.Request)
	if err != nil {
		l.Trace("no valid token presented: continuing with unauthenticated request")
		return
	}
	c.Set(oauth.SessionAuthorizedToken, ti)
	l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedToken, ti)

	// check for user-level token
	if uid := ti.GetUserID(); uid != "" {
		l.Tracef("authenticated user %s with bearer token, scope is %s", uid, ti.GetScope())

		// fetch user's and account for this user id
		user := &model.User{}
		if err := m.db.GetByID(uid, user); err != nil || user == nil {
			l.Warnf("no user found for validated uid %s", uid)
			return
		}
		c.Set(oauth.SessionAuthorizedUser, user)
		l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedUser, user)

		acct := &model.Account{}
		if err := m.db.GetByID(user.AccountID, acct); err != nil || acct == nil {
			l.Warnf("no account found for validated user %s", uid)
			return
		}
		c.Set(oauth.SessionAuthorizedAccount, acct)
		l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedAccount, acct)
	}

	// check for application token
	if cid := ti.GetClientID(); cid != "" {
		l.Tracef("authenticated client %s with bearer token, scope is %s", cid, ti.GetScope())
		app := &model.Application{}
		if err := m.db.GetWhere("client_id", cid, app); err != nil {
			l.Tracef("no app found for client %s", cid)
		}
		c.Set(oauth.SessionAuthorizedApplication, app)
		l.Tracef("set gin context %s to %+v", oauth.SessionAuthorizedApplication, app)
	}
}

/*
	SUB-HANDLERS -- don't serve these directly, they should be attached to the oauth2 server or used inside handler funcs
*/

// validatePassword takes an email address and a password.
// The goal is to authenticate the password against the one for that email
// address stored in the database. If OK, we return the userid (a uuid) for that user,
// so that it can be used in further Oauth flows to generate a token/retreieve an oauth client from the db.
func (m *authModule) validatePassword(email string, password string) (userid string, err error) {
	l := m.log.WithField("func", "ValidatePassword")

	// make sure an email/password was provided and bail if not
	if email == "" || password == "" {
		l.Debug("email or password was not provided")
		return incorrectPassword()
	}

	// first we select the user from the database based on email address, bail if no user found for that email
	gtsUser := &model.User{}

	if err := m.db.GetWhere("email", email, gtsUser); err != nil {
		l.Debugf("user %s was not retrievable from db during oauth authorization attempt: %s", email, err)
		return incorrectPassword()
	}

	// make sure a password is actually set and bail if not
	if gtsUser.EncryptedPassword == "" {
		l.Warnf("encrypted password for user %s was empty for some reason", gtsUser.Email)
		return incorrectPassword()
	}

	// compare the provided password with the encrypted one from the db, bail if they don't match
	if err := bcrypt.CompareHashAndPassword([]byte(gtsUser.EncryptedPassword), []byte(password)); err != nil {
		l.Debugf("password hash didn't match for user %s during login attempt: %s", gtsUser.Email, err)
		return incorrectPassword()
	}

	// If we've made it this far the email/password is correct, so we can just return the id of the user.
	userid = gtsUser.ID
	l.Tracef("returning (%s, %s)", userid, err)
	return
}

// incorrectPassword is just a little helper function to use in the ValidatePassword function
func incorrectPassword() (string, error) {
	return "", errors.New("password/email combination was incorrect")
}

// parseAuthForm parses the OAuthAuthorize form in the gin context, and stores
// the values in the form into the session.
func parseAuthForm(c *gin.Context, l *logrus.Entry) error {
	s := sessions.Default(c)

	// first make sure they've filled out the authorize form with the required values
	form := &mastotypes.OAuthAuthorize{}
	if err := c.ShouldBind(form); err != nil {
		return err
	}
	l.Tracef("parsed form: %+v", form)

	// these fields are *required* so check 'em
	if form.ResponseType == "" || form.ClientID == "" || form.RedirectURI == "" {
		return errors.New("missing one of: response_type, client_id or redirect_uri")
	}

	// set default scope to read
	if form.Scope == "" {
		form.Scope = "read"
	}

	// save these values from the form so we can use them elsewhere in the session
	s.Set("force_login", form.ForceLogin)
	s.Set("response_type", form.ResponseType)
	s.Set("client_id", form.ClientID)
	s.Set("redirect_uri", form.RedirectURI)
	s.Set("scope", form.Scope)
	return s.Save()
}
