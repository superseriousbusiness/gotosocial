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

package mastotypes

// Token represents an OAuth token used for authenticating with the API and performing actions.. See https://docs.joinmastodon.org/entities/token/
type Token struct {
	// An OAuth token to be used for authorization.
	AccessToken string `json:"access_token"`
	// The OAuth token type. Mastodon uses Bearer tokens.
	TokenType string `json:"token_type"`
	// The OAuth scopes granted by this token, space-separated.
	Scope string `json:"scope"`
	// When the token was generated. (UNIX timestamp seconds)
	CreatedAt int64 `json:"created_at"`
}
