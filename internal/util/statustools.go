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

func IsPlausiblyInHashtag(r rune) bool {
	// Marks are allowed during parsing, prior to normalization, but not after,
	// since they may be combined into letters during normalization.
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsMark(r)
}

func IsPermittedInHashtag(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// Decides where to break before or after a #hashtag or @mention
func IsMentionOrHashtagBoundary(r rune) bool {
	return unicode.IsSpace(r) || unicode.IsPunct(r)
}
