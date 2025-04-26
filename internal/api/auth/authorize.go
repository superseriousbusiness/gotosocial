// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package auth

import (
	"net/http"
	"net/url"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthorizeGETHandler should be served as
// GET at https://example.org/oauth/authorize.
//
// The idea here is to present an authorization
// page to the user, informing them of the scopes
// the application is requesting, with a button
// that they have to click to give it permission.
func (m *Module) AuthorizeGETHandler(c *gin.Context) {
	if _, err := apiutil.NegotiateAccept(c, apiutil.HTMLAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	s := sessions.Default(c)

	// UserID will be set in the session by
	// AuthorizePOSTHandler if the caller has
	// already gone through the auth flow.
	//
	// If it's not set, then we don't yet know
	// yet who the user is, so send them to the
	// sign in page first.
	if userID, ok := s.Get(sessionUserID).(string); !ok || userID == "" {
		m.redirectAuthFormToSignIn(c)
		return
	}

	user := m.mustUserFromSession(c, s)
	if user == nil {
		// Error already
		// written.
		return
	}

	// If the user is unconfirmed, waiting approval,
	// or suspended, redirect to an appropriate help page.
	if !m.validateUser(c, user) {
		// Already
		// redirected.
		return
	}

	// Everything looks OK.
	// Start preparing to render the html template.
	instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	redirectURI := m.mustStringFromSession(c, s, sessionRedirectURI)
	if redirectURI == "" {
		// Error already
		// written.
		return
	}

	scope := m.mustStringFromSession(c, s, sessionScope)
	if scope == "" {
		// Error already
		// written.
		return
	}

	app := m.mustAppFromSession(c, s)
	if app == nil {
		// Error already
		// written.
		return
	}

	// The authorize template will display a form
	// to the user where they can see some info
	// about the app that's trying to authorize,
	// and the scope of the request. They can then
	// approve it if it looks OK to them, which
	// will POST to the AuthorizePOSTHandler.
	apiutil.TemplateWebPage(c, apiutil.WebPage{
		Template: "authorize.tmpl",
		Instance: instance,
		Extra: map[string]any{
			"appname":    app.Name,
			"appwebsite": app.Website,
			"redirect":   redirectURI,
			"scope":      scope,
			"user":       user.Account.Username,
		},
	})
}

// AuthorizePOSTHandler should be served as
// POST at https://example.org/oauth/authorize.
//
// At this point we assume that the user has signed
// in and permitted the app to act on their behalf.
// We should proceed with the authentication flow
// and generate an oauth code at the redirect URI.
func (m *Module) AuthorizePOSTHandler(c *gin.Context) {

	// We need to use the session cookie to
	// recreate the original form submitted
	// to the authorizeGEThandler so that it
	// can be validated by the oauth2 library.
	s := sessions.Default(c)

	responseType := m.mustStringFromSession(c, s, sessionResponseType)
	if responseType == "" {
		// Error already
		// written.
		return
	}

	clientID := m.mustStringFromSession(c, s, sessionClientID)
	if clientID == "" {
		// Error already
		// written.
		return
	}

	redirectURI := m.mustStringFromSession(c, s, sessionRedirectURI)
	if redirectURI == "" {
		// Error already
		// written.
		return
	}

	scope := m.mustStringFromSession(c, s, sessionScope)
	if scope == "" {
		// Error already
		// written.
		return
	}

	user := m.mustUserFromSession(c, s)
	if user == nil {
		// Error already
		// written.
		return
	}

	// Force login is optional with default of "false".
	forceLogin, ok := s.Get(sessionForceLogin).(string)
	if !ok || forceLogin == "" {
		forceLogin = "false"
	}

	// Client state is optional with default of "".
	var clientState string
	if cs, ok := s.Get(sessionClientState).(string); ok {
		clientState = cs
	}

	// If the user is unconfirmed, waiting approval,
	// or suspended, redirect to an appropriate help page.
	if !m.validateUser(c, user) {
		// Already
		// redirected.
		return
	}

	// If we're redirecting to our OOB token handler,
	// we need to keep the session around so the OOB
	// handler can extract values from it. Otherwise,
	// we're going to be redirecting somewhere else
	// so we can safely clear the session now.
	if redirectURI != oauth.OOBURI {
		m.mustClearSession(s)
	}

	// Set values on the request form so that
	// they're picked up by the oauth server.
	c.Request.Form = url.Values{
		sessionResponseType: {responseType},
		sessionClientID:     {clientID},
		sessionRedirectURI:  {redirectURI},
		sessionScope:        {scope},
		sessionUserID:       {user.ID},
		sessionForceLogin:   {forceLogin},
	}

	if clientState != "" {
		// If client state was submitted,
		// set it on the form so it can be
		// fed back to the client via a query
		// param at the eventual redirect URL.
		c.Request.Form.Set("state", clientState)
	}

	// If OAuthHandleAuthorizeRequest is successful,
	// it'll handle any further redirects for us,
	// but we do still need to handle any errors.
	errWithCode := m.processor.OAuthHandleAuthorizeRequest(c.Writer, c.Request)
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
	}
}

// redirectAuthFormToSignIn binds an OAuthAuthorize form,
// presumed to be set as url query params, stores the values
// into the session, and redirects the user to the sign in page.
func (m *Module) redirectAuthFormToSignIn(c *gin.Context) {
	s := sessions.Default(c)

	form := &apimodel.OAuthAuthorize{}
	if err := c.ShouldBind(form); err != nil {
		m.clearSessionWithBadRequest(c, s, err, err.Error(), oauth.HelpfulAdvice)
		return
	}

	// If scope isn't set default to read.
	//
	// Else massage submitted scope(s) from
	// '+'-separated to space-separated.
	if form.Scope == "" {
		form.Scope = "read"
	} else {
		form.Scope = strings.ReplaceAll(form.Scope, "+", " ")
	}

	// Save these values from the form so we
	// can use them elsewhere in the session.
	s.Set(sessionForceLogin, form.ForceLogin)
	s.Set(sessionResponseType, form.ResponseType)
	s.Set(sessionClientID, form.ClientID)
	s.Set(sessionRedirectURI, form.RedirectURI)
	s.Set(sessionScope, form.Scope)
	s.Set(sessionInternalState, uuid.NewString())
	s.Set(sessionClientState, form.State)

	m.mustSaveSession(s)
	c.Redirect(http.StatusSeeOther, "/auth"+AuthSignInPath)
}

// validateUser checks if the given user:
//
//  1. Has a confirmed email address.
//  2. Has been approved.
//  3. Is not disabled or suspended.
//
// If all looks OK, returns true. Otherwise,
// redirects to a help page and returns false.
func (m *Module) validateUser(
	c *gin.Context,
	user *gtsmodel.User,
) bool {
	switch {
	case user.ConfirmedAt.IsZero():
		// User email not confirmed yet.
		const redirectTo = "/auth" + AuthCheckYourEmailPath
		c.Redirect(http.StatusSeeOther, redirectTo)
		return false

	case !*user.Approved:
		// User signup not approved yet.
		const redirectTo = "/auth" + AuthWaitForApprovalPath
		c.Redirect(http.StatusSeeOther, redirectTo)
		return false

	case *user.Disabled || !user.Account.SuspendedAt.IsZero():
		// User disabled or suspended.
		const redirectTo = "/auth" + AuthAccountDisabledPath
		c.Redirect(http.StatusSeeOther, redirectTo)
		return false

	default:
		// All good.
		return true
	}
}
