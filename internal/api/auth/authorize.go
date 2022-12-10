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

package auth

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

// AuthorizeGETHandler should be served as GET at https://example.org/oauth/authorize
// The idea here is to present an oauth authorize page to the user, with a button
// that they have to click to accept.
func (m *Module) AuthorizeGETHandler(c *gin.Context) {
	s := sessions.Default(c)

	if _, err := apiutil.NegotiateAccept(c, apiutil.HTMLAcceptHeaders...); err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorNotAcceptable(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// UserID will be set in the session by AuthorizePOSTHandler if the caller has already gone through the authentication flow
	// If it's not set, then we don't know yet who the user is, so we need to redirect them to the sign in page.
	userID, ok := s.Get(sessionUserID).(string)
	if !ok || userID == "" {
		form := &apimodel.OAuthAuthorize{}
		if err := c.ShouldBind(form); err != nil {
			m.clearSession(s)
			apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
			return
		}

		if errWithCode := saveAuthFormToSession(s, form); errWithCode != nil {
			m.clearSession(s)
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
			return
		}

		c.Redirect(http.StatusSeeOther, "/auth"+AuthSignInPath)
		return
	}

	// use session information to validate app, user, and account for this request
	clientID, ok := s.Get(sessionClientID).(string)
	if !ok || clientID == "" {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionClientID)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}

	app := &gtsmodel.Application{}
	if err := m.db.GetWhere(c.Request.Context(), []db.Where{{Key: sessionClientID, Value: clientID}}, app); err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("application for %s %s could not be retrieved", sessionClientID, clientID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	user, err := m.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("user with id %s could not be retrieved", userID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	acct, err := m.db.GetAccountByID(c.Request.Context(), user.AccountID)
	if err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("account with id %s could not be retrieved", user.AccountID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if ensureUserIsAuthorizedOrRedirect(c, user, acct) {
		return
	}

	// Finally we should also get the redirect and scope of this particular request, as stored in the session.
	redirect, ok := s.Get(sessionRedirectURI).(string)
	if !ok || redirect == "" {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionRedirectURI)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}

	scope, ok := s.Get(sessionScope).(string)
	if !ok || scope == "" {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionScope)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}

	instance, errWithCode := m.processor.InstanceGet(c.Request.Context(), config.GetHost())
	if errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	// the authorize template will display a form to the user where they can get some information
	// about the app that's trying to authorize, and the scope of the request.
	// They can then approve it if it looks OK to them, which will POST to the AuthorizePOSTHandler
	c.HTML(http.StatusOK, "authorize.tmpl", gin.H{
		"appname":    app.Name,
		"appwebsite": app.Website,
		"redirect":   redirect,
		"scope":      scope,
		"user":       acct.Username,
		"instance":   instance,
	})
}

// AuthorizePOSTHandler should be served as POST at https://example.org/oauth/authorize
// At this point we assume that the user has A) logged in and B) accepted that the app should act for them,
// so we should proceed with the authentication flow and generate an oauth token for them if we can.
func (m *Module) AuthorizePOSTHandler(c *gin.Context) {
	s := sessions.Default(c)

	// We need to retrieve the original form submitted to the authorizeGEThandler, and
	// recreate it on the request so that it can be used further by the oauth2 library.
	errs := []string{}

	forceLogin, ok := s.Get(sessionForceLogin).(string)
	if !ok {
		forceLogin = "false"
	}

	responseType, ok := s.Get(sessionResponseType).(string)
	if !ok || responseType == "" {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionResponseType))
	}

	clientID, ok := s.Get(sessionClientID).(string)
	if !ok || clientID == "" {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionClientID))
	}

	redirectURI, ok := s.Get(sessionRedirectURI).(string)
	if !ok || redirectURI == "" {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionRedirectURI))
	}

	scope, ok := s.Get(sessionScope).(string)
	if !ok {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionScope))
	}

	var clientState string
	if s, ok := s.Get(sessionClientState).(string); ok {
		clientState = s
	}

	userID, ok := s.Get(sessionUserID).(string)
	if !ok {
		errs = append(errs, fmt.Sprintf("key %s was not found in session", sessionUserID))
	}

	if len(errs) != 0 {
		errs = append(errs, oauth.HelpfulAdvice)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(errors.New("one or more missing keys on session during AuthorizePOSTHandler"), errs...), m.processor.InstanceGet)
		return
	}

	user, err := m.db.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("user with id %s could not be retrieved", userID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	acct, err := m.db.GetAccountByID(c.Request.Context(), user.AccountID)
	if err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("account with id %s could not be retrieved", user.AccountID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	if ensureUserIsAuthorizedOrRedirect(c, user, acct) {
		return
	}

	if redirectURI != oauth.OOBURI {
		// we're done with the session now, so just clear it out
		m.clearSession(s)
	}

	// we have to set the values on the request form
	// so that they're picked up by the oauth server
	c.Request.Form = url.Values{
		sessionForceLogin:   {forceLogin},
		sessionResponseType: {responseType},
		sessionClientID:     {clientID},
		sessionRedirectURI:  {redirectURI},
		sessionScope:        {scope},
		sessionUserID:       {userID},
	}

	if clientState != "" {
		c.Request.Form.Set("state", clientState)
	}

	if errWithCode := m.processor.OAuthHandleAuthorizeRequest(c.Writer, c.Request); errWithCode != nil {
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
	}
}

// saveAuthFormToSession checks the given OAuthAuthorize form,
// and stores the values in the form into the session.
func saveAuthFormToSession(s sessions.Session, form *apimodel.OAuthAuthorize) gtserror.WithCode {
	if form == nil {
		err := errors.New("OAuthAuthorize form was nil")
		return gtserror.NewErrorBadRequest(err, err.Error(), oauth.HelpfulAdvice)
	}

	if form.ResponseType == "" {
		err := errors.New("field response_type was not set on OAuthAuthorize form")
		return gtserror.NewErrorBadRequest(err, err.Error(), oauth.HelpfulAdvice)
	}

	if form.ClientID == "" {
		err := errors.New("field client_id was not set on OAuthAuthorize form")
		return gtserror.NewErrorBadRequest(err, err.Error(), oauth.HelpfulAdvice)
	}

	if form.RedirectURI == "" {
		err := errors.New("field redirect_uri was not set on OAuthAuthorize form")
		return gtserror.NewErrorBadRequest(err, err.Error(), oauth.HelpfulAdvice)
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
	s.Set(sessionInternalState, uuid.NewString())
	s.Set(sessionClientState, form.State)

	if err := s.Save(); err != nil {
		err := fmt.Errorf("error saving form values onto session: %s", err)
		return gtserror.NewErrorInternalError(err, oauth.HelpfulAdvice)
	}

	return nil
}

func ensureUserIsAuthorizedOrRedirect(ctx *gin.Context, user *gtsmodel.User, account *gtsmodel.Account) (redirected bool) {
	if user.ConfirmedAt.IsZero() {
		ctx.Redirect(http.StatusSeeOther, "/auth"+AuthCheckYourEmailPath)
		redirected = true
		return
	}

	if !*user.Approved {
		ctx.Redirect(http.StatusSeeOther, "/auth"+AuthWaitForApprovalPath)
		redirected = true
		return
	}

	if *user.Disabled || !account.SuspendedAt.IsZero() {
		ctx.Redirect(http.StatusSeeOther, "/auth"+AuthAccountDisabledPath)
		redirected = true
		return
	}

	return
}
