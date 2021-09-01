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

import "time"

// Token is a translation of the gotosocial token with the ExpiresIn fields replaced with ExpiresAt.
//
// Explanation for this: gotosocial assumes an in-memory or file database of some kind, where a time-to-live parameter (TTL) can be defined,
// and tokens with expired TTLs are automatically removed. Since some databases don't have that feature, it's easier to set an expiry time and
// then periodically sweep out tokens when that time has passed.
//
// Note that this struct does *not* satisfy the token interface shown here: https://github.com/superseriousbusiness/oauth2/blob/master/model.go#L22
// and implemented here: https://github.com/superseriousbusiness/oauth2/blob/master/models/token.go.
// As such, manual translation is always required between Token and the gotosocial *model.Token. The helper functions oauthTokenToPGToken
// and pgTokenToOauthToken can be used for that.
type Token struct {
	ID                  string `validate:"ulid" bun:"type:CHAR(26),pk,nullzero,notnull"`
	ClientID            string
	UserID              string
	RedirectURI         string
	Scope               string
	Code                string `bun:"default:'',pk"`
	CodeChallenge       string
	CodeChallengeMethod string
	CodeCreateAt        time.Time `bun:",nullzero"`
	CodeExpiresAt       time.Time `bun:",nullzero"`
	Access              string    `bun:"default:'',pk"`
	AccessCreateAt      time.Time `bun:",nullzero"`
	AccessExpiresAt     time.Time `bun:",nullzero"`
	Refresh             string    `bun:"default:'',pk"`
	RefreshCreateAt     time.Time `bun:",nullzero"`
	RefreshExpiresAt    time.Time `bun:",nullzero"`
}
