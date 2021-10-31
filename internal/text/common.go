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
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"html"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// preformat contains some common logic for making a string ready for formatting, which should be used for all user-input text.
func preformat(in string) string {
	// do some preformatting of the text

	// 1. unescape everything that might be html escaped
	s := html.UnescapeString(in)

	// 2. trim leading or trailing whitespace
	s = strings.TrimSpace(s)
	return s
}

// postformat contains some common logic for html sanitization of text, wrapping elements, and trimming newlines and whitespace
func postformat(in string) string {
	// do some postformatting of the text

	// 1. sanitize html to remove potentially dangerous elements
	s := SanitizeHTML(in)

	// 2. the sanitize step tends to escape characters inside codeblocks, which is behavior we don't want, so unescape everything again
	s = html.UnescapeString(s)

	// 3. minify html to remove any trailing newlines, spaces, unnecessary elements, etc etc
	mini, err := MinifyHTML(s)
	if err != nil {
		// if the minify failed, just return what we have
		return s
	}
	// return minified version of the html
	return mini
}

func (f *formatter) ReplaceTags(ctx context.Context, in string, tags []*gtsmodel.Tag) string {
	return regexes.HashtagFinder.ReplaceAllStringFunc(in, func(match string) string {
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

func (f *formatter) ReplaceMentions(ctx context.Context, in string, mentions []*gtsmodel.Mention) string {
	return regexes.MentionFinder.ReplaceAllStringFunc(in, func(match string) string {
		// we have a match
		matchTrimmed := strings.TrimSpace(match)
		// check through mentions to find what we're matching
		for _, menchie := range mentions {
			if strings.EqualFold(matchTrimmed, menchie.NameString) {
				// make sure we have an account attached to this mention
				if menchie.TargetAccount == nil {
					a, err := f.db.GetAccountByID(ctx, menchie.TargetAccountID)
					if err != nil {
						logrus.Errorf("error getting account with id %s from the db: %s", menchie.TargetAccountID, err)
						return match
					}
					menchie.TargetAccount = a
				}
				targetAccount := menchie.TargetAccount

				// replace the mention with the formatted mention content
				mentionContent := fmt.Sprintf(`<span class="h-card"><a href="%s" class="u-url mention">@<span>%s</span></a></span>`, targetAccount.URL, targetAccount.Username)

				// in case the match picked up any previous space or newlines (thanks to the regex), include them as well
				if strings.HasPrefix(match, " ") {
					mentionContent = " " + mentionContent
				} else if strings.HasPrefix(match, "\n") {
					mentionContent = "\n" + mentionContent
				}

				// done
				return mentionContent
			}
		}
		// the match wasn't in the list of mentions for whatever reason, so just return the match as we found it so nothing changes
		return match
	})
}
