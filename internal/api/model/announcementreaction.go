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

// AnnouncementReaction models a user reaction to an announcement.
//
// swagger:model announcementReaction
type AnnouncementReaction struct {
	// The emoji used for the reaction. Either a unicode emoji, or a custom emoji's shortcode.
	// example: blobcat_uwu
	Name string `json:"name"`
	// The total number of users who have added this reaction.
	// example: 5
	Count int `json:"count"`
	// This reaction belongs to the account viewing it.
	Me bool `json:"me"`
	// Web link to the image of the custom emoji.
	// Empty for unicode emojis.
	// example: https://example.org/custom_emojis/original/blobcat_uwu.png
	URL string `json:"url,omitempty"`
	// Web link to a non-animated image of the custom emoji.
	// Empty for unicode emojis.
	// example: https://example.org/custom_emojis/statuc/blobcat_uwu.png
	StaticURL string `json:"static_url,omitempty"`
}
