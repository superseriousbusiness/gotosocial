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

package model

// OAuthAuthorize represents a request sent to https://example.org/oauth/authorize
type OAuthAuthorize struct {
	// Forces the user to re-login, which is necessary for authorizing with multiple accounts from the same instance.
	ForceLogin string `form:"force_login" json:"force_login"`
	// Should be set equal to `code`.
	ResponseType string `form:"response_type" json:"response_type"`
	// Client ID, obtained during app registration.
	ClientID string `form:"client_id" json:"client_id"`
	// Set a URI to redirect the user to.
	// If this parameter is set to urn:ietf:wg:oauth:2.0:oob then the authorization code will be shown instead.
	// Must match one of the redirect URIs declared during app registration.
	RedirectURI string `form:"redirect_uri" json:"redirect_uri"`
	// List of requested OAuth scopes, separated by spaces (or by pluses, if using query parameters).
	// Must be a subset of scopes declared during app registration. If not provided, defaults to read.
	Scope string `form:"scope" json:"scope"`
	// State is used by the application to store request-specific data and/or prevent CSRF attacks.
	// The authorization server must return the unmodified state value back to the application.
	// See https://www.oauth.com/oauth2-servers/authorization/the-authorization-request/
	State string `form:"state" json:"state"`
}
