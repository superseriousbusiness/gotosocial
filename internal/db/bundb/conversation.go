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
	"github.com/uptrace/bun/dialect"
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
	conversation, err := c.state.Caches.DB.Conversation.LoadOne(lookup, func() (*gtsmodel.Conversation, error) {
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

	if conversation.LastStatus == nil && conversation.LastStatusID != "" {
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
	conversationLastStatusIDs, err := c.getAccountConversationLastStatusIDs(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return c.getConversationsByLastStatusIDs(ctx, accountID, conversationLastStatusIDs)
}

func (c *conversationDB) getAccountConversationLastStatusIDs(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&c.state.Caches.DB.ConversationLastStatusIDs, accountID, page, func() ([]string, error) {
		var conversationLastStatusIDs []string

		// Conversation last status IDs not in cache. Perform DB query.
		if _, err := c.db.
			NewSelect().
			Model((*gtsmodel.Conversation)(nil)).
			Column("last_status_id").
			Where("? = ?", bun.Ident("account_id"), accountID).
			OrderExpr("? DESC", bun.Ident("last_status_id")).
			Exec(ctx, &conversationLastStatusIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		return conversationLastStatusIDs, nil
	})
}

func (c *conversationDB) getConversationsByLastStatusIDs(
	ctx context.Context,
	accountID string,
	conversationLastStatusIDs []string,
) ([]*gtsmodel.Conversation, error) {
	// Load all conversation IDs via cache loader callbacks.
	conversations, err := c.state.Caches.DB.Conversation.LoadIDs2Part(
		"AccountID,LastStatusID",
		accountID,
		conversationLastStatusIDs,
		func(accountID string, uncached []string) ([]*gtsmodel.Conversation, error) {
			// Preallocate expected length of uncached conversations.
			conversations := make([]*gtsmodel.Conversation, 0, len(uncached))

			// Perform database query scanning the remaining (uncached) IDs.
			if err := c.db.NewSelect().
				Model(&conversations).
				Where("? = ?", bun.Ident("account_id"), accountID).
				Where("? IN (?)", bun.Ident("last_status_id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return conversations, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the conversations by their last status IDs to ensure correct order.
	getID := func(b *gtsmodel.Conversation) string { return b.ID }
	util.OrderBy(conversations, conversationLastStatusIDs, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return conversations, nil
	}

	// Populate all loaded conversations, removing those we fail to populate.
	conversations = slices.DeleteFunc(conversations, func(conversation *gtsmodel.Conversation) bool {
		if err := c.populateConversation(ctx, conversation); err != nil {
			log.Errorf(ctx, "error populating conversation %s: %v", conversation.ID, err)
			return true
		}
		return false
	})

	return conversations, nil
}

func (c *conversationDB) UpsertConversation(ctx context.Context, conversation *gtsmodel.Conversation, columns ...string) error {
	// If we're updating by column, ensure "updated_at" is included.
	if len(columns) > 0 {
		columns = append(columns, "updated_at")
	}

	return c.state.Caches.DB.Conversation.Store(conversation, func() error {
		_, err := NewUpsert(c.db).
			Model(conversation).
			Constraint("id").
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (c *conversationDB) LinkConversationToStatus(ctx context.Context, conversationID string, statusID string) error {
	conversationToStatus := &gtsmodel.ConversationToStatus{
		ConversationID: conversationID,
		StatusID:       statusID,
	}

	if _, err := c.db.NewInsert().
		Model(conversationToStatus).
		Exec(ctx); // nocollapse
	err != nil {
		return err
	}
	return nil
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
	defer c.state.Caches.DB.Conversation.Invalidate("ID", id)

	// Finally delete conversation from DB.
	_, err = c.db.NewDelete().
		Model((*gtsmodel.Conversation)(nil)).
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx)
	return err
}

func (c *conversationDB) DeleteConversationsByOwnerAccountID(ctx context.Context, accountID string) error {
	defer func() {
		// Invalidate any cached conversations and conversation IDs owned by this account on return.
		// Conversation invalidate hooks only invalidate the conversation ID cache,
		// so we don't need to load all conversations into the cache to run invalidation hooks,
		// as with some other object types (blocks, for example).
		c.state.Caches.DB.Conversation.Invalidate("AccountID", accountID)
		// In case there were no cached conversations,
		// explicitly invalidate the user's conversation last status ID cache.
		c.state.Caches.DB.ConversationLastStatusIDs.Invalidate(accountID)
	}()

	return c.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete conversations matching the account ID.
		deletedConversationIDs := []string{}
		if err := tx.NewDelete().
			Model((*gtsmodel.Conversation)(nil)).
			Where("? = ?", bun.Ident("account_id"), accountID).
			Returning("?", bun.Ident("id")).
			Scan(ctx, &deletedConversationIDs); // nocollapse
		err != nil {
			return gtserror.Newf("error deleting conversations for account %s: %w", accountID, err)
		}

		// Delete any conversation-to-status links matching the deleted conversation IDs.
		if _, err := tx.NewDelete().
			Model((*gtsmodel.ConversationToStatus)(nil)).
			Where("? IN (?)", bun.Ident("conversation_id"), bun.In(deletedConversationIDs)).
			Exec(ctx); // nocollapse
		err != nil {
			return gtserror.Newf("error deleting conversation-to-status links for account %s: %w", accountID, err)
		}

		return nil
	})
}

func (c *conversationDB) DeleteStatusFromConversations(ctx context.Context, statusID string) error {
	// SQL returning the current time.
	var nowSQL string
	switch c.db.Dialect().Name() {
	case dialect.SQLite:
		nowSQL = "DATE('now')"
	case dialect.PG:
		nowSQL = "NOW()"
	default:
		log.Panicf(nil, "db conn %s was neither pg nor sqlite", c.db)
	}

	updatedConversationIDs := []string{}
	deletedConversationIDs := []string{}

	if err := c.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// Delete this status from conversation-to-status links.
		if _, err := tx.NewDelete().
			Model((*gtsmodel.ConversationToStatus)(nil)).
			Where("? = ?", bun.Ident("status_id"), statusID).
			Exec(ctx); // nocollapse
		err != nil {
			return gtserror.Newf("error deleting conversation-to-status links while deleting status %s: %w", statusID, err)
		}

		// Note: Bun doesn't currently support CREATE TABLE … AS SELECT … so we need to use raw queries here.

		// Create a temporary table with all statuses other than the deleted status
		// in each conversation for which the deleted status is the last status
		// (if there are such statuses).
		conversationStatusesTempTable := "conversation_statuses_" + id.NewULID()
		if _, err := tx.NewRaw(
			"CREATE TEMPORARY TABLE ? AS ?",
			bun.Ident(conversationStatusesTempTable),
			tx.NewSelect().
				ColumnExpr(
					"? AS ?",
					bun.Ident("conversations.id"),
					bun.Ident("conversation_id"),
				).
				ColumnExpr(
					"? AS ?",
					bun.Ident("conversation_to_statuses.status_id"),
					bun.Ident("id"),
				).
				Column("statuses.created_at").
				Table("conversations").
				Join("LEFT JOIN ?", bun.Ident("conversation_to_statuses")).
				JoinOn(
					"? = ?",
					bun.Ident("conversations.id"),
					bun.Ident("conversation_to_statuses.conversation_id"),
				).
				JoinOn(
					"? != ?",
					bun.Ident("conversation_to_statuses.status_id"),
					statusID,
				).
				Join("LEFT JOIN ?", bun.Ident("statuses")).
				JoinOn(
					"? = ?",
					bun.Ident("conversation_to_statuses.status_id"),
					bun.Ident("statuses.id"),
				).
				Where(
					"? = ?",
					bun.Ident("conversations.last_status_id"),
					statusID,
				),
		).
			Exec(ctx); // nocollapse
		err != nil {
			return gtserror.Newf("error creating conversationStatusesTempTable while deleting status %s: %w", statusID, err)
		}

		// Create a temporary table with the most recently created status in each conversation
		// for which the deleted status is the last status (if there is such a status).
		latestConversationStatusesTempTable := "latest_conversation_statuses_" + id.NewULID()
		if _, err := tx.NewRaw(
			"CREATE TEMPORARY TABLE ? AS ?",
			bun.Ident(latestConversationStatusesTempTable),
			tx.NewSelect().
				Column(
					"conversation_statuses.conversation_id",
					"conversation_statuses.id",
				).
				TableExpr(
					"? AS ?",
					bun.Ident(conversationStatusesTempTable),
					bun.Ident("conversation_statuses"),
				).
				Join(
					"LEFT JOIN ? AS ?",
					bun.Ident(conversationStatusesTempTable),
					bun.Ident("later_statuses"),
				).
				JoinOn(
					"? = ?",
					bun.Ident("conversation_statuses.conversation_id"),
					bun.Ident("later_statuses.conversation_id"),
				).
				JoinOn(
					"? > ?",
					bun.Ident("later_statuses.created_at"),
					bun.Ident("conversation_statuses.created_at"),
				).
				Where("? IS NULL", bun.Ident("later_statuses.id")),
		).
			Exec(ctx); // nocollapse
		err != nil {
			return gtserror.Newf("error creating latestConversationStatusesTempTable while deleting status %s: %w", statusID, err)
		}

		// For every conversation where the given status was the last one,
		// reset its last status to the most recently created in the conversation other than that one,
		// if there is such a status.
		// Return conversation IDs for invalidation.
		if err := tx.NewUpdate().
			Model((*gtsmodel.Conversation)(nil)).
			SetColumn("last_status_id", "?", bun.Ident("latest_conversation_statuses.id")).
			SetColumn("updated_at", "?", bun.Safe(nowSQL)).
			TableExpr("? AS ?", bun.Ident(latestConversationStatusesTempTable), bun.Ident("latest_conversation_statuses")).
			Where("?TableAlias.? = ?", bun.Ident("id"), bun.Ident("latest_conversation_statuses.conversation_id")).
			Where("? IS NOT NULL", bun.Ident("latest_conversation_statuses.id")).
			Returning("?TableName.?", bun.Ident("id")).
			Scan(ctx, &updatedConversationIDs); // nocollapse
		err != nil {
			return gtserror.Newf("error rolling back last status for conversation while deleting status %s: %w", statusID, err)
		}

		// If there is no such status, delete the conversation.
		// Return conversation IDs for invalidation.
		if err := tx.NewDelete().
			Model((*gtsmodel.Conversation)(nil)).
			Where(
				"? IN (?)",
				bun.Ident("id"),
				tx.NewSelect().
					Table(latestConversationStatusesTempTable).
					Column("conversation_id").
					Where("? IS NULL", bun.Ident("id")),
			).
			Returning("?", bun.Ident("id")).
			Scan(ctx, &deletedConversationIDs); // nocollapse
		err != nil {
			return gtserror.Newf("error deleting conversation while deleting status %s: %w", statusID, err)
		}

		// Clean up.
		for _, tempTable := range []string{
			conversationStatusesTempTable,
			latestConversationStatusesTempTable,
		} {
			if _, err := tx.NewDropTable().Table(tempTable).Exec(ctx); err != nil {
				return gtserror.Newf(
					"error dropping temporary table %s after deleting status %s: %w",
					tempTable,
					statusID,
					err,
				)
			}
		}

		return nil
	}); err != nil {
		return err
	}

	updatedConversationIDs = append(updatedConversationIDs, deletedConversationIDs...)
	c.state.Caches.DB.Conversation.InvalidateIDs("ID", updatedConversationIDs)

	return nil
}
