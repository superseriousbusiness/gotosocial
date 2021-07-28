/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	"fmt"
	"strings"

	"github.com/russross/blackfriday/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

var bfExtensions = blackfriday.NoIntraEmphasis |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough |
	blackfriday.SpaceHeadings |
	blackfriday.BackslashLineBreak

func (f *formatter) FromMarkdown(md string, mentions []*gtsmodel.Mention, tags []*gtsmodel.Tag) string {
	content := preformat(md)

	// do the markdown parsing *first*
	content = string(blackfriday.Run([]byte(content), blackfriday.WithExtensions(bfExtensions)))

	// format tags nicely
	content = util.HashtagFinderRegex.ReplaceAllStringFunc(content, func(match string) string {
		for _, tag := range tags {
			if strings.TrimSpace(match) == fmt.Sprintf("#%s", tag.Name) {
				tagContent := fmt.Sprintf(`<a href="%s" class="mention hashtag" rel="tag">#<span>%s</span></a>`, tag.URL, tag.Name)
				if strings.HasPrefix(match, " ") {
					tagContent = " " + tagContent
				}
				fmt.Println(tagContent)
				return tagContent
			}
		}
		return content
	})

	// format mentions nicely
	for _, menchie := range mentions {
		targetAccount := &gtsmodel.Account{}
		if err := f.db.GetByID(menchie.TargetAccountID, targetAccount); err == nil {
			mentionContent := fmt.Sprintf(`<span class="h-card"><a href="%s" class="u-url mention">@<span>%s</span></a></span>`, targetAccount.URL, targetAccount.Username)
			content = strings.ReplaceAll(content, menchie.NameString, mentionContent)
		}
	}

	return postformat(content)
}
