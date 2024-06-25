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

	// GetConversationsByOwnerAccountID gets all conversations owned by the given account, with optional paging.
	GetConversationsByOwnerAccountID(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Conversation, error)

	// PutConversation creates or updates a conversation.
	PutConversation(ctx context.Context, conversation *gtsmodel.Conversation, columns ...string) error

	// DeleteConversationByID deletes a conversation, removing it from the owning account's conversation list.
	DeleteConversationByID(ctx context.Context, id string) error

	MigrateConversations(ctx context.Context, migrateStatus func(ctx context.Context, status *gtsmodel.Status) error) error

	// TODO: (Vyr) delete conversations when thread is deleted
	// TODO: (Vyr) delete conversations when owning account is deleted
	// TODO: (Vyr) delete conversations when a participating account is deleted
	// TODO: (Vyr) null out conversation last status when that status is deleted (not currently possible, isn't nullable)
	// TODO: (Vyr) or: replace status with its parent, or delete conversation if parent not available
}
