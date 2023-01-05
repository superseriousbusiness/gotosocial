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

package text_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
	simpleMarkdown                  = "# Title\n\nHere's a simple text in markdown.\n\nHere's a [link](https://example.org)."
	simpleMarkdownExpected          = "<h1>Title</h1><p>Here's a simple text in markdown.</p><p>Here's a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p>"
	withCodeBlockExpected           = "<h1>Title</h1><p>Below is some JSON.</p><pre><code class=\"language-json\">{\n  &#34;key&#34;: &#34;value&#34;,\n  &#34;another_key&#34;: [\n    &#34;value1&#34;,\n    &#34;value2&#34;\n  ]\n}\n</code></pre><p>that was some JSON :)</p>"
	withInlineCode                  = "`Nobody tells you about the <code><del>SECRET CODE</del></code>, do they?`"
	withInlineCodeExpected          = "<p><code>Nobody tells you about the &lt;code>&lt;del>SECRET CODE&lt;/del>&lt;/code>, do they?</code></p>"
	withInlineCode2                 = "`Nobody tells you about the </code><del>SECRET CODE</del><code>, do they?`"
	withInlineCode2Expected         = "<p><code>Nobody tells you about the &lt;/code>&lt;del>SECRET CODE&lt;/del>&lt;code>, do they?</code></p>"
	withHashtag                     = "# Title\n\nhere's a simple status that uses hashtag #Hashtag!"
	withHashtagExpected             = "<h1>Title</h1><p>here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a>!</p>"
	mdWithHTML                      = "# Title\n\nHere's a simple text in markdown.\n\nHere's a <a href=\"https://example.org\">link</a>.\n\nHere's an image: <img src=\"https://gts.superseriousbusiness.org/assets/logo.png\" alt=\"The GoToSocial sloth logo.\" width=\"500\" height=\"600\">"
	mdWithHTMLExpected              = "<h1>Title</h1><p>Here's a simple text in markdown.</p><p>Here's a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p><p>Here's an image: <img src=\"https://gts.superseriousbusiness.org/assets/logo.png\" alt=\"The GoToSocial sloth logo.\" width=\"500\" height=\"600\" crossorigin=\"anonymous\"></p>"
	mdWithCheekyHTML                = "# Title\n\nHere's a simple text in markdown.\n\nHere's a cheeky little script: <script>alert(ahhhh)</script>"
	mdWithCheekyHTMLExpected        = "<h1>Title</h1><p>Here's a simple text in markdown.</p><p>Here's a cheeky little script:</p>"
	mdWithHashtagInitial            = "#welcome #Hashtag"
	mdWithHashtagInitialExpected    = "<p><a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a> <a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a></p>"
	mdCodeBlockWithNewlines         = "some code coming up\n\n```\n\n\n\n```\nthat was some code"
	mdCodeBlockWithNewlinesExpected = "<p>some code coming up</p><pre><code>\n\n\n</code></pre><p>that was some code</p>"
	mdWithFootnote                  = "fox mulder,fbi.[^1]\n\n[^1]: federated bureau of investigation"
	mdWithFootnoteExpected          = "<p>fox mulder,fbi.[^1]</p><p>[^1]: federated bureau of investigation</p>"
	mdWithBlockQuote                = "get ready, there's a block quote coming:\n\n>line1\n>line2\n>\n>line3\n\n"
	mdWithBlockQuoteExpected        = "<p>get ready, there's a block quote coming:</p><blockquote><p>line1<br>line2</p><p>line3</p></blockquote>"
	mdHashtagAndCodeBlock           = "#Hashtag\n\n```\n#Hashtag\n```"
	mdHashtagAndCodeBlockExpected   = "<p><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a></p><pre><code>#Hashtag\n</code></pre>"
	mdMentionAndCodeBlock           = "@the_mighty_zork\n\n```\n@the_mighty_zork\n```"
	mdMentionAndCodeBlockExpected   = "<p><span class=\"h-card\"><a href=\"http://localhost:8080/@the_mighty_zork\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>the_mighty_zork</span></a></span></p><pre><code>@the_mighty_zork\n</code></pre>"
	mdWithSmartypants               = "\"you have to quargle the bleepflorp\" they said with 1/2 of nominal speed and 1/3 of the usual glumping"
	mdWithSmartypantsExpected       = "<p>\"you have to quargle the bleepflorp\" they said with 1/2 of nominal speed and 1/3 of the usual glumping</p>"
	mdWithAsciiHeart                = "hello <3 old friend <3 i loved u </3 :(( you stole my heart"
	mdWithAsciiHeartExpected        = "<p>hello &lt;3 old friend &lt;3 i loved u &lt;/3 :(( you stole my heart</p>"
	mdWithStrikethrough             = "I have ~~mdae~~ made an error"
	mdWithStrikethroughExpected     = "<p>I have <del>mdae</del> made an error</p>"
	mdWithLink                      = "Check out this code, i heard it was written by a sloth https://github.com/superseriousbusiness/gotosocial"
	mdWithLinkExpected              = "<p>Check out this code, i heard it was written by a sloth <a href=\"https://github.com/superseriousbusiness/gotosocial\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">https://github.com/superseriousbusiness/gotosocial</a></p>"
)

