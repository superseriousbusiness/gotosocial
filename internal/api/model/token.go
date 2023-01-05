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

// Token represents an OAuth token used for authenticating with the GoToSocial API and performing actions.
//
// swagger:model oauthToken
type Token struct {
	// Access token used for authorization.
	AccessToken string `json:"access_token"`
	// OAuth token type. Will always be 'Bearer'.
	// example: bearer
	TokenType string `json:"token_type"`
	// OAuth scopes granted by this token, space-separated.
	// example: read write admin
	Scope string `json:"scope"`
	// When the OAuth token was generated (UNIX timestamp seconds).
	// example: 1627644520
	CreatedAt int64 `json:"created_at"`
}
