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

package cache

import (
	"codeberg.org/gruf/go-cache/v3/result"
	"codeberg.org/gruf/go-cache/v3/ttl"
	"github.com/superseriousbusiness/gotosocial/internal/cache/domain"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type GTSCaches struct {
	account *result.Cache[*gtsmodel.Account]
	block   *result.Cache[*gtsmodel.Block]
	// TODO: maybe should be moved out of here since it's
	// not actually doing anything with gtsmodel.DomainBlock.
	domainBlock   *domain.BlockCache
	emoji         *result.Cache[*gtsmodel.Emoji]
	emojiCategory *result.Cache[*gtsmodel.EmojiCategory]
	follow        *result.Cache[*gtsmodel.Follow]
	followRequest *result.Cache[*gtsmodel.FollowRequest]
	instance      *result.Cache[*gtsmodel.Instance]
	list          *result.Cache[*gtsmodel.List]
	listEntry     *result.Cache[*gtsmodel.ListEntry]
	media         *result.Cache[*gtsmodel.MediaAttachment]
	mention       *result.Cache[*gtsmodel.Mention]
	notification  *result.Cache[*gtsmodel.Notification]
	report        *result.Cache[*gtsmodel.Report]
	status        *result.Cache[*gtsmodel.Status]
	statusFave    *result.Cache[*gtsmodel.StatusFave]
	tombstone     *result.Cache[*gtsmodel.Tombstone]
	user          *result.Cache[*gtsmodel.User]
	// TODO: move out of GTS caches since not using database models.
	webfinger *ttl.Cache[string, string]
}

// Init will initialize all the gtsmodel caches in this collection.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *GTSCaches) Init() {
	c.initAccount()
	c.initBlock()
	c.initDomainBlock()
	c.initEmoji()
	c.initEmojiCategory()
	c.initFollow()
	c.initFollowRequest()
	c.initInstance()
	c.initList()
	c.initListEntry()
	c.initMedia()
	c.initMention()
	c.initNotification()
	c.initReport()
	c.initStatus()
	c.initStatusFave()
	c.initTombstone()
	c.initUser()
	c.initWebfinger()
}

// Start will attempt to start all of the gtsmodel caches, or panic.
func (c *GTSCaches) Start() {
	tryStart(c.account, config.GetCacheGTSAccountSweepFreq())
	tryStart(c.block, config.GetCacheGTSBlockSweepFreq())
	tryStart(c.emoji, config.GetCacheGTSEmojiSweepFreq())
	tryStart(c.emojiCategory, config.GetCacheGTSEmojiCategorySweepFreq())
	tryStart(c.follow, config.GetCacheGTSFollowSweepFreq())
	tryStart(c.followRequest, config.GetCacheGTSFollowRequestSweepFreq())
	tryStart(c.instance, config.GetCacheGTSInstanceSweepFreq())
	tryStart(c.list, config.GetCacheGTSListSweepFreq())
	tryStart(c.listEntry, config.GetCacheGTSListEntrySweepFreq())
	tryStart(c.media, config.GetCacheGTSMediaSweepFreq())
	tryStart(c.mention, config.GetCacheGTSMentionSweepFreq())
	tryStart(c.notification, config.GetCacheGTSNotificationSweepFreq())
	tryStart(c.report, config.GetCacheGTSReportSweepFreq())
	tryStart(c.status, config.GetCacheGTSStatusSweepFreq())
	tryStart(c.statusFave, config.GetCacheGTSStatusFaveSweepFreq())
	tryStart(c.tombstone, config.GetCacheGTSTombstoneSweepFreq())
	tryStart(c.user, config.GetCacheGTSUserSweepFreq())
	tryUntil("starting *gtsmodel.Webfinger cache", 5, func() bool {
		if sweep := config.GetCacheGTSWebfingerSweepFreq(); sweep > 0 {
			return c.webfinger.Start(sweep)
		}
		return true
	})
}

