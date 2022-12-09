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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// extraInfo wraps a form-submitted username and transmitted name
type extraInfo struct {
	Username string `form:"username"`
	Name     string `form:"name"` // note that this is only used for re-rendering the page in case of an error
}

// CallbackGETHandler parses a token from an external auth provider.
func (m *Module) CallbackGETHandler(c *gin.Context) {
	s := sessions.Default(c)

	// check the query vs session state parameter to mitigate csrf
	// https://auth0.com/docs/secure/attack-protection/state-parameters

	returnedInternalState := c.Query(callbackStateParam)
	if returnedInternalState == "" {
		m.clearSession(s)
		err := fmt.Errorf("%s parameter not found on callback query", callbackStateParam)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	savedInternalStateI := s.Get(sessionInternalState)
	savedInternalState, ok := savedInternalStateI.(string)
	if !ok {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionInternalState)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if returnedInternalState != savedInternalState {
		m.clearSession(s)
		err := errors.New("mismatch between callback state and saved state")
		api.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGet)
		return
	}

	// retrieve stored claims using code
	code := c.Query(callbackCodeParam)
	if code == "" {
		m.clearSession(s)
		err := fmt.Errorf("%s parameter not found on callback query", callbackCodeParam)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	claims, errWithCode := m.idp.HandleCallback(c.Request.Context(), code)
	if errWithCode != nil {
		m.clearSession(s)
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	// We can use the client_id on the session to retrieve
	// info about the app associated with the client_id
	clientID, ok := s.Get(sessionClientID).(string)
	if !ok || clientID == "" {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionClientID)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
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
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	user, errWithCode := m.fetchUserForClaims(c.Request.Context(), claims, net.IP(c.ClientIP()), app.ID)
	if errWithCode != nil {
		m.clearSession(s)
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}
	if user == nil {
		// no user exists yet - let's ask them for their preferred username
		instance, errWithCode := m.processor.InstanceGet(c.Request.Context(), config.GetHost())
		if errWithCode != nil {
			api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
			return
		}

		// store the claims in the session - that way we know the user is authenticated when processing the form later
		s.Set(sessionClaims, claims)
		s.Set(sessionAppID, app.ID)
		if err := s.Save(); err != nil {
			m.clearSession(s)
			api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
			return
		}
		c.HTML(http.StatusOK, "finalize.tmpl", gin.H{
			"instance":          instance,
			"name":              claims.Name,
			"preferredUsername": claims.PreferredUsername,
		})
		return
	}
	s.Set(sessionUserID, user.ID)
	if err := s.Save(); err != nil {
		m.clearSession(s)
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}
	c.Redirect(http.StatusFound, OauthAuthorizePath)
}

// FinalizePOSTHandler registers the user after additional data has been provided
func (m *Module) FinalizePOSTHandler(c *gin.Context) {
	s := sessions.Default(c)

	form := &extraInfo{}
	if err := c.ShouldBind(form); err != nil {
		m.clearSession(s)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}

	// since we have multiple possible validation error, `validationError` is a shorthand for rendering them
	validationError := func(err error) {
		instance, errWithCode := m.processor.InstanceGet(c.Request.Context(), config.GetHost())
		if errWithCode != nil {
			api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
			return
		}
		c.HTML(http.StatusOK, "finalize.tmpl", gin.H{
			"instance":          instance,
			"name":              form.Name,
			"preferredUsername": form.Username,
			"error":             err,
		})
	}

	// check if the username conforms to the spec
	if err := validate.Username(form.Username); err != nil {
		validationError(err)
		return
	}

	// see if the username is still available
	usernameAvailable, err := m.db.IsUsernameAvailable(c.Request.Context(), form.Username)
	if err != nil {
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}
	if !usernameAvailable {
		validationError(fmt.Errorf("Username %s is already taken", form.Username))
		return
	}

	// retrieve the information previously set by the oidc logic
	appID, ok := s.Get(sessionAppID).(string)
	if !ok {
		err := fmt.Errorf("key %s was not found in session", sessionAppID)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}

	// retrieve the claims returned by the IDP. Having this present means that we previously already verified these claims
	claims, ok := s.Get(sessionClaims).(*oidc.Claims)
	if !ok {
		err := fmt.Errorf("key %s was not found in session", sessionClaims)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGet)
		return
	}

	// we're now ready to actually create the user
	user, errWithCode := m.createUserFromOIDC(c.Request.Context(), claims, form, net.IP(c.ClientIP()), appID)
	if errWithCode != nil {
		m.clearSession(s)
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}
	s.Delete(sessionClaims)
	s.Delete(sessionAppID)
	s.Set(sessionUserID, user.ID)
	if err := s.Save(); err != nil {
		m.clearSession(s)
		api.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGet)
		return
	}
	c.Redirect(http.StatusFound, OauthAuthorizePath)
}

