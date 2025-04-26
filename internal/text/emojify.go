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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
)

// EmojifyWeb replaces emoji shortcodes like `:example:` in the given HTML
// fragment with `<img>` tags suitable for rendering on the web frontend.
func EmojifyWeb(emojis []apimodel.Emoji, html template.HTML) template.HTML {
	out := emojify(
		emojis,
		string(html),
		func(url, staticURL, code string, buf *bytes.Buffer) {
			// Open a picture tag so we
			// can present multiple options.
			buf.WriteString(`<picture>`)

			// Static version.
			buf.WriteString(`<source `)
			{
				buf.WriteString(`class="emoji" `)
				buf.WriteString(`srcset="` + staticURL + `" `)
				buf.WriteString(`type="image/png" `)
				// Show this version when user
				// doesn't want an animated emoji.
				buf.WriteString(`media="(prefers-reduced-motion: reduce)" `)
				// Limit size to avoid showing
				// huge emojis when unstyled.
				buf.WriteString(`width="25" height="25" `)
			}
			buf.WriteString(`/>`)

			// Original image source.
			buf.WriteString(`<img `)
			{
				buf.WriteString(`class="emoji" `)
				buf.WriteString(`src="` + url + `" `)
				buf.WriteString(`title=":` + code + `:" `)
				buf.WriteString(`alt=":` + code + `:" `)
				// Lazy load emojis when
				// they scroll into view.
				buf.WriteString(`loading="lazy" `)
				// Limit size to avoid showing
				// huge emojis when unstyled.
				buf.WriteString(`width="25" height="25" `)
			}
			buf.WriteString(`/>`)

			// Close the picture tag.
			buf.WriteString(`</picture>`)
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
		func(url, staticURL, code string, buf *bytes.Buffer) {
			// Original image source.
			buf.WriteString(`<img `)
			{
				buf.WriteString(`src="` + url + `" `)
				buf.WriteString(`title=":` + code + `:" `)
				buf.WriteString(`alt=":` + code + `:" `)
				// Limit size to avoid showing
				// huge emojis in RSS readers.
				buf.WriteString(`width="25" height="25" `)
			}
			buf.WriteString(`/>`)
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
	write func(url, staticURL, code string, buf *bytes.Buffer),
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
			staticURL := html.EscapeString(emoji.StaticURL)
			code := html.EscapeString(emoji.Shortcode)

			// Write emoji repr to buffer.
			write(url, staticURL, code, buf)
			return buf.String()
		},
	)
}
