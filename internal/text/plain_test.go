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
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

const (
	simple         = "this is a plain and simple status"
	simpleExpected = "<p>this is a plain and simple status</p>"

	withTag         = "here's a simple status that uses hashtag #welcome!"
	withTagExpected = "<p>here's a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!</p>"

	moreComplex = `Another test @foss_satan@fossbros-anonymous.io

#Hashtag

Text`
	moreComplexFull = "<p>Another test <span class=\"h-card\"><a href=\"http://fossbros-anonymous.io/@foss_satan\" class=\"u-url mention\" rel=\"nofollow noreferrer noopener\" target=\"_blank\">@<span>foss_satan</span></a></span><br><br><a href=\"http://localhost:8080/tags/Hashtag\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>Hashtag</span></a><br><br>Text</p>"
)

type PlainTestSuite struct {
	TextStandardTestSuite
}

func (suite *PlainTestSuite) SetupSuite() {
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

func (suite *PlainTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.formatter = text.NewFormatter(suite.config, suite.db)

	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *PlainTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *PlainTestSuite) TestParseSimple() {
	f := suite.formatter.FromPlain(context.Background(), simple, nil, nil)
	assert.Equal(suite.T(), simpleExpected, f)
}

func (suite *PlainTestSuite) TestParseWithTag() {

	foundTags := []*gtsmodel.Tag{
		suite.testTags["welcome"],
	}

	f := suite.formatter.FromPlain(context.Background(), withTag, nil, foundTags)
	assert.Equal(suite.T(), withTagExpected, f)
}

func (suite *PlainTestSuite) TestParseMoreComplex() {

	foundTags := []*gtsmodel.Tag{
		suite.testTags["Hashtag"],
	}

	foundMentions := []*gtsmodel.Mention{
		suite.testMentions["zork_mention_foss_satan"],
	}

	f := suite.formatter.FromPlain(context.Background(), moreComplex, foundMentions, foundTags)

	fmt.Println(f)

	assert.Equal(suite.T(), moreComplexFull, f)
}

func TestPlainTestSuite(t *testing.T) {
	suite.Run(t, new(PlainTestSuite))
}
