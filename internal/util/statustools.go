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
	"unicode"
	"unicode/utf8"

	"github.com/superseriousbusiness/gotosocial/internal/regexes"
)

const (
	maximumHashtagLength = 30
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

type Pair[A, B any] struct {
	First  A
	Second B
}

// Byte index in original string
// `First` includes `#`.
type Span = Pair[int, int]

// Takes a plaintext (ie., not HTML-formatted) text,
// and returns a slice of unique hashtags.
func DeriveHashtagsFromText(text string) []string {
	tagsMap := make(map[string]bool)
	tags := []string{}

	for _, v := range FindHashtagSpansInText(text) {
		t := text[v.First+1 : v.Second]
		if _, value := tagsMap[t]; !value {
			tagsMap[t] = true
			tags = append(tags, t)
		}
	}

	return tags
}

// Takes a plaintext (ie., not HTML-formatted) text,
// and returns a list of pairs of indices into the original string, where
// hashtags are located.
func FindHashtagSpansInText(text string) []Span {
	tags := []Span{}
	start := 0
	// Keep one rune of lookbehind.
	prev := ' '
	inTag := false

	for i, r := range text {
		if r == '#' && IsHashtagBoundary(prev) {
			// Start of hashtag.
			inTag = true
			start = i
		} else if inTag && !IsPermittedInHashtag(r) && !IsHashtagBoundary(r) {
			// Inside the hashtag, but it was a phoney, gottem.
			inTag = false
		} else if inTag && IsHashtagBoundary(r) {
			// End of hashtag.
			inTag = false
			appendTag(&tags, text, start, i)
		} else if irl := i + utf8.RuneLen(r); inTag && irl == len(text) {
			// End of text.
			appendTag(&tags, text, start, irl)
		}

		prev = r
	}

	return tags
}

func appendTag(tags *[]Span, text string, start int, end int) {
	l := end - start - 1
	// This check could be moved out into the parsing loop if necessary!
	if 0 < l && l <= maximumHashtagLength {
		*tags = append(*tags, Span{First: start, Second: end})
	}
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

func IsPermittedInHashtag(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// Decides where to break before or after a hashtag.
func IsHashtagBoundary(r rune) bool {
	return r == '#' || // `###lol` should work
		unicode.IsSpace(r) || // All kinds of Unicode whitespace.
		unicode.IsControl(r) || // All kinds of control characters, like tab.
		// Most kinds of punctuation except "Pc" ("Punctuation, connecting", like `_`).
		// But `someurl/#fragment` should not match, neither should HTML entities like `&#35;`.
		('/' != r && '&' != r && !unicode.Is(unicode.Categories["Pc"], r) && unicode.IsPunct(r))
}
