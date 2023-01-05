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

// History represents daily usage history of a hashtag.
type History struct {
	// UNIX timestamp on midnight of the given day (string cast from integer).
	Day string `json:"day"`
	// The counted usage of the tag within that day (string cast from integer).
	Uses string `json:"uses"`
	// The total of accounts using the tag within that day (string cast from integer).
	Accounts string `json:"accounts"`
}