// Stop will attempt to stop all of the gtsmodel caches, or panic.
func (c *GTSCaches) Stop() {
	tryStop(c.account, config.GetCacheGTSAccountSweepFreq())
	tryStop(c.block, config.GetCacheGTSBlockSweepFreq())
	tryStop(c.emoji, config.GetCacheGTSEmojiSweepFreq())
	tryStop(c.emojiCategory, config.GetCacheGTSEmojiCategorySweepFreq())
	tryStop(c.follow, config.GetCacheGTSFollowSweepFreq())
	tryStop(c.followRequest, config.GetCacheGTSFollowRequestSweepFreq())
	tryStop(c.instance, config.GetCacheGTSInstanceSweepFreq())
	tryStop(c.list, config.GetCacheGTSListSweepFreq())
	tryStop(c.listEntry, config.GetCacheGTSListEntrySweepFreq())
	tryStop(c.media, config.GetCacheGTSMediaSweepFreq())
	tryStop(c.mention, config.GetCacheGTSNotificationSweepFreq())
	tryStop(c.notification, config.GetCacheGTSNotificationSweepFreq())
	tryStop(c.report, config.GetCacheGTSReportSweepFreq())
	tryStop(c.status, config.GetCacheGTSStatusSweepFreq())
	tryStop(c.statusFave, config.GetCacheGTSStatusFaveSweepFreq())
	tryStop(c.tombstone, config.GetCacheGTSTombstoneSweepFreq())
	tryStop(c.user, config.GetCacheGTSUserSweepFreq())
	tryUntil("stopping *gtsmodel.Webfinger cache", 5, c.webfinger.Stop)
}

// Account provides access to the gtsmodel Account database cache.
func (c *GTSCaches) Account() *result.Cache[*gtsmodel.Account] {
	return c.account
}

// Block provides access to the gtsmodel Block (account) database cache.
func (c *GTSCaches) Block() *result.Cache[*gtsmodel.Block] {
	return c.block
}

// DomainBlock provides access to the domain block database cache.
func (c *GTSCaches) DomainBlock() *domain.BlockCache {
	return c.domainBlock
}

// Emoji provides access to the gtsmodel Emoji database cache.
func (c *GTSCaches) Emoji() *result.Cache[*gtsmodel.Emoji] {
	return c.emoji
}

// EmojiCategory provides access to the gtsmodel EmojiCategory database cache.
func (c *GTSCaches) EmojiCategory() *result.Cache[*gtsmodel.EmojiCategory] {
	return c.emojiCategory
}

// Follow provides access to the gtsmodel Follow database cache.
func (c *GTSCaches) Follow() *result.Cache[*gtsmodel.Follow] {
	return c.follow
}

// FollowRequest provides access to the gtsmodel FollowRequest database cache.
func (c *GTSCaches) FollowRequest() *result.Cache[*gtsmodel.FollowRequest] {
	return c.followRequest
}

// Instance provides access to the gtsmodel Instance database cache.
func (c *GTSCaches) Instance() *result.Cache[*gtsmodel.Instance] {
	return c.instance
}

// List provides access to the gtsmodel List database cache.
func (c *GTSCaches) List() *result.Cache[*gtsmodel.List] {
	return c.list
}

// ListEntry provides access to the gtsmodel ListEntry database cache.
func (c *GTSCaches) ListEntry() *result.Cache[*gtsmodel.ListEntry] {
	return c.listEntry
}

// Media provides access to the gtsmodel Media database cache.
func (c *GTSCaches) Media() *result.Cache[*gtsmodel.MediaAttachment] {
	return c.media
}

// Mention provides access to the gtsmodel Mention database cache.
func (c *GTSCaches) Mention() *result.Cache[*gtsmodel.Mention] {
	return c.mention
}

// Notification provides access to the gtsmodel Notification database cache.
func (c *GTSCaches) Notification() *result.Cache[*gtsmodel.Notification] {
	return c.notification
}

