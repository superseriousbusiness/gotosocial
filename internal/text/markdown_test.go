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

package text_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

var withCodeBlock = `# Title

Below is some JSON.

` + "```" + `json
{
  "key": "value",
  "another_key": [
    "value1",
    "value2"
  ]
}
` + "```" + `

that was some JSON :)
`

const (
	simpleMarkdown                     = "# Title\n\nHere's a simple text in markdown.\n\nHere's a [link](https://example.org)."
	simpleMarkdownExpected             = "<h1>Title</h1><p>Here's a simple text in markdown.</p><p>Here's a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p>"
	withCodeBlockExpected              = "<h1>Title</h1><p>Below is some JSON.</p><pre><code class=\"language-json\">{\n  &#34;key&#34;: &#34;value&#34;,\n  &#34;another_key&#34;: [\n    &#34;value1&#34;,\n    &#34;value2&#34;\n  ]\n}\n</code></pre><p>that was some JSON :)</p>"
	withInlineCode                     = "`Nobody tells you about the <code><del>SECRET CODE</del></code>, do they?`"
	withInlineCodeExpected             = "<p><code>Nobody tells you about the &lt;code>&lt;del>SECRET CODE&lt;/del>&lt;/code>, do they?</code></p>"
	withInlineCode2                    = "`Nobody tells you about the </code><del>SECRET CODE</del><code>, do they?`"
	withInlineCode2Expected            = "<p><code>Nobody tells you about the &lt;/code>&lt;del>SECRET CODE&lt;/del>&lt;code>, do they?</code></p>"
	withHashtag                        = "# Title\n\nhere's a simple status that uses hashtag #Hashtag!"
	withHashtagExpected                = "<h1>Title</h1><p>here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a>!</p>"
	withTamilHashtag                   = "here's a simple status that uses a hashtag in Tamil #தமிழ்"
	withTamilHashtagExpected           = "<p>here's a simple status that uses a hashtag in Tamil <a href=\"http://localhost:8080/tags/%E0%AE%A4%E0%AE%AE%E0%AE%BF%E0%AE%B4%E0%AF%8D\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>தமிழ்</span></a></p>"
	mdWithHTML                         = "# Title\n\nHere's a simple text in markdown.\n\nHere's a <a href=\"https://example.org\">link</a>.\n\nHere's an image: <img src=\"https://gts.superseriousbusiness.org/assets/logo.png\" alt=\"The GoToSocial sloth logo.\" width=\"500\" height=\"600\">"
	mdWithHTMLExpected                 = "<h1>Title</h1><p>Here's a simple text in markdown.</p><p>Here's a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p><p>Here's an image:</p>"
	mdWithCheekyHTML                   = "# Title\n\nHere's a simple text in markdown.\n\nHere's a cheeky little script: <script>alert(ahhhh)</script>"
	mdWithCheekyHTMLExpected           = "<h1>Title</h1><p>Here's a simple text in markdown.</p><p>Here's a cheeky little script:</p>"
	mdWithHashtagInitial               = "#welcome #Hashtag"
	mdWithHashtagInitialExpected       = "<p><a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a> <a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a></p>"
	mdCodeBlockWithNewlines            = "some code coming up\n\n```\n\n\n\n```\nthat was some code"
	mdCodeBlockWithNewlinesExpected    = "<p>some code coming up</p><pre><code>\n\n\n</code></pre><p>that was some code</p>"
	mdWithFootnote                     = "fox mulder,fbi.[^1]\n\n[^1]: federated bureau of investigation"
	mdWithFootnoteExpected             = "<p>fox mulder,fbi.<sup id=\"dummy_status_ID-fnref:1\"><a href=\"#dummy_status_ID-fn:1\" class=\"footnote-ref\" role=\"doc-noteref\">1</a></sup></p><div><hr><ol><li id=\"dummy_status_ID-fn:1\"><p>federated bureau of investigation <a href=\"#dummy_status_ID-fnref:1\" class=\"footnote-backref\" role=\"doc-backlink\">↩︎</a></p></li></ol></div>"
	mdWithAttemptedRelative            = "hello this is a cheeky relative link: <a href=\"../sneaky.html\">click it!</a>"
	mdWithAttemptedRelativeExpected    = "<p>hello this is a cheeky relative link: click it!</p>"
	mdWithBlockQuote                   = "get ready, there's a block quote coming:\n\n>line1\n>line2\n>\n>line3\n\n"
	mdWithBlockQuoteExpected           = "<p>get ready, there's a block quote coming:</p><blockquote><p>line1<br>line2</p><p>line3</p></blockquote>"
	mdHashtagAndCodeBlock              = "#Hashtag\n\n```\n#Hashtag\n```"
	mdHashtagAndCodeBlockExpected      = "<p><a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a></p><pre><code>#Hashtag\n</code></pre>"
	mdMentionAndCodeBlock              = "@the_mighty_zork\n\n```\n@the_mighty_zork\n```"
	mdMentionAndCodeBlockExpected      = "<p><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span></p><pre><code>@the_mighty_zork\n</code></pre>"
	mdMentionAndCodeBlockBasicExpected = "<p>@the_mighty_zork</p><pre><code>@the_mighty_zork\n</code></pre>"
	mdWithSmartypants                  = "\"you have to quargle the bleepflorp\" they said with 1/2 of nominal speed and 1/3 of the usual glumping"
	mdWithSmartypantsExpected          = "<p>\"you have to quargle the bleepflorp\" they said with 1/2 of nominal speed and 1/3 of the usual glumping</p>"
	mdWithAsciiHeart                   = "hello <3 old friend <3 i loved u </3 :(( you stole my heart"
	mdWithAsciiHeartExpected           = "<p>hello &lt;3 old friend &lt;3 i loved u &lt;/3 :(( you stole my heart</p>"
	mdWithStrikethrough                = "I have ~~mdae~~ made an error"
	mdWithStrikethroughExpected        = "<p>I have <del>mdae</del> made an error</p>"
	mdWithLink                         = "Check out this code, i heard it was written by a sloth https://codeberg.org/superseriousbusiness/gotosocial"
	mdWithLinkExpected                 = "<p>Check out this code, i heard it was written by a sloth <a href=\"https://codeberg.org/superseriousbusiness/gotosocial\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://codeberg.org/superseriousbusiness/gotosocial</a></p>"
	mdWithLinkBasicExpected            = "Check out this code, i heard it was written by a sloth <a href=\"https://codeberg.org/superseriousbusiness/gotosocial\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://codeberg.org/superseriousbusiness/gotosocial</a>"
	mdObjectInCodeBlock                = "@foss_satan@fossbros-anonymous.io this is how to mention a user\n```\n@the_mighty_zork hey bud! nice #ObjectOrientedProgramming software you've been writing lately! :rainbow:\n```\nhope that helps"
	mdObjectInCodeBlockExpected        = "<p><span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span> this is how to mention a user</p><pre><code>@the_mighty_zork hey bud! nice #ObjectOrientedProgramming software you&#39;ve been writing lately! :rainbow:\n</code></pre><p>hope that helps</p>"
	// Hashtags can be italicized but only with *, not _.
	mdItalicHashtag          = "*#hashtag*"
	mdItalicHashtagExpected  = "<p><em><a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag</span></a></em></p>"
	mdItalicHashtags         = "*#hashtag #hashtag #hashtag*"
	mdItalicHashtagsExpected = "<p><em><a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag</span></a> <a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag</span></a> <a href=\"http://localhost:8080/tags/hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag</span></a></em></p>"
	// Hashtags can end with or contain _ but not start with it.
	mdUnderscorePrefixHashtag         = "_#hashtag"
	mdUnderscorePrefixHashtagExpected = "<p>_#hashtag</p>"
	mdUnderscoreSuffixHashtag         = "#hashtag_"
	mdUnderscoreSuffixHashtagExpected = "<p><a href=\"http://localhost:8080/tags/hashtag_\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hashtag_</span></a></p>"
	// BEWARE: sneaky unicode business going on.
	// the first ö is one rune, the second ö is an o with a combining diacritic.
	mdUnnormalizedHashtag         = "#hellöthere #hellöthere"
	mdUnnormalizedHashtagExpected = "<p><a href=\"http://localhost:8080/tags/hell%C3%B6there\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hellöthere</span></a> <a href=\"http://localhost:8080/tags/hell%C3%B6there\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>hellöthere</span></a></p>"
)

