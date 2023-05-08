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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func (f *formatter) fromPlain(
	ctx context.Context,
	ptParser parser.Parser,
	pmf gtsmodel.ParseMentionFunc,
	authorID string,
	statusID string,
	plain string,
) *FormatResult {
	result := &FormatResult{
		Mentions: []*gtsmodel.Mention{},
		Tags:     []*gtsmodel.Tag{},
		Emojis:   []*gtsmodel.Emoji{},
	}

	// Parse markdown into html, using custom renderer
	// to add hashtag/mention links and emoji images.
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithHardWraps(),
		),
		goldmark.WithParser(ptParser), // use parser we were passed
		goldmark.WithExtensions(
			&customRenderer{f, ctx, pmf, authorID, statusID, false, result},
			extension.Linkify, // turns URLs into links
		),
	)

	var htmlContentBytes bytes.Buffer
	if err := md.Convert([]byte(plain), &htmlContentBytes); err != nil {
		log.Errorf(ctx, "error formatting plaintext to HTML: %s", err)
	}
	result.HTML = htmlContentBytes.String()

	// Clean anything dangerous out of resulting HTML.
	result.HTML = SanitizeHTML(result.HTML)

	// Shrink ray!
	var err error
	if result.HTML, err = m.String("text/html", result.HTML); err != nil {
		log.Errorf(ctx, "error minifying HTML: %s", err)
	}

	return result
}

func (f *formatter) FromPlain(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, plain string) *FormatResult {
	ptParser := parser.NewParser(
		parser.WithBlockParsers(
			util.Prioritized(newPlaintextParser(), 500),
		),
	)

	return f.fromPlain(ctx, ptParser, pmf, authorID, statusID, plain)
}

func (f *formatter) FromPlainNoParagraph(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, plain string) *FormatResult {
	ptParser := parser.NewParser(
		parser.WithBlockParsers(
			// Initialize block parser that doesn't wrap in <p> tags.
			util.Prioritized(newPlaintextParserNoParagraph(), 500),
		),
	)

	return f.fromPlain(ctx, ptParser, pmf, authorID, statusID, plain)
}