// Report provides access to the gtsmodel Report database cache.
func (c *GTSCaches) Report() *result.Cache[*gtsmodel.Report] {
	return c.report
}

// Status provides access to the gtsmodel Status database cache.
func (c *GTSCaches) Status() *result.Cache[*gtsmodel.Status] {
	return c.status
}

// StatusFave provides access to the gtsmodel StatusFave database cache.
func (c *GTSCaches) StatusFave() *result.Cache[*gtsmodel.StatusFave] {
	return c.statusFave
}

// Tombstone provides access to the gtsmodel Tombstone database cache.
func (c *GTSCaches) Tombstone() *result.Cache[*gtsmodel.Tombstone] {
	return c.tombstone
}

// User provides access to the gtsmodel User database cache.
func (c *GTSCaches) User() *result.Cache[*gtsmodel.User] {
	return c.user
}

// Webfinger provides access to the webfinger URL cache.
func (c *GTSCaches) Webfinger() *ttl.Cache[string, string] {
	return c.webfinger
}

func (c *GTSCaches) initAccount() {
	c.account = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "Username.Domain"},
		{Name: "PublicKeyURI"},
		{Name: "InboxURI"},
		{Name: "OutboxURI"},
		{Name: "FollowersURI"},
		{Name: "FollowingURI"},
	}, func(a1 *gtsmodel.Account) *gtsmodel.Account {
		a2 := new(gtsmodel.Account)
		*a2 = *a1
		return a2
	}, config.GetCacheGTSAccountMaxSize())
	c.account.SetTTL(config.GetCacheGTSAccountTTL(), true)
	c.account.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initBlock() {
	c.block = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
	}, func(b1 *gtsmodel.Block) *gtsmodel.Block {
		b2 := new(gtsmodel.Block)
		*b2 = *b1
		return b2
	}, config.GetCacheGTSBlockMaxSize())
	c.block.SetTTL(config.GetCacheGTSBlockTTL(), true)
	c.block.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initDomainBlock() {
	c.domainBlock = new(domain.BlockCache)
}

func (c *GTSCaches) initEmoji() {
	c.emoji = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "Shortcode.Domain"},
		{Name: "ImageStaticURL"},
	}, func(e1 *gtsmodel.Emoji) *gtsmodel.Emoji {
		e2 := new(gtsmodel.Emoji)
		*e2 = *e1
		return e2
	}, config.GetCacheGTSEmojiMaxSize())
	c.emoji.SetTTL(config.GetCacheGTSEmojiTTL(), true)
	c.emoji.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initEmojiCategory() {
	c.emojiCategory = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "Name"},
	}, func(c1 *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
		c2 := new(gtsmodel.EmojiCategory)
		*c2 = *c1
		return c2
	}, config.GetCacheGTSEmojiCategoryMaxSize())
	c.emojiCategory.SetTTL(config.GetCacheGTSEmojiCategoryTTL(), true)
	c.emojiCategory.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initFollow() {
	c.follow = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
	}, func(f1 *gtsmodel.Follow) *gtsmodel.Follow {
		f2 := new(gtsmodel.Follow)
		*f2 = *f1
		return f2
	}, config.GetCacheGTSFollowMaxSize())
	c.follow.SetTTL(config.GetCacheGTSFollowTTL(), true)
}

func (c *GTSCaches) initFollowRequest() {
	c.followRequest = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
	}, func(f1 *gtsmodel.FollowRequest) *gtsmodel.FollowRequest {
		f2 := new(gtsmodel.FollowRequest)
		*f2 = *f1
		return f2
	}, config.GetCacheGTSFollowRequestMaxSize())
	c.followRequest.SetTTL(config.GetCacheGTSFollowRequestTTL(), true)
}

