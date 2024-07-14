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
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

func (suite *ConversationsTestSuite) TestGetAll() {
	conversation := suite.NewTestConversation(suite.testAccount, 0)

	resp, err := suite.conversationsProcessor.GetAll(context.Background(), suite.testAccount, nil)
	if suite.NoError(err) && suite.Len(resp.Items, 1) && suite.IsType((*apimodel.Conversation)(nil), resp.Items[0]) {
		apiConversation := resp.Items[0].(*apimodel.Conversation)
		suite.Equal(conversation.ID, apiConversation.ID)
		suite.True(apiConversation.Unread)
	}
}

// Test that conversations with newer last status IDs are returned earlier.
func (suite *ConversationsTestSuite) TestGetAllOrder() {
	// Create a new conversation.
	conversation1 := suite.NewTestConversation(suite.testAccount, 0)

	// Create another new conversation with a last status newer than conversation1's.
	conversation2 := suite.NewTestConversation(suite.testAccount, 1*time.Second)

	// Add an even newer status than that to conversation1.
	conversation1Status2 := suite.NewTestStatus(suite.testAccount, conversation1.LastStatus.ThreadID, 2*time.Second, conversation1.LastStatus)
	conversation1.LastStatusID = conversation1Status2.ID
	if err := suite.db.UpsertConversation(context.Background(), conversation1, "last_status_id"); err != nil {
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
