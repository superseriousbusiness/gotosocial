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

package bundb_test

import (
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type BunDBStandardTestSuite struct {
	// standard suite interfaces
	suite.Suite
	db    db.DB
	state state.State

	// standard suite models
	testTokens              map[string]*gtsmodel.Token
	testApplications        map[string]*gtsmodel.Application
	testUsers               map[string]*gtsmodel.User
	testAccounts            map[string]*gtsmodel.Account
	testAttachments         map[string]*gtsmodel.MediaAttachment
	testStatuses            map[string]*gtsmodel.Status
	testTags                map[string]*gtsmodel.Tag
	testMentions            map[string]*gtsmodel.Mention
	testFollows             map[string]*gtsmodel.Follow
	testEmojis              map[string]*gtsmodel.Emoji
	testReports             map[string]*gtsmodel.Report
	testBookmarks           map[string]*gtsmodel.StatusBookmark
	testFaves               map[string]*gtsmodel.StatusFave
	testLists               map[string]*gtsmodel.List
	testListEntries         map[string]*gtsmodel.ListEntry
	testAccountNotes        map[string]*gtsmodel.AccountNote
	testMarkers             map[string]*gtsmodel.Marker
	testRules               map[string]*gtsmodel.Rule
	testThreads             map[string]*gtsmodel.Thread
	testPolls               map[string]*gtsmodel.Poll
	testPollVotes           map[string]*gtsmodel.PollVote
	testInteractionRequests map[string]*gtsmodel.InteractionRequest
	testStatusEdits         map[string]*gtsmodel.StatusEdit
}

func (suite *BunDBStandardTestSuite) SetupSuite() {
	suite.testTokens = testrig.NewTestTokens()
	suite.testApplications = testrig.NewTestApplications()
	suite.testUsers = testrig.NewTestUsers()
	suite.testAccounts = testrig.NewTestAccounts()
	suite.testAttachments = testrig.NewTestAttachments()
	suite.testStatuses = testrig.NewTestStatuses()
	suite.testTags = testrig.NewTestTags()
	suite.testMentions = testrig.NewTestMentions()
	suite.testFollows = testrig.NewTestFollows()
	suite.testEmojis = testrig.NewTestEmojis()
	suite.testReports = testrig.NewTestReports()
	suite.testBookmarks = testrig.NewTestBookmarks()
	suite.testFaves = testrig.NewTestFaves()
	suite.testLists = testrig.NewTestLists()
	suite.testListEntries = testrig.NewTestListEntries()
	suite.testAccountNotes = testrig.NewTestAccountNotes()
	suite.testMarkers = testrig.NewTestMarkers()
	suite.testRules = testrig.NewTestRules()
	suite.testThreads = testrig.NewTestThreads()
	suite.testPolls = testrig.NewTestPolls()
	suite.testPollVotes = testrig.NewTestPollVotes()
	suite.testInteractionRequests = testrig.NewTestInteractionRequests()
	suite.testStatusEdits = testrig.NewTestStatusEdits()
}

func (suite *BunDBStandardTestSuite) SetupTest() {
	testrig.InitTestConfig()
	testrig.InitTestLog()
	suite.state.Caches.Init()
	suite.db = testrig.NewTestDB(&suite.state)
	testrig.StandardDBSetup(suite.db, suite.testAccounts)
}

func (suite *BunDBStandardTestSuite) TearDownTest() {
	if suite.db != nil {
		testrig.StandardDBTeardown(suite.db)
	}
}
