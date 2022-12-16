/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/tdewolff/minify/v2"
	minifyHtml "github.com/tdewolff/minify/v2/html"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	m *minify.M
)

func (f *formatter) FromMarkdown(ctx context.Context, markdownText string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag, emojis []*gtsmodel.Emoji) string {

	// Temporarily replace all found emoji shortcodes in the markdown text with
	// their ID so that they're not parsed as anything by the markdown parser -
	// this fixes cases where emojis with some underscores in them are parsed as
	// words with emphasis, eg `:_some_emoji:` becomes `:<em>some</em>emoji:`
	//
	// Since the IDs of the emojis are just uppercase letters + numbers they should
	// be safe to pass through the markdown parser without unexpected effects.
	for _, e := range emojis {
		markdownText = strings.ReplaceAll(markdownText, ":"+e.Shortcode+":", ":"+e.ID+":")
	}

	// parse markdown text into html, using custom renderer to add hashtag/mention links
	md := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithHardWraps(),
			html.WithUnsafe(), // allows raw HTML
		),
		goldmark.WithExtensions(
			&customRenderer{f, ctx, mentions, tags},
			extension.Linkify, // turns URLs into links
			extension.Strikethrough,
		),
	)

	var htmlContentBytes bytes.Buffer
	err := md.Convert([]byte(markdownText), &htmlContentBytes)
	if err != nil {
		log.Errorf("error rendering markdown to HTML: %s", err)
	}
	htmlContent := htmlContentBytes.String()

	// Replace emoji IDs in the parsed html content with their shortcodes again
	for _, e := range emojis {
		htmlContent = strings.ReplaceAll(htmlContent, ":"+e.ID+":", ":"+e.Shortcode+":")
	}

	// clean anything dangerous out of the html
	htmlContent = SanitizeHTML(htmlContent)

	if m == nil {
		m = minify.New()
		m.Add("text/html", &minifyHtml.Minifier{
			KeepEndTags: true,
			KeepQuotes:  true,
		})
	}

	minified, err := m.String("text/html", htmlContent)
	if err != nil {
		log.Errorf("error minifying markdown text: %s", err)
	}

	return minified
}
