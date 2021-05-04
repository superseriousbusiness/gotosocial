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

// Activity represents the mastodon-api Activity type. See here: https://docs.joinmastodon.org/entities/activity/
type Activity struct {
	// Midnight at the first day of the week. (UNIX Timestamp as string)
	Week string `json:"week"`
	// Statuses created since the week began. Integer cast to string.
	Statuses string `json:"statuses"`
	// User logins since the week began. Integer cast as string.
	Logins string `json:"logins"`
	// User registrations since the week began. Integer cast as string.
	Registrations string `json:"registrations"`
}
