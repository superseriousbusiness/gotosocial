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

	"codeberg.org/gruf/go-byteutil"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// FromMarkdown fulfils FormatFunc by parsing
// the given markdown input into a FormatResult.
func (f *Formatter) FromMarkdown(
	ctx context.Context,
	parseMention gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	input string,
) *FormatResult {
	result := new(FormatResult)

	// Instantiate goldmark parser for
	// markdown, using custom renderer
	// to add hashtag/mention links.
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithHardWraps(),
			// Allows raw HTML. We sanitize
			// at the end so this is OK.
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			&customRenderer{
				ctx,
				f.db,
				parseMention,
				authorID,
				statusID,
				false, // emojiOnly = false.
				result,
			},
			// Turns URLs into links.
			extension.NewLinkify(
				extension.WithLinkifyURLRegexp(regexes.LinkScheme),
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
	result.HTML = SanitizeToHTML(result.HTML)
	result.HTML = MinifyHTML(result.HTML)

	return result
}
