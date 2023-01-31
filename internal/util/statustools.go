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

package util

import (
	"unicode"
)

func IsPermittedInHashtag(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// Decides where to break before or after a hashtag.
func IsHashtagBoundary(r rune) bool {
	return r == '#' || // `###lol` should work
		unicode.IsSpace(r) || // All kinds of Unicode whitespace.
		unicode.IsControl(r) || // All kinds of control characters, like tab.
		// Most kinds of punctuation except "Pc" ("Punctuation, connecting", like `_`).
		// But `someurl/#fragment` should not match, neither should HTML entities like `&#35;`.
		('/' != r && '&' != r && !unicode.Is(unicode.Categories["Pc"], r) && unicode.IsPunct(r))
}
