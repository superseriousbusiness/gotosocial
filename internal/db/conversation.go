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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

type Conversation interface {
	// GetConversationByID gets a single conversation by ID.
	GetConversationByID(ctx context.Context, id string) (*gtsmodel.Conversation, error)

	// GetConversationByThreadAndAccountIDs retrieves a conversation by thread ID and participant account IDs, if it exists.
	GetConversationByThreadAndAccountIDs(ctx context.Context, threadID string, accountID string, otherAccountIDs []string) (*gtsmodel.Conversation, error)

	// GetConversationsByOwnerAccountID gets all conversations owned by the given account,
	// with optional paging based on last status ID.
	GetConversationsByOwnerAccountID(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Conversation, error)

	// UpsertConversation creates or updates a conversation.
	UpsertConversation(ctx context.Context, conversation *gtsmodel.Conversation, columns ...string) error

	// LinkConversationToStatus creates a conversation-to-status link.
	LinkConversationToStatus(ctx context.Context, statusID string, conversationID string) error

	// DeleteConversationByID deletes a conversation, removing it from the owning account's conversation list.
	DeleteConversationByID(ctx context.Context, id string) error

	// DeleteConversationsByOwnerAccountID deletes all conversations owned by the given account.
	DeleteConversationsByOwnerAccountID(ctx context.Context, accountID string) error

	// DeleteStatusFromConversations handles when a status is deleted by updating or deleting conversations for which it was the last status.
	DeleteStatusFromConversations(ctx context.Context, statusID string) error
}
