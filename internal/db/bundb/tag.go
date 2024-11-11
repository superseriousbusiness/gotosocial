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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type tagDB struct {
	db    *bun.DB
	state *state.State
}

func (t *tagDB) GetTag(ctx context.Context, id string) (*gtsmodel.Tag, error) {
	return t.state.Caches.DB.Tag.LoadOne("ID", func() (*gtsmodel.Tag, error) {
		var tag gtsmodel.Tag

		q := t.db.
			NewSelect().
			Model(&tag).
			Where("? = ?", bun.Ident("tag.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &tag, nil
	}, id)
}

func (t *tagDB) GetTagByName(ctx context.Context, name string) (*gtsmodel.Tag, error) {
	// Normalize 'name' string.
	name = strings.TrimSpace(name)
	name = strings.ToLower(name)

	return t.state.Caches.DB.Tag.LoadOne("Name", func() (*gtsmodel.Tag, error) {
		var tag gtsmodel.Tag

		q := t.db.
			NewSelect().
			Model(&tag).
			Where("? = ?", bun.Ident("tag.name"), name)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &tag, nil
	}, name)
}

func (t *tagDB) GetTags(ctx context.Context, ids []string) ([]*gtsmodel.Tag, error) {
	// Load all tag IDs via cache loader callbacks.
	tags, err := t.state.Caches.DB.Tag.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Tag, error) {
			// Preallocate expected length of uncached tags.
			tags := make([]*gtsmodel.Tag, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := t.db.NewSelect().
				Model(&tags).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return tags, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the tags by their
	// IDs to ensure in correct order.
	getID := func(t *gtsmodel.Tag) string { return t.ID }
	xslices.OrderBy(tags, ids, getID)

	return tags, nil
}

func (t *tagDB) PutTag(ctx context.Context, tag *gtsmodel.Tag) error {
	// Normalize 'name' string before it enters
	// the db, without changing tag we were given.
	//
	// First copy tag to new pointer.
	t2 := new(gtsmodel.Tag)
	*t2 = *tag

	// Normalize name on new pointer.
	t2.Name = strings.TrimSpace(t2.Name)
	t2.Name = strings.ToLower(t2.Name)

	// Insert the copy.
	if err := t.state.Caches.DB.Tag.Store(t2, func() error {
		_, err := t.db.NewInsert().Model(t2).Exec(ctx)
		return err
	}); err != nil {
		return err // err already processed
	}

	// Update original tag with
	// field values populated by db.
	tag.CreatedAt = t2.CreatedAt
	tag.UpdatedAt = t2.UpdatedAt
	tag.Useable = t2.Useable
	tag.Listable = t2.Listable

	return nil
}

func (t *tagDB) GetFollowedTags(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Tag, error) {
	tagIDs, err := t.getTagIDsFollowedByAccount(ctx, accountID, page)
	if err != nil {
		return nil, err
	}
	return t.GetTags(ctx, tagIDs)
}

func (t *tagDB) getTagIDsFollowedByAccount(ctx context.Context, accountID string, page *paging.Page) ([]string, error) {
	return loadPagedIDs(&t.state.Caches.DB.FollowingTagIDs, ">"+accountID, page, func() ([]string, error) {
		var tagIDs []string

		// Tag IDs not in cache. Perform DB query.
		if _, err := t.db.
			NewSelect().
			Model((*gtsmodel.FollowedTag)(nil)).
			Column("tag_id").
			Where("? = ?", bun.Ident("account_id"), accountID).
			OrderExpr("? DESC", bun.Ident("tag_id")).
			Exec(ctx, &tagIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf("error getting tag IDs followed by account %s: %w", accountID, err)
		}

		return tagIDs, nil
	})
}

func (t *tagDB) getAccountIDsFollowingTag(ctx context.Context, tagID string) ([]string, error) {
	return loadPagedIDs(&t.state.Caches.DB.FollowingTagIDs, "<"+tagID, nil, func() ([]string, error) {
		var accountIDs []string

		// Account IDs not in cache. Perform DB query.
		if _, err := t.db.
			NewSelect().
			Model((*gtsmodel.FollowedTag)(nil)).
			Column("account_id").
			Where("? = ?", bun.Ident("tag_id"), tagID).
			OrderExpr("? DESC", bun.Ident("account_id")).
			Exec(ctx, &accountIDs); // nocollapse
		err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf("error getting account IDs following tag %s: %w", tagID, err)
		}

		return accountIDs, nil
	})
}

func (t *tagDB) IsAccountFollowingTag(ctx context.Context, accountID string, tagID string) (bool, error) {
	followingTagIDs, err := t.getTagIDsFollowedByAccount(ctx, accountID, nil)
	if err != nil {
		return false, err
	}
	return slices.Contains(followingTagIDs, tagID), nil
}

func (t *tagDB) PutFollowedTag(ctx context.Context, accountID string, tagID string) error {
	// Insert the followed tag.
	result, err := t.db.NewInsert().
		Model(&gtsmodel.FollowedTag{
			AccountID: accountID,
			TagID:     tagID,
		}).
		On("CONFLICT (?, ?) DO NOTHING", bun.Ident("account_id"), bun.Ident("tag_id")).
		Exec(ctx)
	if err != nil {
		return gtserror.Newf("error inserting followed tag: %w", err)
	}

	// If it fails because that account already follows that tag, that's fine, and we're done.
	rows, err := result.RowsAffected()
	if err != nil {
		return gtserror.Newf("error getting inserted row count: %w", err)
	}
	if rows == 0 {
		return nil
	}

	// We updated something, invalidate caches.
	t.state.Caches.DB.FollowingTagIDs.Invalidate(

		// tag IDs followed by account
		">"+accountID,

		// account IDs following tag
		"<"+tagID,
	)

	return nil
}

func (t *tagDB) DeleteFollowedTag(ctx context.Context, accountID string, tagID string) error {
	result, err := t.db.NewDelete().
		Model((*gtsmodel.FollowedTag)(nil)).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Where("? = ?", bun.Ident("tag_id"), tagID).
		Exec(ctx)
	if err != nil {
		return gtserror.Newf("error deleting followed tag %s for account %s: %w", tagID, accountID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return gtserror.Newf("error getting inserted row count: %w", err)
	}
	if rows == 0 {
		return nil
	}

	// We deleted something, invalidate caches.
	t.state.Caches.DB.FollowingTagIDs.Invalidate(

		// tag IDs followed by account
		">"+accountID,

		// account IDs following tag
		"<"+tagID,
	)

	return err
}

func (t *tagDB) DeleteFollowedTagsByAccountID(ctx context.Context, accountID string) error {
	// Delete followed tags from the database, returning the list of tag IDs affected.
	tagIDs := []string{}
	if err := t.db.NewDelete().
		Model((*gtsmodel.FollowedTag)(nil)).
		Where("? = ?", bun.Ident("account_id"), accountID).
		Returning("?", bun.Ident("tag_id")).
		Scan(ctx, &tagIDs); // nocollapse
	err != nil {
		return gtserror.Newf("error deleting followed tags for account %s: %w", accountID, err)
	}

	// Convert tag IDs to the keys
	// we use for caching tag follow
	// and following IDs.
	keys := tagIDs
	for i := range keys {
		keys[i] = "<" + keys[i]
	}
	keys = append(keys, ">"+accountID)

	// If we deleted anything, invalidate caches with keys.
	t.state.Caches.DB.FollowingTagIDs.Invalidate(keys...)

	return nil
}

func (t *tagDB) GetAccountIDsFollowingTagIDs(ctx context.Context, tagIDs []string) ([]string, error) {
	// Make conservative estimate for no. accounts.
	accountIDs := make([]string, 0, len(tagIDs))

	// Gather all accounts following tags.
	for _, tagID := range tagIDs {
		tagAccountIDs, err := t.getAccountIDsFollowingTag(ctx, tagID)
		if err != nil {
			return nil, err
		}
		accountIDs = append(accountIDs, tagAccountIDs...)
	}

	// Accounts might be following multiple tags in list,
	// but we only want to return each account once.
	return xslices.Deduplicate(accountIDs), nil
}
