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

// Mention represents the mastodon-api mention type, as documented here: https://docs.joinmastodon.org/entities/mention/
type Mention struct {
	// The account id of the mentioned user.
	ID string `json:"id"`
	// The username of the mentioned user.
	Username string `json:"username"`
	// The location of the mentioned user's profile.
	URL string `json:"url"`
	// The webfinger acct: URI of the mentioned user. Equivalent to username for local users, or username@domain for remote users.
	Acct string `json:"acct"`
}
