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
	simple         = "this is a plain and simple status"
	simpleExpected = "<p>this is a plain and simple status</p>"

	withTag         = "this is a simple status that uses hashtag #welcome!"
	withTagExpected = "<p>this is a simple status that uses hashtag <a href=\"http://localhost:8080/tags/welcome\" class=\"mention hashtag\" rel=\"tag nofollow noreferrer noopener\" target=\"_blank\">#<span>welcome</span></a>!</p>"
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
}

func (suite *PlainTestSuite) SetupTest() {
	suite.config = testrig.NewTestConfig()
	suite.db = testrig.NewTestDB()
	suite.log = testrig.NewTestLog()
	suite.formatter = text.NewFormatter(suite.config, suite.db, suite.log)

	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *PlainTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *PlainTestSuite) TestParseSimple() {
	f := suite.formatter.FromPlain(simple, nil, nil)
	assert.Equal(suite.T(), simpleExpected, f)
}

func (suite *PlainTestSuite) TestParseWithTag() {

	foundTags := []*gtsmodel.Tag{
		suite.testTags["welcome"],
	}

	f := suite.formatter.FromPlain(withTag, nil, foundTags)
	assert.Equal(suite.T(), withTagExpected, f)
}

func TestPlainTestSuite(t *testing.T) {
	suite.Run(t, new(PlainTestSuite))
}
