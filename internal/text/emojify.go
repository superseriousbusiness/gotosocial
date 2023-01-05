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

package text

import (
	"bytes"
	"html"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// Emojify replaces shortcodes in `inputText` with the emoji in `emojis`.
//
// Callers should ensure that inputText and resulting text are escaped
// appropriately depending on what they're used for.
func Emojify(emojis []apimodel.Emoji, inputText string) string {
	emojisMap := make(map[string]apimodel.Emoji, len(emojis))

	for _, emoji := range emojis {
		shortcode := ":" + emoji.Shortcode + ":"
		emojisMap[shortcode] = emoji
	}

	return regexes.ReplaceAllStringFunc(
		regexes.EmojiFinder,
		inputText,
		func(shortcode string, buf *bytes.Buffer) string {
			// Look for emoji according to this shortcode
			emoji, ok := emojisMap[shortcode]
			if !ok {
				return shortcode
			}

			// Escape raw emoji content
			safeURL := html.EscapeString(emoji.URL)
			safeCode := html.EscapeString(emoji.Shortcode)

			// Write HTML emoji repr to buffer
			buf.WriteString(`<img src="`)
			buf.WriteString(safeURL)
			buf.WriteString(`" title=":`)
			buf.WriteString(safeCode)
			buf.WriteString(`:" alt=":`)
			buf.WriteString(safeCode)
			buf.WriteString(`:" class="emoji"/>`)

			return buf.String()
		},
	)
}
