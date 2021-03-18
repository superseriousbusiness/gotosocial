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

package gtsmodel

type Application struct {
	ID string `pg:"type:uuid,default:gen_random_uuid(),pk,notnull"`
	Name string `json:"name"`
	// The website associated with your application (url)
	Website string `json:"website,omitempty"`
	// Where the user should be redirected after authorization.
	RedirectURI string `json:"redirect_uri"`
	// ClientID to use when obtaining an oauth token for this application (ie., in client_id parameter of https://docs.joinmastodon.org/methods/apps/)
	ClientID string `json:"client_id"`
	// Client secret to use when obtaining an auth token for this application (ie., in client_secret parameter of https://docs.joinmastodon.org/methods/apps/)
	ClientSecret string `json:"client_secret"`
	// Used for Push Streaming API. Returned with POST /api/v1/apps. Equivalent to https://docs.joinmastodon.org/entities/pushsubscription/#server_key
	VapidKey string `json:"vapid_key"`
}
