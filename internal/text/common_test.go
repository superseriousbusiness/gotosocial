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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

const (
	replaceMentionsString = `Another test @foss_satan@fossbros-anonymous.io

#Hashtag

Text`
	replaceMentionsExpected = `Another test <span class="h-card"><a href="http://fossbros-anonymous.io/@foss_satan" class="u-url mention">@<span>foss_satan</span></a></span>

#Hashtag

Text`

	replaceHashtagsExpected = `Another test @foss_satan@fossbros-anonymous.io

<a href="http://localhost:8080/tags/Hashtag" class="mention hashtag" rel="tag">#<span>Hashtag</span></a>

Text`

	replaceHashtagsAfterMentionsExpected = `Another test <span class="h-card"><a href="http://fossbros-anonymous.io/@foss_satan" class="u-url mention">@<span>foss_satan</span></a></span>

<a href="http://localhost:8080/tags/Hashtag" class="mention hashtag" rel="tag">#<span>Hashtag</span></a>

Text`
)

type CommonTestSuite struct {
	TextStandardTestSuite
}

func (suite *CommonTestSuite) SetupSuite() {
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

func (suite *CommonTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.formatter = text.NewFormatter(suite.config, suite.db, suite.log)

	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *CommonTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *CommonTestSuite) TestReplaceMentions() {
	foundMentions := []*gtsmodel.Mention{
		suite.testMentions["zork_mention_foss_satan"],
	}

	f := suite.formatter.ReplaceMentions(replaceMentionsString, foundMentions)
	assert.Equal(suite.T(), replaceMentionsExpected, f)
}

func (suite *CommonTestSuite) TestReplaceHashtags() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	f := suite.formatter.ReplaceTags(replaceMentionsString, foundTags)

	assert.Equal(suite.T(), replaceHashtagsExpected, f)
}

func (suite *CommonTestSuite) TestReplaceHashtagsAfterReplaceMentions() {
	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	f := suite.formatter.ReplaceTags(replaceMentionsExpected, foundTags)

	assert.Equal(suite.T(), replaceHashtagsAfterMentionsExpected, f)
}

func TestCommonTestSuite(t *testing.T) {
	suite.Run(t, new(CommonTestSuite))
}