func (m *Module) fetchUserForClaims(ctx context.Context, claims *oidc.Claims, ip net.IP, appID string) (*gtsmodel.User, gtserror.WithCode) {
	if claims.Sub == "" {
		err := errors.New("no sub claim found - is your provider OIDC compliant?")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
	user, err := m.db.GetUserByExternalID(ctx, claims.Sub)
	if err == nil {
		return user, nil
	}
	if err != db.ErrNoEntries {
		err := fmt.Errorf("error checking database for externalID %s: %s", claims.Sub, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !config.GetOIDCLinkExisting() {
		return nil, nil
	}
	// fallback to email if we want to link existing users
	user, err = m.db.GetUserByEmailAddress(ctx, claims.Email)
	if err == db.ErrNoEntries {
		return nil, nil
	} else if err != nil {
		err := fmt.Errorf("error checking database for email %s: %s", claims.Email, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	// at this point we have found a matching user but still need to link the newly received external ID

	user.ExternalID = claims.Sub
	err = m.db.UpdateUser(ctx, user, "external_id")
	if err != nil {
		err := fmt.Errorf("error linking existing user %s: %s", claims.Email, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return user, nil
}

func (m *Module) createUserFromOIDC(ctx context.Context, claims *oidc.Claims, extraInfo *extraInfo, ip net.IP, appID string) (*gtsmodel.User, gtserror.WithCode) {
	// check if the email address is available for use; if it's not there's nothing we can so
	emailAvailable, err := m.db.IsEmailAvailable(ctx, claims.Email)
	if err != nil {
		return nil, gtserror.NewErrorBadRequest(err)
	}
	if !emailAvailable {
		help := "The email address given to us by your authentication provider already exists in our records and the server administrator has not enabled account migration"
		return nil, gtserror.NewErrorConflict(fmt.Errorf("email address %s is not available", claims.Email), help)
	}

	// check if the user is in any recognised admin groups
	var admin bool
	for _, g := range claims.Groups {
		if strings.EqualFold(g, "admin") || strings.EqualFold(g, "admins") {
			admin = true
		}
	}

	// We still need to set *a* password even if it's not a password the user will end up using, so set something random.
	// We'll just set two uuids on top of each other, which should be long + random enough to baffle any attempts to crack.
	//
	// If the user ever wants to log in using gts password rather than oidc flow, they'll have to request a password reset, which is fine
	password := uuid.NewString() + uuid.NewString()

	// Since this user is created via oidc, which has been set up by the admin, we can assume that the account is already
	// implicitly approved, and that the email address has already been verified: otherwise, we end up in situations where
	// the admin first approves the user in OIDC, and then has to approve them again in GoToSocial, which doesn't make sense.
	//
	// In other words, if a user logs in via OIDC, they should be able to use their account straight away.
	//
	// See: https://github.com/superseriousbusiness/gotosocial/issues/357
	requireApproval := false
	emailVerified := true

	// create the user! this will also create an account and store it in the database so we don't need to do that here
	user, err := m.db.NewSignup(ctx, extraInfo.Username, "", requireApproval, claims.Email, password, ip, "", appID, emailVerified, claims.Sub, admin)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return user, nil
}
