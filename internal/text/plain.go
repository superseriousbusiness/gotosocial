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
	"context"
	gohtml "html"
	"strings"

	"codeberg.org/gruf/go-byteutil"
	"github.com/k3a/html2text"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// FromPlain fulfils FormatFunc by parsing
// the given plaintext input into a FormatResult.
func (f *Formatter) FromPlain(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	// Initialize standard block parser
	// that wraps result in <p> tags.
	plainTextParser := parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(newPlaintextParser(), 500),
		),
	)

	return f.fromPlain(
		ctx,
		plainTextParser,
		false, // basic = false
		parseMention,
		authorID,
		statusID,
		input,
	)
}

// FromPlainNoParagraph fulfils FormatFunc by parsing
// the given plaintext input into a FormatResult.
//
// Unlike FromPlain, it will not wrap the resulting
// HTML in <p> tags, making it useful for parsing
// short fragments of text that oughtn't be formally
// wrapped as a paragraph.
func (f *Formatter) FromPlainNoParagraph(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	// Initialize block parser that
	// doesn't wrap result in <p> tags.
	plainTextParser := parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(newPlaintextParserNoParagraph(), 500),
		),
	)

	return f.fromPlain(
		ctx,
		plainTextParser,
		false, // basic = false
		parseMention,
		authorID,
		statusID,
		input,
	)
}

// FromPlainBasic fulfils FormatFunc by parsing
// the given plaintext input into a FormatResult.
//
// Unlike FromPlain, it will only parse emojis with
// the custom renderer, leaving aside mentions and tags.
//
// Resulting HTML will also NOT be wrapped in <p> tags.
func (f *Formatter) FromPlainBasic(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	// Initialize block parser that
	// doesn't wrap result in <p> tags.
	plainTextParser := parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(newPlaintextParserNoParagraph(), 500),
		),
	)

	return f.fromPlain(
		ctx,
		plainTextParser,
		true, // basic = true
		parseMention,
		authorID,
		statusID,
		input,
	)
}

// fromPlain parses the given input text
// using the given plainTextParser, and
// returns the result.
func (f *Formatter) fromPlain(
	ctx context.Context,
	plainTextParser parser.Parser,
	basic bool,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	result := new(FormatResult)

	// Instantiate goldmark parser for
	// plaintext, using custom renderer
	// to add hashtag/mention links.
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithHardWraps(),
		),
		// Use whichever plaintext
		// parser we were passed.
		goldmark.WithParser(plainTextParser),
		goldmark.WithExtensions(
			&customRenderer{
				ctx,
				f.db,
				parseMention,
				authorID,
				statusID,
				// If basic, pass
				// emojiOnly = true.
				basic,
				result,
			},
			// Turns URLs into links.
			extension.NewLinkify(
				extension.WithLinkifyURLRegexp(regexes.URLLike),
			),
		),
	)

	// Convert input string to bytes
	// without performing any allocs.
	bInput := byteutil.S2B(input)

	// Parse input into HTML.
	var htmlBytes bytes.Buffer
	if err := md.Convert(
		bInput,
		&htmlBytes,
	); err != nil {
		log.Errorf(ctx, "error formatting plaintext input to HTML: %s", err)
	}

	// Clean and shrink HTML.
	result.HTML = byteutil.B2S(htmlBytes.Bytes())
	result.HTML = SanitizeHTML(result.HTML)
	result.HTML = MinifyHTML(result.HTML)

	return result
}

// ParseHTMLToPlain parses the given HTML string, then
// outputs it to equivalent plaintext while trying to
// keep as much of the smenantic intent of the input
// HTML as possible, ie., titles are placed on separate
// lines, `<br>`s are converted to newlines, text inside
// `<strong>` and `<em>` tags is retained, but without
// emphasis, `<a>` links are unnested and the URL they
// link to is placed in angle brackets next to them,
// lists are replaced with newline-separated indented
// items, etc.
//
// This function is useful when you need to filter on
// HTML and want to avoid catching tags in the filter,
// or when you want to serve something in a plaintext
// format that may contain HTML tags (eg., CWs).
func ParseHTMLToPlain(html string) string {
	plain := html2text.HTML2TextWithOptions(
		html,
		html2text.WithLinksInnerText(),
		html2text.WithUnixLineBreaks(),
		html2text.WithListSupport(),
	)
	return strings.TrimSpace(plain)
}

// StripHTMLFromText runs text through strict sanitization
// to completely remove any HTML from the input without
// trying to preserve the semantic intent of any HTML tags.
//
// This is useful in cases where the input was not allowed
// to contain HTML at all, and the output isn't either.
func StripHTMLFromText(text string) string {
	// Unescape first to catch any tricky critters.
	content := gohtml.UnescapeString(text)

	// Remove all detected HTML.
	content = strict.Sanitize(content)

	// Unescape again to return plaintext.
	content = gohtml.UnescapeString(content)
	return strings.TrimSpace(content)
}
