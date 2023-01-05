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

// List represents a list of some users that the authenticated user follows.
type List struct {
	// The internal database ID of the list.
	ID string `json:"id"`
	// The user-defined title of the list.
	Title string `json:"title"`
	// followed = Show replies to any followed user
	//	list = Show replies to members of the list
	//	none = Show replies to no one
	RepliesPolicy string `json:"replies_policy"`
}
