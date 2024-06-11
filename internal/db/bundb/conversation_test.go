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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

type ConversationTestSuite struct {
	BunDBStandardTestSuite

	// account is the owner of statuses and conversations in these tests (must be local).
	account *gtsmodel.Account
	// now is the timestamp used as a base for creating new statuses in any given test.
	now time.Time
	// threadID is the thread used for statuses in any given test.
	threadID string
}

func (suite *ConversationTestSuite) SetupSuite() {
	suite.BunDBStandardTestSuite.SetupSuite()

	suite.account = suite.testAccounts["local_account_1"]
}

func (suite *ConversationTestSuite) SetupTest() {
	suite.BunDBStandardTestSuite.SetupTest()

	suite.now = time.Now()
	suite.threadID = id.NewULID()
}

// newStatus creates a new status in the DB that would be eligible for a conversation, optionally replying to a previous status.
func (suite *ConversationTestSuite) newStatus(nowOffset time.Duration, inReplyTo *gtsmodel.Status) *gtsmodel.Status {
	statusID := id.NewULID()
	createdAt := suite.now.Add(nowOffset)
	status := &gtsmodel.Status{
		ID:                  statusID,
		CreatedAt:           createdAt,
		UpdatedAt:           createdAt,
		URI:                 "http://localhost:8080/users/" + suite.account.Username + "/statuses/" + statusID,
		AccountID:           suite.account.ID,
		AccountURI:          suite.account.URI,
		Local:               util.Ptr(true),
		ThreadID:            suite.threadID,
		Visibility:          gtsmodel.VisibilityDirect,
		ActivityStreamsType: ap.ObjectNote,
		Federated:           util.Ptr(true),
		Boostable:           util.Ptr(true),
		Replyable:           util.Ptr(true),
		Likeable:            util.Ptr(true),
	}
	if inReplyTo != nil {
		status.InReplyToID = inReplyTo.ID
		status.InReplyToURI = inReplyTo.URI
		status.InReplyToAccountID = inReplyTo.AccountID
	}
	if err := suite.db.PutStatus(context.Background(), status); err != nil {
		suite.FailNow(err.Error())
	}

	return status
}

// newConversation creates a new conversation not yet in the DB.
func (suite *ConversationTestSuite) newConversation() *gtsmodel.Conversation {
	return &gtsmodel.Conversation{
		ID:        id.NewULID(),
		AccountID: suite.account.ID,
		ThreadID:  suite.threadID,
		Read:      util.Ptr(true),
	}
}

// addStatus adds a status to a conversation and ends the test if that fails.
func (suite *ConversationTestSuite) addStatus(
	conversation *gtsmodel.Conversation,
	status *gtsmodel.Status,
) *gtsmodel.Conversation {
	conversation, err := suite.db.AddStatusToConversation(context.Background(), conversation, status)
	if err != nil {
		suite.FailNow(err.Error())
	}
	return conversation
}

// deleteStatus deletes a status from conversations and ends the test if that fails.
func (suite *ConversationTestSuite) deleteStatus(statusID string) {
	err := suite.db.DeleteStatusFromConversations(context.Background(), statusID)
	if err != nil {
		suite.FailNow(err.Error())
	}
}

// getConversation fetches a conversation by ID and ends the test if that fails.
func (suite *ConversationTestSuite) getConversation(conversationID string) *gtsmodel.Conversation {
	conversation, err := suite.db.GetConversationByID(context.Background(), conversationID)
	if err != nil {
		suite.FailNow(err.Error())
	}
	return conversation
}

// Adding a status to a new conversation should set the last status.
func (suite *ConversationTestSuite) TestAddStatusToNewConversation() {
	initial := suite.newStatus(0, nil)
	conversation := suite.addStatus(suite.newConversation(), initial)
	suite.Equal(initial.ID, conversation.LastStatusID)
	if suite.NotNil(conversation.Read) {
		// In this test suite, the author of the statuses is also the owner of the conversation,
		// so the conversation should be marked as read.
		suite.True(*conversation.Read)
	}
}

// Adding a newer status to an existing conversation should update the last status.
func (suite *ConversationTestSuite) TestAddStatusToExistingConversation() {
	initial := suite.newStatus(0, nil)
	conversation := suite.addStatus(suite.newConversation(), initial)

	reply := suite.newStatus(1, initial)
	conversation = suite.addStatus(conversation, reply)
	suite.Equal(reply.ID, conversation.LastStatusID)
	if suite.NotNil(conversation.Read) {
		suite.True(*conversation.Read)
	}
}

// If we delete a status that is in a conversation but not the last status,
// the conversation's last status should not change.
func (suite *ConversationTestSuite) TestDeleteNonLastStatus() {
	initial := suite.newStatus(0, nil)
	conversation := suite.addStatus(suite.newConversation(), initial)
	reply := suite.newStatus(1, initial)
	conversation = suite.addStatus(conversation, reply)

	suite.deleteStatus(initial.ID)
	conversation = suite.getConversation(conversation.ID)
	suite.Equal(reply.ID, conversation.LastStatusID)
}

// If we delete the last status in a conversation that has other statuses,
// a previous status should become the new last status.
func (suite *ConversationTestSuite) TestDeleteLastStatus() {
	initial := suite.newStatus(0, nil)
	conversation := suite.addStatus(suite.newConversation(), initial)
	reply := suite.newStatus(1, initial)
	conversation = suite.addStatus(conversation, reply)

	suite.deleteStatus(reply.ID)
	conversation = suite.getConversation(conversation.ID)
	suite.Equal(initial.ID, conversation.LastStatusID)
}

// If we delete the only status in a conversation,
// the conversation should be deleted as well.
func (suite *ConversationTestSuite) TestDeleteOnlyStatus() {
	initial := suite.newStatus(0, nil)
	conversation := suite.addStatus(suite.newConversation(), initial)

	suite.deleteStatus(initial.ID)
	_, err := suite.db.GetConversationByID(context.Background(), conversation.ID)
	suite.ErrorIs(err, db.ErrNoEntries)
}

func TestConversationTestSuite(t *testing.T) {
	suite.Run(t, new(ConversationTestSuite))
}
