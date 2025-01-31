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

package text

import "unicode"

func isPermittedInHashtag(r rune) bool {
	return unicode.IsLetter(r) || isPermittedIfNotEntireHashtag(r)
}

// isPermittedIfNotEntireHashtag is true for characters that may be in a hashtag
// but are not allowed to be the only characters making up the hashtag.
func isPermittedIfNotEntireHashtag(r rune) bool {
	return unicode.IsNumber(r) || unicode.IsMark(r) || r == '_'
}

// isHashtagBoundary returns true if rune r
// is a recognized break character for before
// or after a #hashtag.
func isHashtagBoundary(r rune) bool {
	switch {

	// Zero width space.
	case r == '\u200B':
		return true

	// Zero width no-break space.
	case r == '\uFEFF':
		return true

	// Pipe character sometimes
	// used as workaround.
	case r == '|':
		return true

	// Standard Unicode white space.
	case unicode.IsSpace(r):
		return true

	// Non-underscore punctuation.
	case unicode.IsPunct(r) && r != '_':
		return true

	// Not recognized
	// hashtag boundary.
	default:
		return false
	}
}

// isMentionBoundary returns true if rune r
// is a recognized break character for before
// or after a @mention.
func isMentionBoundary(r rune) bool {
	return unicode.IsSpace(r) ||
		unicode.IsPunct(r)
}
