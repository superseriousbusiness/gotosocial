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
	poll             *result.Cache[*gtsmodel.Poll]
	pollVote         *result.Cache[*gtsmodel.PollVote]
	pollVoteIDs      *SliceCache[string]
	report           *result.Cache[*gtsmodel.Report]
	status           *result.Cache[*gtsmodel.Status]
	statusFave       *result.Cache[*gtsmodel.StatusFave]
	statusFaveIDs    *SliceCache[string]
	tag              *result.Cache[*gtsmodel.Tag]
	threadMute       *result.Cache[*gtsmodel.ThreadMute]
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
	c.initPoll()
	c.initPollVote()
	c.initPollVoteIDs()
	c.initReport()
	c.initStatus()
	c.initStatusFave()
	c.initTag()
	c.initThreadMute()
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

// Poll provides access to the gtsmodel Poll database cache.
func (c *GTSCaches) Poll() *result.Cache[*gtsmodel.Poll] {
	return c.poll
}

// PollVote provides access to the gtsmodel PollVote database cache.
func (c *GTSCaches) PollVote() *result.Cache[*gtsmodel.PollVote] {
	return c.pollVote
}

// PollVoteIDs provides access to the poll vote IDs list database cache.
func (c *GTSCaches) PollVoteIDs() *SliceCache[string] {
	return c.pollVoteIDs
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

// StatusFaveIDs provides access to the status fave IDs list database cache.
func (c *GTSCaches) StatusFaveIDs() *SliceCache[string] {
	return c.statusFaveIDs
}

// Tag provides access to the gtsmodel Tag database cache.
func (c *GTSCaches) Tag() *result.Cache[*gtsmodel.Tag] {
	return c.tag
}

// Tombstone provides access to the gtsmodel Tombstone database cache.
func (c *GTSCaches) Tombstone() *result.Cache[*gtsmodel.Tombstone] {
	return c.tombstone
}

// ThreadMute provides access to the gtsmodel ThreadMute database cache.
func (c *GTSCaches) ThreadMute() *result.Cache[*gtsmodel.ThreadMute] {
	return c.threadMute
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

	copyF := func(a1 *gtsmodel.Account) *gtsmodel.Account {
		a2 := new(gtsmodel.Account)
		*a2 = *a1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/account.go.
		a2.AvatarMediaAttachment = nil
		a2.HeaderMediaAttachment = nil
		a2.Emojis = nil

		return a2
	}

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
	}, copyF, cap)

	c.account.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initAccountNote() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofAccountNote(), // model in-mem size.
		config.GetCacheAccountNoteMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(n1 *gtsmodel.AccountNote) *gtsmodel.AccountNote {
		n2 := new(gtsmodel.AccountNote)
		*n2 = *n1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/relationship_note.go.
		n2.Account = nil
		n2.TargetAccount = nil

		return n2
	}

	c.accountNote = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.TargetAccountID"},
	}, copyF, cap)

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

	copyF := func(b1 *gtsmodel.Block) *gtsmodel.Block {
		b2 := new(gtsmodel.Block)
		*b2 = *b1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/relationship_block.go.
		b2.Account = nil
		b2.TargetAccount = nil

		return b2
	}

	c.block = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "AccountID", Multi: true},
		{Name: "TargetAccountID", Multi: true},
	}, copyF, cap)

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

	copyF := func(e1 *gtsmodel.Emoji) *gtsmodel.Emoji {
		e2 := new(gtsmodel.Emoji)
		*e2 = *e1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/emoji.go.
		e2.Category = nil

		return e2
	}

	c.emoji = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "Shortcode.Domain", AllowZero: true /* domain can be zero i.e. "" */},
		{Name: "ImageStaticURL"},
		{Name: "CategoryID", Multi: true},
	}, copyF, cap)

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

	copyF := func(f1 *gtsmodel.Follow) *gtsmodel.Follow {
		f2 := new(gtsmodel.Follow)
		*f2 = *f1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/relationship_follow.go.
		f2.Account = nil
		f2.TargetAccount = nil

		return f2
	}

	c.follow = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "AccountID", Multi: true},
		{Name: "TargetAccountID", Multi: true},
	}, copyF, cap)

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

	copyF := func(f1 *gtsmodel.FollowRequest) *gtsmodel.FollowRequest {
		f2 := new(gtsmodel.FollowRequest)
		*f2 = *f1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/relationship_follow_req.go.
		f2.Account = nil
		f2.TargetAccount = nil

		return f2
	}

	c.followRequest = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "AccountID.TargetAccountID"},
		{Name: "AccountID", Multi: true},
		{Name: "TargetAccountID", Multi: true},
	}, copyF, cap)

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

	copyF := func(i1 *gtsmodel.Instance) *gtsmodel.Instance {
		i2 := new(gtsmodel.Instance)
		*i2 = *i1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/instance.go.
		i2.DomainBlock = nil
		i2.ContactAccount = nil

		return i1
	}

	c.instance = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "Domain"},
	}, copyF, cap)

	c.instance.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initList() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofList(), // model in-mem size.
		config.GetCacheListMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(l1 *gtsmodel.List) *gtsmodel.List {
		l2 := new(gtsmodel.List)
		*l2 = *l1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/list.go.
		l2.Account = nil
		l2.ListEntries = nil

		return l2
	}

	c.list = result.New([]result.Lookup{
		{Name: "ID"},
	}, copyF, cap)

	c.list.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initListEntry() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofListEntry(), // model in-mem size.
		config.GetCacheListEntryMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(l1 *gtsmodel.ListEntry) *gtsmodel.ListEntry {
		l2 := new(gtsmodel.ListEntry)
		*l2 = *l1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/list.go.
		l2.Follow = nil

		return l2
	}

	c.listEntry = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "ListID", Multi: true},
		{Name: "FollowID", Multi: true},
	}, copyF, cap)

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

	copyF := func(m1 *gtsmodel.Mention) *gtsmodel.Mention {
		m2 := new(gtsmodel.Mention)
		*m2 = *m1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/mention.go.
		m2.Status = nil
		m2.OriginAccount = nil
		m2.TargetAccount = nil

		return m2
	}

	c.mention = result.New([]result.Lookup{
		{Name: "ID"},
	}, copyF, cap)

	c.mention.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initNotification() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofNotification(), // model in-mem size.
		config.GetCacheNotificationMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(n1 *gtsmodel.Notification) *gtsmodel.Notification {
		n2 := new(gtsmodel.Notification)
		*n2 = *n1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/notification.go.
		n2.Status = nil
		n2.OriginAccount = nil
		n2.TargetAccount = nil

		return n2
	}

	c.notification = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "NotificationType.TargetAccountID.OriginAccountID.StatusID"},
	}, copyF, cap)

	c.notification.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initPoll() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofPoll(), // model in-mem size.
		config.GetCachePollMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(p1 *gtsmodel.Poll) *gtsmodel.Poll {
		p2 := new(gtsmodel.Poll)
		*p2 = *p1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/poll.go.
		p2.Status = nil

		// Don't include ephemeral fields
		// which are only expected to be
		// set on ONE poll instance.
		p2.Closing = false

		return p2
	}

	c.poll = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "StatusID"},
	}, copyF, cap)

	c.poll.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initPollVote() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofPollVote(), // model in-mem size.
		config.GetCachePollVoteMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(v1 *gtsmodel.PollVote) *gtsmodel.PollVote {
		v2 := new(gtsmodel.PollVote)
		*v2 = *v1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/poll.go.
		v2.Account = nil
		v2.Poll = nil

		return v2
	}

	c.pollVote = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "PollID.AccountID"},
		{Name: "PollID", Multi: true},
	}, copyF, cap)

	c.pollVote.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initPollVoteIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCachePollVoteIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.pollVoteIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *GTSCaches) initReport() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofReport(), // model in-mem size.
		config.GetCacheReportMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(r1 *gtsmodel.Report) *gtsmodel.Report {
		r2 := new(gtsmodel.Report)
		*r2 = *r1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/report.go.
		r2.Account = nil
		r2.TargetAccount = nil
		r2.Statuses = nil
		r2.Rules = nil
		r2.ActionTakenByAccount = nil

		return r2
	}

	c.report = result.New([]result.Lookup{
		{Name: "ID"},
	}, copyF, cap)

	c.report.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatus() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofStatus(), // model in-mem size.
		config.GetCacheStatusMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(s1 *gtsmodel.Status) *gtsmodel.Status {
		s2 := new(gtsmodel.Status)
		*s2 = *s1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/status.go.
		s2.Account = nil
		s2.InReplyTo = nil
		s2.InReplyToAccount = nil
		s2.BoostOf = nil
		s2.BoostOfAccount = nil
		s2.Poll = nil
		s2.Attachments = nil
		s2.Tags = nil
		s2.Mentions = nil
		s2.Emojis = nil
		s2.CreatedWithApplication = nil

		return s2
	}

	c.status = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "URI"},
		{Name: "URL"},
		{Name: "PollID"},
		{Name: "BoostOfID.AccountID"},
		{Name: "ThreadID", Multi: true},
	}, copyF, cap)

	c.status.IgnoreErrors(ignoreErrors)
}

