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
	"context"

	"github.com/russross/blackfriday/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

var (
	bfExtensions = blackfriday.CommonExtensions | blackfriday.HardLineBreak | blackfriday.Footnotes
	m            *minify.M
)

func (f *formatter) FromMarkdown(ctx context.Context, md string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string {
	// format tags nicely
	content := f.ReplaceTags(ctx, md, tags)

	// format mentions nicely
	content = f.ReplaceMentions(ctx, content, mentions)

	// parse markdown
	contentBytes := blackfriday.Run([]byte(content), blackfriday.WithExtensions(bfExtensions))

	// clean anything dangerous out of it
	content = SanitizeHTML(string(contentBytes))

	if m == nil {
		m = minify.New()
		m.Add("text/html", &html.Minifier{
			KeepEndTags: true,
			KeepQuotes:  true,
		})
	}

	minified, err := m.String("text/html", content)
	if err != nil {
		log.Errorf("error minifying markdown text: %s", err)
	}

	return minified
}
