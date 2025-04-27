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
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"

	apiutil "code.superseriousbusiness.org/gotosocial/internal/api/util"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/oauth"
	"code.superseriousbusiness.org/gotosocial/internal/oidc"
	"code.superseriousbusiness.org/gotosocial/internal/validate"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// extraInfo wraps a form-submitted username and transmitted name
type extraInfo struct {
	Username string `form:"username"`
	Name     string `form:"name"` // note that this is only used for re-rendering the page in case of an error
}

// CallbackGETHandler parses a token from an external auth provider.
func (m *Module) CallbackGETHandler(c *gin.Context) {
	if !config.GetOIDCEnabled() {
		err := errors.New("oidc is not enabled for this server")
		apiutil.ErrorHandler(c, gtserror.NewErrorNotFound(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	s := sessions.Default(c)

	// check the query vs session state parameter to mitigate csrf
	// https://auth0.com/docs/secure/attack-protection/state-parameters

	returnedInternalState := c.Query(callbackStateParam)
	if returnedInternalState == "" {
		m.mustClearSession(s)
		err := fmt.Errorf("%s parameter not found on callback query", callbackStateParam)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	savedInternalStateI := s.Get(sessionInternalState)
	savedInternalState, ok := savedInternalStateI.(string)
	if !ok {
		m.mustClearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionInternalState)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	if returnedInternalState != savedInternalState {
		m.mustClearSession(s)
		err := errors.New("mismatch between callback state and saved state")
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	// retrieve stored claims using code
	code := c.Query(callbackCodeParam)
	if code == "" {
		m.mustClearSession(s)
		err := fmt.Errorf("%s parameter not found on callback query", callbackCodeParam)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	claims, errWithCode := m.idp.HandleCallback(c.Request.Context(), code)
	if errWithCode != nil {
		m.mustClearSession(s)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	// We can use the client_id on the session to retrieve
	// info about the app associated with the client_id
	clientID, ok := s.Get(sessionClientID).(string)
	if !ok || clientID == "" {
		m.mustClearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionClientID)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
		return
	}

	app, err := m.state.DB.GetApplicationByClientID(c.Request.Context(), clientID)
	if err != nil {
		m.mustClearSession(s)
		safe := fmt.Sprintf("application for %s %s could not be retrieved", sessionClientID, clientID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, oauth.HelpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, oauth.HelpfulAdvice)
		}
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}

	user, errWithCode := m.fetchUserForClaims(c.Request.Context(), claims, net.IP(c.ClientIP()), app.ID)
	if errWithCode != nil {
		m.mustClearSession(s)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}
	if user == nil {
		// no user exists yet - let's ask them for their preferred username
		instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

		// store the claims in the session - that way we know the user is authenticated when processing the form later
		s.Set(sessionClaims, claims)
		s.Set(sessionAppID, app.ID)
		if err := s.Save(); err != nil {
			m.mustClearSession(s)
			apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
			return
		}

		// Since we require lowercase usernames at this point, lowercase the one
		// from the claims and use this to autofill the form with a suggestion.
		//
		// Pending https://codeberg.org/superseriousbusiness/gotosocial/issues/1813
		suggestedUsername := strings.ToLower(claims.PreferredUsername)

		page := apiutil.WebPage{
			Template: "finalize.tmpl",
			Instance: instance,
			Extra: map[string]any{
				"name":              claims.Name,
				"suggestedUsername": suggestedUsername,
			},
		}

		apiutil.TemplateWebPage(c, page)
		return
	}

	// Check user permissions on login
	if !allowedGroup(claims.Groups) {
		err := fmt.Errorf("User groups %+v do not include an allowed group", claims.Groups)
		apiutil.ErrorHandler(c, gtserror.NewErrorUnauthorized(err, err.Error()), m.processor.InstanceGetV1)
		return
	}

	s.Set(sessionUserID, user.ID)
	if err := s.Save(); err != nil {
		m.mustClearSession(s)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}
	c.Redirect(http.StatusFound, "/oauth"+OauthAuthorizePath)
}

// FinalizePOSTHandler registers the user after additional data has been provided
func (m *Module) FinalizePOSTHandler(c *gin.Context) {
	s := sessions.Default(c)

	form := &extraInfo{}
	if err := c.ShouldBind(form); err != nil {
		m.mustClearSession(s)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
		return
	}

	// since we have multiple possible validation error, `validationError` is a shorthand for rendering them
	validationError := func(err error) {
		instance, errWithCode := m.processor.InstanceGetV1(c.Request.Context())
		if errWithCode != nil {
			apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
			return
		}

		page := apiutil.WebPage{
			Template: "finalize.tmpl",
			Instance: instance,
			Extra: map[string]any{
				"name":              form.Name,
				"preferredUsername": form.Username,
				"error":             err,
			},
		}

		apiutil.TemplateWebPage(c, page)
	}

	// check if the username conforms to the spec
	if err := validate.Username(form.Username); err != nil {
		validationError(err)
		return
	}

	// see if the username is still available
	usernameAvailable, err := m.state.DB.IsUsernameAvailable(c.Request.Context(), form.Username)
	if err != nil {
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
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
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
		return
	}

	// retrieve the claims returned by the IDP. Having this present means that we previously already verified these claims
	claims, ok := s.Get(sessionClaims).(*oidc.Claims)
	if !ok {
		err := fmt.Errorf("key %s was not found in session", sessionClaims)
		apiutil.ErrorHandler(c, gtserror.NewErrorBadRequest(err, oauth.HelpfulAdvice), m.processor.InstanceGetV1)
		return
	}

	// we're now ready to actually create the user
	user, errWithCode := m.createUserFromOIDC(c.Request.Context(), claims, form, net.IP(c.ClientIP()), appID)
	if errWithCode != nil {
		m.mustClearSession(s)
		apiutil.ErrorHandler(c, errWithCode, m.processor.InstanceGetV1)
		return
	}
	s.Delete(sessionClaims)
	s.Delete(sessionAppID)
	s.Set(sessionUserID, user.ID)
	if err := s.Save(); err != nil {
		m.mustClearSession(s)
		apiutil.ErrorHandler(c, gtserror.NewErrorInternalError(err), m.processor.InstanceGetV1)
		return
	}
	c.Redirect(http.StatusFound, "/oauth"+OauthAuthorizePath)
}

func (m *Module) fetchUserForClaims(ctx context.Context, claims *oidc.Claims, ip net.IP, appID string) (*gtsmodel.User, gtserror.WithCode) {
	if claims.Sub == "" {
		err := errors.New("no sub claim found - is your provider OIDC compliant?")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}
	user, err := m.state.DB.GetUserByExternalID(ctx, claims.Sub)
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
	user, err = m.state.DB.GetUserByEmailAddress(ctx, claims.Email)
	if err == db.ErrNoEntries {
		return nil, nil
	} else if err != nil {
		err := fmt.Errorf("error checking database for email %s: %s", claims.Email, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	// at this point we have found a matching user but still need to link the newly received external ID

	user.ExternalID = claims.Sub
	err = m.state.DB.UpdateUser(ctx, user, "external_id")
	if err != nil {
		err := fmt.Errorf("error linking existing user %s: %s", claims.Email, err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return user, nil
}

func (m *Module) createUserFromOIDC(ctx context.Context, claims *oidc.Claims, extraInfo *extraInfo, ip net.IP, appID string) (*gtsmodel.User, gtserror.WithCode) {
	// Check if the claimed email address is available for use.
	emailAvailable, err := m.state.DB.IsEmailAvailable(ctx, claims.Email)
	if err != nil {
		err := gtserror.Newf("db error checking email availability: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if !emailAvailable {
		const help = "The email address given to us by your authentication provider already exists in our records and the server administrator has not enabled account migration"
		err := gtserror.Newf("email address %s is not available", claims.Email)
		return nil, gtserror.NewErrorConflict(err, help)
	}

	if !allowedGroup(claims.Groups) {
		err := fmt.Errorf("User groups %+v do not include an allowed group", claims.Groups)
		return nil, gtserror.NewErrorUnauthorized(err, err.Error())
	}

	// We still need to set something as a password, even
	// if it's not a password the user will end up using.
	//
	// We'll just set two uuids on top of each other, which
	// should be long + random enough to baffle any attempts
	// to crack, and which is also, conveniently, 72 bytes,
	// which is the maximum length that bcrypt can handle.
	//
	// If the user ever wants to log in using a password
	// rather than oidc flow, they'll have to request a
	// password reset, which is fine.
	password := uuid.NewString() + uuid.NewString()

	// Since this user is created via OIDC, we can assume
	// that the account should be preapproved, and the email
	// address should be considered as verified already,
	// since the OIDC login was successful.
	//
	// If we don't assume this, we end up in a situation
	// where the admin first adds a user to OIDC, then has
	// to approve them again in GoToSocial when they log in
	// there for the first time, which doesn't make sense.
	//
	// In other words, if a user logs in via OIDC, they
	// should be able to use their account straight away.
	var (
		preApproved   = true
		emailVerified = true
	)

	// If one of the claimed groups corresponds to one of
	// the configured admin OIDC groups, create this user
	// as an admin.
	admin := adminGroup(claims.Groups)

	// Create the user! This will also create an account and
	// store it in the database, so we don't need to do that.
	user, err := m.state.DB.NewSignup(ctx, gtsmodel.NewSignup{
		Username:      extraInfo.Username,
		Email:         claims.Email,
		Password:      password,
		SignUpIP:      ip,
		AppID:         appID,
		ExternalID:    claims.Sub,
		PreApproved:   preApproved,
		EmailVerified: emailVerified,
		Admin:         admin,
	})
	if err != nil {
		err := gtserror.Newf("db error doing new signup: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return user, nil
}

// adminGroup returns true if one of the given OIDC
// groups is equal to at least one admin OIDC group.
func adminGroup(groups []string) bool {
	adminGroups := config.GetOIDCAdminGroups()
	for _, claimedGroup := range groups {
		if slices.ContainsFunc(adminGroups, func(allowedGroup string) bool {
			return strings.EqualFold(claimedGroup, allowedGroup)
		}) {
			return true
		}
	}

	// User is in no admin groups,
	// ∴ user is not an admin.
	return false
}

// allowedGroup returns true if one of the given OIDC
// groups is equal to at least one allowed OIDC group.
func allowedGroup(groups []string) bool {
	allowedGroups := config.GetOIDCAllowedGroups()
	if len(allowedGroups) == 0 {
		// If no groups are configured, allow access (for backwards compatibility)
		return true
	}
	for _, claimedGroup := range groups {
		if slices.ContainsFunc(allowedGroups, func(allowedGroup string) bool {
			return strings.EqualFold(claimedGroup, allowedGroup)
		}) {
			return true
		}
	}

	// User is in no allowed groups,
	// ∴ user is not allowed to log in
	return false
}
