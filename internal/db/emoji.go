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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// EmojiAllDomains can be used as the `domain` value in a GetEmojis
// query to indicate that emojis from all domains should be returned.
const EmojiAllDomains string = "all"

// Emoji contains functions for getting emoji in the database.
type Emoji interface {
	// PutEmoji puts one emoji in the database.
	PutEmoji(ctx context.Context, emoji *gtsmodel.Emoji) error

	// UpdateEmoji updates the given columns of one emoji.
	// If no columns are specified, every column is updated.
	UpdateEmoji(ctx context.Context, emoji *gtsmodel.Emoji, columns ...string) error

	// DeleteEmojiByID deletes one emoji by its database ID.
	DeleteEmojiByID(ctx context.Context, id string) error

	// GetEmojisByIDs gets emojis for the given IDs.
	GetEmojisByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Emoji, error)

	// GetUseableEmojis gets all emojis which are useable by accounts on this instance.
	GetUseableEmojis(ctx context.Context) ([]*gtsmodel.Emoji, error)

	// GetEmojis fetches all emojis with IDs less than 'maxID', up to a maximum of 'limit' emojis.
	GetEmojis(ctx context.Context, page *paging.Page) ([]*gtsmodel.Emoji, error)

	// GetRemoteEmojis fetches all remote emojis with IDs less than 'maxID', up to a maximum of 'limit' emojis.
	GetRemoteEmojis(ctx context.Context, page *paging.Page) ([]*gtsmodel.Emoji, error)

	// GetCachedEmojisOlderThan fetches all cached remote emojis with 'updated_at' greater than 'olderThan', up to a maximum of 'limit' emojis.
	GetCachedEmojisOlderThan(ctx context.Context, olderThan time.Time, limit int) ([]*gtsmodel.Emoji, error)

	// GetEmojisBy gets emojis based on given parameters. Useful for admin actions.
	GetEmojisBy(ctx context.Context, domain string, includeDisabled bool, includeEnabled bool, shortcode string, maxShortcodeDomain string, minShortcodeDomain string, limit int) ([]*gtsmodel.Emoji, error)

	// GetEmojiByID gets a specific emoji by its database ID.
	GetEmojiByID(ctx context.Context, id string) (*gtsmodel.Emoji, error)

	// PopulateEmoji populates the struct pointers on the given emoji.
	PopulateEmoji(ctx context.Context, emoji *gtsmodel.Emoji) error

	// GetEmojiByShortcodeDomain gets an emoji based on its shortcode and domain.
	// For local emoji, domain should be an empty string.
	GetEmojiByShortcodeDomain(ctx context.Context, shortcode string, domain string) (*gtsmodel.Emoji, error)

	// GetEmojiByURI returns one emoji based on its ActivityPub URI.
	GetEmojiByURI(ctx context.Context, uri string) (*gtsmodel.Emoji, error)

	// GetEmojiByStaticURL gets an emoji using the URL of the static version of the emoji image.
	GetEmojiByStaticURL(ctx context.Context, imageStaticURL string) (*gtsmodel.Emoji, error)

	// PutEmojiCategory puts one new emoji category in the database.
	PutEmojiCategory(ctx context.Context, emojiCategory *gtsmodel.EmojiCategory) error

	// GetEmojiCategoriesByIDs gets emoji categories for given IDs.
	GetEmojiCategoriesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.EmojiCategory, error)

	// GetEmojiCategories gets a slice of the names of all existing emoji categories.
	GetEmojiCategories(ctx context.Context) ([]*gtsmodel.EmojiCategory, error)

	// GetEmojiCategory gets one emoji category by its id.
	GetEmojiCategory(ctx context.Context, id string) (*gtsmodel.EmojiCategory, error)

	// GetEmojiCategoryByName gets one emoji category by its name.
	GetEmojiCategoryByName(ctx context.Context, name string) (*gtsmodel.EmojiCategory, error)
}
