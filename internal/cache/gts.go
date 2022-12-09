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
	"codeberg.org/gruf/go-cache/v3/result"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type GTSCaches interface {
	// Init will initialize all the gtsmodel caches in this collection.
	// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
	Init()

	// Start will attempt to start all of the gtsmodel caches, or panic.
	Start()

	// Stop will attempt to stop all of the gtsmodel caches, or panic.
	Stop()

	// Account provides access to the gtsmodel Account database cache.
	Account() *result.Cache[*gtsmodel.Account]

	// Block provides access to the gtsmodel Block (account) database cache.
	Block() *result.Cache[*gtsmodel.Block]

	// DomainBlock provides access to the gtsmodel DomainBlock database cache.
	DomainBlock() *result.Cache[*gtsmodel.DomainBlock]

	// Emoji provides access to the gtsmodel Emoji database cache.
	Emoji() *result.Cache[*gtsmodel.Emoji]

	// EmojiCategory provides access to the gtsmodel EmojiCategory database cache.
	EmojiCategory() *result.Cache[*gtsmodel.EmojiCategory]

	// Mention provides access to the gtsmodel Mention database cache.
	Mention() *result.Cache[*gtsmodel.Mention]

	// Notification provides access to the gtsmodel Notification database cache.
	Notification() *result.Cache[*gtsmodel.Notification]

	// Status provides access to the gtsmodel Status database cache.
	Status() *result.Cache[*gtsmodel.Status]

	// Tombstone provides access to the gtsmodel Tombstone database cache.
	Tombstone() *result.Cache[*gtsmodel.Tombstone]

	// User provides access to the gtsmodel User database cache.
	User() *result.Cache[*gtsmodel.User]
}

// NewGTS returns a new default implementation of GTSCaches.
func NewGTS() GTSCaches {
	return &gtsCaches{}
}

type gtsCaches struct {
	account       *result.Cache[*gtsmodel.Account]
	block         *result.Cache[*gtsmodel.Block]
	domainBlock   *result.Cache[*gtsmodel.DomainBlock]
	emoji         *result.Cache[*gtsmodel.Emoji]
	emojiCategory *result.Cache[*gtsmodel.EmojiCategory]
	mention       *result.Cache[*gtsmodel.Mention]
	notification  *result.Cache[*gtsmodel.Notification]
	status        *result.Cache[*gtsmodel.Status]
	tombstone     *result.Cache[*gtsmodel.Tombstone]
	user          *result.Cache[*gtsmodel.User]
}

func (c *gtsCaches) Init() {
	c.initAccount()
	c.initBlock()
	c.initDomainBlock()
	c.initEmoji()
	c.initEmojiCategory()
	c.initMention()
	c.initNotification()
	c.initStatus()
	c.initTombstone()
	c.initUser()
}

func (c *gtsCaches) Start() {
	tryUntil("starting gtsmodel.Account cache", 5, func() bool {
		return c.account.Start(config.GetCacheAccountSweepFreq())
	})
	tryUntil("starting gtsmodel.Block cache", 5, func() bool {
		return c.block.Start(config.GetCacheBlockSweepFreq())
	})
	tryUntil("starting gtsmodel.DomainBlock cache", 5, func() bool {
		return c.domainBlock.Start(config.GetCacheDomainBlockSweepFreq())
	})
	tryUntil("starting gtsmodel.Emoji cache", 5, func() bool {
		return c.emoji.Start(config.GetCacheEmojiSweepFreq())
	})
	tryUntil("starting gtsmodel.EmojiCategory cache", 5, func() bool {
		return c.emojiCategory.Start(config.GetCacheEmojiCategorySweepFreq())
	})
	tryUntil("starting gtsmodel.Mention cache", 5, func() bool {
		return c.mention.Start(config.GetCacheMentionSweepFreq())
	})
	tryUntil("starting gtsmodel.Notification cache", 5, func() bool {
		return c.notification.Start(config.GetCacheNotificationSweepFreq())
	})
	tryUntil("starting gtsmodel.Status cache", 5, func() bool {
		return c.status.Start(config.GetCacheStatusSweepFreq())
	})
	tryUntil("starting gtsmodel.Tombstone cache", 5, func() bool {
		return c.tombstone.Start(config.GetCacheTombstoneSweepFreq())
	})
	tryUntil("starting gtsmodel.User cache", 5, func() bool {
		return c.user.Start(config.GetCacheUserSweepFreq())
	})
}

func (c *gtsCaches) Stop() {
	tryUntil("stopping gtsmodel.Account cache", 5, c.account.Stop)
	tryUntil("stopping gtsmodel.Block cache", 5, c.block.Stop)
	tryUntil("stopping gtsmodel.DomainBlock cache", 5, c.domainBlock.Stop)
	tryUntil("stopping gtsmodel.Emoji cache", 5, c.emoji.Stop)
	tryUntil("stopping gtsmodel.EmojiCategory cache", 5, c.emojiCategory.Stop)
	tryUntil("stopping gtsmodel.Mention cache", 5, c.mention.Stop)
	tryUntil("stopping gtsmodel.Notification cache", 5, c.notification.Stop)
	tryUntil("stopping gtsmodel.Status cache", 5, c.status.Stop)
	tryUntil("stopping gtsmodel.Tombstone cache", 5, c.tombstone.Stop)
	tryUntil("stopping gtsmodel.User cache", 5, c.user.Stop)
}

