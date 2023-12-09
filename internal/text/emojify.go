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
	"bytes"
	"html"
	"html/template"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// EmojifyWeb replaces emoji shortcodes like `:example:` in the given HTML
// fragment with `<img>` tags suitable for rendering on the web frontend.
func EmojifyWeb(emojis []apimodel.Emoji, html template.HTML) template.HTML {
	out := emojify(
		emojis,
		string(html),
		func(url, code string, buf *bytes.Buffer) {
			buf.WriteString(`<img src="`)
			buf.WriteString(url)
			buf.WriteString(`" title=":`)
			buf.WriteString(code)
			buf.WriteString(`:" alt=":`)
			buf.WriteString(code)
			buf.WriteString(`:" class="emoji" `)
			// Lazy load emojis when
			// they scroll into view.
			buf.WriteString(`loading="lazy"/>`)
		},
	)

	// If input was safe,
	// we can trust output.
	return template.HTML(out) // #nosec G203
}

// EmojifyRSS replaces emoji shortcodes like `:example:` in the given text
// fragment with `<img>` tags suitable for rendering as RSS content.
func EmojifyRSS(emojis []apimodel.Emoji, text string) string {
	return emojify(
		emojis,
		text,
		func(url, code string, buf *bytes.Buffer) {
			buf.WriteString(`<img src="`)
			buf.WriteString(url)
			buf.WriteString(`" title=":`)
			buf.WriteString(code)
			buf.WriteString(`:" alt=":`)
			buf.WriteString(code)
			buf.WriteString(`:" class="emoji" `)
			// Limit size to avoid showing
			// huge emojis in RSS readers.
			buf.WriteString(`width="50" height="50"/>`)
		},
	)
}

// Demojify replaces emoji shortcodes like `:example:` in the given text
// fragment with empty strings, essentially stripping them from the text.
// This is useful for text used in OG Meta headers.
func Demojify(text string) string {
	return regexes.EmojiFinder.ReplaceAllString(text, "")
}

func emojify(
	emojis []apimodel.Emoji,
	input string,
	write func(url, code string, buf *bytes.Buffer),
) string {
	// Build map of shortcodes. Normalize each
	// shortcode by readding closing colons.
	emojisMap := make(map[string]apimodel.Emoji, len(emojis))
	for _, emoji := range emojis {
		shortcode := ":" + emoji.Shortcode + ":"
		emojisMap[shortcode] = emoji
	}

	return regexes.ReplaceAllStringFunc(
		regexes.EmojiFinder,
		input,
		func(shortcode string, buf *bytes.Buffer) string {
			// Look for emoji with this shortcode.
			emoji, ok := emojisMap[shortcode]
			if !ok {
				return shortcode
			}

			// Escape raw emoji content.
			url := html.EscapeString(emoji.URL)
			code := html.EscapeString(emoji.Shortcode)

			// Write emoji repr to buffer.
			write(url, code, buf)
			return buf.String()
		},
	)
}
