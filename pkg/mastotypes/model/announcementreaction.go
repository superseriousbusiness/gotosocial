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

// AnnouncementReaction represents a user reaction to admin/moderator announcement. See here: https://docs.joinmastodon.org/entities/announcementreaction/
type AnnouncementReaction struct {
	// The emoji used for the reaction. Either a unicode emoji, or a custom emoji's shortcode.
	Name      string `json:"name"`
	// The total number of users who have added this reaction.
	Count     int    `json:"count"`
	// Whether the authorized user has added this reaction to the announcement.
	Me        bool   `json:"me"`
	// A link to the custom emoji.
	URL       string `json:"url,omitempty"`
	// A link to a non-animated version of the custom emoji.
	StaticURL string `json:"static_url,omitempty"`
}
