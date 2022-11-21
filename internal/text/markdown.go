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
	"io"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

var (
	bfExtensions = blackfriday.NoIntraEmphasis | blackfriday.FencedCode | blackfriday.Autolink | blackfriday.Strikethrough | blackfriday.SpaceHeadings | blackfriday.HardLineBreak
	m            *minify.M
)

type renderer struct {
	f        *formatter
	ctx      context.Context
	mentions []*gtsmodel.Mention
	tags     []*gtsmodel.Tag
	blackfriday.HTMLRenderer
}

func (r *renderer) RenderNode(w io.Writer, node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
	if node.Type == blackfriday.Text {
		// call RenderNode to do the html escaping
		var buff bytes.Buffer
		status := r.HTMLRenderer.RenderNode(&buff, node, entering)

		html := buff.String()
		html = r.f.ReplaceTags(r.ctx, html, r.tags)
		html = r.f.ReplaceMentions(r.ctx, html, r.mentions)

		// we don't have much recourse if this fails
		if _, err := io.WriteString(w, html); err != nil {
			log.Errorf("error outputting markdown text: %s", err)
		}
		return status
	}
	return r.HTMLRenderer.RenderNode(w, node, entering)
}

func (f *formatter) FromMarkdown(ctx context.Context, markdownText string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag, emojis []*gtsmodel.Emoji) string {

	renderer := &renderer{
		f:        f,
		ctx:      ctx,
		mentions: mentions,
		tags:     tags,
		HTMLRenderer: *blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{
			// same as blackfriday.CommonHTMLFlags, but with SmartypantsFractions disabled
			// ref: https://github.com/superseriousbusiness/gotosocial/issues/1028
			Flags: blackfriday.UseXHTML | blackfriday.Smartypants | blackfriday.SmartypantsDashes | blackfriday.SmartypantsLatexDashes,
		}),
	}

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
	htmlContentBytes := blackfriday.Run([]byte(markdownText), blackfriday.WithExtensions(bfExtensions), blackfriday.WithRenderer(renderer))
	htmlContent := string(htmlContentBytes)

	// Replace emoji IDs in the parsed html content with their shortcodes again
	for _, e := range emojis {
		htmlContent = strings.ReplaceAll(htmlContent, ":"+e.ID+":", ":"+e.Shortcode+":")
	}

	// clean anything dangerous out of the html
	htmlContent = SanitizeHTML(htmlContent)

	if m == nil {
		m = minify.New()
		m.Add("text/html", &html.Minifier{
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
