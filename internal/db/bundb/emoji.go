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
	"errors"
	"strings"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type emojiDB struct {
	conn  *DBConn
	state *state.State
}

func (e *emojiDB) PutEmoji(ctx context.Context, emoji *gtsmodel.Emoji) error {
	return e.state.Caches.GTS.Emoji().Store(emoji, func() error {
		_, err := e.conn.NewInsert().Model(emoji).Exec(ctx)
		return e.conn.ProcessError(err)
	})
}

func (e *emojiDB) UpdateEmoji(ctx context.Context, emoji *gtsmodel.Emoji, columns ...string) error {
	emoji.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	// Update the emoji model in the database.
	return e.state.Caches.GTS.Emoji().Store(emoji, func() error {
		_, err := e.conn.
			NewUpdate().
			Model(emoji).
			Where("? = ?", bun.Ident("emoji.id"), emoji.ID).
			Column(columns...).
			Exec(ctx)
		return e.conn.ProcessError(err)
	})
}

func (e *emojiDB) DeleteEmojiByID(ctx context.Context, id string) error {
	var (
		accountIDs []string
		statusIDs  []string
	)

	defer func() {
		// Invalidate cached emoji.
		e.state.Caches.GTS.
			Emoji().
			Invalidate("ID", id)

		for _, id := range accountIDs {
			// Invalidate cached account.
			e.state.Caches.GTS.
				Account().
				Invalidate("ID", id)
		}

		for _, id := range statusIDs {
			// Invalidate cached account.
			e.state.Caches.GTS.
				Status().
				Invalidate("ID", id)
		}
	}()

	// Load emoji into cache before attempting a delete,
	// as we need it cached in order to trigger the invalidate
	// callback. This in turn invalidates others.
	_, err := e.GetEmojiByID(
		gtscontext.SetBarebones(ctx),
		id,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// NOTE: even if db.ErrNoEntries is returned, we
		// still run the below transaction to ensure related
		// objects are appropriately deleted.
		return err
	}

	return e.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// delete links between this emoji and any statuses that use it
		// TODO: remove when we delete this table
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("status_to_emojis"), bun.Ident("status_to_emoji")).
			Where("? = ?", bun.Ident("status_to_emoji.emoji_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// delete links between this emoji and any accounts that use it
		// TODO: remove when we delete this table
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("account_to_emojis"), bun.Ident("account_to_emoji")).
			Where("? = ?", bun.Ident("account_to_emoji.emoji_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// Prepare SELECT accounts query.
		aq := tx.NewSelect().
			Table("accounts").
			Column("id")

		// Append a WHERE LIKE clause to the query
		// that checks the `emoji` column for any
		// text containing this specific emoji ID.
		//
		// (see GetStatusesUsingEmoji() for details.)
		aq = whereLike(aq, "emojis", id)

		// Select all accounts using this emoji into accountIDss.
		if _, err := aq.Exec(ctx, &accountIDs); err != nil {
			return err
		}

		for _, id := range accountIDs {
			var emojiIDs []string

			// Select account with ID.
			if _, err := tx.NewSelect().
				Table("accounts").
				Column("emojis").
				Where("id = ?", id).
				Exec(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}

			// Drop ID from account emojis.
			emojiIDs = dropID(emojiIDs, id)

			// Update account emoji IDs.
			if _, err := tx.NewUpdate().
				Table("accounts").
				Where("id = ?", id).
				Set("emojis = ?", emojiIDs).
				Exec(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}
		}

		// Prepare SELECT statuses query.
		sq := tx.NewSelect().
			Table("statuses").
			Column("id")

		// Append a WHERE LIKE clause to the query
		// that checks the `emoji` column for any
		// text containing this specific emoji ID.
		//
		// (see GetStatusesUsingEmoji() for details.)
		sq = whereLike(sq, "emojis", id)

		// Select all statuses using this emoji into statusIDs.
		if _, err := sq.Exec(ctx, &statusIDs); err != nil {
			return err
		}

		for _, id := range statusIDs {
			var emojiIDs []string

			// Select statuses with ID.
			if _, err := tx.NewSelect().
				Table("statuses").
				Column("emojis").
				Where("id = ?", id).
				Exec(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}

			// Drop ID from account emojis.
			emojiIDs = dropID(emojiIDs, id)

			// Update status emoji IDs.
			if _, err := tx.NewUpdate().
				Table("statuses").
				Where("id = ?", id).
				Set("emojis = ?", emojiIDs).
				Exec(ctx); err != nil &&
				err != sql.ErrNoRows {
				return err
			}
		}

		// Delete emoji from database.
		if _, err := tx.NewDelete().
			Table("emojis").
			Where("id = ?", id).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
}

func (e *emojiDB) GetEmojisBy(ctx context.Context, domain string, includeDisabled bool, includeEnabled bool, shortcode string, maxShortcodeDomain string, minShortcodeDomain string, limit int) ([]*gtsmodel.Emoji, error) {
	emojiIDs := []string{}

	subQuery := e.conn.
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
	switch e.conn.Dialect().Name() {
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
	if err := e.conn.
		NewSelect().
		Column("subquery.emoji_ids").
		TableExpr("(?) AS ?", subQuery, bun.Ident("subquery")).
		Scan(ctx, &emojiIDs); err != nil {
		return nil, e.conn.ProcessError(err)
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

func (e *emojiDB) GetEmojis(ctx context.Context, maxID string, limit int) ([]*gtsmodel.Emoji, error) {
	var emojiIDs []string

	q := e.conn.NewSelect().
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
		return nil, e.conn.ProcessError(err)
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetRemoteEmojis(ctx context.Context, maxID string, limit int) ([]*gtsmodel.Emoji, error) {
	var emojiIDs []string

	q := e.conn.NewSelect().
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
		return nil, e.conn.ProcessError(err)
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetCachedEmojisOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.Emoji, error) {
	var emojiIDs []string

	q := e.conn.NewSelect().
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
		return nil, e.conn.ProcessError(err)
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetUseableEmojis(ctx context.Context) ([]*gtsmodel.Emoji, error) {
	emojiIDs := []string{}

	q := e.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("emojis"), bun.Ident("emoji")).
		Column("emoji.id").
		Where("? = ?", bun.Ident("emoji.visible_in_picker"), true).
		Where("? = ?", bun.Ident("emoji.disabled"), false).
		Where("? IS NULL", bun.Ident("emoji.domain")).
		Order("emoji.shortcode ASC")

	if err := q.Scan(ctx, &emojiIDs); err != nil {
		return nil, e.conn.ProcessError(err)
	}

	return e.GetEmojisByIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetEmojiByID(ctx context.Context, id string) (*gtsmodel.Emoji, error) {
	return e.getEmoji(
		ctx,
		"ID",
		func(emoji *gtsmodel.Emoji) error {
			return e.conn.
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
			return e.conn.
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
		"Shortcode.Domain",
		func(emoji *gtsmodel.Emoji) error {
			q := e.conn.
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
			return e.conn.
				NewSelect().
				Model(emoji).
				Where("? = ?", bun.Ident("emoji.image_static_url"), imageStaticURL).
				Scan(ctx)
		},
		imageStaticURL,
	)
}

func (e *emojiDB) PutEmojiCategory(ctx context.Context, emojiCategory *gtsmodel.EmojiCategory) error {
	return e.state.Caches.GTS.EmojiCategory().Store(emojiCategory, func() error {
		_, err := e.conn.NewInsert().Model(emojiCategory).Exec(ctx)
		return e.conn.ProcessError(err)
	})
}

func (e *emojiDB) GetEmojiCategories(ctx context.Context) ([]*gtsmodel.EmojiCategory, error) {
	emojiCategoryIDs := []string{}

	q := e.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("emoji_categories"), bun.Ident("emoji_category")).
		Column("emoji_category.id").
		Order("emoji_category.name ASC")

	if err := q.Scan(ctx, &emojiCategoryIDs); err != nil {
		return nil, e.conn.ProcessError(err)
	}

	return e.GetEmojiCategoriesByIDs(ctx, emojiCategoryIDs)
}

func (e *emojiDB) GetEmojiCategory(ctx context.Context, id string) (*gtsmodel.EmojiCategory, error) {
	return e.getEmojiCategory(
		ctx,
		"ID",
		func(emojiCategory *gtsmodel.EmojiCategory) error {
			return e.conn.
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
			return e.conn.
				NewSelect().
				Model(emojiCategory).
				Where("LOWER(?) = ?", bun.Ident("emoji_category.name"), strings.ToLower(name)).Scan(ctx)
		},
		name,
	)
}

func (e *emojiDB) getEmoji(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Emoji) error, keyParts ...any) (*gtsmodel.Emoji, error) {
	// Fetch emoji from database cache with loader callback
	emoji, err := e.state.Caches.GTS.Emoji().Load(lookup, func() (*gtsmodel.Emoji, error) {
		var emoji gtsmodel.Emoji

		// Not cached! Perform database query
		if err := dbQuery(&emoji); err != nil {
			return nil, e.conn.ProcessError(err)
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

func (e *emojiDB) GetEmojisByIDs(ctx context.Context, emojiIDs []string) ([]*gtsmodel.Emoji, error) {
	if len(emojiIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	emojis := make([]*gtsmodel.Emoji, 0, len(emojiIDs))

	for _, id := range emojiIDs {
		emoji, err := e.GetEmojiByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "emojisFromIDs: error getting emoji %q: %v", id, err)
			continue
		}

		emojis = append(emojis, emoji)
	}

	return emojis, nil
}

func (e *emojiDB) getEmojiCategory(ctx context.Context, lookup string, dbQuery func(*gtsmodel.EmojiCategory) error, keyParts ...any) (*gtsmodel.EmojiCategory, error) {
	return e.state.Caches.GTS.EmojiCategory().Load(lookup, func() (*gtsmodel.EmojiCategory, error) {
		var category gtsmodel.EmojiCategory

		// Not cached! Perform database query
		if err := dbQuery(&category); err != nil {
			return nil, e.conn.ProcessError(err)
		}

		return &category, nil
	}, keyParts...)
}

func (e *emojiDB) GetEmojiCategoriesByIDs(ctx context.Context, emojiCategoryIDs []string) ([]*gtsmodel.EmojiCategory, error) {
	if len(emojiCategoryIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	emojiCategories := make([]*gtsmodel.EmojiCategory, 0, len(emojiCategoryIDs))

	for _, id := range emojiCategoryIDs {
		emojiCategory, err := e.GetEmojiCategory(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting emoji category %q: %v", id, err)
			continue
		}

		emojiCategories = append(emojiCategories, emojiCategory)
	}

	return emojiCategories, nil
}

// dropIDs drops given ID string from IDs slice.
func dropID(ids []string, id string) []string {
	for i := 0; i < len(ids); {
		if ids[i] == id {
			// Remove this reference.
			copy(ids[i:], ids[i+1:])
			ids = ids[:len(ids)-1]
			continue
		}
		i++
	}
	return ids
}
