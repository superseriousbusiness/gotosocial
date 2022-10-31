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

package cache

import (
	"time"

	"codeberg.org/gruf/go-cache/v2"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// EmojiCategoryCache is a cache wrapper to provide ID lookups for gtsmodel.EmojiCategory
type EmojiCategoryCache struct {
	cache cache.Cache[string, *gtsmodel.EmojiCategory]
}

// NewEmojiCategoryCache returns a new instantiated EmojiCategoryCache object
func NewEmojiCategoryCache() *EmojiCategoryCache {
	c := &EmojiCategoryCache{}
	c.cache = cache.New[string, *gtsmodel.EmojiCategory]()
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

// GetByID attempts to fetch an emojiCategory from the cache by its ID, you will receive a copy for thread-safety
func (c *EmojiCategoryCache) GetByID(id string) (*gtsmodel.EmojiCategory, bool) {
	return c.cache.Get(id)
}

// Put places an emojiCategory in the cache, ensuring that the object place is a copy for thread-safety
func (c *EmojiCategoryCache) Put(emoji *gtsmodel.EmojiCategory) {
	if emoji == nil || emoji.ID == "" {
		panic("invalid emoji")
	}
	c.cache.Set(emoji.ID, copyEmojiCategory(emoji))
}

func (c *EmojiCategoryCache) Invalidate(emojiID string) {
	c.cache.Invalidate(emojiID)
}

func copyEmojiCategory(emojiCategory *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
	return &gtsmodel.EmojiCategory{
		ID:        emojiCategory.ID,
		CreatedAt: emojiCategory.CreatedAt,
		UpdatedAt: emojiCategory.UpdatedAt,
		Name:      emojiCategory.Name,
	}
}
