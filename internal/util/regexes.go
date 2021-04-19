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

package util

import "regexp"

var (
	// mention regex can be played around with here: https://regex101.com/r/qwM9D3/1
	mentionRegexString = `(?: |^|\W)(@[a-zA-Z0-9_]+(?:@[a-zA-Z0-9_\-\.]+)?)(?: |\n)`
	mentionRegex       = regexp.MustCompile(mentionRegexString)
	// hashtag regex can be played with here: https://regex101.com/r/Vhy8pg/1
	hashtagRegexString = `(?: |^|\W)?#([a-zA-Z0-9]{1,30})(?:\b|\r)`
	hashtagRegex       = regexp.MustCompile(hashtagRegexString)
	// emoji regex can be played with here: https://regex101.com/r/478XGM/1
	emojiRegexString = `(?: |^|\W)?:([a-zA-Z0-9_]{2,30}):(?:\b|\r)?`
	emojiRegex       = regexp.MustCompile(emojiRegexString)
	// emoji shortcode regex can be played with here: https://regex101.com/r/zMDRaG/1
	emojiShortcodeString = `^[a-z0-9_]{2,30}$`
	emojiShortcodeRegex  = regexp.MustCompile(emojiShortcodeString)
)
