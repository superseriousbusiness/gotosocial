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

package text

import (
	"fmt"
	"strings"
)

// preformat contains some common logic for making a string ready for formatting, which should be used for all user-input text.
func preformat(in string) string {
	// do some preformatting of the text
	// 1. Trim all the whitespace
	s := strings.TrimSpace(in)
	return s
}

// postformat contains some common logic for html sanitization of text, wrapping elements, and trimming newlines and whitespace
func postformat(in string) string {
	// do some postformatting of the text
	// 1. sanitize html to remove any dodgy scripts or other disallowed elements
	s := SanitizeOutgoing(in)
	// 2. wrap the whole thing in a paragraph
	s = fmt.Sprintf(`<p>%s</p>`, s)
	// 3. remove any cheeky newlines
	s = strings.ReplaceAll(s, "\n", "")
	// 4. remove any whitespace added as a result of the formatting
	s = strings.TrimSpace(s)
	return s
}
