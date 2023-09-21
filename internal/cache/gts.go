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
	"time"

	"codeberg.org/gruf/go-cache/v3/result"
	"codeberg.org/gruf/go-cache/v3/simple"
	"codeberg.org/gruf/go-cache/v3/ttl"
	"github.com/superseriousbusiness/gotosocial/internal/cache/domain"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type GTSCaches struct {
	account          *result.Cache[*gtsmodel.Account]
	accountNote      *result.Cache[*gtsmodel.AccountNote]
	application      *result.Cache[*gtsmodel.Application]
	block            *result.Cache[*gtsmodel.Block]
	blockIDs         *SliceCache[string]
	boostOfIDs       *SliceCache[string]
	domainAllow      *domain.Cache
	domainBlock      *domain.Cache
	emoji            *result.Cache[*gtsmodel.Emoji]
	emojiCategory    *result.Cache[*gtsmodel.EmojiCategory]
	follow           *result.Cache[*gtsmodel.Follow]
	followIDs        *SliceCache[string]
	followRequest    *result.Cache[*gtsmodel.FollowRequest]
	followRequestIDs *SliceCache[string]
	instance         *result.Cache[*gtsmodel.Instance]
	inReplyToIDs     *SliceCache[string]
	list             *result.Cache[*gtsmodel.List]
	listEntry        *result.Cache[*gtsmodel.ListEntry]
	marker           *result.Cache[*gtsmodel.Marker]
	media            *result.Cache[*gtsmodel.MediaAttachment]
	mention          *result.Cache[*gtsmodel.Mention]
	notification     *result.Cache[*gtsmodel.Notification]
	report           *result.Cache[*gtsmodel.Report]
	status           *result.Cache[*gtsmodel.Status]
	statusFave       *result.Cache[*gtsmodel.StatusFave]
	statusFaveIDs    *SliceCache[string]
	tag              *result.Cache[*gtsmodel.Tag]
	tombstone        *result.Cache[*gtsmodel.Tombstone]
	user             *result.Cache[*gtsmodel.User]

	// TODO: move out of GTS caches since unrelated to DB.
	webfinger *ttl.Cache[string, string] // TTL=24hr, sweep=5min
}

// Init will initialize all the gtsmodel caches in this collection.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *GTSCaches) Init() {
	c.initAccount()
	c.initAccountNote()
	c.initApplication()
	c.initBlock()
	c.initBlockIDs()
	c.initBoostOfIDs()
	c.initDomainAllow()
	c.initDomainBlock()
	c.initEmoji()
	c.initEmojiCategory()
	c.initFollow()
	c.initFollowIDs()
	c.initFollowRequest()
	c.initFollowRequestIDs()
	c.initInReplyToIDs()
	c.initInstance()
	c.initList()
	c.initListEntry()
	c.initMarker()
	c.initMedia()
	c.initMention()
	c.initNotification()
	c.initReport()
	c.initStatus()
	c.initStatusFave()
	c.initTag()
	c.initStatusFaveIDs()
	c.initTombstone()
	c.initUser()
	c.initWebfinger()
}

// Start will attempt to start all of the gtsmodel caches, or panic.
func (c *GTSCaches) Start() {
	tryUntil("starting *gtsmodel.Webfinger cache", 5, func() bool {
		return c.webfinger.Start(5 * time.Minute)
	})
}

// Stop will attempt to stop all of the gtsmodel caches, or panic.
func (c *GTSCaches) Stop() {
	tryUntil("stopping *gtsmodel.Webfinger cache", 5, c.webfinger.Stop)
}

// Account provides access to the gtsmodel Account database cache.
func (c *GTSCaches) Account() *result.Cache[*gtsmodel.Account] {
	return c.account
}

// AccountNote provides access to the gtsmodel Note database cache.
func (c *GTSCaches) AccountNote() *result.Cache[*gtsmodel.AccountNote] {
	return c.accountNote
}

// Application provides access to the gtsmodel Application database cache.
func (c *GTSCaches) Application() *result.Cache[*gtsmodel.Application] {
	return c.application
}

// Block provides access to the gtsmodel Block (account) database cache.
func (c *GTSCaches) Block() *result.Cache[*gtsmodel.Block] {
	return c.block
}

// FollowIDs provides access to the block IDs database cache.
func (c *GTSCaches) BlockIDs() *SliceCache[string] {
	return c.blockIDs
}

// BoostOfIDs provides access to the boost of IDs list database cache.
func (c *GTSCaches) BoostOfIDs() *SliceCache[string] {
	return c.boostOfIDs
}

// DomainAllow provides access to the domain allow database cache.
func (c *GTSCaches) DomainAllow() *domain.Cache {
	return c.domainAllow
}

