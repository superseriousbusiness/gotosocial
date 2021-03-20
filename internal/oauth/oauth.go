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
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/gotosocial/gotosocial/internal/api"
	"github.com/gotosocial/gotosocial/internal/gtsmodel"
	"github.com/gotosocial/gotosocial/pkg/mastotypes"
	"github.com/gotosocial/oauth2/v4"
	"github.com/gotosocial/oauth2/v4/errors"
	"github.com/gotosocial/oauth2/v4/manage"
	"github.com/gotosocial/oauth2/v4/server"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	outOfBandRedirect  = "urn:ietf:wg:oauth:2.0:oob"
	appsPath           = "/api/v1/apps"
	authSignInPath     = "/auth/sign_in"
	oauthTokenPath     = "/oauth/token"
	oauthAuthorizePath = "/oauth/authorize"
)

type API struct {
	manager *manage.Manager
	server  *server.Server
	conn    *pg.DB
	log     *logrus.Logger
}

type login struct {
	Email    string `form:"username"`
	Password string `form:"password"`
}

type code struct {
	Code string `form:"code"`
}

func New(ts oauth2.TokenStore, cs oauth2.ClientStore, conn *pg.DB, log *logrus.Logger) *API {
	manager := manage.NewDefaultManager()
	manager.MapTokenStorage(ts)
	manager.MapClientStorage(cs)
	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	sc := &server.Config{
		TokenType: "Bearer",
		// Must follow the spec.
		AllowGetAccessRequest: false,
		// Support only the non-implicit flow.
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code},
		// Allow:
		// - Authorization Code (for first & third parties)
		// - Refreshing Tokens
		//
		// Deny:
		// - Resource owner secrets (password grant)
		// - Client secrets
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengePlain,
		},
	}

	srv := server.NewServer(sc, manager)
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

	api.server.SetUserAuthorizationHandler(api.UserAuthorizationHandler)
	api.server.SetClientInfoHandler(server.ClientFormHandler)
	return api
}

func (a *API) AddRoutes(s api.Server) error {
	s.AttachHandler(http.MethodPost, appsPath, a.AppsPOSTHandler)

	s.AttachHandler(http.MethodGet, authSignInPath, a.SignInGETHandler)
	s.AttachHandler(http.MethodPost, authSignInPath, a.SignInPOSTHandler)

	s.AttachHandler(http.MethodPost, oauthTokenPath, a.TokenPOSTHandler)

	s.AttachHandler(http.MethodGet, oauthAuthorizePath, a.AuthorizeGETHandler)
	s.AttachHandler(http.MethodPost, oauthAuthorizePath, a.AuthorizePOSTHandler)

	return nil
}

func incorrectPassword() (string, error) {
	return "", errors.New("password/email combination was incorrect")
}

/*
	MAIN HANDLERS -- serve these through a server/router
*/

