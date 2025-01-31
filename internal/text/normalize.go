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

import (
	"strings"

	"golang.org/x/text/unicode/norm"
)

const (
	maximumHashtagLength = 100
)

// NormalizeHashtag normalizes the given hashtag text by
// removing the initial '#' symbol, and then decomposing
// and canonically recomposing chars + combining diacritics
// in the text to single unicode characters, following
// Normalization Form C (https://unicode.org/reports/tr15/).
//
// Finally, it will do a check on the normalized string to
// ensure that it's below maximumHashtagLength chars, and
// contains only letters, numbers, and underscores (and not
// *JUST* underscores).
//
// If all this passes, returned bool will be true.
func NormalizeHashtag(text string) (string, bool) {
	// This normalization is specifically to avoid cases
	// where visually-identical hashtags are stored with
	// different unicode representations (e.g. with combining
	// diacritics). It allows a tasteful number of combining
	// diacritics to be used, as long as they can be combined
	// with parent characters to form regular letter symbols.
	normalized := norm.NFC.String(strings.TrimPrefix(text, "#"))

	// Validate normalized result.
	var (
		atLeastOneRequiredChar = false
		onlyPermittedChars     = true
		lengthOK               = true
	)

	for i, r := range normalized {
		if !isPermittedIfNotEntireHashtag(r) {
			// This isn't an underscore, mark, etc,
			// so the hashtag contains at least one
			atLeastOneRequiredChar = true
		}

		if i >= maximumHashtagLength {
			lengthOK = false
			break
		}

		if !isPermittedInHashtag(r) {
			onlyPermittedChars = false
			break
		}
	}

	return normalized, lengthOK && onlyPermittedChars && atLeastOneRequiredChar
}