func (c *gtsCaches) Account() *result.Cache[*gtsmodel.Account] {
	return c.account
}

func (c *gtsCaches) Block() *result.Cache[*gtsmodel.Block] {
	return c.block
}

func (c *gtsCaches) DomainBlock() *result.Cache[*gtsmodel.DomainBlock] {
	return c.domainBlock
}

func (c *gtsCaches) Emoji() *result.Cache[*gtsmodel.Emoji] {
	return c.emoji
}

func (c *gtsCaches) EmojiCategory() *result.Cache[*gtsmodel.EmojiCategory] {
	return c.emojiCategory
}

func (c *gtsCaches) Mention() *result.Cache[*gtsmodel.Mention] {
	return c.mention
}

func (c *gtsCaches) Notification() *result.Cache[*gtsmodel.Notification] {
	return c.notification
}

func (c *gtsCaches) Status() *result.Cache[*gtsmodel.Status] {
	return c.status
}

func (c *gtsCaches) Tombstone() *result.Cache[*gtsmodel.Tombstone] {
	return c.tombstone
}

func (c *gtsCaches) User() *result.Cache[*gtsmodel.User] {
	return c.user
}

func (c *gtsCaches) initAccount() {
	c.account = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "Username.Domain"},
		{Name: "PublicKeyURI"},
	}, func(a1 *gtsmodel.Account) *gtsmodel.Account {
		a2 := new(gtsmodel.Account)
		*a2 = *a1
		return a2
	}, config.GetCacheAccountMaxSize())
	c.account.SetTTL(config.GetCacheAccountTTL(), true)
}

func (c *gtsCaches) initBlock() {
	c.block = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "URI"},
	}, func(b1 *gtsmodel.Block) *gtsmodel.Block {
		b2 := new(gtsmodel.Block)
		*b2 = *b1
		return b2
	}, config.GetCacheBlockMaxSize())
	c.block.SetTTL(config.GetCacheBlockTTL(), true)
}

func (c *gtsCaches) initDomainBlock() {
	c.domainBlock = result.NewSized([]result.Lookup{
		{Name: "Domain"},
	}, func(d1 *gtsmodel.DomainBlock) *gtsmodel.DomainBlock {
		d2 := new(gtsmodel.DomainBlock)
		*d2 = *d1
		return d2
	}, config.GetCacheDomainBlockMaxSize())
	c.domainBlock.SetTTL(config.GetCacheDomainBlockTTL(), true)
}

func (c *gtsCaches) initEmoji() {
	c.emoji = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "Shortcode.Domain"},
		{Name: "ImageStaticURL"},
	}, func(e1 *gtsmodel.Emoji) *gtsmodel.Emoji {
		e2 := new(gtsmodel.Emoji)
		*e2 = *e1
		return e2
	}, config.GetCacheEmojiMaxSize())
	c.emoji.SetTTL(config.GetCacheEmojiTTL(), true)
}

func (c *gtsCaches) initEmojiCategory() {
	c.emojiCategory = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "Name"},
	}, func(c1 *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
		c2 := new(gtsmodel.EmojiCategory)
		*c2 = *c1
		return c2
	}, config.GetCacheEmojiCategoryMaxSize())
	c.emojiCategory.SetTTL(config.GetCacheEmojiCategoryTTL(), true)
}

func (c *gtsCaches) initMention() {
	c.mention = result.NewSized([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.Mention) *gtsmodel.Mention {
		m2 := new(gtsmodel.Mention)
		*m2 = *m1
		return m2
	}, config.GetCacheMentionMaxSize())
	c.mention.SetTTL(config.GetCacheMentionTTL(), true)
}

func (c *gtsCaches) initNotification() {
	c.notification = result.NewSized([]result.Lookup{
		{Name: "ID"},
	}, func(n1 *gtsmodel.Notification) *gtsmodel.Notification {
		n2 := new(gtsmodel.Notification)
		*n2 = *n1
		return n2
	}, config.GetCacheNotificationMaxSize())
	c.notification.SetTTL(config.GetCacheNotificationTTL(), true)
}

func (c *gtsCaches) initStatus() {
	c.status = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
	}, func(s1 *gtsmodel.Status) *gtsmodel.Status {
		s2 := new(gtsmodel.Status)
		*s2 = *s1
		return s2
	}, config.GetCacheStatusMaxSize())
	c.status.SetTTL(config.GetCacheStatusTTL(), true)
}

// initTombstone will initialize the gtsmodel.Tombstone cache.
func (c *gtsCaches) initTombstone() {
	c.tombstone = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
	}, func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}, config.GetCacheTombstoneMaxSize())
	c.tombstone.SetTTL(config.GetCacheTombstoneTTL(), true)
}

func (c *gtsCaches) initUser() {
	c.user = result.NewSized([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID"},
		{Name: "Email"},
		{Name: "ConfirmationToken"},
		{Name: "ExternalID"},
	}, func(u1 *gtsmodel.User) *gtsmodel.User {
		u2 := new(gtsmodel.User)
		*u2 = *u1
		return u2
	}, config.GetCacheUserMaxSize())
	c.user.SetTTL(config.GetCacheUserTTL(), true)
}
