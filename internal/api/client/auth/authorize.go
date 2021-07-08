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

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// AuthorizeGETHandler should be served as GET at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept. See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (m *Module) AuthorizeGETHandler(c *gin.Context) {
	l := m.log.WithField("func", "AuthorizeGETHandler")
	s := sessions.Default(c)

	// UserID will be set in the session by AuthorizePOSTHandler if the caller has already gone through the authentication flow
	// If it's not set, then we don't know yet who the user is, so we need to redirect them to the sign in page.
	userID, ok := s.Get(sessionUserID).(string)
	if !ok || userID == "" {
		l.Trace("userid was empty, parsing form then redirecting to sign in page")
		form := &model.OAuthAuthorize{}
		if err := c.Bind(form); err != nil {
			l.Debugf("invalid auth form: %s", err)
			return
		}
		l.Tracef("parsed auth form: %+v", form)

		if err := extractAuthForm(s, form); err != nil {
			l.Debugf(fmt.Sprintf("error parsing form at /oauth/authorize: %s", err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, AuthSignInPath)
		return
	}

	// We can use the client_id on the session to retrieve info about the app associated with the client_id
	clientID, ok := s.Get(sessionClientID).(string)
	if !ok || clientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no client_id found in session"})
		return
	}
	app := &gtsmodel.Application{
		ClientID: clientID,
	}
	if err := m.db.GetWhere([]db.Where{{Key: sessionClientID, Value: app.ClientID}}, app); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("no application found for client id %s", clientID)})
		return
	}

	// we can also use the userid of the user to fetch their username from the db to greet them nicely <3
	user := &gtsmodel.User{
		ID: userID,
	}
	if err := m.db.GetByID(user.ID, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acct := &gtsmodel.Account{
		ID: user.AccountID,
	}

	if err := m.db.GetByID(acct.ID, acct); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Finally we should also get the redirect and scope of this particular request, as stored in the session.
	redirect, ok := s.Get(sessionRedirectURI).(string)
	if !ok || redirect == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no redirect_uri found in session"})
		return
	}
	scope, ok := s.Get(sessionScope).(string)
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
		sessionScope: scope,
		"user":       acct.Username,
	})
}

// AuthorizePOSTHandler should be served as POST at https://example.org/oauth/authorize
// At this point we assume that the user has A) logged in and B) accepted that the app should act for them,
// so we should proceed with the authentication flow and generate an oauth token for them if we can.
// See here: https://docs.joinmastodon.org/methods/apps/oauth/#authorize-a-user
func (m *Module) AuthorizePOSTHandler(c *gin.Context) {
	l := m.log.WithField("func", "AuthorizePOSTHandler")
	s := sessions.Default(c)

	// We need to retrieve the original form submitted to the authorizeGEThandler, and
	// recreate it on the request so that it can be used further by the oauth2 library.
	// So first fetch all the values from the session.

	forceLogin, ok := s.Get(sessionForceLogin).(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing force_login"})
		return
	}

	responseType, ok := s.Get(sessionResponseType).(string)
	if !ok || responseType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing response_type"})
		return
	}

	clientID, ok := s.Get(sessionClientID).(string)
	if !ok || clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing client_id"})
		return
	}

	redirectURI, ok := s.Get(sessionRedirectURI).(string)
	if !ok || redirectURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing redirect_uri"})
		return
	}

	scope, ok := s.Get(sessionScope).(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing scope"})
		return
	}

	userID, ok := s.Get(sessionUserID).(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session missing userid"})
		return
	}

	// we're done with the session so we can clear it now
	for _, key := range sessionKeys {
		s.Delete(key)
	}
	if err := s.Save(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// now set the values on the request
	values := url.Values{}
	values.Set(sessionForceLogin, forceLogin)
	values.Set(sessionResponseType, responseType)
	values.Set(sessionClientID, clientID)
	values.Set(sessionRedirectURI, redirectURI)
	values.Set(sessionScope, scope)
	values.Set(sessionUserID, userID)
	c.Request.Form = values
	l.Tracef("values on request set to %+v", c.Request.Form)

	// and proceed with authorization using the oauth2 library
	if err := m.server.HandleAuthorizeRequest(c.Writer, c.Request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// extractAuthForm checks the given OAuthAuthorize form, and stores
// the values in the form into the session.
func extractAuthForm(s sessions.Session, form *model.OAuthAuthorize) error {
	// these fields are *required* so check 'em
	if form.ResponseType == "" || form.ClientID == "" || form.RedirectURI == "" {
		return errors.New("missing one of: response_type, client_id or redirect_uri")
	}

	// set default scope to read
	if form.Scope == "" {
		form.Scope = "read"
	}

	// save these values from the form so we can use them elsewhere in the session
	s.Set(sessionForceLogin, form.ForceLogin)
	s.Set(sessionResponseType, form.ResponseType)
	s.Set(sessionClientID, form.ClientID)
	s.Set(sessionRedirectURI, form.RedirectURI)
	s.Set(sessionScope, form.Scope)
	return s.Save()
}
