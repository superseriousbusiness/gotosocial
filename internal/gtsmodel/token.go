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

package gtsmodel

import "time"

// Token is a translation of the gotosocial token with the ExpiresIn fields replaced with ExpiresAt.
type Token struct {
	ID                  string    `validate:"required,ulid" bun:"type:CHAR(26),pk,nullzero,notnull,unique"`        // id of this item in the database
	CreatedAt           time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item created
	UpdatedAt           time.Time `validate:"-" bun:"type:timestamptz,nullzero,notnull,default:current_timestamp"` // when was item last updated
	ClientID            string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero,notnull"`                  // ID of the client who owns this token
	UserID              string    `validate:"required,ulid" bun:"type:CHAR(26),nullzero"`                          // ID of the user who owns this token
	RedirectURI         string    `validate:"required,uri" bun:",nullzero,notnull"`                                // Oauth redirect URI for this token
	Scope               string    `validate:"required" bun:",notnull"`                                             // Oauth scope
	Code                string    `validate:"-" bun:",pk,nullzero,notnull,default:''"`                             // Code, if present
	CodeChallenge       string    `validate:"-" bun:",nullzero"`                                                   // Code challenge, if code present
	CodeChallengeMethod string    `validate:"-" bun:",nullzero"`                                                   // Code challenge method, if code present
	CodeCreateAt        time.Time `validate:"required_with=Code" bun:"type:timestamptz,nullzero"`                  // Code created time, if code present
	CodeExpiresAt       time.Time `validate:"-" bun:"type:timestamptz,nullzero"`                                   // Code expires at -- null means the code never expires
	Access              string    `validate:"-" bun:",pk,nullzero,notnull,default:''"`                             // User level access token, if present
	AccessCreateAt      time.Time `validate:"required_with=Access" bun:"type:timestamptz,nullzero"`                // User level access token created time, if access present
	AccessExpiresAt     time.Time `validate:"-" bun:"type:timestamptz,nullzero"`                                   // User level access token expires at -- null means the token never expires
	Refresh             string    `validate:"-" bun:",pk,nullzero,notnull,default:''"`                             // Refresh token, if present
	RefreshCreateAt     time.Time `validate:"required_with=Refresh" bun:"type:timestamptz,nullzero"`               // Refresh created at, if refresh present
	RefreshExpiresAt    time.Time `validate:"-" bun:"type:timestamptz,nullzero"`                                   // Refresh expires at -- null means the refresh token never expires
}
