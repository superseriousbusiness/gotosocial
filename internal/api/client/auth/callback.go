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
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/validate"
)

// CallbackGETHandler parses a token from an external auth provider.
func (m *Module) CallbackGETHandler(c *gin.Context) {
	s := sessions.Default(c)

	// check the query vs session state parameter to mitigate csrf
	// https://auth0.com/docs/secure/attack-protection/state-parameters

	state := c.Query(callbackStateParam)
	if state == "" {
		m.clearSession(s)
		err := fmt.Errorf("%s parameter not found on callback query", callbackStateParam)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	savedStateI := s.Get(sessionState)
	savedState, ok := savedStateI.(string)
	if !ok {
		m.clearSession(s)
		err := fmt.Errorf("key %s was not found in session", sessionState)
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, err.Error()), m.processor.InstanceGet)
		return
	}

	if state != savedState {
		m.clearSession(s)
		err := errors.New("mismatch between query state and session state")
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
		api.ErrorHandler(c, gtserror.NewErrorBadRequest(err, helpfulAdvice), m.processor.InstanceGet)
		return
	}

	app := &gtsmodel.Application{}
	if err := m.db.GetWhere(c.Request.Context(), []db.Where{{Key: sessionClientID, Value: clientID}}, app); err != nil {
		m.clearSession(s)
		safe := fmt.Sprintf("application for %s %s could not be retrieved", sessionClientID, clientID)
		var errWithCode gtserror.WithCode
		if err == db.ErrNoEntries {
			errWithCode = gtserror.NewErrorBadRequest(err, safe, helpfulAdvice)
		} else {
			errWithCode = gtserror.NewErrorInternalError(err, safe, helpfulAdvice)
		}
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
		return
	}

	user, errWithCode := m.parseUserFromClaims(c.Request.Context(), claims, net.IP(c.ClientIP()), app.ID)
	if errWithCode != nil {
		m.clearSession(s)
		api.ErrorHandler(c, errWithCode, m.processor.InstanceGet)
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

func (m *Module) parseUserFromClaims(ctx context.Context, claims *oidc.Claims, ip net.IP, appID string) (*gtsmodel.User, gtserror.WithCode) {
	if claims.Email == "" {
		err := errors.New("no email returned in claims")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// see if we already have a user for this email address
	// if so, we don't need to continue + create one
	user := &gtsmodel.User{}
	err := m.db.GetWhere(ctx, []db.Where{{Key: "email", Value: claims.Email}}, user)
	if err == nil {
		return user, nil
	}

	if err != db.ErrNoEntries {
		err := fmt.Errorf("error checking database for email %s: %s", claims.Email, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// maybe we have an unconfirmed user
	err = m.db.GetWhere(ctx, []db.Where{{Key: "unconfirmed_email", Value: claims.Email}}, user)
	if err == nil {
		err := fmt.Errorf("user with email address %s is unconfirmed", claims.Email)
		return nil, gtserror.NewErrorForbidden(err, err.Error())
	}

	if err != db.ErrNoEntries {
		err := fmt.Errorf("error checking database for email %s: %s", claims.Email, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// we don't have a confirmed or unconfirmed user with the claimed email address
	// however, because we trust the OIDC provider, we should now create a user + account with the provided claims

	// check if the email address is available for use; if it's not there's nothing we can so
	emailAvailable, err := m.db.IsEmailAvailable(ctx, claims.Email)
	if err != nil {
		return nil, gtserror.NewErrorBadRequest(err)
	}
	if !emailAvailable {
		return nil, gtserror.NewErrorConflict(fmt.Errorf("email address %s is not available", claims.Email))
	}

	// now we need a username
	var username string

	// make sure claims.Name is defined since we'll be using that for the username
	if claims.Name == "" {
		err := errors.New("no name returned in claims")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// check if we can just use claims.Name as-is
	if err = validate.Username(claims.Name); err == nil {
		// the name we have on the claims is already a valid username
		username = claims.Name
	} else {
		// not a valid username so we have to fiddle with it to try to make it valid
		// first trim leading and trailing whitespace
		trimmed := strings.TrimSpace(claims.Name)
		// underscore any spaces in the middle of the name
		underscored := strings.ReplaceAll(trimmed, " ", "_")
		// lowercase the whole thing
		lower := strings.ToLower(underscored)
		// see if this is valid....
		if err := validate.Username(lower); err != nil {
			err := fmt.Errorf("couldn't parse a valid username from claims.Name value of %s: %s", claims.Name, err)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}
		// we managed to get a valid username
		username = lower
	}

	var iString string
	var found bool
	// if the username isn't available we need to iterate on it until we find one that is
	// we should try to do this in a predictable way so we just keep iterating i by one and trying
	// the username with that number on the end
	//
	// note that for the first iteration, iString is still "" when the check is made, so our first choice
	// is still the raw username with no integer stuck on the end
	for i := 1; !found; i++ {
		usernameAvailable, err := m.db.IsUsernameAvailable(ctx, username+iString)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		if usernameAvailable {
			// no error so we've found a username that works
			found = true
			username += iString
			continue
		}
		iString = strconv.Itoa(i)
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
	user, err = m.db.NewSignup(ctx, username, "", requireApproval, claims.Email, password, ip, "", appID, emailVerified, admin)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return user, nil
}