func (c *GTSCaches) initInstance() {
	c.instance = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "Domain"},
	}, func(i1 *gtsmodel.Instance) *gtsmodel.Instance {
		i2 := new(gtsmodel.Instance)
		*i2 = *i1
		return i1
	}, config.GetCacheGTSInstanceMaxSize())
	c.instance.SetTTL(config.GetCacheGTSInstanceTTL(), true)
	c.emojiCategory.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initList() {
	c.list = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(l1 *gtsmodel.List) *gtsmodel.List {
		l2 := new(gtsmodel.List)
		*l2 = *l1
		return l2
	}, config.GetCacheGTSListMaxSize())
	c.list.SetTTL(config.GetCacheGTSListTTL(), true)
	c.list.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initListEntry() {
	c.listEntry = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(l1 *gtsmodel.ListEntry) *gtsmodel.ListEntry {
		l2 := new(gtsmodel.ListEntry)
		*l2 = *l1
		return l2
	}, config.GetCacheGTSListEntryMaxSize())
	c.list.SetTTL(config.GetCacheGTSListEntryTTL(), true)
	c.list.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initMedia() {
	c.media = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.MediaAttachment) *gtsmodel.MediaAttachment {
		m2 := new(gtsmodel.MediaAttachment)
		*m2 = *m1
		return m2
	}, config.GetCacheGTSMediaMaxSize())
	c.media.SetTTL(config.GetCacheGTSMediaTTL(), true)
	c.media.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initMention() {
	c.mention = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.Mention) *gtsmodel.Mention {
		m2 := new(gtsmodel.Mention)
		*m2 = *m1
		return m2
	}, config.GetCacheGTSMentionMaxSize())
	c.mention.SetTTL(config.GetCacheGTSMentionTTL(), true)
	c.mention.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initNotification() {
	c.notification = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "NotificationType.TargetAccountID.OriginAccountID.StatusID"},
	}, func(n1 *gtsmodel.Notification) *gtsmodel.Notification {
		n2 := new(gtsmodel.Notification)
		*n2 = *n1
		return n2
	}, config.GetCacheGTSNotificationMaxSize())
	c.notification.SetTTL(config.GetCacheGTSNotificationTTL(), true)
	c.notification.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initReport() {
	c.report = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(r1 *gtsmodel.Report) *gtsmodel.Report {
		r2 := new(gtsmodel.Report)
		*r2 = *r1
		return r2
	}, config.GetCacheGTSReportMaxSize())
	c.report.SetTTL(config.GetCacheGTSReportTTL(), true)
	c.report.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatus() {
	c.status = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
	}, func(s1 *gtsmodel.Status) *gtsmodel.Status {
		s2 := new(gtsmodel.Status)
		*s2 = *s1
		return s2
	}, config.GetCacheGTSStatusMaxSize())
	c.status.SetTTL(config.GetCacheGTSStatusTTL(), true)
	c.status.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatusFave() {
	c.statusFave = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.StatusID"},
	}, func(f1 *gtsmodel.StatusFave) *gtsmodel.StatusFave {
		f2 := new(gtsmodel.StatusFave)
		*f2 = *f1
		return f2
	}, config.GetCacheGTSStatusFaveMaxSize())
	c.status.SetTTL(config.GetCacheGTSStatusFaveTTL(), true)
	c.status.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initTombstone() {
	c.tombstone = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
	}, func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}, config.GetCacheGTSTombstoneMaxSize())
	c.tombstone.SetTTL(config.GetCacheGTSTombstoneTTL(), true)
	c.tombstone.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initUser() {
	c.user = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID"},
		{Name: "Email"},
		{Name: "ConfirmationToken"},
		{Name: "ExternalID"},
	}, func(u1 *gtsmodel.User) *gtsmodel.User {
		u2 := new(gtsmodel.User)
		*u2 = *u1
		return u2
	}, config.GetCacheGTSUserMaxSize())
	c.user.SetTTL(config.GetCacheGTSUserTTL(), true)
	c.user.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initWebfinger() {
	c.webfinger = ttl.New[string, string](
		0,
		config.GetCacheGTSWebfingerMaxSize(),
		config.GetCacheGTSWebfingerTTL())
}
