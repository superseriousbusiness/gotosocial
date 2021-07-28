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
	"github.com/microcosm-cc/bluemonday"
)

// '[A]llows a broad selection of HTML elements and attributes that are safe for user generated content.
// Note that this policy does not allow iframes, object, embed, styles, script, etc.
// An example usage scenario would be blog post bodies where a variety of formatting is expected along with the potential for TABLEs and IMGs.'
//
// Source: https://github.com/microcosm-cc/bluemonday#usage
var regular *bluemonday.Policy = bluemonday.UGCPolicy().
	RequireNoReferrerOnLinks(true).
	RequireNoFollowOnLinks(true).
	RequireCrossOriginAnonymous(true).
	AddTargetBlankToFullyQualifiedLinks(true).
	AllowAttrs("class").OnElements("a")

// outgoing policy should be used on statuses we've already parsed and added our own elements etc to. It is less strict than regular.
var outgoing *bluemonday.Policy = regular.
	AllowAttrs("class", "href", "rel").OnElements("a").
	AllowAttrs("class").OnElements("span")

// '[C]an be thought of as equivalent to stripping all HTML elements and their attributes as it has nothing on its allowlist.
// An example usage scenario would be blog post titles where HTML tags are not expected at all
// and if they are then the elements and the content of the elements should be stripped. This is a very strict policy.'
//
// Source: https://github.com/microcosm-cc/bluemonday#usage
var strict *bluemonday.Policy = bluemonday.StrictPolicy()

// SanitizeHTML cleans up HTML in the given string, allowing through only safe HTML elements.
func SanitizeHTML(in string) string {
	return regular.Sanitize(in)
}

// RemoveHTML removes all HTML from the given string.
func RemoveHTML(in string) string {
	return strict.Sanitize(in)
}

// SanitizeOutgoing cleans up HTML in the given string, allowing through only safe elements and elements that were added during the parsing process.
// This should be used on text that we've already converted into HTML, just to catch any weirdness.
func SanitizeOutgoing(in string) string {
	return outgoing.Sanitize(in)
}
