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

package db

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// Emoji contains functions for getting emoji in the database.
type Emoji interface {
	// GetCustomEmojis gets all custom emoji for the instance
	GetCustomEmojis(ctx context.Context) ([]*gtsmodel.Emoji, Error)
	// GetEmojiByID gets a specific emoji by its database ID.
	GetEmojiByID(ctx context.Context, id string) (*gtsmodel.Emoji, Error)
	// GetEmojiByShortcodeDomain gets an emoji based on its shortcode and domain.
	// For local emoji, domain should be an empty string.
	GetEmojiByShortcodeDomain(ctx context.Context, shortcode string, domain string) (*gtsmodel.Emoji, Error)
}
