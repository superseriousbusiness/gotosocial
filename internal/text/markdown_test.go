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

package text_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

const (
	simpleMarkdown = `# Title

Here's a simple text in markdown.

Here's a [link](https://example.org).`

	simpleMarkdownExpected = "<h1>Title</h1><p>Here’s a simple text in markdown.</p><p>Here’s a <a href=\"https://example.org\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">link</a>.</p>"

	withCodeBlock         = "# Title\n\n``` text\nhere's some code!\n```\n\nthat was some code :)"
	withCodeBlockExpected = "<h1>Title</h1><pre><code class=\"language-text\">here&#39;s some code!</code></pre><p>that was some code :)</p>"

	withHashtag         = "# Title\n\nhere's a simple status that uses hashtag #Hashtag!"
	withHashtagExpected = "<h1>Title</h1><p>here’s a simple status that uses hashtag <a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a>!</p>"
)

type MarkdownTestSuite struct {
	TextStandardTestSuite
}

func (suite *MarkdownTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
}

func (suite *MarkdownTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.formatter = text.NewFormatter(suite.config, suite.db, suite.log)

	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *MarkdownTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *MarkdownTestSuite) TestParseSimple() {
	s := suite.formatter.FromMarkdown(simpleMarkdown, nil, nil)
	suite.Equal(simpleMarkdownExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithCodeBlock() {
	s := suite.formatter.FromMarkdown(withCodeBlock, nil, nil)
	suite.Equal(withCodeBlockExpected, s)
}

func (suite *MarkdownTestSuite) TestParseWithHashtag() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	s := suite.formatter.FromMarkdown(withHashtag, nil, foundTags)
	suite.Equal(withHashtagExpected, s)
}

func TestMarkdownTestSuite(t *testing.T) {
	suite.Run(t, new(MarkdownTestSuite))
}
