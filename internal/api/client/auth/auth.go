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
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/api"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
	"github.com/superseriousbusiness/gotosocial/internal/oidc"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/router"
)

/* #nosec G101 */
const (
	// AuthSignInPath is the API path for users to sign in through
	AuthSignInPath = "/auth/sign_in"

	// CheckYourEmailPath users land here after registering a new account, instructs them to confirm thier email
	CheckYourEmailPath = "/check_your_email"

	// WaitForApprovalPath users land here after confirming thier email but before an admin approves thier account
	// (if such is required)
	WaitForApprovalPath = "/wait_for_approval"

	// AccountDisabledPath users land here when thier account is suspended by an admin
	AccountDisabledPath = "/account_disabled"

	// OauthTokenPath is the API path to use for granting token requests to users with valid credentials
	OauthTokenPath = "/oauth/token"

	// OauthAuthorizePath is the API path for authorization requests (eg., authorize this app to act on my behalf as a user)
	OauthAuthorizePath = "/oauth/authorize"

	// OauthFinalizePath is the API path for completing user registration with additional user details
	OauthFinalizePath = "/oauth/finalize"

	// CallbackPath is the API path for receiving callback tokens from external OIDC providers
	CallbackPath = oidc.CallbackPath

	callbackStateParam = "state"
	callbackCodeParam  = "code"

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

// Module implements the ClientAPIModule interface for
type Module struct {
	db        db.DB
	idp       oidc.IDP
	processor processing.Processor
}

// New returns a new auth module
func New(db db.DB, idp oidc.IDP, processor processing.Processor) api.ClientModule {
	return &Module{
		db:        db,
		idp:       idp,
		processor: processor,
	}
}

// Route satisfies the RESTAPIModule interface
func (m *Module) Route(s router.Router) error {
	s.AttachHandler(http.MethodGet, AuthSignInPath, m.SignInGETHandler)
	s.AttachHandler(http.MethodPost, AuthSignInPath, m.SignInPOSTHandler)

	s.AttachHandler(http.MethodPost, OauthTokenPath, m.TokenPOSTHandler)

	s.AttachHandler(http.MethodGet, OauthAuthorizePath, m.AuthorizeGETHandler)
	s.AttachHandler(http.MethodPost, OauthAuthorizePath, m.AuthorizePOSTHandler)

	s.AttachHandler(http.MethodGet, CallbackPath, m.CallbackGETHandler)
	s.AttachHandler(http.MethodPost, OauthFinalizePath, m.FinalizePOSTHandler)

	s.AttachHandler(http.MethodGet, oauth.OOBTokenPath, m.OobHandler)
	return nil
}