// AppsPOSTHandler should be served at https://example.org/api/v1/apps
// It is equivalent to: https://docs.joinmastodon.org/methods/apps/
func (a *API) AppsPOSTHandler(c *gin.Context) {
	l := a.log.WithField("func", "AppsPOSTHandler")
	l.Trace("entering AppsPOSTHandler")

	form := &mastotypes.ApplicationPOSTRequest{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	// permitted length for most fields
	permittedLength := 64
	// redirect can be a bit bigger because we probably need to encode data in the redirect uri
	permittedRedirect := 256

	// check lengths of fields before proceeding so the user can't spam huge entries into the database
	if len(form.ClientName) > permittedLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("client_name must be less than %d bytes", permittedLength)})
		return
	}
	if len(form.Website) > permittedLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("website must be less than %d bytes", permittedLength)})
		return
	}
	if len(form.RedirectURIs) > permittedRedirect {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("redirect_uris must be less than %d bytes", permittedRedirect)})
		return
	}
	if len(form.Scopes) > permittedLength {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("scopes must be less than %d bytes", permittedLength)})
		return
	}

	// set default 'read' for scopes if it's not set
	var scopes string
	if form.Scopes == "" {
		scopes = "read"
	} else {
		scopes = form.Scopes
	}

	// generate new IDs for this application and its associated client
	clientID := uuid.NewString()
	clientSecret := uuid.NewString()
	vapidKey := uuid.NewString()

	// generate the application to put in the database
	app := &gtsmodel.Application{
		Name:         form.ClientName,
		Website:      form.Website,
		RedirectURI:  form.RedirectURIs,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		VapidKey:     vapidKey,
	}

	// chuck it in the db
	if _, err := a.conn.Model(app).Insert(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// now we need to model an oauth client from the application that the oauth library can use
	oc := &oauthClient{
		ID:     clientID,
		Secret: clientSecret,
		Domain: form.RedirectURIs,
		UserID: "", // This client isn't yet associated with a specific user,  it's just an app client right now
	}

	// chuck it in the db
	if _, err := a.conn.Model(oc).Insert(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// done, return the new app information per the spec here: https://docs.joinmastodon.org/methods/apps/
	c.JSON(http.StatusOK, app)
}

// SignInGETHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The form will then POST to the sign in page, which will be handled by SignInPOSTHandler
func (a *API) SignInGETHandler(c *gin.Context) {
	a.log.WithField("func", "SignInGETHandler").Trace("serving sign in html")
	c.HTML(http.StatusOK, "sign-in.tmpl", gin.H{})
}

// SignInPOSTHandler should be served at https://example.org/auth/sign_in.
// The idea is to present a sign in page to the user, where they can enter their username and password.
// The handler will then redirect to the auth handler served at /auth
func (a *API) SignInPOSTHandler(c *gin.Context) {
	l := a.log.WithField("func", "SignInPOSTHandler")
	s := sessions.Default(c)
	form := &login{}
	if err := c.ShouldBind(form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	l.Tracef("parsed form: %+v", form)

	userid, err := a.ValidatePassword(form.Email, form.Password)
	if err != nil {
		c.String(http.StatusForbidden, err.Error())
		return
	}

	s.Set("username", userid)
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	l.Trace("redirecting to auth page")
	c.Redirect(http.StatusFound, oauthAuthorizePath)
}

// TokenPOSTHandler should be served as a POST at https://example.org/oauth/token
// The idea here is to serve an oauth access token to a user, which can be used for authorizing against non-public APIs.
// See https://docs.joinmastodon.org/methods/apps/oauth/#obtain-a-token
func (a *API) TokenPOSTHandler(c *gin.Context) {
	l := a.log.WithField("func", "TokenPOSTHandler")
	l.Trace("entered TokenPOSTHandler")

	// The commented-out code below doesn't work yet because the oauth2 library can't handle OOB properly!

	// // make sure redirect_uri is actually set first (we don't accept empty)
	// if v, ok := c.GetPostForm("redirect_uri"); !ok || v == "" {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "session missing redirect_uri"})
	// 	return
	// } else if v == outOfBandRedirect {
	// 	// If redirect_uri is set to out of band, redirect to this endpoint, where we can display the code later
	// 	// This is a bit of a workaround because the oauth library doesn't recognise oob redirect URIs
	// 	c.Request.Form.Set("redirect_uri", fmt.Sprintf("%s://%s%s", a.config.Protocol, a.config.Host, oauthTokenPath))
	// }

	if err := a.server.HandleTokenRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

// AuthorizeGETHandler should be served as GET at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept. See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (a *API) AuthorizeGETHandler(c *gin.Context) {
	l := a.log.WithField("func", "AuthorizeGETHandler")
	s := sessions.Default(c)

	// Username will be set in the session by AuthorizePOSTHandler if the caller has already gone through the authentication flow
	// If it's not set, then we don't know yet who the user is, so we need to redirect them to the sign in page.
	v := s.Get("username")
	if username, ok := v.(string); !ok || username == "" {
		l.Trace("username was empty, parsing form then redirecting to sign in page")

		// first make sure they've filled out the authorize form with the required values
		form := &mastotypes.OAuthAuthorize{}
		if err := c.ShouldBind(form); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		l.Tracef("parsed form: %+v", form)

		// these fields are *required* so check 'em
		if form.ResponseType == "" || form.ClientID == "" || form.RedirectURI == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing one of: response_type, client_id or redirect_uri"})
			return
		}

		// save these values from the form so we can use them elsewhere in the session
		s.Set("force_login", form.ForceLogin)
		s.Set("response_type", form.ResponseType)
		s.Set("client_id", form.ClientID)
		s.Set("redirect_uri", form.RedirectURI)
		s.Set("scope", form.Scope)
		if err := s.Save(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// send them to the sign in page so we can tell who they are
		c.Redirect(http.StatusFound, authSignInPath)
		return
	}

	// Check if we have a code already. If we do, it means the user used urn:ietf:wg:oauth:2.0:oob as their redirect URI
	// and were sent here, which means they just want the code displayed so they can use it out of band.
	code := &code{}
	if err := c.Bind(code); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// the authorize template will either:
	// 1. Display the code to the user if they're already authorized and were redirected here because they selected urn:ietf:wg:oauth:2.0:oob.
	// 2. Display a form where they can get some information about the app that's trying to authorize, and approve it, which will then go to AuthorizePOSTHandler
	l.Trace("serving authorize html")
	c.HTML(http.StatusOK, "authorize.tmpl", gin.H{
		"code": code.Code,
	})
}

// AuthorizePOSTHandler should be served as POST at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept. See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (a *API) AuthorizePOSTHandler(c *gin.Context) {
	l := a.log.WithField("func", "AuthorizePOSTHandler")
	s := sessions.Default(c)

	v := s.Get("username")
	if username, ok := v.(string); !ok || username == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "you are not signed in"})
	}

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
		// The commented-out code below doesn't work yet because the oauth2 library can't handle OOB properly!

		// if the client requests this particular redirect URI, it means they want to be able to authenticate out of band,
		// ie., just have their access_code shown to them so they can do what they want with it later.
		//
		// But we can't just show the code yet; there's still an authorization flow to go through.
		// What we can do is set the redirect uri to the /oauth/authorize page, do the auth
		// flow as normal, and then handle showing the code there. See AuthorizeGETHandler.
		// if v == outOfBandRedirect {
		// 	v = fmt.Sprintf("%s://%s%s", a.config.Protocol, a.config.Host, oauthAuthorizePath)
		// }
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
	l.Tracef("values on request set to %+v", c.Request.Form)

	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
func (a *API) ValidatePassword(email string, password string) (userid string, err error) {
	l := a.log.WithField("func", "PasswordAuthorizationHandler")

	// make sure an email/password was provided and bail if not
	if email == "" || password == "" {
		l.Debug("email or password was not provided")
		return incorrectPassword()
	}

	// first we select the user from the database based on email address, bail if no user found for that email
	gtsUser := &gtsmodel.User{}
	if err := a.conn.Model(gtsUser).Where("email = ?", email).Select(); err != nil {
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

// UserAuthorizationHandler gets the user's ID from the 'username' field of the request form,
// or redirects to the /auth/sign_in page, if this key is not present.
func (a *API) UserAuthorizationHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	l := a.log.WithField("func", "UserAuthorizationHandler")
	userID = r.FormValue("username")
	if userID == "" {
		l.Trace("username was empty, redirecting to sign in page")
		http.Redirect(w, r, authSignInPath, http.StatusFound)
		return "", nil
	}
	l.Tracef("returning (%s, %s)", userID, err)
	return userID, err
}
