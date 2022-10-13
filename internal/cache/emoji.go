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

// EmojiCache is a cache wrapper to provide ID and URI lookups for gtsmodel.Emoji
type EmojiCache struct {
	cache cache.LookupCache[string, string, *gtsmodel.Emoji]
}

// NewEmojiCache returns a new instantiated EmojiCache object
func NewEmojiCache() *EmojiCache {
	c := &EmojiCache{}
	c.cache = cache.NewLookup(cache.LookupCfg[string, string, *gtsmodel.Emoji]{
		RegisterLookups: func(lm *cache.LookupMap[string, string]) {
			lm.RegisterLookup("uri")
			lm.RegisterLookup("shortcodedomain")
			lm.RegisterLookup("imagestaticurl")
		},

		AddLookups: func(lm *cache.LookupMap[string, string], emoji *gtsmodel.Emoji) {
			lm.Set("shortcodedomain", shortcodeDomainKey(emoji.Shortcode, emoji.Domain), emoji.ID)
			if uri := emoji.URI; uri != "" {
				lm.Set("uri", uri, emoji.ID)
			}
			if imageStaticURL := emoji.ImageStaticURL; imageStaticURL != "" {
				lm.Set("imagestaticurl", imageStaticURL, emoji.ID)
			}
		},

		DeleteLookups: func(lm *cache.LookupMap[string, string], emoji *gtsmodel.Emoji) {
			lm.Delete("shortcodedomain", shortcodeDomainKey(emoji.Shortcode, emoji.Domain))
			if uri := emoji.URI; uri != "" {
				lm.Delete("uri", uri)
			}
			if imageStaticURL := emoji.ImageStaticURL; imageStaticURL != "" {
				lm.Delete("imagestaticurl", imageStaticURL)
			}
		},
	})
	c.cache.SetTTL(time.Minute*5, false)
	c.cache.Start(time.Second * 10)
	return c
}

// GetByID attempts to fetch an emoji from the cache by its ID, you will receive a copy for thread-safety
func (c *EmojiCache) GetByID(id string) (*gtsmodel.Emoji, bool) {
	return c.cache.Get(id)
}

// GetByURI attempts to fetch an emoji from the cache by its URI, you will receive a copy for thread-safety
func (c *EmojiCache) GetByURI(uri string) (*gtsmodel.Emoji, bool) {
	return c.cache.GetBy("uri", uri)
}

func (c *EmojiCache) GetByShortcodeDomain(shortcode string, domain string) (*gtsmodel.Emoji, bool) {
	return c.cache.GetBy("shortcodedomain", shortcodeDomainKey(shortcode, domain))
}

func (c *EmojiCache) GetByImageStaticURL(imageStaticURL string) (*gtsmodel.Emoji, bool) {
	return c.cache.GetBy("imagestaticurl", imageStaticURL)
}

// Put places an emoji in the cache, ensuring that the object place is a copy for thread-safety
func (c *EmojiCache) Put(emoji *gtsmodel.Emoji) {
	if emoji == nil || emoji.ID == "" {
		panic("invalid emoji")
	}
	c.cache.Set(emoji.ID, copyEmoji(emoji))
}

func (c *EmojiCache) Invalidate(emojiID string) {
	c.cache.Invalidate(emojiID)
}

// copyEmoji performs a surface-level copy of emoji, only keeping attached IDs intact, not the objects.
// due to all the data being copied being 99% primitive types or strings (which are immutable and passed by ptr)
// this should be a relatively cheap process
func copyEmoji(emoji *gtsmodel.Emoji) *gtsmodel.Emoji {
	return &gtsmodel.Emoji{
		ID:                     emoji.ID,
		CreatedAt:              emoji.CreatedAt,
		UpdatedAt:              emoji.UpdatedAt,
		Shortcode:              emoji.Shortcode,
		Domain:                 emoji.Domain,
		ImageRemoteURL:         emoji.ImageRemoteURL,
		ImageStaticRemoteURL:   emoji.ImageStaticRemoteURL,
		ImageURL:               emoji.ImageURL,
		ImageStaticURL:         emoji.ImageStaticURL,
		ImagePath:              emoji.ImagePath,
		ImageStaticPath:        emoji.ImageStaticPath,
		ImageContentType:       emoji.ImageContentType,
		ImageStaticContentType: emoji.ImageStaticContentType,
		ImageFileSize:          emoji.ImageFileSize,
		ImageStaticFileSize:    emoji.ImageStaticFileSize,
		ImageUpdatedAt:         emoji.ImageUpdatedAt,
		Disabled:               copyBoolPtr(emoji.Disabled),
		URI:                    emoji.URI,
		VisibleInPicker:        copyBoolPtr(emoji.VisibleInPicker),
		CategoryID:             emoji.CategoryID,
	}
}

func shortcodeDomainKey(shortcode string, domain string) string {
	if domain != "" {
		return shortcode + "@" + domain
	}
	return shortcode
}
