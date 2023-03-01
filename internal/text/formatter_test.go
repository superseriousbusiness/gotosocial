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
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/concurrency"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/processing"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type TextStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db           db.DB
	parseMention gtsmodel.ParseMentionFunc

	// standard suite models
	testTokens       map[string]*gtsmodel.Token
	testClients      map[string]*gtsmodel.Client
	testApplications map[string]*gtsmodel.Application
	testUsers        map[string]*gtsmodel.User
	testAccounts     map[string]*gtsmodel.Account
	testAttachments  map[string]*gtsmodel.MediaAttachment
	testStatuses     map[string]*gtsmodel.Status
	testTags         map[string]*gtsmodel.Tag
	testMentions     map[string]*gtsmodel.Mention
	testEmojis       map[string]*gtsmodel.Emoji

	// module being tested
	formatter text.Formatter
}

func (suite *TextStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testClients = testrig.NewTestClients()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
	suite.testEmojis = testrig.NewTestEmojis()
}

func (suite *TextStandardTestSuite) SetupTest() {
	testrig.InitTestLog()
	testrig.InitTestConfig()

	suite.db = testrig.NewTestDB()

	fedWorker := concurrency.NewWorkerPool[messages.FromFederator](-1, -1)
	federator := testrig.NewTestFederator(suite.db, testrig.NewTestTransportController(testrig.NewMockHTTPClient(nil, "../../testrig/media"), suite.db, fedWorker), nil, nil, fedWorker)
	suite.parseMention = processing.GetParseMentionFunc(suite.db, federator)

	suite.formatter = text.NewFormatter(suite.db)

	testrig.StandardDBSetup(suite.db, nil)
}

func (suite *TextStandardTestSuite) TearDownTest() {
	testrig.StandardDBTeardown(suite.db)
}

func (suite *TextStandardTestSuite) FromMarkdown(text string) *text.FormatResult {
	return suite.formatter.FromMarkdown(context.Background(), suite.parseMention, suite.testAccounts["local_account_1"].ID, "status_ID", text)
}

func (suite *TextStandardTestSuite) FromPlain(text string) *text.FormatResult {
	return suite.formatter.FromPlain(context.Background(), suite.parseMention, suite.testAccounts["local_account_1"].ID, "status_ID", text)
}
