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

// TokenInfo represents metadata about one user-level access token.
// The actual access token itself will never be sent via the API.
//
// swagger:model tokenInfo
type TokenInfo struct {
	// Database ID of this token.
	// example: 01JMW7QBAZYZ8T8H73PCEX12XG
	ID string `json:"id"`
	// When the token was created (ISO 8601 Datetime).
	// example: 2021-07-30T09:20:25+00:00
	CreatedAt string `json:"created_at"`
	// Approximate time (accurate to within an hour) when the token was last used (ISO 8601 Datetime).
	// Omitted if token has never been used, or it is not known when it was last used (eg., it was last used before tracking "last_used" became a thing).
	// example: 2021-07-30T09:20:25+00:00
	LastUsed string `json:"last_used,omitempty"`
	// OAuth scopes granted by the token, space-separated.
	// example: read write admin
	Scope string `json:"scope"`
	// Application used to create this token.
	Application *Application `json:"application"`
}
