/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

// Application represents a mastodon-api Application, as defined here: https://docs.joinmastodon.org/entities/application/.
// Primarily, application is used for allowing apps like Tusky etc to connect to Mastodon on behalf of a user.
// See https://docs.joinmastodon.org/methods/apps/
type Application struct {
	// The application ID in the db
	ID string `json:"id,omitempty"`
	// The name of your application.
	Name string `json:"name"`
	// The website associated with your application (url)
	Website string `json:"website,omitempty"`
	// Where the user should be redirected after authorization.
	RedirectURI string `json:"redirect_uri,omitempty"`
	// ClientID to use when obtaining an oauth token for this application (ie., in client_id parameter of https://docs.joinmastodon.org/methods/apps/)
	ClientID string `json:"client_id,omitempty"`
	// Client secret to use when obtaining an auth token for this application (ie., in client_secret parameter of https://docs.joinmastodon.org/methods/apps/)
	ClientSecret string `json:"client_secret,omitempty"`
	// Used for Push Streaming API. Returned with POST /api/v1/apps. Equivalent to https://docs.joinmastodon.org/entities/pushsubscription/#server_key
	VapidKey string `json:"vapid_key,omitempty"`
}

// ApplicationCreateRequest represents a POST request to https://example.org/api/v1/apps.
// See here: https://docs.joinmastodon.org/methods/apps/
// And here: https://docs.joinmastodon.org/client/token/
type ApplicationCreateRequest struct {
	// A name for your application
	ClientName string `form:"client_name" binding:"required"`
	// Where the user should be redirected after authorization.
	// To display the authorization code to the user instead of redirecting
	// to a web page, use urn:ietf:wg:oauth:2.0:oob in this parameter.
	RedirectURIs string `form:"redirect_uris" binding:"required"`
	// Space separated list of scopes. If none is provided, defaults to read.
	Scopes string `form:"scopes"`
	// A URL to the homepage of your app
	Website string `form:"website"`
}
