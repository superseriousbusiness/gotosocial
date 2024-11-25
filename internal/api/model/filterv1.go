// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package model

// FilterV1 represents a user-defined filter for determining which statuses should not be shown to the user.
// Note that v1 filters are mapped to v2 filters and v2 filter keywords internally.
// If whole_word is true, client app should do:
// Define ‘word constituent character’ for your app. In the official implementation, it’s [A-Za-z0-9_] in JavaScript, and [[:word:]] in Ruby.
// Ruby uses the POSIX character class (Letter | Mark | Decimal_Number | Connector_Punctuation).
// If the phrase starts with a word character, and if the previous character before matched range is a word character, its matched range should be treated to not match.
// If the phrase ends with a word character, and if the next character after matched range is a word character, its matched range should be treated to not match.
// Please check app/javascript/mastodon/selectors/index.js and app/lib/feed_manager.rb in the Mastodon source code for more details.
//
// swagger:model filterV1
//
// ---
// tags:
// - filters
type FilterV1 struct {
	// The ID of the filter in the database.
	ID string `json:"id"`
	// The text to be filtered.
	//
	// Example: fnord
	Phrase string `json:"phrase"`
	// The contexts in which the filter should be applied.
	//
	// Minimum items: 1
	// Unique: true
	// Enum:
	//	- home
	//	- notifications
	//	- public
	//	- thread
	//	- account
	// Example: ["home", "public"]
	Context []FilterContext `json:"context"`
	// Should the filter consider word boundaries?
	//
	// Example: true
	WholeWord bool `json:"whole_word"`
	// Should matching entities be removed from the user's timelines/views, instead of hidden?
	//
	// Example: false
	Irreversible bool `json:"irreversible"`
	// When the filter should no longer be applied. Null if the filter does not expire.
	//
	// Example: 2024-02-01T02:57:49Z
	ExpiresAt *string `json:"expires_at"`
}

// FilterCreateUpdateRequestV1 captures params for creating or updating a v1 filter.
//
// swagger:ignore
type FilterCreateUpdateRequestV1 struct {
	// The text to be filtered.
	//
	// Required: true
	// Maximum length: 40
	// Example: fnord
	Phrase string `form:"phrase" json:"phrase" xml:"phrase"`
	// The contexts in which the filter should be applied.
	//
	// Required: true
	// Minimum length: 1
	// Unique: true
	// Enum: home,notifications,public,thread,account
	// Example: ["home", "public"]
	Context []FilterContext `form:"context[]" json:"context" xml:"context"`
	// Should matching entities be removed from the user's timelines/views, instead of hidden?
	//
	// Example: false
	Irreversible *bool `form:"irreversible" json:"irreversible" xml:"irreversible"`
	// Should the filter consider word boundaries?
	//
	// Example: true
	WholeWord *bool `form:"whole_word" json:"whole_word" xml:"whole_word"`
	// Number of seconds from now that the filter should expire. If omitted, filter never expires.
	ExpiresIn *int `json:"-" form:"expires_in" xml:"expires_in"`
	// Number of seconds from now that the filter should expire. If omitted, filter never expires.
	//
	// Example: 86400
	ExpiresInI Nullable[any] `json:"expires_in"`
}
