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
	simpleMarkdown           = "# Title\n\nHere's a simple text in markdown.\n\nHere's a [link](https://example.org)."
	simpleMarkdownExpected   = "<h1>Title</h1>\n\n<p>Here’s a simple text in markdown.</p>\n\n<p>Here’s a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p>\n"
	withCodeBlockExpected    = "<h1>Title</h1>\n\n<p>Below is some JSON.</p>\n\n<pre><code class=\"language-json\">{\n  &#34;key&#34;: &#34;value&#34;,\n  &#34;another_key&#34;: [\n    &#34;value1&#34;,\n    &#34;value2&#34;\n  ]\n}\n</code></pre>\n\n<p>that was some JSON :)</p>\n"
	withInlineCode           = "`Nobody tells you about the <code><del>SECRET CODE</del></code>, do they?`"
	withInlineCodeExpected   = "<p><code>Nobody tells you about the &lt;code&gt;&lt;del&gt;SECRET CODE&lt;/del&gt;&lt;/code&gt;, do they?</code></p>\n"
	withInlineCode2          = "`Nobody tells you about the </code><del>SECRET CODE</del><code>, do they?`"
	withInlineCode2Expected  = "<p><code>Nobody tells you about the &lt;/code&gt;&lt;del&gt;SECRET CODE&lt;/del&gt;&lt;code&gt;, do they?</code></p>\n"
	withHashtag              = "# Title\n\nhere's a simple status that uses hashtag #Hashtag!"
	withHashtagExpected      = "<h1>Title</h1>\n\n<p>here’s a simple status that uses hashtag <a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a>!</p>\n"
	mdWithHTML               = "# Title\n\nHere's a simple text in markdown.\n\nHere's a <a href=\"https://example.org\">link</a>.\n\nHere's an image: <img src=\"https://gts.superseriousbusiness.org/assets/logo.png\" alt=\"The GoToSocial sloth logo.\" width=\"500\" height=\"600\">"
	mdWithHTMLExpected       = "<h1>Title</h1>\n\n<p>Here’s a simple text in markdown.</p>\n\n<p>Here’s a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p>\n\n<p>Here’s an image: <img src=\"https://gts.superseriousbusiness.org/assets/logo.png\" alt=\"The GoToSocial sloth logo.\" width=\"500\" height=\"600\" crossorigin=\"anonymous\"></p>\n"
	mdWithCheekyHTML         = "# Title\n\nHere's a simple text in markdown.\n\nHere's a cheeky little script: <script>alert(ahhhh)</script>"
	mdWithCheekyHTMLExpected = "<h1>Title</h1>\n\n<p>Here’s a simple text in markdown.</p>\n\n<p>Here’s a cheeky little script: </p>\n"
)

type MarkdownTestSuite struct {
	TextStandardTestSuite
}

func (suite *MarkdownTestSuite) TestParseSimple() {
	s := suite.formatter.FromMarkdown(context.Background(), simpleMarkdown, nil, nil)
	suite.Equal(simpleMarkdownExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithCodeBlock() {
	s := suite.formatter.FromMarkdown(context.Background(), withCodeBlock, nil, nil)
	suite.Equal(withCodeBlockExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithInlineCode() {
	s := suite.formatter.FromMarkdown(context.Background(), withInlineCode, nil, nil)
	suite.Equal(withInlineCodeExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithInlineCode2() {
	s := suite.formatter.FromMarkdown(context.Background(), withInlineCode2, nil, nil)
	suite.Equal(withInlineCode2Expected, s)
}

func (suite *MarkdownTestSuite) TestParseWithHashtag() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	s := suite.formatter.FromMarkdown(context.Background(), withHashtag, nil, foundTags)
	suite.Equal(withHashtagExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithHTML() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithHTML, nil, nil)
	suite.Equal(mdWithHTMLExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithCheekyHTML() {
	s := suite.formatter.FromMarkdown(context.Background(), mdWithCheekyHTML, nil, nil)
	suite.Equal(mdWithCheekyHTMLExpected, s)
}

func TestMarkdownTestSuite(t *testing.T) {
	suite.Run(t, new(MarkdownTestSuite))
}
