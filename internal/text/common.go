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
	// 1. remove any cheeky newlines
	s := strings.ReplaceAll(in, "\n", "")
	// 2. remove any whitespace added as a result of the formatting
	s = strings.TrimSpace(s)
	// 3. sanitize
	s = regular.Sanitize(s)
	return s
}

func (f *formatter) ReplaceTags(in string, tags []*gtsmodel.Tag) string {
	return util.HashtagFinderRegex.ReplaceAllStringFunc(in, func(match string) string {
		// we have a match
		matchTrimmed := strings.TrimSpace(match)
		tagAsEntered := strings.Split(matchTrimmed, "#")[1]

		// check through the tags to find what we're matching
		for _, tag := range tags {

			if strings.EqualFold(matchTrimmed, fmt.Sprintf("#%s", tag.Name)) {
				// replace the #tag with the formatted tag content
				tagContent := fmt.Sprintf(`<a href="%s" class="mention hashtag" rel="tag">#<span>%s</span></a>`, tag.URL, tagAsEntered)

				// in case the match picked up any previous space or newlines (thanks to the regex), include them as well
				if strings.HasPrefix(match, " ") {
					tagContent = " " + tagContent
				} else if strings.HasPrefix(match, "\n") {
					tagContent = "\n" + tagContent
				}

				// done
				return tagContent
			}
		}
		// the match wasn't in the list of tags for whatever reason, so just return the match as we found it so nothing changes
		return match
	})
}

func (f *formatter) ReplaceMentions(in string, mentions []*gtsmodel.Mention) string {
	for _, menchie := range mentions {
		// make sure we have a target account, either by getting one pinned on the mention,
		// or by pulling it from the database
		var targetAccount *gtsmodel.Account
		if menchie.GTSAccount != nil {
			// got it from the mention
			targetAccount = menchie.GTSAccount
		} else {
			a := &gtsmodel.Account{}
			if err := f.db.GetByID(menchie.TargetAccountID, a); err == nil {
				// got it from the db
				targetAccount = a
			} else {
				// couldn't get it so we can't do replacement
				return in
			}
		}

		mentionContent := fmt.Sprintf(`<span class="h-card"><a href="%s" class="u-url mention">@<span>%s</span></a></span>`, targetAccount.URL, targetAccount.Username)
		in = strings.ReplaceAll(in, menchie.NameString, mentionContent)
	}
	return in
}
