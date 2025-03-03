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

package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"codeberg.org/superseriousbusiness/oauth2/v4"
	oautherr "codeberg.org/superseriousbusiness/oauth2/v4/errors"
	"codeberg.org/superseriousbusiness/oauth2/v4/manage"
	"codeberg.org/superseriousbusiness/oauth2/v4/server"
	apiutil "github.com/superseriousbusiness/gotosocial/internal/api/util"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// GetClientScopeHandler returns a handler for testing scope on a TokenGenerateRequest.
func GetClientScopeHandler(ctx context.Context, state *state.State) server.ClientScopeHandler {
	return func(tgr *oauth2.TokenGenerateRequest) (allowed bool, err error) {
		application, err := state.DB.GetApplicationByClientID(
			gtscontext.SetBarebones(ctx),
			tgr.ClientID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf(ctx, "database error getting application: %v", err)
			return false, err
		}

		if application == nil {
			err := gtserror.Newf("no application found with client id %s", tgr.ClientID)
			return false, err
		}

		// Normalize scope.
		if strings.TrimSpace(tgr.Scope) == "" {
			tgr.Scope = "read"
		}

		// Make sure requested scopes are all
		// within scopes permitted by application.
		hasScopes := strings.Split(application.Scopes, " ")
		wantsScopes := strings.Split(tgr.Scope, " ")
		for _, wantsScope := range wantsScopes {
			thisOK := slices.ContainsFunc(
				hasScopes,
				func(hasScope string) bool {
					has := apiutil.Scope(hasScope)
					wants := apiutil.Scope(wantsScope)
					return has.Permits(wants)
				},
			)

			if !thisOK {
				// Requested unpermitted
				// scope for this app.
				return false, nil
			}
		}

		// All OK.
		return true, nil
	}
}

func GetValidateURIHandler(ctx context.Context) manage.ValidateURIHandler {
	return func(hasRedirects string, wantsRedirect string) error {
		// Normalize the wantsRedirect URI
		// string by parsing + reserializing.
		wantsRedirectURI, err := url.Parse(wantsRedirect)
		if err != nil {
			return err
		}
		wantsRedirect = wantsRedirectURI.String()

		// Redirect URIs are given to us as
		// a list of URIs, newline-separated.
		//
		// They're already normalized on input so
		// we don't need to parse + reserialize them.
		//
		// Ensure that one of them matches.
		if slices.ContainsFunc(
			strings.Split(hasRedirects, "\n"),
			func(hasRedirect string) bool {
				// Want an exact match.
				// See: https://www.oauth.com/oauth2-servers/redirect-uris/redirect-uri-validation/
				return wantsRedirect == hasRedirect
			},
		) {
			return nil
		}

		return oautherr.ErrInvalidRedirectURI
	}
}

func GetAuthorizeScopeHandler() server.AuthorizeScopeHandler {
	return func(_ http.ResponseWriter, r *http.Request) (string, error) {
		// Use provided scope or
		// fall back to default "read".
		scope := r.FormValue("scope")
		if strings.TrimSpace(scope) == "" {
			scope = "read"
		}
		return scope, nil
	}
}

func GetInternalErrorHandler(ctx context.Context) server.InternalErrorHandler {
	return func(err error) *oautherr.Response {
		log.Errorf(ctx, "internal oauth error: %v", err)
		return nil
	}
}

func GetResponseErrorHandler(ctx context.Context) server.ResponseErrorHandler {
	return func(re *oautherr.Response) {
		log.Errorf(ctx, "internal response error: %v", re.Error)
	}
}

func GetUserAuthorizationHandler() server.UserAuthorizationHandler {
	return func(w http.ResponseWriter, r *http.Request) (string, error) {
		userID := r.FormValue("userid")
		if userID == "" {
			return "", errors.New("userid was empty")
		}
		return userID, nil
	}
}
