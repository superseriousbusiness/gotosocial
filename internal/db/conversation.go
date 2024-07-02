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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

type Conversation interface {
	// GetConversationByID gets a single conversation by ID.
	GetConversationByID(ctx context.Context, id string) (*gtsmodel.Conversation, error)

	// GetConversationByThreadAndAccountIDs retrieves a conversation by thread ID and participant account IDs, if it exists.
	GetConversationByThreadAndAccountIDs(ctx context.Context, threadID string, accountID string, otherAccountIDs []string) (*gtsmodel.Conversation, error)

	// GetConversationsByOwnerAccountID gets all conversations owned by the given account,
	// with optional paging based on last status ID.
	GetConversationsByOwnerAccountID(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Conversation, error)

	// UpdateConversation updates an existing conversation.
	UpdateConversation(ctx context.Context, conversation *gtsmodel.Conversation, columns ...string) error

	// AddStatusToConversation takes a conversation (which may or may not exist in the DB yet) and a status.
	// It will link the status to the conversation, and if the status is newer than the last status,
	// it will become the last status. This happens in a transaction.
	AddStatusToConversation(ctx context.Context, conversation *gtsmodel.Conversation, status *gtsmodel.Status) (*gtsmodel.Conversation, error)

	// DeleteConversationByID deletes a conversation, removing it from the owning account's conversation list.
	DeleteConversationByID(ctx context.Context, id string) error

	// DeleteConversationsByOwnerAccountID deletes all conversations owned by the given account.
	DeleteConversationsByOwnerAccountID(ctx context.Context, accountID string) error

	// DeleteStatusFromConversations handles when a status is deleted by nulling out the last status for
	// any conversations in which it was the last status.
	DeleteStatusFromConversations(ctx context.Context, statusID string) error

	MigrateConversations(ctx context.Context, migrateStatus func(ctx context.Context, status *gtsmodel.Status) error) error
}
