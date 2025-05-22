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
	"testing"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/db/test"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type ConversationTestSuite struct {
	BunDBStandardTestSuite

	cf test.ConversationFactory

	// testAccount is the owner of statuses and conversations in these tests (must be local).
	testAccount *gtsmodel.Account
	// threadID is the thread used for statuses in any given test.
	threadID string
}

func (suite *ConversationTestSuite) SetupSuite() {
	suite.BunDBStandardTestSuite.SetupSuite()

	suite.cf.SetupSuite(suite)

	suite.testAccount = suite.testAccounts["local_account_1"]
}

func (suite *ConversationTestSuite) SetupTest() {
	suite.BunDBStandardTestSuite.SetupTest()

	suite.cf.SetupTest(suite.db)

	suite.threadID = suite.cf.NewULID(0)
}

// deleteStatus deletes a status from conversations and ends the test if that fails.
func (suite *ConversationTestSuite) deleteStatus(statusID string) {
	err := suite.db.DeleteStatusFromConversations(suite.T().Context(), statusID)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// getConversation fetches a conversation by ID and ends the test if that fails.
func (suite *ConversationTestSuite) getConversation(conversationID string) *gtsmodel.Conversation {
	conversation, err := suite.db.GetConversationByID(suite.T().Context(), conversationID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	return conversation
}

// If we delete a status that is in a conversation but not the last status,
// the conversation's last status should not change.
func (suite *ConversationTestSuite) TestDeleteNonLastStatus() {
	conversation := suite.cf.NewTestConversation(suite.testAccount, 0)
	initial := conversation.LastStatus
	reply := suite.cf.NewTestStatus(suite.testAccount, conversation.ThreadID, 1*time.Second, initial)
	conversation = suite.cf.SetLastStatus(conversation, reply)

	suite.deleteStatus(initial.ID)
	conversation = suite.getConversation(conversation.ID)
	suite.Equal(reply.ID, conversation.LastStatusID)
}

// If we delete the last status in a conversation that has other statuses,
// a previous status should become the new last status.
func (suite *ConversationTestSuite) TestDeleteLastStatus() {
	conversation := suite.cf.NewTestConversation(suite.testAccount, 0)
	initial := conversation.LastStatus
	reply := suite.cf.NewTestStatus(suite.testAccount, conversation.ThreadID, 1*time.Second, initial)
	conversation = suite.cf.SetLastStatus(conversation, reply)
	conversation = suite.getConversation(conversation.ID)

	suite.deleteStatus(reply.ID)
	conversation = suite.getConversation(conversation.ID)
	suite.Equal(initial.ID, conversation.LastStatusID)
}

// If we delete the only status in a conversation,
// the conversation should be deleted as well.
func (suite *ConversationTestSuite) TestDeleteOnlyStatus() {
	conversation := suite.cf.NewTestConversation(suite.testAccount, 0)
	initial := conversation.LastStatus

	suite.deleteStatus(initial.ID)
	_, err := suite.db.GetConversationByID(suite.T().Context(), conversation.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func TestConversationTestSuite(t *testing.T) {
	suite.Run(t, new(ConversationTestSuite))
}
