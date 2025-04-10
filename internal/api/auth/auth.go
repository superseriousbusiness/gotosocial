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

	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

const (
	/*
		paths prefixed with 'auth'
	*/

	AuthSignInPath          = "/sign_in"
	Auth2FAPath             = "/2fa"
	AuthCheckYourEmailPath  = "/check_your_email"
	AuthWaitForApprovalPath = "/wait_for_approval"
	AuthAccountDisabledPath = "/account_disabled"
	AuthCallbackPath        = "/callback"

	/*
		paths prefixed with 'oauth'
	*/

	OauthAuthorizePath = "/authorize"
	OauthFinalizePath  = "/finalize"
	OauthOOBTokenPath  = "/oob"   // #nosec G101 else we get a hardcoded credentials warning
	OauthTokenPath     = "/token" // #nosec G101 else we get a hardcoded credentials warning
	OauthRevokePath    = "/revoke"

	/*
		params / session keys
	*/

	callbackStateParam       = "state"
	callbackCodeParam        = "code"
	sessionUserID            = "userid"
	sessionUserIDAwaiting2FA = "userid_awaiting_2fa"
	sessionClientID          = "client_id"
	sessionRedirectURI       = "redirect_uri"
	sessionForceLogin        = "force_login"
	sessionResponseType      = "response_type"
	sessionScope             = "scope"
	sessionInternalState     = "internal_state"
	sessionClientState       = "client_state"
	sessionClaims            = "claims"
	sessionAppID             = "app_id"
)

type Module struct {
	state     *state.State
	processor *processing.Processor
	idp       oidc.IDP
}

// New returns an Auth module which provides
// both 'oauth' and 'auth' endpoints.
//
// It is safe to pass a nil idp if oidc is disabled.
func New(
	state *state.State,
	processor *processing.Processor,
	idp oidc.IDP,
) *Module {
	return &Module{
		state:     state,
		processor: processor,
		idp:       idp,
	}
}

// RouteAuth routes all paths that should have an 'auth' prefix
func (m *Module) RouteAuth(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, AuthSignInPath, m.SignInGETHandler)
	attachHandler(http.MethodPost, AuthSignInPath, m.SignInPOSTHandler)
	attachHandler(http.MethodGet, Auth2FAPath, m.TwoFactorCodeGETHandler)
	attachHandler(http.MethodPost, Auth2FAPath, m.TwoFactorCodePOSTHandler)
	attachHandler(http.MethodGet, AuthCallbackPath, m.CallbackGETHandler)
}

// RouteOAuth routes all paths that should have an 'oauth' prefix
func (m *Module) RouteOAuth(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodPost, OauthTokenPath, m.TokenPOSTHandler)
	attachHandler(http.MethodPost, OauthRevokePath, m.TokenRevokePOSTHandler)
	attachHandler(http.MethodGet, OauthAuthorizePath, m.AuthorizeGETHandler)
	attachHandler(http.MethodPost, OauthAuthorizePath, m.AuthorizePOSTHandler)
	attachHandler(http.MethodPost, OauthFinalizePath, m.FinalizePOSTHandler)
	attachHandler(http.MethodGet, OauthOOBTokenPath, m.OOBTokenGETHandler)
}
