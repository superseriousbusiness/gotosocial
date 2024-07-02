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

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

func (suite *ConversationsTestSuite) TestGetAll() {
	resp, err := suite.conversationsProcessor.GetAll(context.Background(), suite.testAccount, nil)
	if suite.NoError(err) && suite.Len(resp.Items, 1) && suite.IsType((*apimodel.Conversation)(nil), resp.Items[0]) {
		apiConversation := resp.Items[0].(*apimodel.Conversation)
		suite.Equal(suite.testConversation.ID, apiConversation.ID)
		suite.True(apiConversation.Unread)
	}
}

// Test that conversations with newer last status IDs are returned earlier.
func (suite *ConversationsTestSuite) TestGetAllOrder() {
	// Get our previously created conversation.
	conversation1 := suite.testConversation

	// Create a new conversation with a last status newer than conversation1's.
	conversation2 := suite.newTestConversation(1)

	// Add an even newer status than that to conversation1.
	conversation1Status2 := suite.newTestStatus(conversation1.LastStatus.ThreadID, 2, conversation1.LastStatus)
	conversation1, err := suite.db.AddStatusToConversation(context.Background(), conversation1, conversation1Status2)
	if err != nil {
		suite.FailNow(err.Error())
	}

	resp, err := suite.conversationsProcessor.GetAll(context.Background(), suite.testAccount, nil)
	if suite.NoError(err) && suite.Len(resp.Items, 2) {
		// conversation1 should be the first conversation returned.
		apiConversation1 := resp.Items[0].(*apimodel.Conversation)
		suite.Equal(conversation1.ID, apiConversation1.ID)
		// It should have the newest status added to it.
		suite.Equal(conversation1.LastStatusID, conversation1Status2.ID)

		// conversation2 should be the second conversation returned.
		apiConversation2 := resp.Items[1].(*apimodel.Conversation)
		suite.Equal(conversation2.ID, apiConversation2.ID)
	}
}
