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
	"unicode"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/regexes"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (f *formatter) ReplaceTags(ctx context.Context, in string, tags []*gtsmodel.Tag) string {
	spans := util.FindHashtagSpansInText(in)

	if len(spans) == 0 {
		return in
	}

	var b strings.Builder
	i := 0

spans:
	for _, t := range spans {
		b.WriteString(in[i:t.First])
		i = t.Second
		tagAsEntered := in[t.First+1 : t.Second]

		for _, tag := range tags {
			if strings.EqualFold(tagAsEntered, tag.Name) {
				// replace the #tag with the formatted tag content
				// `<a href="tag.URL" class="mention hashtag" rel="tag">#<span>tagAsEntered</span></a>
				b.WriteString(`<a href="`)
				b.WriteString(tag.URL)
				b.WriteString(`" class="mention hashtag" rel="tag">#<span>`)
				b.WriteString(tagAsEntered)
				b.WriteString(`</span></a>`)
				continue spans
			}
		}

		b.WriteString(in[t.First:t.Second])
	}

	// Get the last bits.
	i = spans[len(spans)-1].Second
	b.WriteString(in[i:])

	return b.String()
}

func (f *formatter) ReplaceMentions(ctx context.Context, in string, mentions []*gtsmodel.Mention) string {
	return regexes.ReplaceAllStringFunc(regexes.MentionFinder, in, func(match string, buf *bytes.Buffer) string {
		// we have a match, trim any spaces
		matchTrimmed := strings.TrimSpace(match)

		// check through mentions to find what we're matching
		for _, menchie := range mentions {
			if strings.EqualFold(matchTrimmed, menchie.NameString) {
				// make sure we have an account attached to this mention
				if menchie.TargetAccount == nil {
					a, err := f.db.GetAccountByID(ctx, menchie.TargetAccountID)
					if err != nil {
						log.Errorf("error getting account with id %s from the db: %s", menchie.TargetAccountID, err)
						return match
					}
					menchie.TargetAccount = a
				}

				// The mention's target is our target
				targetAccount := menchie.TargetAccount

				// Add any dropped space from match
				if unicode.IsSpace(rune(match[0])) {
					buf.WriteByte(match[0])
				}

				// replace the mention with the formatted mention content
				// <span class="h-card"><a href="targetAccount.URL" class="u-url mention">@<span>targetAccount.Username</span></a></span>
				buf.WriteString(`<span class="h-card"><a href="`)
				buf.WriteString(targetAccount.URL)
				buf.WriteString(`" class="u-url mention">@<span>`)
				buf.WriteString(targetAccount.Username)
				buf.WriteString(`</span></a></span>`)
				return buf.String()
			}
		}

		// the match wasn't in the list of mentions for whatever reason, so just return the match as we found it so nothing changes
		return match
	})
}
