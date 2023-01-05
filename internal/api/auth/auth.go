/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
)

const (
	/*
		paths prefixed with 'auth'
	*/

	// AuthSignInPath is the API path for users to sign in through
	AuthSignInPath = "/sign_in"
	// AuthCheckYourEmailPath users land here after registering a new account, instructs them to confirm their email
	AuthCheckYourEmailPath = "/check_your_email"
	// AuthWaitForApprovalPath users land here after confirming their email
	// but before an admin approves their account (if such is required)
	AuthWaitForApprovalPath = "/wait_for_approval"
	// AuthAccountDisabledPath users land here when their account is suspended by an admin
	AuthAccountDisabledPath = "/account_disabled"
	// AuthCallbackPath is the API path for receiving callback tokens from external OIDC providers
	AuthCallbackPath = "/callback"

	/*
		paths prefixed with 'oauth'
	*/

	// OauthTokenPath is the API path to use for granting token requests to users with valid credentials
	OauthTokenPath = "/token" // #nosec G101 else we get a hardcoded credentials warning
	// OauthAuthorizePath is the API path for authorization requests (eg., authorize this app to act on my behalf as a user)
	OauthAuthorizePath = "/authorize"
	// OauthFinalizePath is the API path for completing user registration with additional user details
	OauthFinalizePath = "/finalize"
	// OauthOobTokenPath is the path for serving an html representation of an oob token page.
	OauthOobTokenPath = "/oob" // #nosec G101 else we get a hardcoded credentials warning

	/*
		params / session keys
	*/

	callbackStateParam   = "state"
	callbackCodeParam    = "code"
	sessionUserID        = "userid"
	sessionClientID      = "client_id"
	sessionRedirectURI   = "redirect_uri"
	sessionForceLogin    = "force_login"
	sessionResponseType  = "response_type"
	sessionScope         = "scope"
	sessionInternalState = "internal_state"
	sessionClientState   = "client_state"
	sessionClaims        = "claims"
	sessionAppID         = "app_id"
)

type Module struct {
	db        db.DB
	processor processing.Processor
	idp       oidc.IDP
}

// New returns an Auth module which provides both 'oauth' and 'auth' endpoints.
//
// It is safe to pass a nil idp if oidc is disabled.
func New(db db.DB, processor processing.Processor, idp oidc.IDP) *Module {
	return &Module{
		db:        db,
		processor: processor,
		idp:       idp,
	}
}

// RouteAuth routes all paths that should have an 'auth' prefix
func (m *Module) RouteAuth(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodGet, AuthSignInPath, m.SignInGETHandler)
	attachHandler(http.MethodPost, AuthSignInPath, m.SignInPOSTHandler)
	attachHandler(http.MethodGet, AuthCallbackPath, m.CallbackGETHandler)
}

// RouteOauth routes all paths that should have an 'oauth' prefix
func (m *Module) RouteOauth(attachHandler func(method string, path string, f ...gin.HandlerFunc) gin.IRoutes) {
	attachHandler(http.MethodPost, OauthTokenPath, m.TokenPOSTHandler)
	attachHandler(http.MethodGet, OauthAuthorizePath, m.AuthorizeGETHandler)
	attachHandler(http.MethodPost, OauthAuthorizePath, m.AuthorizePOSTHandler)
	attachHandler(http.MethodPost, OauthFinalizePath, m.FinalizePOSTHandler)
	attachHandler(http.MethodGet, OauthOobTokenPath, m.OobHandler)
}

func (m *Module) clearSession(s sessions.Session) {
	s.Clear()
	if err := s.Save(); err != nil {
		panic(err)
	}
}