type MarkdownTestSuite struct {
	TextStandardTestSuite
}

func (suite *MarkdownTestSuite) TestParseSimple() {
	formatted := suite.FromMarkdown(simpleMarkdown)
	suite.Equal(simpleMarkdownExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithCodeBlock() {
	formatted := suite.FromMarkdown(withCodeBlock)
	suite.Equal(withCodeBlockExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithInlineCode() {
	formatted := suite.FromMarkdown(withInlineCode)
	suite.Equal(withInlineCodeExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithInlineCode2() {
	formatted := suite.FromMarkdown(withInlineCode2)
	suite.Equal(withInlineCode2Expected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithHashtag() {
	formatted := suite.FromMarkdown(withHashtag)
	suite.Equal(withHashtagExpected, formatted.HTML)
}

// Regressiom test for https://codeberg.org/superseriousbusiness/gotosocial/issues/3618
func (suite *MarkdownTestSuite) TestParseWithTamilHashtag() {
	formatted := suite.FromMarkdown(withTamilHashtag)
	suite.Equal(withTamilHashtagExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithHTML() {
	formatted := suite.FromMarkdown(mdWithHTML)
	suite.Equal(mdWithHTMLExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithCheekyHTML() {
	formatted := suite.FromMarkdown(mdWithCheekyHTML)
	suite.Equal(mdWithCheekyHTMLExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithHashtagInitial() {
	formatted := suite.FromMarkdown(mdWithHashtagInitial)
	suite.Equal(mdWithHashtagInitialExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseCodeBlockWithNewlines() {
	formatted := suite.FromMarkdown(mdCodeBlockWithNewlines)
	suite.Equal(mdCodeBlockWithNewlinesExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithFootnote() {
	formatted := suite.FromMarkdown(mdWithFootnote)
	suite.Equal(mdWithFootnoteExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithAttemptedRelative() {
	formatted := suite.FromMarkdown(mdWithAttemptedRelative)
	suite.Equal(mdWithAttemptedRelativeExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseWithBlockquote() {
	formatted := suite.FromMarkdown(mdWithBlockQuote)
	suite.Equal(mdWithBlockQuoteExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseHashtagWithCodeBlock() {
	formatted := suite.FromMarkdown(mdHashtagAndCodeBlock)
	suite.Equal(mdHashtagAndCodeBlockExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseMentionWithCodeBlock() {
	formatted := suite.FromMarkdown(mdMentionAndCodeBlock)
	suite.Equal(mdMentionAndCodeBlockExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseMentionWithCodeBlockBasic() {
	formatted := suite.FromMarkdownBasic(mdMentionAndCodeBlock)
	suite.Equal(mdMentionAndCodeBlockBasicExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseSmartypants() {
	formatted := suite.FromMarkdown(mdWithSmartypants)
	suite.Equal(mdWithSmartypantsExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseAsciiHeart() {
	formatted := suite.FromMarkdown(mdWithAsciiHeart)
	suite.Equal(mdWithAsciiHeartExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseStrikethrough() {
	formatted := suite.FromMarkdown(mdWithStrikethrough)
	suite.Equal(mdWithStrikethroughExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseLink() {
	formatted := suite.FromMarkdown(mdWithLink)
	suite.Equal(mdWithLinkExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseLinkBasic() {
	formatted := suite.FromMarkdownBasic(mdWithLink)
	suite.Equal(mdWithLinkBasicExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseObjectInCodeBlock() {
	formatted := suite.FromMarkdown(mdObjectInCodeBlock)
	suite.Equal(mdObjectInCodeBlockExpected, formatted.HTML)
	suite.Len(formatted.Mentions, 1)
	suite.Equal("@foss_satan@fossbros-anonymous.io", formatted.Mentions[0].NameString)
	suite.Empty(formatted.Tags)
	suite.Empty(formatted.Emojis)
}

func (suite *MarkdownTestSuite) TestParseItalicHashtag() {
	formatted := suite.FromMarkdown(mdItalicHashtag)
	suite.Equal(mdItalicHashtagExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseItalicHashtags() {
	formatted := suite.FromMarkdown(mdItalicHashtags)
	suite.Equal(mdItalicHashtagsExpected, formatted.HTML)
}

func (suite *MarkdownTestSuite) TestParseHashtagUnderscorePrefix() {
	formatted := suite.FromMarkdown(mdUnderscorePrefixHashtag)
	suite.Equal(mdUnderscorePrefixHashtagExpected, formatted.HTML)
	suite.Empty(formatted.Tags)
}

func (suite *MarkdownTestSuite) TestParseHashtagUnderscoreSuffix() {
	formatted := suite.FromMarkdown(mdUnderscoreSuffixHashtag)
	suite.Equal(mdUnderscoreSuffixHashtagExpected, formatted.HTML)
	suite.NotEmpty(formatted.Tags)
	suite.Equal("hashtag_", formatted.Tags[0].Name)
}

func (suite *MarkdownTestSuite) TestParseUnnormalizedHashtag() {
	formatted := suite.FromMarkdown(mdUnnormalizedHashtag)
	suite.Equal(mdUnnormalizedHashtagExpected, formatted.HTML)
}

func TestMarkdownTestSuite(t *testing.T) {
	suite.Run(t, new(MarkdownTestSuite))
}
