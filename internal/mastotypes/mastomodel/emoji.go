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

import "mime/multipart"

// Emoji represents a custom emoji. See https://docs.joinmastodon.org/entities/emoji/
type Emoji struct {
	// REQUIRED

	// The name of the custom emoji.
	Shortcode string `json:"shortcode"`
	// A link to the custom emoji.
	URL string `json:"url"`
	// A link to a static copy of the custom emoji.
	StaticURL string `json:"static_url"`
	// Whether this Emoji should be visible in the picker or unlisted.
	VisibleInPicker bool `json:"visible_in_picker"`

	// OPTIONAL

	// Used for sorting custom emoji in the picker.
	Category string `json:"category,omitempty"`
}

// EmojiCreateRequest represents a request to create a custom emoji made through the admin API.
type EmojiCreateRequest struct {
	// Desired shortcode for the emoji, without surrounding colons. This must be unique for the domain.
	Shortcode string                `form:"shortcode" validation:"required"`
	// Image file to use for the emoji. Must be png or gif and no larger than 50kb.
	Image     *multipart.FileHeader `form:"image" validation:"required"`
}
