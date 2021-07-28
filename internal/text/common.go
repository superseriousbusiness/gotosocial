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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// preformat contains some common logic for making a string ready for formatting, which should be used for all user-input text.
func preformat(in string) string {
	// do some preformatting of the text
	// 1. Trim all the whitespace
	s := strings.TrimSpace(in)
	return s
}

// postformat contains some common logic for html sanitization of text, wrapping elements, and trimming newlines and whitespace
func postformat(in string) string {
	// do some postformatting of the text
	// 1. sanitize html to remove any dodgy scripts or other disallowed elements
	s := SanitizeOutgoing(in)
	// 2. wrap the whole thing in a paragraph
	s = fmt.Sprintf(`<p>%s</p>`, s)
	// 3. remove any cheeky newlines
	s = strings.ReplaceAll(s, "\n", "")
	// 4. remove any whitespace added as a result of the formatting
	s = strings.TrimSpace(s)
	return s
}

func (f *formatter) ReplaceTags(in string, tags []*gtsmodel.Tag) string {
	return util.HashtagFinderRegex.ReplaceAllStringFunc(in, func(match string) string {
		for _, tag := range tags {
			if strings.TrimSpace(match) == fmt.Sprintf("#%s", tag.Name) {
				tagContent := fmt.Sprintf(`<a href="%s" class="mention hashtag" rel="tag">#<span>%s</span></a>`, tag.URL, tag.Name)
				if strings.HasPrefix(match, " ") {
					tagContent = " " + tagContent
				}
				return tagContent
			}
		}
		return in
	})
}

func (f *formatter) ReplaceMentions(in string, mentions []*gtsmodel.Mention) string {
	for _, menchie := range mentions {
		targetAccount := &gtsmodel.Account{}
		if err := f.db.GetByID(menchie.TargetAccountID, targetAccount); err == nil {
			mentionContent := fmt.Sprintf(`<span class="h-card"><a href="%s" class="u-url mention">@<span>%s</span></a></span>`, targetAccount.URL, targetAccount.Username)
			in = strings.ReplaceAll(in, menchie.NameString, mentionContent)
		}
	}
	return in
}
