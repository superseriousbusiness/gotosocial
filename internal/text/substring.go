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
	"github.com/rivo/uniseg"
)

// FirstNBytesByWords produces a prefix substring of up to n bytes from a given string, respecting Unicode grapheme and
// word boundaries. The substring may be empty, and may include leading or trailing whitespace.
func FirstNBytesByWords(s string, n int) string {
	substringEnd := 0

	graphemes := uniseg.NewGraphemes(s)
	for graphemes.Next() {

		if !graphemes.IsWordBoundary() {
			continue
		}

		_, end := graphemes.Positions()
		if end > n {
			break
		}

		substringEnd = end
	}

	return s[0:substringEnd]
}
