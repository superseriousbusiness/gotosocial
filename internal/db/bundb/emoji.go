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
	"database/sql"
	"slices"
	"strings"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type emojiDB struct {
	db    *bun.DB
	state *state.State
}

func (e *emojiDB) PutEmoji(ctx context.Context, emoji *gtsmodel.Emoji) error {
	return e.state.Caches.DB.Emoji.Store(emoji, func() error {
		_, err := e.db.NewInsert().Model(emoji).Exec(ctx)
		return err
	})
}

func (e *emojiDB) UpdateEmoji(ctx context.Context, emoji *gtsmodel.Emoji, columns ...string) error {
	emoji.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	// Update the emoji model in the database.
	return e.state.Caches.DB.Emoji.Store(emoji, func() error {
		_, err := e.db.
			NewUpdate().
			Model(emoji).
			Where("? = ?", bun.Ident("emoji.id"), emoji.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (e *emojiDB) DeleteEmojiByID(ctx context.Context, id string) error {
	var (
		// Gather necessary fields from
		// deleted for cache invaliation.
		accountIDs []string
		statusIDs  []string
	)

	// Delete the emoji and all related links to it in a singular transaction.
	if err := e.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// Delete relational links between this emoji
		// and any statuses using it, returning the
		// status IDs so we can later update them.
		if _, err := tx.NewDelete().
			Table("status_to_emojis").
			Where("? = ?", bun.Ident("emoji_id"), id).
			Returning("status_id").
			Exec(ctx, &statusIDs); err != nil {
			return err
		}

		// Delete relational links between this emoji
		// and any accounts using it, returning the
		// account IDs so we can later update them.
		if _, err := tx.NewDelete().
			Table("account_to_emojis").
			Where("? = ?", bun.Ident("emoji_id"), id).
			Returning("account_id").
			Exec(ctx, &accountIDs); err != nil {
			return err
		}

		for _, statusID := range statusIDs {
			status := new(gtsmodel.Status)

			// Select status emoji IDs.
			if err := tx.NewSelect().
				Model(status).
				Column("emojis").
				Where("? = ?", bun.Ident("id"), statusID).
				Scan(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}

			// Delete all instances of this
			// emoji ID from status emoji IDs.
			status.EmojiIDs = slices.DeleteFunc(
				status.EmojiIDs,
				func(emojiID string) bool {
					return emojiID == id
				},
			)

			// Update status emoji IDs.
			if _, err := tx.NewUpdate().
				Model(status).
				Where("? = ?", bun.Ident("id"), statusID).
				Column("emojis").
				Exec(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}
		}

		for _, accountID := range accountIDs {
			account := new(gtsmodel.Account)

			// Select account emoji IDs.
			if err := tx.NewSelect().
				Model(account).
				Column("emojis").
				Where("? = ?", bun.Ident("id"), accountID).
				Scan(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}

			// Delete all instances of this
			// emoji ID from account emoji IDs.
			account.EmojiIDs = slices.DeleteFunc(
				account.EmojiIDs,
				func(emojiID string) bool {
					return emojiID == id
				},
			)

			// Update account emoji IDs.
			if _, err := tx.NewUpdate().
				Model(account).
				Where("? = ?", bun.Ident("id"), accountID).
				Column("emojis").
				Exec(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}
		}

		// Finally, delete emoji from database.
		if _, err := tx.NewDelete().
			Table("emojis").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Invalidate emoji, and any effected statuses / accounts.
	e.state.Caches.DB.Emoji.Invalidate("ID", id)
	e.state.Caches.DB.Account.InvalidateIDs("ID", accountIDs)
	e.state.Caches.DB.Status.InvalidateIDs("ID", statusIDs)

	return nil
}

func (e *emojiDB) GetEmojisBy(ctx context.Context, domain string, includeDisabled bool, includeEnabled bool, shortcode string, maxShortcodeDomain string, minShortcodeDomain string, limit int) ([]*gtsmodel.Emoji, error) {
	emojiIDs := []string{}

	subQuery := e.db.
		NewSelect().
		ColumnExpr("? AS ?", bun.Ident("emoji.id"), bun.Ident("emoji_ids"))

	// To ensure consistent ordering and make paging possible, we sort not by shortcode
	// but by [shortcode]@[domain]. Because sqlite and postgres have different syntax
	// for concatenation, that means we need to switch here. Depending on which driver
	// is in use, query will look something like this (sqlite):
	//
	//	SELECT
	//		"emoji"."id" AS "emoji_ids",
	//		lower("emoji"."shortcode" || '@' || COALESCE("emoji"."domain", '')) AS "shortcode_domain"
	//	FROM
	//		"emojis" AS "emoji"
	//	ORDER BY
	//		"shortcode_domain" ASC
	//
	// Or like this (postgres):
	//
	//	SELECT
	//		"emoji"."id" AS "emoji_ids",
	//		LOWER(CONCAT("emoji"."shortcode", '@', COALESCE("emoji"."domain", ''))) AS "shortcode_domain"
	//	FROM
	//		"emojis" AS "emoji"
	//	ORDER BY
	//		"shortcode_domain" ASC
	switch e.db.Dialect().Name() {
	case dialect.SQLite:
		subQuery = subQuery.ColumnExpr("LOWER(? || ? || COALESCE(?, ?)) AS ?", bun.Ident("emoji.shortcode"), "@", bun.Ident("emoji.domain"), "", bun.Ident("shortcode_domain"))
	case dialect.PG:
		subQuery = subQuery.ColumnExpr("LOWER(CONCAT(?, ?, COALESCE(?, ?))) AS ?", bun.Ident("emoji.shortcode"), "@", bun.Ident("emoji.domain"), "", bun.Ident("shortcode_domain"))
	default:
		panic("db conn was neither pg not sqlite")
	}

	subQuery = subQuery.TableExpr("? AS ?", bun.Ident("emojis"), bun.Ident("emoji"))

	if domain == "" {
		subQuery = subQuery.Where("? IS NULL", bun.Ident("emoji.domain"))
	} else if domain != db.EmojiAllDomains {
		subQuery = subQuery.Where("? = ?", bun.Ident("emoji.domain"), domain)
	}

	switch {
	case includeDisabled && !includeEnabled:
		// show only disabled emojis
		subQuery = subQuery.Where("? = ?", bun.Ident("emoji.disabled"), true)
	case includeEnabled && !includeDisabled:
		// show only enabled emojis
		subQuery = subQuery.Where("? = ?", bun.Ident("emoji.disabled"), false)
	default:
		// show emojis regardless of emoji.disabled value
	}

	if shortcode != "" {
		subQuery = subQuery.Where("LOWER(?) = LOWER(?)", bun.Ident("emoji.shortcode"), shortcode)
	}

	// assume we want to sort ASC (a-z) unless informed otherwise
	order := "ASC"

	if maxShortcodeDomain != "" {
		subQuery = subQuery.Where("? > LOWER(?)", bun.Ident("shortcode_domain"), maxShortcodeDomain)
	}

	if minShortcodeDomain != "" {
		subQuery = subQuery.Where("? < LOWER(?)", bun.Ident("shortcode_domain"), minShortcodeDomain)
		// if we have a minShortcodeDomain we're paging upwards/backwards
		order = "DESC"
	}

	subQuery = subQuery.Order("shortcode_domain " + order)

	if limit > 0 {
		subQuery = subQuery.Limit(limit)
	}

	// Wrap the subQuery in a query, since we don't need to select the shortcode_domain column.
	//
	// The final query will come out looking something like...
	//
	//	SELECT
	//		"subquery"."emoji_ids"
	//	FROM (
	//		SELECT
	//			"emoji"."id" AS "emoji_ids",
	//			LOWER("emoji"."shortcode" || '@' || COALESCE("emoji"."domain", '')) AS "shortcode_domain"
	//		FROM
	//			"emojis" AS "emoji"
	//		ORDER BY
	//			"shortcode_domain" ASC
	//	) AS "subquery"
	if err := e.db.
		NewSelect().
		Column("subquery.emoji_ids").
		TableExpr("(?) AS ?", subQuery, bun.Ident("subquery")).
		Scan(ctx, &emojiIDs); err != nil {
		return nil, err
	}

	if order == "DESC" {
		// Reverse the slice order so the caller still
		// gets emojis in expected a-z alphabetical order.
		//
		// See https://github.com/golang/go/wiki/SliceTricks#reversing
		for i := len(emojiIDs)/2 - 1; i >= 0; i-- {
			opp := len(emojiIDs) - 1 - i
			emojiIDs[i], emojiIDs[opp] = emojiIDs[opp], emojiIDs[i]
		}
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetEmojis(ctx context.Context, page *paging.Page) ([]*gtsmodel.Emoji, error) {
	maxID := page.GetMax()
	limit := page.GetLimit()

	emojiIDs := make([]string, 0, limit)

	q := e.db.NewSelect().
		Table("emojis").
		Column("id").
		Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &emojiIDs); err != nil {
		return nil, err
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetRemoteEmojis(ctx context.Context, page *paging.Page) ([]*gtsmodel.Emoji, error) {
	maxID := page.GetMax()
	limit := page.GetLimit()

	emojiIDs := make([]string, 0, limit)

	q := e.db.NewSelect().
		Table("emojis").
		Column("id").
		Where("domain IS NOT NULL").
		Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &emojiIDs); err != nil {
		return nil, err
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetCachedEmojisOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.Emoji, error) {
	var emojiIDs []string

	q := e.db.NewSelect().
		Table("emojis").
		Column("id").
		Where("cached = true").
		Where("domain IS NOT NULL").
		Where("created_at < ?", olderThan).
		Order("created_at DESC")

	if limit != 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &emojiIDs); err != nil {
		return nil, err
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetUseableEmojis(ctx context.Context) ([]*gtsmodel.Emoji, error) {
	emojiIDs := []string{}

	q := e.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("emojis"), bun.Ident("emoji")).
		Column("emoji.id").
		Where("? = ?", bun.Ident("emoji.visible_in_picker"), true).
		Where("? = ?", bun.Ident("emoji.disabled"), false).
		Where("? IS NULL", bun.Ident("emoji.domain")).
		Order("emoji.shortcode ASC")

	if err := q.Scan(ctx, &emojiIDs); err != nil {
		return nil, err
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetEmojiByID(ctx context.Context, id string) (*gtsmodel.Emoji, error) {
	return e.getEmoji(
		ctx,
		"ID",
		func(emoji *gtsmodel.Emoji) error {
			return e.db.
				NewSelect().
				Model(emoji).
				Where("? = ?", bun.Ident("emoji.id"), id).Scan(ctx)
		},
		id,
	)
}

func (e *emojiDB) GetEmojiByURI(ctx context.Context, uri string) (*gtsmodel.Emoji, error) {
	return e.getEmoji(
		ctx,
		"URI",
		func(emoji *gtsmodel.Emoji) error {
			return e.db.
				NewSelect().
				Model(emoji).
				Where("? = ?", bun.Ident("emoji.uri"), uri).Scan(ctx)
		},
		uri,
	)
}

func (e *emojiDB) GetEmojiByShortcodeDomain(ctx context.Context, shortcode string, domain string) (*gtsmodel.Emoji, error) {
	return e.getEmoji(
		ctx,
		"Shortcode,Domain",
		func(emoji *gtsmodel.Emoji) error {
			q := e.db.
				NewSelect().
				Model(emoji)

			if domain != "" {
				q = q.Where("? = ?", bun.Ident("emoji.shortcode"), shortcode)
				q = q.Where("? = ?", bun.Ident("emoji.domain"), domain)
			} else {
				q = q.Where("? = ?", bun.Ident("emoji.shortcode"), strings.ToLower(shortcode))
				q = q.Where("? IS NULL", bun.Ident("emoji.domain"))
			}

			return q.Scan(ctx)
		},
		shortcode,
		domain,
	)
}

func (e *emojiDB) GetEmojiByStaticURL(ctx context.Context, imageStaticURL string) (*gtsmodel.Emoji, error) {
	return e.getEmoji(
		ctx,
		"ImageStaticURL",
		func(emoji *gtsmodel.Emoji) error {
			return e.db.
				NewSelect().
				Model(emoji).
				Where("? = ?", bun.Ident("emoji.image_static_url"), imageStaticURL).
				Scan(ctx)
		},
		imageStaticURL,
	)
}

func (e *emojiDB) PutEmojiCategory(ctx context.Context, emojiCategory *gtsmodel.EmojiCategory) error {
	return e.state.Caches.DB.EmojiCategory.Store(emojiCategory, func() error {
		_, err := e.db.NewInsert().Model(emojiCategory).Exec(ctx)
		return err
	})
}

func (e *emojiDB) GetEmojiCategories(ctx context.Context) ([]*gtsmodel.EmojiCategory, error) {
	emojiCategoryIDs := []string{}

	q := e.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("emoji_categories"), bun.Ident("emoji_category")).
		Column("emoji_category.id").
		Order("emoji_category.name ASC")

	if err := q.Scan(ctx, &emojiCategoryIDs); err != nil {
		return nil, err
	}

	return e.GetEmojiCategoriesByIDs(ctx, emojiCategoryIDs)
}

func (e *emojiDB) GetEmojiCategory(ctx context.Context, id string) (*gtsmodel.EmojiCategory, error) {
	return e.getEmojiCategory(
		ctx,
		"ID",
		func(emojiCategory *gtsmodel.EmojiCategory) error {
			return e.db.
				NewSelect().
				Model(emojiCategory).
				Where("? = ?", bun.Ident("emoji_category.id"), id).Scan(ctx)
		},
		id,
	)
}

func (e *emojiDB) GetEmojiCategoryByName(ctx context.Context, name string) (*gtsmodel.EmojiCategory, error) {
	return e.getEmojiCategory(
		ctx,
		"Name",
		func(emojiCategory *gtsmodel.EmojiCategory) error {
			return e.db.
				NewSelect().
				Model(emojiCategory).
				Where("LOWER(?) = ?", bun.Ident("emoji_category.name"), strings.ToLower(name)).Scan(ctx)
		},
		name,
	)
}

func (e *emojiDB) getEmoji(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Emoji) error, keyParts ...any) (*gtsmodel.Emoji, error) {
	// Fetch emoji from database cache with loader callback
	emoji, err := e.state.Caches.DB.Emoji.LoadOne(lookup, func() (*gtsmodel.Emoji, error) {
		var emoji gtsmodel.Emoji

		// Not cached! Perform database query
		if err := dbQuery(&emoji); err != nil {
			return nil, err
		}

		return &emoji, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return emoji, nil
	}

	if emoji.CategoryID != "" {
		emoji.Category, err = e.GetEmojiCategory(ctx, emoji.CategoryID)
		if err != nil {
			log.Errorf(ctx, "error getting emoji category %s: %v", emoji.CategoryID, err)
		}
	}

	return emoji, nil
}

func (e *emojiDB) PopulateEmoji(ctx context.Context, emoji *gtsmodel.Emoji) error {
	var (
		errs = gtserror.NewMultiError(1)
		err  error
	)

	if emoji.CategoryID != "" && emoji.Category == nil {
		emoji.Category, err = e.GetEmojiCategory(
			ctx, // these are already barebones
			emoji.CategoryID,
		)
		if err != nil {
			errs.Appendf("error populating emoji category: %w", err)
		}
	}

	return errs.Combine()
}

func (e *emojiDB) GetEmojisByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Emoji, error) {
	if len(ids) == 0 {
		return nil, db.ErrNoEntries
	}

	// Load all emoji IDs via cache loader callbacks.
	emojis, err := e.state.Caches.DB.Emoji.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Emoji, error) {
			// Preallocate expected length of uncached emojis.
			emojis := make([]*gtsmodel.Emoji, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := e.db.NewSelect().
				Model(&emojis).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return emojis, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the emojis by their
	// IDs to ensure in correct order.
	getID := func(e *gtsmodel.Emoji) string { return e.ID }
	xslices.OrderBy(emojis, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return emojis, nil
	}

	// Populate all loaded emojis, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	emojis = slices.DeleteFunc(emojis, func(emoji *gtsmodel.Emoji) bool {
		if err := e.PopulateEmoji(ctx, emoji); err != nil {
			log.Errorf(ctx, "error populating emoji %s: %v", emoji.ID, err)
			return true
		}
		return false
	})

	return emojis, nil
}

func (e *emojiDB) getEmojiCategory(ctx context.Context, lookup string, dbQuery func(*gtsmodel.EmojiCategory) error, keyParts ...any) (*gtsmodel.EmojiCategory, error) {
	return e.state.Caches.DB.EmojiCategory.LoadOne(lookup, func() (*gtsmodel.EmojiCategory, error) {
		var category gtsmodel.EmojiCategory

		// Not cached! Perform database query
		if err := dbQuery(&category); err != nil {
			return nil, err
		}

		return &category, nil
	}, keyParts...)
}

func (e *emojiDB) GetEmojiCategoriesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.EmojiCategory, error) {
	if len(ids) == 0 {
		return nil, db.ErrNoEntries
	}

	// Load all category IDs via cache loader callbacks.
	categories, err := e.state.Caches.DB.EmojiCategory.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.EmojiCategory, error) {
			// Preallocate expected length of uncached categories.
			categories := make([]*gtsmodel.EmojiCategory, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := e.db.NewSelect().
				Model(&categories).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return categories, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the categories by their
	// IDs to ensure in correct order.
	getID := func(c *gtsmodel.EmojiCategory) string { return c.ID }
	xslices.OrderBy(categories, ids, getID)

	return categories, nil
}
