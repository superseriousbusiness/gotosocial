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

// Application models an api application.
//
// swagger:model application
type Application struct {
	// The ID of the application.
	// example: 01FBVD42CQ3ZEEVMW180SBX03B
	ID string `json:"id,omitempty"`
	// The name of the application.
	// example: Tusky
	Name string `json:"name"`
	// The website associated with the application (url)
	// example: https://tusky.app
	Website string `json:"website,omitempty"`
	// Post-authorization redirect URI for the application (OAuth2).
	// example: https://example.org/callback?some=query
	RedirectURI string `json:"redirect_uri,omitempty"`
	// Client ID associated with this application.
	ClientID string `json:"client_id,omitempty"`
	// Client secret associated with this application.
	ClientSecret string `json:"client_secret,omitempty"`
	// Push API key for this application.
	VapidKey string `json:"vapid_key,omitempty"`
}

// ApplicationCreateRequest models app create parameters.
//
// swagger:parameters appCreate
type ApplicationCreateRequest struct {
	// The name of the application.
	//
	// in: formData
	// required: true
	ClientName string `form:"client_name" json:"client_name" xml:"client_name" binding:"required"`
	// Where the user should be redirected after authorization.
	//
	// To display the authorization code to the user instead of redirecting to a web page, use `urn:ietf:wg:oauth:2.0:oob` in this parameter.
	//
	// in: formData
	// required: true
	RedirectURIs string `form:"redirect_uris" json:"redirect_uris" xml:"redirect_uris" binding:"required"`
	// Space separated list of scopes.
	//
	// If no scopes are provided, defaults to `read`.
	//
	// in: formData
	Scopes string `form:"scopes" json:"scopes" xml:"scopes"`
	// A URL to the web page of the app (optional).
	//
	// in: formData
	Website string `form:"website" json:"website" xml:"website"`
}