// DomainBlock provides access to the domain block database cache.
func (c *GTSCaches) DomainBlock() *domain.Cache {
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

// FollowIDs provides access to the follower / following IDs database cache.
// THIS CACHE IS KEYED AS THE FOLLOWING {prefix}{accountID} WHERE PREFIX IS:
// - '>'  for following IDs
// - 'l>' for local following IDs
// - '<'  for follower IDs
// - 'l<' for local follower IDs
func (c *GTSCaches) FollowIDs() *SliceCache[string] {
	return c.followIDs
}

// FollowRequest provides access to the gtsmodel FollowRequest database cache.
func (c *GTSCaches) FollowRequest() *result.Cache[*gtsmodel.FollowRequest] {
	return c.followRequest
}

// FollowRequestIDs provides access to the follow requester / requesting IDs database
// cache. THIS CACHE IS KEYED AS THE FOLLOWING {prefix}{accountID} WHERE PREFIX IS:
// - '>'  for following IDs
// - '<'  for follower IDs
func (c *GTSCaches) FollowRequestIDs() *SliceCache[string] {
	return c.followRequestIDs
}

// Instance provides access to the gtsmodel Instance database cache.
func (c *GTSCaches) Instance() *result.Cache[*gtsmodel.Instance] {
	return c.instance
}

// InReplyToIDs provides access to the status in reply to IDs list database cache.
func (c *GTSCaches) InReplyToIDs() *SliceCache[string] {
	return c.inReplyToIDs
}

// List provides access to the gtsmodel List database cache.
func (c *GTSCaches) List() *result.Cache[*gtsmodel.List] {
	return c.list
}

// ListEntry provides access to the gtsmodel ListEntry database cache.
func (c *GTSCaches) ListEntry() *result.Cache[*gtsmodel.ListEntry] {
	return c.listEntry
}

// Marker provides access to the gtsmodel Marker database cache.
func (c *GTSCaches) Marker() *result.Cache[*gtsmodel.Marker] {
	return c.marker
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

// Tag provides access to the gtsmodel Tag database cache.
func (c *GTSCaches) Tag() *result.Cache[*gtsmodel.Tag] {
	return c.tag
}

// StatusFaveIDs provides access to the status fave IDs list database cache.
func (c *GTSCaches) StatusFaveIDs() *SliceCache[string] {
	return c.statusFaveIDs
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
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofAccount(), // model in-mem size.
		config.GetCacheAccountMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.account = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "Username.Domain", AllowZero: true /* domain can be zero i.e. "" */},
		{Name: "PublicKeyURI"},
		{Name: "InboxURI"},
		{Name: "OutboxURI"},
		{Name: "FollowersURI"},
		{Name: "FollowingURI"},
	}, func(a1 *gtsmodel.Account) *gtsmodel.Account {
		a2 := new(gtsmodel.Account)
		*a2 = *a1
		return a2
	}, cap)

	c.account.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initAccountNote() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofAccountNote(), // model in-mem size.
		config.GetCacheAccountNoteMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.accountNote = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.TargetAccountID"},
	}, func(n1 *gtsmodel.AccountNote) *gtsmodel.AccountNote {
		n2 := new(gtsmodel.AccountNote)
		*n2 = *n1
		return n2
	}, cap)

	c.accountNote.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initApplication() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofApplication(), // model in-mem size.
		config.GetCacheApplicationMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.application = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "ClientID"},
	}, func(a1 *gtsmodel.Application) *gtsmodel.Application {
		a2 := new(gtsmodel.Application)
		*a2 = *a1
		return a2
	}, cap)

	c.application.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initBlock() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofBlock(), // model in-mem size.
		config.GetCacheBlockMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.block = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "AccountID", Multi: true},
		{Name: "TargetAccountID", Multi: true},
	}, func(b1 *gtsmodel.Block) *gtsmodel.Block {
		b2 := new(gtsmodel.Block)
		*b2 = *b1
		return b2
	}, cap)

	c.block.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initBlockIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheBlockIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.blockIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initBoostOfIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheBoostOfIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.boostOfIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initDomainAllow() {
	c.domainAllow = new(domain.Cache)
}

func (c *GTSCaches) initDomainBlock() {
	c.domainBlock = new(domain.Cache)
}

