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

package conversations_test

// Test that we can create conversations when a new status comes in.
func (suite *ConversationsTestSuite) TestUpdateConversationsForStatus() {
	ctx := suite.T().Context()

	// Precondition: the test user shouldn't have any conversations yet.
	conversations, err := suite.db.GetConversationsByOwnerAccountID(ctx, suite.testAccount.ID, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.Empty(conversations)

	// Create a status.
	threadID := suite.NewULID(0)
	status := suite.NewTestStatus(suite.testAccount, threadID, 0, nil)

	// Update conversations for it.
	notifications, err := suite.conversationsProcessor.UpdateConversationsForStatus(ctx, status)
	if err != nil {
		suite.FailNow(err.Error())
	}

	// In this test, the user is DMing themself, and should not receive a notification from that.
	suite.Empty(notifications)

	// The test user should have a conversation now.
	conversations, err = suite.db.GetConversationsByOwnerAccountID(ctx, suite.testAccount.ID, nil)
	if err != nil {
		suite.FailNow(err.Error())
	}
	suite.NotEmpty(conversations)
}