type MarkdownTestSuite struct {
	TextStandardTestSuite
}

func (suite *MarkdownTestSuite) TestParseSimple() {
	s := suite.formatter.FromMarkdown(context.Background(), simpleMarkdown, nil, nil, nil)
	suite.Equal(simpleMarkdownExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithCodeBlock() {
	s := suite.formatter.FromMarkdown(context.Background(), withCodeBlock, nil, nil, nil)
	suite.Equal(withCodeBlockExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithInlineCode() {
	s := suite.formatter.FromMarkdown(context.Background(), withInlineCode, nil, nil, nil)
	suite.Equal(withInlineCodeExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithInlineCode2() {
	s := suite.formatter.FromMarkdown(context.Background(), withInlineCode2, nil, nil, nil)
	suite.Equal(withInlineCode2Expected, s)
}

func (suite *MarkdownTestSuite) TestParseWithHashtag() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	s := suite.formatter.FromMarkdown(context.Background(), withHashtag, nil, foundTags, nil)
	suite.Equal(withHashtagExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithHTML() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithHTML, nil, nil, nil)
	suite.Equal(mdWithHTMLExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithCheekyHTML() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithCheekyHTML, nil, nil, nil)
	suite.Equal(mdWithCheekyHTMLExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithHashtagInitial() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithHashtagInitial, nil, []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
		suite.testTags["welcome"],
	}, nil)
	suite.Equal(mdWithHashtagInitialExpected, s)
}

func (suite *MarkdownTestSuite) TestParseCodeBlockWithNewlines() {
	s := suite.formatter.FromMarkdown(context.Background(), mdCodeBlockWithNewlines, nil, nil, nil)
	suite.Equal(mdCodeBlockWithNewlinesExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithFootnote() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithFootnote, nil, nil, nil)
	suite.Equal(mdWithFootnoteExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithBlockquote() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithBlockQuote, nil, nil, nil)
	suite.Equal(mdWithBlockQuoteExpected, s)
}

func (suite *MarkdownTestSuite) TestParseHashtagWithCodeBlock() {
	s := suite.formatter.FromMarkdown(context.Background(), mdHashtagAndCodeBlock, nil, []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}, nil)
	suite.Equal(mdHashtagAndCodeBlockExpected, s)
}

func (suite *MarkdownTestSuite) TestParseMentionWithCodeBlock() {
	s := suite.formatter.FromMarkdown(context.Background(), mdMentionAndCodeBlock, []*gtsmodel.Mention{
		suite.testMentions["local_user_2_mention_zork"],
	}, nil, nil)
	suite.Equal(mdMentionAndCodeBlockExpected, s)
}

func (suite *MarkdownTestSuite) TestParseSmartypants() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithSmartypants, []*gtsmodel.Mention{
		suite.testMentions["local_user_2_mention_zork"],
	}, nil, nil)
	suite.Equal(mdWithSmartypantsExpected, s)
}

func (suite *MarkdownTestSuite) TestParseAsciiHeart() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithAsciiHeart, nil, nil, nil)
	suite.Equal(mdWithAsciiHeartExpected, s)
}

func (suite *MarkdownTestSuite) TestParseStrikethrough() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithStrikethrough, nil, nil, nil)
	suite.Equal(mdWithStrikethroughExpected, s)
}

func (suite *MarkdownTestSuite) TestParseLink() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithLink, nil, nil, nil)
	suite.Equal(mdWithLinkExpected, s)
}

func TestMarkdownTestSuite(t *testing.T) {
	suite.Run(t, new(MarkdownTestSuite))
}
