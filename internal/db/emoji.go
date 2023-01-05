/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// EmojiAllDomains can be used as the `domain` value in a GetEmojis
// query to indicate that emojis from all domains should be returned.
const EmojiAllDomains string = "all"

// Emoji contains functions for getting emoji in the database.
type Emoji interface {
	// PutEmoji puts one emoji in the database.
	PutEmoji(ctx context.Context, emoji *gtsmodel.Emoji) Error
	// UpdateEmoji updates the given columns of one emoji.
	// If no columns are specified, every column is updated.
	UpdateEmoji(ctx context.Context, emoji *gtsmodel.Emoji, columns ...string) (*gtsmodel.Emoji, Error)
	// DeleteEmojiByID deletes one emoji by its database ID.
	DeleteEmojiByID(ctx context.Context, id string) Error
	// GetEmojisByIDs gets emojis for the given IDs.
	GetEmojisByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Emoji, Error)
	// GetUseableEmojis gets all emojis which are useable by accounts on this instance.
	GetUseableEmojis(ctx context.Context) ([]*gtsmodel.Emoji, Error)
	// GetEmojis gets emojis based on given parameters. Useful for admin actions.
	GetEmojis(ctx context.Context, domain string, includeDisabled bool, includeEnabled bool, shortcode string, maxShortcodeDomain string, minShortcodeDomain string, limit int) ([]*gtsmodel.Emoji, Error)
	// GetEmojiByID gets a specific emoji by its database ID.
	GetEmojiByID(ctx context.Context, id string) (*gtsmodel.Emoji, Error)
	// GetEmojiByShortcodeDomain gets an emoji based on its shortcode and domain.
	// For local emoji, domain should be an empty string.
	GetEmojiByShortcodeDomain(ctx context.Context, shortcode string, domain string) (*gtsmodel.Emoji, Error)
	// GetEmojiByURI returns one emoji based on its ActivityPub URI.
	GetEmojiByURI(ctx context.Context, uri string) (*gtsmodel.Emoji, Error)
	// GetEmojiByStaticURL gets an emoji using the URL of the static version of the emoji image.
	GetEmojiByStaticURL(ctx context.Context, imageStaticURL string) (*gtsmodel.Emoji, Error)
	// PutEmojiCategory puts one new emoji category in the database.
	PutEmojiCategory(ctx context.Context, emojiCategory *gtsmodel.EmojiCategory) Error
	// GetEmojiCategoriesByIDs gets emoji categories for given IDs.
	GetEmojiCategoriesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.EmojiCategory, Error)
	// GetEmojiCategories gets a slice of the names of all existing emoji categories.
	GetEmojiCategories(ctx context.Context) ([]*gtsmodel.EmojiCategory, Error)
	// GetEmojiCategory gets one emoji category by its id.
	GetEmojiCategory(ctx context.Context, id string) (*gtsmodel.EmojiCategory, Error)
	// GetEmojiCategoryByName gets one emoji category by its name.
	GetEmojiCategoryByName(ctx context.Context, name string) (*gtsmodel.EmojiCategory, Error)
}
