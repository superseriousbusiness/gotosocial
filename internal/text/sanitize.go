// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package text

import (
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// Regular HTML policy is an adapted version of the default
// bluemonday UGC policy, with some tweaks of our own.
// See: https://github.com/microcosm-cc/bluemonday#usage
var regular *bluemonday.Policy = func() *bluemonday.Policy {
	p := bluemonday.NewPolicy()

	// AllowStandardAttributes will enable "id", "title" and
	// the language specific attributes "dir" and "lang" on
	// all elements that are allowed
	p.AllowStandardAttributes()

	/*
		LAYOUT AND FORMATTING
	*/

	// "aside" is permitted and takes no attributes.
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/aside
	p.AllowElements("article", "aside")

	// "details" is permitted, including the "open" attribute
	// which can either be blank or the value "open".
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/details
	p.AllowAttrs("open").Matching(regexp.MustCompile(`(?i)^(|open)$`)).OnElements("details")

	// "section" is permitted and takes no attributes.
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/section
	p.AllowElements("section")

	// "summary" is permitted and takes no attributes.
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/summary
	p.AllowElements("summary")

	// "h1" through "h6" are permitted and take no attributes.
	p.AllowElements("h1", "h2", "h3", "h4", "h5", "h6")

	// "hgroup" is permitted and takes no attributes.
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/hgroup
	p.AllowElements("hgroup")

	// "blockquote" is permitted, including the "cite"
	// attribute which must be a standard URL.
	p.AllowAttrs("cite").OnElements("blockquote")

	// "br" "div" "hr" "p" "span" "wbr" are permitted and take no attributes
	p.AllowElements("br", "div", "hr", "p", "span", "wbr")

	// The following are all inline phrasing elements:
	p.AllowElements("abbr", "acronym", "cite", "code", "dfn", "em",
		"figcaption", "mark", "s", "samp", "strong", "sub", "sup", "var")

	// "q" is permitted and "cite" is a URL and handled by URL policies
	p.AllowAttrs("cite").OnElements("q")

	// "time" is permitted
	p.AllowAttrs("datetime").Matching(bluemonday.ISO8601).OnElements("time")

	// Block and inline elements that impart no
	// semantic meaning but style the document.
	// Underlines, italics, bold, strikethrough etc.
	p.AllowElements("b", "i", "pre", "small", "strike", "tt", "u")

	// "del" "ins" are permitted
	p.AllowAttrs("cite").Matching(bluemonday.Paragraph).OnElements("del", "ins")
	p.AllowAttrs("datetime").Matching(bluemonday.ISO8601).OnElements("del", "ins")

	// Enable ordered, unordered, and definition lists.
	p.AllowLists()

	// Class needed on span for mentions, which look like this when assembled:
	// `<span class="h-card"><a href="https://example.org/users/targetAccount" class="u-url mention">@<span>someusername</span></a></span>`
	p.AllowAttrs("class").OnElements("span")

	/*
		LANGUAGE FORMATTING
	*/

	// "bdi" "bdo" are permitted on "dir".
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/dir
	p.AllowAttrs("dir").Matching(bluemonday.Direction).OnElements("bdi", "bdo")

	// "rp" "rt" "ruby" are permitted. See:
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/rp
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/rt
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/ruby
	p.AllowElements("rp", "rt", "ruby")

	/*
		CODE BLOCKS
	*/

	// Permit language tags for code elements.
	p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")

	// Don't sanitize HTML inside code blocks.
	p.SkipElementsContent("code", "pre")

	/*
		LINKS AND LINK SAFETY.
	*/

	// Permit hyperlinks.
	p.AllowAttrs("class", "rel").OnElements("a")

	// Permit footnote roles on anchor elements.
	p.AllowAttrs("role").Matching(regexp.MustCompile("^doc-noteref$")).OnElements("a")
	p.AllowAttrs("role").Matching(regexp.MustCompile("^doc-backlink$")).OnElements("a")

	// URLs must be parseable by net/url.Parse().
	p.RequireParseableURLs(true)

	// Relative URLs are OK as we
	// need fragments for footnotes.
	p.AllowRelativeURLs(true)

	// However *only* allow common schemes, and also
	// relative URLs beginning with "#", ie., fragments.
	// We don't want URL's like "../../peepee.html".
	p.AllowURLSchemes("mailto", "http", "https")
	p.AllowAttrs("href").Matching(regexp.MustCompile("^(?:#|mailto|https://|http://).+$")).OnElements("a")

	// Force rel="noreferrer".
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel/noreferrer
	p.RequireNoReferrerOnFullyQualifiedLinks(true)

	// Add rel="nofollow" on all fully qualified (not relative) links.
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/rel#nofollow
	p.RequireNoFollowOnFullyQualifiedLinks(true)

	// Force crossorigin="anonymous"
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes/crossorigin#anonymous
	p.RequireCrossOriginAnonymous(true)

	// Force target="_blank".
	// See: https://developer.mozilla.org/en-US/docs/Web/HTML/Element/a#target
	p.AddTargetBlankToFullyQualifiedLinks(true)

	return p
}()

// '[C]an be thought of as equivalent to stripping all HTML
// elements and their attributes as it has nothing on its allowlist.
// An example usage scenario would be blog post titles where HTML
// tags are not expected at all and if they are then the elements
// and the content of the elements should be stripped. This is a
// very strict policy.'
//
// Source: https://github.com/microcosm-cc/bluemonday#usage
var strict *bluemonday.Policy = bluemonday.StrictPolicy()

// SanitizeHTML sanitizes only risky html elements
// from the given string, allowing safe ones through.
//
// It returns an HTML string.
func SanitizeHTML(html string) string {
	return regular.Sanitize(html)
}
