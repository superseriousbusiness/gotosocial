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

// Filter represents a user-defined filter for determining which statuses should not be shown to the user.
// If whole_word is true , client app should do:
// Define ‘word constituent character’ for your app. In the official implementation, it’s [A-Za-z0-9_] in JavaScript, and [[:word:]] in Ruby.
// Ruby uses the POSIX character class (Letter | Mark | Decimal_Number | Connector_Punctuation).
// If the phrase starts with a word character, and if the previous character before matched range is a word character, its matched range should be treated to not match.
// If the phrase ends with a word character, and if the next character after matched range is a word character, its matched range should be treated to not match.
// Please check app/javascript/mastodon/selectors/index.js and app/lib/feed_manager.rb in the Mastodon source code for more details.
type Filter struct {
	// The ID of the filter in the database.
	ID string `json:"id"`
	// The text to be filtered.
	Phrase string `json:"text"`
	// The contexts in which the filter should be applied.
	// Array of String (Enumerable anyOf)
	// 	home = home timeline and lists
	// 	notifications = notifications timeline
	// 	public = public timelines
	// 	thread = expanded thread of a detailed status
	Context []string `json:"context"`
	// Should the filter consider word boundaries?
	WholeWord bool `json:"whole_word"`
	// When the filter should no longer be applied (ISO 8601 Datetime), or null if the filter does not expire
	ExpiresAt string `json:"expires_at,omitempty"`
	// Should matching entities in home and notifications be dropped by the server?
	Irreversible bool `json:"irreversible"`
}
