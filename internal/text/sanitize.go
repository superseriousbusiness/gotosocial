/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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
	"html"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// '[A]llows a broad selection of HTML elements and attributes that are safe for user generated content.
// Note that this policy does not allow iframes, object, embed, styles, script, etc.
// An example usage scenario would be blog post bodies where a variety of formatting is expected along with the potential for TABLEs and IMGs.'
//
// Source: https://github.com/microcosm-cc/bluemonday#usage
var regular *bluemonday.Policy = bluemonday.UGCPolicy().
	RequireNoReferrerOnLinks(true).
	RequireNoFollowOnLinks(false).              // remove the global default which adds rel="nofollow" to all links including local relative
	RequireNoFollowOnFullyQualifiedLinks(true). // add rel="nofollow" on all external links
	RequireCrossOriginAnonymous(true).
	AddTargetBlankToFullyQualifiedLinks(true).
	AllowAttrs("class", "href", "rel").OnElements("a").
	AllowAttrs("class").OnElements("span").
	AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code").
	SkipElementsContent("code", "pre")

// '[C]an be thought of as equivalent to stripping all HTML elements and their attributes as it has nothing on its allowlist.
// An example usage scenario would be blog post titles where HTML tags are not expected at all
// and if they are then the elements and the content of the elements should be stripped. This is a very strict policy.'
//
// Source: https://github.com/microcosm-cc/bluemonday#usage
var strict *bluemonday.Policy = bluemonday.StrictPolicy()

// removeHTML strictly removes *all* recognized HTML elements from the given string.
func removeHTML(in string) string {
	return strict.Sanitize(in)
}

// SanitizeHTML sanitizes risky html elements from the given string, allowing only safe ones through.
func SanitizeHTML(in string) string {
	return regular.Sanitize(in)
}

// SanitizePlaintext runs text through basic sanitization. This removes
// any html elements that were in the string, and returns clean plaintext.
func SanitizePlaintext(in string) string {
	content := html.UnescapeString(in)
	content = removeHTML(content)
	content = html.UnescapeString(content)
	return strings.TrimSpace(content)
}
