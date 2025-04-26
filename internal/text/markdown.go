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
	"regexp"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/regexes"
	"codeberg.org/gruf/go-byteutil"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
)

// FromMarkdown fulfils FormatFunc by parsing
// the given markdown input into a FormatResult.
//
// Inline (aka unsafe) HTML elements are allowed,
// as they should be sanitized afterwards anyway.
func (f *Formatter) FromMarkdown(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	return f.fromMarkdown(
		ctx,
		false, // basic = false
		parseMention,
		authorID,
		statusID,
		input,
	)
}

// FromMarkdownBasic fulfils FormatFunc by parsing
// the given markdown input into a FormatResult.
//
// Unlike FromMarkdown, it will only parse emojis with
// the custom renderer, leaving aside mentions and tags.
//
// Inline (aka unsafe) HTML elements are not allowed.
//
// If the result is a single paragraph,
// it will not be wrapped in <p> tags.
func (f *Formatter) FromMarkdownBasic(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	res := f.fromMarkdown(
		ctx,
		true, // basic = true
		parseMention,
		authorID,
		statusID,
		input,
	)

	res.HTML = unwrapParagraph(res.HTML)
	return res
}

// fromMarkdown parses the given input text either
// with or without emojis, and returns the result.
func (f *Formatter) fromMarkdown(
	ctx context.Context,
	basic bool,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	var (
		result = new(FormatResult)
		opts   []renderer.Option
	)

	if basic {
		// Don't allow raw HTML tags,
		// markdown syntax only.
		opts = []renderer.Option{
			html.WithXHTML(),
			html.WithHardWraps(),
		}
	} else {
		opts = []renderer.Option{
			html.WithXHTML(),
			html.WithHardWraps(),

			// Allow raw HTML tags, we
			// sanitize at the end anyway.
			html.WithUnsafe(),
		}
	}

	// Instantiate goldmark parser for
	// markdown, using custom renderer
	// to add hashtag/mention links.
	md := goldmark.New(
		goldmark.WithRendererOptions(
			opts...,
		),
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
			extension.Strikethrough,
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
		log.Errorf(ctx, "error formatting markdown input to HTML: %s", err)
	}

	// Clean and shrink HTML.
	result.HTML = byteutil.B2S(htmlBytes.Bytes())
	result.HTML = SanitizeHTML(result.HTML)
	result.HTML = MinifyHTML(result.HTML)

	return result
}

var parasRegexp = regexp.MustCompile(`</?p>`)

// unwrapParagraph removes opening and closing paragraph tags
// of input HTML, if input html is a single paragraph only.
func unwrapParagraph(html string) string {
	if !strings.HasPrefix(html, "<p>") {
		return html
	}

	if !strings.HasSuffix(html, "</p>") {
		return html
	}

	// Make a substring excluding the
	// opening and closing paragraph tags.
	sub := html[3 : len(html)-4]

	// If there are still other paragraph tags left
	// inside the substring, return html unchanged.
	containsOtherParas := parasRegexp.MatchString(sub)
	if containsOtherParas {
		return html
	}

	// Return the substring.
	return sub
}