func (c *GTSCaches) initStatusFave() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofStatusFave(), // model in-mem size.
		config.GetCacheStatusFaveMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(f1 *gtsmodel.StatusFave) *gtsmodel.StatusFave {
		f2 := new(gtsmodel.StatusFave)
		*f2 = *f1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/statusfave.go.
		f2.Account = nil
		f2.TargetAccount = nil
		f2.Status = nil

		return f2
	}

	c.statusFave = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID.StatusID"},
		{Name: "StatusID", Multi: true},
	}, copyF, cap)

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

func (c *GTSCaches) initThreadMute() {
	cap := calculateResultCacheMax(
		sizeOfThreadMute(), // model in-mem size.
		config.GetCacheThreadMuteMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.threadMute = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "ThreadID", Multi: true},
		{Name: "AccountID", Multi: true},
		{Name: "ThreadID.AccountID"},
	}, func(t1 *gtsmodel.ThreadMute) *gtsmodel.ThreadMute {
		t2 := new(gtsmodel.ThreadMute)
		*t2 = *t1
		return t2
	}, cap)

	c.threadMute.IgnoreErrors(ignoreErrors)
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

	copyF := func(u1 *gtsmodel.User) *gtsmodel.User {
		u2 := new(gtsmodel.User)
		*u2 = *u1

		// Don't include ptr fields that
		// will be populated separately.
		// See internal/db/bundb/user.go.
		u2.Account = nil

		return u2
	}

	c.user = result.New([]result.Lookup{
		{Name: "ID"},
		{Name: "AccountID"},
		{Name: "Email"},
		{Name: "ConfirmationToken"},
		{Name: "ExternalID"},
	}, copyF, cap)

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
