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

package util

import (
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

// DeriveMentionNamesFromText takes a plaintext (ie., not html-formatted) text,
// and applies a regex to it to return a deduplicated list of account names
// mentioned in that text, in the format "@user@example.org" or "@username" for
// local users.
func DeriveMentionNamesFromText(text string) []string {
	mentionedAccounts := []string{}
	for _, m := range regexes.MentionFinder.FindAllStringSubmatch(text, -1) {
		mentionedAccounts = append(mentionedAccounts, m[1])
	}
	return UniqueStrings(mentionedAccounts)
}

// DeriveHashtagsFromText takes a plaintext (ie., not html-formatted) text,
// and applies a regex to it to return a deduplicated list of hashtags
// used in that text, without the leading #. The case of the returned
// tags will be lowered, for consistency.
func DeriveHashtagsFromText(text string) []string {
	tags := []string{}
	for _, m := range regexes.HashtagFinder.FindAllStringSubmatch(text, -1) {
		tags = append(tags, strings.TrimPrefix(m[1], "#"))
	}
	return UniqueStrings(tags)
}

// DeriveEmojisFromText takes a plaintext (ie., not html-formatted) text,
// and applies a regex to it to return a deduplicated list of emojis
// used in that text, without the surrounding `::`
func DeriveEmojisFromText(text string) []string {
	emojis := []string{}
	for _, m := range regexes.EmojiFinder.FindAllStringSubmatch(text, -1) {
		emojis = append(emojis, m[1])
	}
	return UniqueStrings(emojis)
}
