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

package bundb

import (
	"context"
	"errors"
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type conversationDB struct {
	db    *bun.DB
	state *state.State
}

func (c *conversationDB) GetConversationByID(ctx context.Context, id string) (*gtsmodel.Conversation, error) {
	return c.getConversation(
		ctx,
		"ID",
		func(conversation *gtsmodel.Conversation) error {
			return c.db.
				NewSelect().
				Model(conversation).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (c *conversationDB) GetConversationByThreadAndAccountIDs(ctx context.Context, threadID string, accountID string, otherAccountIDs []string) (*gtsmodel.Conversation, error) {
	otherAccountsKey := gtsmodel.ConversationOtherAccountsKey(otherAccountIDs)
	return c.getConversation(
		ctx,
		"ThreadID,AccountID,OtherAccountsKey",
		func(conversation *gtsmodel.Conversation) error {
			return c.db.
				NewSelect().
				Model(conversation).
				Where("? = ?", bun.Ident("thread_id"), threadID).
				Where("? = ?", bun.Ident("account_id"), accountID).
				Where("? = ?", bun.Ident("other_accounts_key"), otherAccountsKey).
				Scan(ctx)
		},
		threadID,
		accountID,
		otherAccountsKey,
	)
}

func (c *conversationDB) getConversation(
	ctx context.Context,
	lookup string,
	dbQuery func(conversation *gtsmodel.Conversation) error,
	keyParts ...any,
) (*gtsmodel.Conversation, error) {
	// Fetch conversation from cache with loader callback
	conversation, err := c.state.Caches.GTS.Conversation.LoadOne(lookup, func() (*gtsmodel.Conversation, error) {
		var conversation gtsmodel.Conversation

		// Not cached! Perform database query
		if err := dbQuery(&conversation); err != nil {
			return nil, err
		}

		return &conversation, nil
	}, keyParts...)
	if err != nil {
		// already processe
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return conversation, nil
	}

	if err := c.populateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	return conversation, nil
}

func (c *conversationDB) populateConversation(ctx context.Context, conversation *gtsmodel.Conversation) error {
	var (
		errs gtserror.MultiError
		err  error
	)

	if conversation.Account == nil {
		conversation.Account, err = c.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			conversation.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating conversation owner account: %w", err)
		}
	}

	if conversation.OtherAccounts == nil {
		conversation.OtherAccounts, err = c.state.DB.GetAccountsByIDs(
			gtscontext.SetBarebones(ctx),
			conversation.OtherAccountIDs,
		)
		if err != nil {
			errs.Appendf("error populating other conversation accounts: %w", err)
		}
	}

	// TODO: (Vyr) threads are currently not used for anything other than lookup by ID

	if conversation.LastStatus == nil {
		conversation.LastStatus, err = c.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			conversation.LastStatusID,
		)
		if err != nil {
			errs.Appendf("error populating conversation last status: %w", err)
		}
	}

	return errs.Combine()
}

func (c *conversationDB) GetConversationsByOwnerAccountID(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Conversation, error) {
	conversationIDs, err := c.getAccountConversationIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return c.getConversationsByIDs(ctx, conversationIDs)
}

func (c *conversationDB) getAccountConversationIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&c.state.Caches.GTS.ConversationIDs, accountID, page, func() ([]string, error) {
		var conversationIDs []string

		// Conversation IDs not in cache. Perform DB query.
		if _, err := c.db.
			NewSelect().
			TableExpr("?", bun.Ident("conversations")).
			ColumnExpr("?", bun.Ident("id")).
			Where("? = ?", bun.Ident("account_id"), accountID).
			OrderExpr("? DESC", bun.Ident("id")).
			Exec(ctx, &conversationIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return conversationIDs, nil
	})
}

func (c *conversationDB) getConversationsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Conversation, error) {
	// Load all conversation IDs via cache loader callbacks.
	conversations, err := c.state.Caches.GTS.Conversation.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Conversation, error) {
			// Preallocate expected length of uncached conversations.
			conversations := make([]*gtsmodel.Conversation, 0, len(uncached))

			// Perform database query scanning the remaining (uncached) IDs.
			if err := c.db.NewSelect().
				Model(&conversations).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return conversations, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the conversations by their IDs to ensure in correct order.
	getID := func(b *gtsmodel.Conversation) string { return b.ID }
	util.OrderBy(conversations, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return conversations, nil
	}

	// Populate all loaded conversations, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	conversations = slices.DeleteFunc(conversations, func(conversation *gtsmodel.Conversation) bool {
		if err := c.populateConversation(ctx, conversation); err != nil {
			log.Errorf(ctx, "error populating conversation %s: %v", conversation.ID, err)
			return true
		}
		return false
	})

	return conversations, nil
}

func (c *conversationDB) PutConversation(ctx context.Context, conversation *gtsmodel.Conversation, columns ...string) error {
	return c.state.Caches.GTS.Conversation.Store(conversation, func() error {
		_, err := NewUpsert(c.db).
			Model(conversation).
			Constraint("id").
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (c *conversationDB) DeleteConversationByID(ctx context.Context, id string) error {
	// Load conversation into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := c.GetConversationByID(gtscontext.SetBarebones(ctx), id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// not an issue.
			err = nil
		}
		return err
	}

	// Drop this now-cached conversation on return after delete.
	defer c.state.Caches.GTS.Conversation.Invalidate("ID", id)

	// Finally delete conversation from DB.
	_, err = c.db.NewDelete().
		Model((*gtsmodel.Conversation)(nil)).
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return err
}

func (c *conversationDB) MigrateConversations(ctx context.Context, migrateStatus func(ctx context.Context, status *gtsmodel.Status) error) error {
	log.Info(ctx, "migrating DMs to conversations…")

	batchSize := 100
	statuses := make([]*gtsmodel.Status, 0, batchSize)
	minID := id.Lowest
	for {
		if err := c.db.
			NewSelect().
			Model(&statuses).
			Where("? = ?", bun.Ident("visibility"), gtsmodel.VisibilityDirect).
			Where("? > ?", bun.Ident("id"), minID).
			Order("id ASC").
			Limit(batchSize).
			Scan(ctx); err != nil {
			return err
		}

		if len(statuses) == 0 {
			break
		}
		log.Infof(ctx, "migrating %d DMs starting past %s", len(statuses), minID)
		minID = statuses[len(statuses)-1].ID

		for _, status := range statuses {
			if err := c.state.DB.PopulateStatus(ctx, status); err != nil {
				log.Errorf(ctx, "couldn't populate DM %s for conversation migration: %v", status.ID, err)
				continue
			}
			if err := migrateStatus(ctx, status); err != nil {
				log.Errorf(ctx, "couldn't process DM %s for conversation migration: %v", status.ID, err)
				continue
			}
		}
	}

	log.Info(ctx, "finished migrating DMs to conversations.")

	return nil
}
