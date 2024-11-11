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
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

func (r *relationshipDB) IsBlocked(ctx context.Context, sourceAccountID string, targetAccountID string) (bool, error) {
	block, err := r.GetBlock(
		gtscontext.SetBarebones(ctx),
		sourceAccountID,
		targetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return (block != nil), nil
}

func (r *relationshipDB) IsEitherBlocked(ctx context.Context, accountID1 string, accountID2 string) (bool, error) {
	// Look for a block in direction of account1->account2
	b1, err := r.IsBlocked(ctx, accountID1, accountID2)
	if err != nil || b1 {
		return true, err
	}

	// Look for a block in direction of account2->account1
	b2, err := r.IsBlocked(ctx, accountID2, accountID1)
	if err != nil || b2 {
		return true, err
	}

	return false, nil
}

func (r *relationshipDB) GetBlockByID(ctx context.Context, id string) (*gtsmodel.Block, error) {
	return r.getBlock(
		ctx,
		"ID",
		func(block *gtsmodel.Block) error {
			return r.db.NewSelect().Model(block).
				Where("? = ?", bun.Ident("block.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (r *relationshipDB) GetBlockByURI(ctx context.Context, uri string) (*gtsmodel.Block, error) {
	return r.getBlock(
		ctx,
		"URI",
		func(block *gtsmodel.Block) error {
			return r.db.NewSelect().Model(block).
				Where("? = ?", bun.Ident("block.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (r *relationshipDB) GetBlock(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.Block, error) {
	return r.getBlock(
		ctx,
		"AccountID,TargetAccountID",
		func(block *gtsmodel.Block) error {
			return r.db.NewSelect().Model(block).
				Where("? = ?", bun.Ident("block.account_id"), sourceAccountID).
				Where("? = ?", bun.Ident("block.target_account_id"), targetAccountID).
				Scan(ctx)
		},
		sourceAccountID,
		targetAccountID,
	)
}

func (r *relationshipDB) GetBlocksByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Block, error) {
	// Load all blocks IDs via cache loader callbacks.
	blocks, err := r.state.Caches.DB.Block.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Block, error) {
			// Preallocate expected length of uncached blocks.
			blocks := make([]*gtsmodel.Block, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := r.db.NewSelect().
				Model(&blocks).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return blocks, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the blocks by their
	// IDs to ensure in correct order.
	getID := func(b *gtsmodel.Block) string { return b.ID }
	xslices.OrderBy(blocks, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return blocks, nil
	}

	// Populate all loaded blocks, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	blocks = slices.DeleteFunc(blocks, func(block *gtsmodel.Block) bool {
		if err := r.PopulateBlock(ctx, block); err != nil {
			log.Errorf(ctx, "error populating block %s: %v", block.ID, err)
			return true
		}
		return false
	})

	return blocks, nil
}

func (r *relationshipDB) getBlock(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Block) error, keyParts ...any) (*gtsmodel.Block, error) {
	// Fetch block from cache with loader callback
	block, err := r.state.Caches.DB.Block.LoadOne(lookup, func() (*gtsmodel.Block, error) {
		var block gtsmodel.Block

		// Not cached! Perform database query
		if err := dbQuery(&block); err != nil {
			return nil, err
		}

		return &block, nil
	}, keyParts...)
	if err != nil {
		// already processe
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return block, nil
	}

	if err := r.state.DB.PopulateBlock(ctx, block); err != nil {
		return nil, err
	}

	return block, nil
}

func (r *relationshipDB) PopulateBlock(ctx context.Context, block *gtsmodel.Block) error {
	var (
		errs gtserror.MultiError
		err  error
	)

	if block.Account == nil {
		// Block origin account is not set, fetch from database.
		block.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			block.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating block account: %w", err)
		}
	}

	if block.TargetAccount == nil {
		// Block target account is not set, fetch from database.
		block.TargetAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			block.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating block target account: %w", err)
		}
	}

	return errs.Combine()
}

func (r *relationshipDB) PutBlock(ctx context.Context, block *gtsmodel.Block) error {
	return r.state.Caches.DB.Block.Store(block, func() error {
		_, err := r.db.NewInsert().Model(block).Exec(ctx)
		return err
	})
}

func (r *relationshipDB) DeleteBlockByID(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.Block

	// Delete block with given ID,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("id"), id).
		Returning("?, ?",
			bun.Ident("account_id"),
			bun.Ident("target_account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached block with ID, manually
	// call invalidate hook in case not cached.
	r.state.Caches.DB.Block.Invalidate("ID", id)
	r.state.Caches.OnInvalidateBlock(&deleted)

	return nil
}

func (r *relationshipDB) DeleteBlockByURI(ctx context.Context, uri string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.Block

	// Delete block with given URI,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		Where("? = ?", bun.Ident("uri"), uri).
		Returning("?, ?",
			bun.Ident("account_id"),
			bun.Ident("target_account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate cached block with URI, manually
	// call invalidate hook in case not cached.
	r.state.Caches.DB.Block.Invalidate("URI", uri)
	r.state.Caches.OnInvalidateBlock(&deleted)

	return nil
}

func (r *relationshipDB) DeleteAccountBlocks(ctx context.Context, accountID string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted []*gtsmodel.Block

	// Delete all blocks either from
	// account, or targeting account,
	// returning the deleted models.
	if _, err := r.db.NewDelete().
		Model(&deleted).
		WhereOr("? = ? OR ? = ?",
			bun.Ident("account_id"),
			accountID,
			bun.Ident("target_account_id"),
			accountID,
		).
		Returning("?, ?",
			bun.Ident("account_id"),
			bun.Ident("target_account_id"),
		).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate all account's incoming / outoing blocks.
	r.state.Caches.DB.Block.Invalidate("AccountID", accountID)
	r.state.Caches.DB.Block.Invalidate("TargetAccountID", accountID)

	// In case not all blocks were in
	// cache, manually call invalidate hooks.
	for _, block := range deleted {
		r.state.Caches.OnInvalidateBlock(block)
	}

	return nil
}
