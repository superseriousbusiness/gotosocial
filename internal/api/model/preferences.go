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

// Preferences represents a user's preferences.
type Preferences struct {
	// Default visibility for new posts.
	// 	public = Public post
	// 	unlisted = Unlisted post
	// 	private = Followers-only post
	// 	direct = Direct post
	PostingDefaultVisibility string `json:"posting:default:visibility"`
	// Default sensitivity flag for new posts.
	PostingDefaultSensitive bool `json:"posting:default:sensitive"`
	// Default language for new posts. (ISO 639-1 language two-letter code), or null
	PostingDefaultLanguage string `json:"posting:default:language,omitempty"`
	// Whether media attachments should be automatically displayed or blurred/hidden.
	// 	default = Hide media marked as sensitive
	// 	show_all = Always show all media by default, regardless of sensitivity
	// 	hide_all = Always hide all media by default, regardless of sensitivity
	ReadingExpandMedia string `json:"reading:expand:media"`
	// Whether CWs should be expanded by default.
	ReadingExpandSpoilers bool `json:"reading:expand:spoilers"`
}
