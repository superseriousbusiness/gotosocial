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

// Mention represents a mention of another account.
type Mention struct {
	// The ID of the mentioned account.
	// example: 01FBYJHQWQZAVWFRK9PDYTKGMB
	ID string `json:"id"`
	// The username of the mentioned account.
	// example: some_user
	Username string `json:"username"`
	// The web URL of the mentioned account's profile.
	// example: https://example.org/@some_user
	URL string `json:"url"`
	// The account URI as discovered via webfinger.
	// Equal to username for local users, or username@domain for remote users.
	// example: some_user@example.org
	Acct string `json:"acct"`
}
