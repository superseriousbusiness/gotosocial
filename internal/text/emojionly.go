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
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func (f *formatter) FromPlainEmojiOnly(ctx context.Context, pmf gtsmodel.ParseMentionFunc, authorID string, statusID string, plain string) *FormatResult {
	result := &FormatResult{
		Mentions: []*gtsmodel.Mention{},
		Tags:     []*gtsmodel.Tag{},
		Emojis:   []*gtsmodel.Emoji{},
	}
	// parse markdown text into html, using custom renderer to add hashtag/mention links
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithHardWraps(),
		),
		goldmark.WithParser(
			parser.NewParser(
				parser.WithBlockParsers(
					util.Prioritized(newPlaintextParser(), 500),
				),
			),
		),
		goldmark.WithExtensions(
			&customRenderer{f, ctx, pmf, authorID, statusID, true, result},
		),
	)

	var htmlContentBytes bytes.Buffer
	err := md.Convert([]byte(plain), &htmlContentBytes)
	if err != nil {
		log.Errorf("error formatting plaintext to HTML: %s", err)
	}
	result.HTML = htmlContentBytes.String()

	// clean anything dangerous out of the HTML
	result.HTML = SanitizeHTML(result.HTML)

	// shrink ray
	result.HTML = minifyHTML(result.HTML)

	return result
}