func (c *GTSCaches) initEmoji() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofEmoji(), // model in-mem size.
		config.GetCacheEmojiMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.emoji = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "Shortcode.Domain", AllowZero: true /* domain can be zero i.e. "" */},
		{Name: "ImageStaticURL"},
		{Name: "CategoryID", Multi: true},
	}, func(e1 *gtsmodel.Emoji) *gtsmodel.Emoji {
		e2 := new(gtsmodel.Emoji)
		*e2 = *e1
		return e2
	}, cap)

	c.emoji.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initEmojiCategory() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofEmojiCategory(), // model in-mem size.
		config.GetCacheEmojiCategoryMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.emojiCategory = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "Name"},
	}, func(c1 *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
		c2 := new(gtsmodel.EmojiCategory)
		*c2 = *c1
		return c2
	}, cap)

	c.emojiCategory.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initFollow() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofFollow(), // model in-mem size.
		config.GetCacheFollowMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.follow = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "AccountID", Multi: true},
		{Name: "TargetAccountID", Multi: true},
	}, func(f1 *gtsmodel.Follow) *gtsmodel.Follow {
		f2 := new(gtsmodel.Follow)
		*f2 = *f1
		return f2
	}, cap)

	c.follow.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initFollowIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheFollowIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.followIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initFollowRequest() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofFollowRequest(), // model in-mem size.
		config.GetCacheFollowRequestMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.followRequest = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "AccountID", Multi: true},
		{Name: "TargetAccountID", Multi: true},
	}, func(f1 *gtsmodel.FollowRequest) *gtsmodel.FollowRequest {
		f2 := new(gtsmodel.FollowRequest)
		*f2 = *f1
		return f2
	}, cap)

	c.followRequest.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initFollowRequestIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheFollowRequestIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.followRequestIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initInReplyToIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheInReplyToIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.inReplyToIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initInstance() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofInstance(), // model in-mem size.
		config.GetCacheInstanceMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.instance = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "Domain"},
	}, func(i1 *gtsmodel.Instance) *gtsmodel.Instance {
		i2 := new(gtsmodel.Instance)
		*i2 = *i1
		return i1
	}, cap)

	c.instance.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initList() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofList(), // model in-mem size.
		config.GetCacheListMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.list = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(l1 *gtsmodel.List) *gtsmodel.List {
		l2 := new(gtsmodel.List)
		*l2 = *l1
		return l2
	}, cap)

	c.list.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initListEntry() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofListEntry(), // model in-mem size.
		config.GetCacheListEntryMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.listEntry = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "ListID", Multi: true},
		{Name: "FollowID", Multi: true},
	}, func(l1 *gtsmodel.ListEntry) *gtsmodel.ListEntry {
		l2 := new(gtsmodel.ListEntry)
		*l2 = *l1
		return l2
	}, cap)

	c.listEntry.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initMarker() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofMarker(), // model in-mem size.
		config.GetCacheMarkerMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.marker = result.New([]result.Lookup{
		{Name: "AccountID.Name"},
	}, func(m1 *gtsmodel.Marker) *gtsmodel.Marker {
		m2 := new(gtsmodel.Marker)
		*m2 = *m1
		return m2
	}, cap)

	c.marker.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initMedia() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofMedia(), // model in-mem size.
		config.GetCacheMediaMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.media = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.MediaAttachment) *gtsmodel.MediaAttachment {
		m2 := new(gtsmodel.MediaAttachment)
		*m2 = *m1
		return m2
	}, cap)

	c.media.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initMention() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofMention(), // model in-mem size.
		config.GetCacheMentionMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.mention = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(m1 *gtsmodel.Mention) *gtsmodel.Mention {
		m2 := new(gtsmodel.Mention)
		*m2 = *m1
		return m2
	}, cap)

	c.mention.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initNotification() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofNotification(), // model in-mem size.
		config.GetCacheNotificationMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.notification = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "NotificationType.TargetAccountID.OriginAccountID.StatusID"},
	}, func(n1 *gtsmodel.Notification) *gtsmodel.Notification {
		n2 := new(gtsmodel.Notification)
		*n2 = *n1
		return n2
	}, cap)

	c.notification.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initReport() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofReport(), // model in-mem size.
		config.GetCacheReportMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.report = result.New([]result.Lookup{
		{Name: "ID"},
	}, func(r1 *gtsmodel.Report) *gtsmodel.Report {
		r2 := new(gtsmodel.Report)
		*r2 = *r1
		return r2
	}, cap)

	c.report.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatus() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofStatus(), // model in-mem size.
		config.GetCacheStatusMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.status = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "BoostOfID.AccountID"},
	}, func(s1 *gtsmodel.Status) *gtsmodel.Status {
		s2 := new(gtsmodel.Status)
		*s2 = *s1
		return s2
	}, cap)

	c.status.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatusFave() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofStatusFave(), // model in-mem size.
		config.GetCacheStatusFaveMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.statusFave = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.StatusID"},
		{Name: "StatusID", Multi: true},
	}, func(f1 *gtsmodel.StatusFave) *gtsmodel.StatusFave {
		f2 := new(gtsmodel.StatusFave)
		*f2 = *f1
		return f2
	}, cap)

	c.statusFave.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatusFaveIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheStatusFaveIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.statusFaveIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initTag() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofTag(), // model in-mem size.
		config.GetCacheTagMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.tag = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "Name"},
	}, func(m1 *gtsmodel.Tag) *gtsmodel.Tag {
		m2 := new(gtsmodel.Tag)
		*m2 = *m1
		return m2
	}, cap)

	c.tag.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initTombstone() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofTombstone(), // model in-mem size.
		config.GetCacheTombstoneMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.tombstone = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
	}, func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}, cap)

	c.tombstone.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initUser() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofUser(), // model in-mem size.
		config.GetCacheUserMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

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
	}, cap)

	c.user.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initWebfinger() {
	// Calculate maximum cache size.
	cap := calculateCacheMax(
		sizeofURIStr, sizeofURIStr,
		config.GetCacheWebfingerMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.webfinger = ttl.New[string, string](
		0,
		cap,
		24*time.Hour,
	)
}
