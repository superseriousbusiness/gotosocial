/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb

import (
	"context"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type emojiDB struct {
	conn  *DBConn
	cache *cache.EmojiCache
}

func (e *emojiDB) newEmojiQ(emoji *gtsmodel.Emoji) *bun.SelectQuery {
	return e.conn.
		NewSelect().
		Model(emoji)
}

func (e *emojiDB) PutEmoji(ctx context.Context, emoji *gtsmodel.Emoji) db.Error {
	if _, err := e.conn.NewInsert().Model(emoji).Exec(ctx); err != nil {
		return e.conn.ProcessError(err)
	}

	e.cache.Put(emoji)
	return nil
}

func (e *emojiDB) GetEmojis(ctx context.Context, domain string, includeDisabled bool, includeEnabled bool, shortcode string) ([]*gtsmodel.Emoji, db.Error) {
	emojiIDs := []string{}

	q := e.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("emojis"), bun.Ident("emoji")).
		Column("emoji.id")

	switch e.conn.Dialect().Name() {
	case dialect.SQLite:
		q = q.ColumnExpr("? || ? || COALESCE(?, ?) AS ?", bun.Ident("emoji.shortcode"), "@", bun.Ident("emoji.domain"), "", bun.Ident("shortcodedomain"))
	case dialect.PG:
		q = q.ColumnExpr("CONCAT(?, ?, COALESCE(?, ?)) AS ?", bun.Ident("emoji.shortcode"), "@", bun.Ident("emoji.domain"), "", bun.Ident("shortcodedomain"))
	default:
		panic("db conn was neither pg not sqlite")
	}

	q = q.Order("shortcodedomain")

	if domain == "" {
		q = q.Where("? IS NULL", bun.Ident("emoji.domain"))
	} else if domain != db.EmojiAllDomains {
		q = q.Where("? = ?", bun.Ident("emoji.domain"), domain)
	}

	switch {
	case includeDisabled && !includeEnabled:
		// show only disabled emojis
		q = q.Where("? = ?", bun.Ident("emoji.disabled"), true)
	case includeEnabled && !includeDisabled:
		// show only enabled emojis
		q = q.Where("? = ?", bun.Ident("emoji.disabled"), false)
	default:
		// show emojis regardless of emoji.disabled value
	}

	if shortcode != "" {
		q = q.Where("? = ?", bun.Ident("emoji.shortcode"), shortcode)
	}

	if err := q.Scan(ctx, &emojiIDs, new([]string)); err != nil {
		return nil, e.conn.ProcessError(err)
	}

	return e.emojisFromIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetUseableCustomEmojis(ctx context.Context) ([]*gtsmodel.Emoji, db.Error) {
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

	return e.emojisFromIDs(ctx, emojiIDs)
}

func (e *emojiDB) GetEmojiByID(ctx context.Context, id string) (*gtsmodel.Emoji, db.Error) {
	return e.getEmoji(
		ctx,
		func() (*gtsmodel.Emoji, bool) {
			return e.cache.GetByID(id)
		},
		func(emoji *gtsmodel.Emoji) error {
			return e.newEmojiQ(emoji).Where("? = ?", bun.Ident("emoji.id"), id).Scan(ctx)
		},
	)
}

func (e *emojiDB) GetEmojiByURI(ctx context.Context, uri string) (*gtsmodel.Emoji, db.Error) {
	return e.getEmoji(
		ctx,
		func() (*gtsmodel.Emoji, bool) {
			return e.cache.GetByURI(uri)
		},
		func(emoji *gtsmodel.Emoji) error {
			return e.newEmojiQ(emoji).Where("? = ?", bun.Ident("emoji.uri"), uri).Scan(ctx)
		},
	)
}

func (e *emojiDB) GetEmojiByShortcodeDomain(ctx context.Context, shortcode string, domain string) (*gtsmodel.Emoji, db.Error) {
	return e.getEmoji(
		ctx,
		func() (*gtsmodel.Emoji, bool) {
			return e.cache.GetByShortcodeDomain(shortcode, domain)
		},
		func(emoji *gtsmodel.Emoji) error {
			q := e.newEmojiQ(emoji)

			if domain != "" {
				q = q.Where("? = ?", bun.Ident("emoji.shortcode"), shortcode)
				q = q.Where("? = ?", bun.Ident("emoji.domain"), domain)
			} else {
				q = q.Where("? = ?", bun.Ident("emoji.shortcode"), strings.ToLower(shortcode))
				q = q.Where("? IS NULL", bun.Ident("emoji.domain"))
			}

			return q.Scan(ctx)
		},
	)
}

func (e *emojiDB) getEmoji(ctx context.Context, cacheGet func() (*gtsmodel.Emoji, bool), dbQuery func(*gtsmodel.Emoji) error) (*gtsmodel.Emoji, db.Error) {
	// Attempt to fetch cached emoji
	emoji, cached := cacheGet()

	if !cached {
		emoji = &gtsmodel.Emoji{}

		// Not cached! Perform database query
		err := dbQuery(emoji)
		if err != nil {
			return nil, e.conn.ProcessError(err)
		}

		// Place in the cache
		e.cache.Put(emoji)
	}

	return emoji, nil
}

func (e *emojiDB) emojisFromIDs(ctx context.Context, emojiIDs []string) ([]*gtsmodel.Emoji, db.Error) {
	// Catch case of no emojis early
	if len(emojiIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	emojis := make([]*gtsmodel.Emoji, 0, len(emojiIDs))

	for _, id := range emojiIDs {
		emoji, err := e.GetEmojiByID(ctx, id)
		if err != nil {
			log.Errorf("emojisFromIDs: error getting emoji %q: %v", id, err)
		}

		emojis = append(emojis, emoji)
	}

	return emojis, nil
}
