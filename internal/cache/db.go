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

	"codeberg.org/gruf/go-cache/v3/simple"
	"codeberg.org/gruf/go-cache/v3/ttl"
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/cache/domain"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type GTSCaches struct {
	// Account provides access to the gtsmodel Account database cache.
	Account structr.Cache[*gtsmodel.Account]

	// AccountNote provides access to the gtsmodel Note database cache.
	AccountNote structr.Cache[*gtsmodel.AccountNote]

	// Application provides access to the gtsmodel Application database cache.
	Application structr.Cache[*gtsmodel.Application]

	// Block provides access to the gtsmodel Block (account) database cache.
	Block structr.Cache[*gtsmodel.Block]

	// FollowIDs provides access to the block IDs database cache.
	BlockIDs *SliceCache[string]

	// BoostOfIDs provides access to the boost of IDs list database cache.
	BoostOfIDs *SliceCache[string]

	// DomainAllow provides access to the domain allow database cache.
	DomainAllow *domain.Cache

	// DomainBlock provides access to the domain block database cache.
	DomainBlock *domain.Cache

	// Emoji provides access to the gtsmodel Emoji database cache.
	Emoji structr.Cache[*gtsmodel.Emoji]

	// EmojiCategory provides access to the gtsmodel EmojiCategory database cache.
	EmojiCategory structr.Cache[*gtsmodel.EmojiCategory]

	// Follow provides access to the gtsmodel Follow database cache.
	Follow structr.Cache[*gtsmodel.Follow]

	// FollowIDs provides access to the follower / following IDs database cache.
	// THIS CACHE IS KEYED AS THE FOLLOWING {prefix}{accountID} WHERE PREFIX IS:
	// - '>'  for following IDs
	// - 'l>' for local following IDs
	// - '<'  for follower IDs
	// - 'l<' for local follower IDs
	FollowIDs *SliceCache[string]

	// FollowRequest provides access to the gtsmodel FollowRequest database cache.
	FollowRequest structr.Cache[*gtsmodel.FollowRequest]

	// FollowRequestIDs provides access to the follow requester / requesting IDs database
	// cache. THIS CACHE IS KEYED AS THE FOLLOWING {prefix}{accountID} WHERE PREFIX IS:
	// - '>'  for following IDs
	// - '<'  for follower IDs
	FollowRequestIDs *SliceCache[string]

	// Instance provides access to the gtsmodel Instance database cache.
	Instance structr.Cache[*gtsmodel.Instance]

	// InReplyToIDs provides access to the status in reply to IDs list database cache.
	InReplyToIDs *SliceCache[string]

	// List provides access to the gtsmodel List database cache.
	List structr.Cache[*gtsmodel.List]

	// ListEntry provides access to the gtsmodel ListEntry database cache.
	ListEntry structr.Cache[*gtsmodel.ListEntry]

	// Marker provides access to the gtsmodel Marker database cache.
	Marker structr.Cache[*gtsmodel.Marker]

	// Media provides access to the gtsmodel Media database cache.
	Media structr.Cache[*gtsmodel.MediaAttachment]

	// Mention provides access to the gtsmodel Mention database cache.
	Mention structr.Cache[*gtsmodel.Mention]

	// Notification provides access to the gtsmodel Notification database cache.
	Notification structr.Cache[*gtsmodel.Notification]

	// Poll provides access to the gtsmodel Poll database cache.
	Poll structr.Cache[*gtsmodel.Poll]

	// PollVote provides access to the gtsmodel PollVote database cache.
	PollVote structr.Cache[*gtsmodel.PollVote]

	// PollVoteIDs provides access to the poll vote IDs list database cache.
	PollVoteIDs *SliceCache[string]

	// Report provides access to the gtsmodel Report database cache.
	Report structr.Cache[*gtsmodel.Report]

	// Status provides access to the gtsmodel Status database cache.
	Status structr.Cache[*gtsmodel.Status]

	// StatusFave provides access to the gtsmodel StatusFave database cache.
	StatusFave structr.Cache[*gtsmodel.StatusFave]

	// StatusFaveIDs provides access to the status fave IDs list database cache.
	StatusFaveIDs *SliceCache[string]

	// Tag provides access to the gtsmodel Tag database cache.
	Tag structr.Cache[*gtsmodel.Tag]

	// Tombstone provides access to the gtsmodel Tombstone database cache.
	Tombstone structr.Cache[*gtsmodel.Tombstone]

	// ThreadMute provides access to the gtsmodel ThreadMute database cache.
	ThreadMute structr.Cache[*gtsmodel.ThreadMute]

	// User provides access to the gtsmodel User database cache.
	User structr.Cache[*gtsmodel.User]

	// Webfinger provides access to the webfinger URL cache.
	// TODO: move out of GTS caches since unrelated to DB.
	Webfinger *ttl.Cache[string, string] // TTL=24hr, sweep=5min
}

// Start will attempt to start all of the gtsmodel caches, or panic.
func (c *GTSCaches) Start() {
	tryUntil("starting *gtsmodel.Webfinger cache", 5, func() bool {
		return c.Webfinger.Start(5 * time.Minute)
	})
}

// Stop will attempt to stop all of the gtsmodel caches, or panic.
func (c *GTSCaches) Stop() {
	tryUntil("stopping *gtsmodel.Webfinger cache", 5, c.Webfinger.Stop)
}

// NOTE:
// all of the below init functions
// are receivers to the main cache
// struct type, not the database cache
// struct type, in order to get access
// to the full suite of caches for
// our invalidate function hooks.

func (c *Caches) initAccount() {
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

	c.GTS.Account.Init(structr.Config[*gtsmodel.Account]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
			{Fields: "URL"},
			{Fields: "Username,Domain", AllowZero: true},
			{Fields: "PublicKeyURI"},
			{Fields: "InboxURI"},
			{Fields: "OutboxURI"},
			{Fields: "FollowersURI"},
			{Fields: "FollowingURI"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateAccount,
	})
}

func (c *Caches) initAccountNote() {
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

	c.GTS.AccountNote.Init(structr.Config[*gtsmodel.AccountNote]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "AccountID,TargetAccountID"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initApplication() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofApplication(), // model in-mem size.
		config.GetCacheApplicationMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(a1 *gtsmodel.Application) *gtsmodel.Application {
		a2 := new(gtsmodel.Application)
		*a2 = *a1
		return a2
	}

	c.GTS.Application.Init(structr.Config[*gtsmodel.Application]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "ClientID"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initBlock() {
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

	c.GTS.Block.Init(structr.Config[*gtsmodel.Block]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
			{Fields: "AccountID,TargetAccountID"},
			{Fields: "AccountID", Multiple: true},
			{Fields: "TargetAccountID", Multiple: true},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateBlock,
	})
}

func (c *Caches) initBlockIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheBlockIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.BlockIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initBoostOfIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheBoostOfIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.BoostOfIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initDomainAllow() {
	c.GTS.DomainAllow = new(domain.Cache)
}

func (c *Caches) initDomainBlock() {
	c.GTS.DomainBlock = new(domain.Cache)
}

func (c *Caches) initEmoji() {
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

	c.GTS.Emoji.Init(structr.Config[*gtsmodel.Emoji]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
			{Fields: "Shortcode,Domain", AllowZero: true},
			{Fields: "ImageStaticURL"},
			{Fields: "CategoryID", Multiple: true},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initEmojiCategory() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofEmojiCategory(), // model in-mem size.
		config.GetCacheEmojiCategoryMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(c1 *gtsmodel.EmojiCategory) *gtsmodel.EmojiCategory {
		c2 := new(gtsmodel.EmojiCategory)
		*c2 = *c1
		return c2
	}

	c.GTS.EmojiCategory.Init(structr.Config[*gtsmodel.EmojiCategory]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "Name"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateEmojiCategory,
	})
}

func (c *Caches) initFollow() {
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

	c.GTS.Follow.Init(structr.Config[*gtsmodel.Follow]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
			{Fields: "AccountID,TargetAccountID"},
			{Fields: "AccountID", Multiple: true},
			{Fields: "TargetAccountID", Multiple: true},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateFollow,
	})
}

func (c *Caches) initFollowIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheFollowIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.FollowIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initFollowRequest() {
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

	c.GTS.FollowRequest.Init(structr.Config[*gtsmodel.FollowRequest]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
			{Fields: "AccountID,TargetAccountID"},
			{Fields: "AccountID", Multiple: true},
			{Fields: "TargetAccountID", Multiple: true},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateFollowRequest,
	})
}

func (c *Caches) initFollowRequestIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheFollowRequestIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.FollowRequestIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initInReplyToIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheInReplyToIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.InReplyToIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initInstance() {
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

	c.GTS.Instance.Init(structr.Config[*gtsmodel.Instance]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "Domain"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initList() {
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

	c.GTS.List.Init(structr.Config[*gtsmodel.List]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateList,
	})
}

func (c *Caches) initListEntry() {
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

	c.GTS.ListEntry.Init(structr.Config[*gtsmodel.ListEntry]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "ListID", Multiple: true},
			{Fields: "FollowID", Multiple: true},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initMarker() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofMarker(), // model in-mem size.
		config.GetCacheMarkerMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(m1 *gtsmodel.Marker) *gtsmodel.Marker {
		m2 := new(gtsmodel.Marker)
		*m2 = *m1
		return m2
	}

	c.GTS.Marker.Init(structr.Config[*gtsmodel.Marker]{
		Indices: []structr.IndexConfig{
			{Fields: "AccountID,Name"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initMedia() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofMedia(), // model in-mem size.
		config.GetCacheMediaMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(m1 *gtsmodel.MediaAttachment) *gtsmodel.MediaAttachment {
		m2 := new(gtsmodel.MediaAttachment)
		*m2 = *m1
		return m2
	}

	c.GTS.Media.Init(structr.Config[*gtsmodel.MediaAttachment]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateMedia,
	})
}

func (c *Caches) initMention() {
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

	c.GTS.Mention.Init(structr.Config[*gtsmodel.Mention]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initNotification() {
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

	c.GTS.Notification.Init(structr.Config[*gtsmodel.Notification]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "NotificationType,TargetAccountID,OriginAccountID,StatusID", AllowZero: true},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initPoll() {
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

	c.GTS.Poll.Init(structr.Config[*gtsmodel.Poll]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "StatusID"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidatePoll,
	})
}

func (c *Caches) initPollVote() {
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

	c.GTS.PollVote.Init(structr.Config[*gtsmodel.PollVote]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "PollID", Multiple: true},
			{Fields: "PollID,AccountID"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidatePollVote,
	})
}

func (c *Caches) initPollVoteIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCachePollVoteIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.PollVoteIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initReport() {
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

	c.GTS.Report.Init(structr.Config[*gtsmodel.Report]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initStatus() {
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

	c.GTS.Status.Init(structr.Config[*gtsmodel.Status]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
			{Fields: "URL"},
			{Fields: "PollID"},
			{Fields: "BoostOfID,AccountID"},
			{Fields: "ThreadID", Multiple: true},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateStatus,
	})
}

func (c *Caches) initStatusFave() {
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

	c.GTS.StatusFave.Init(structr.Config[*gtsmodel.StatusFave]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "AccountID,StatusID"},
			{Fields: "StatusID", Multiple: true},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateStatusFave,
	})
}

func (c *Caches) initStatusFaveIDs() {
	// Calculate maximum cache size.
	cap := calculateSliceCacheMax(
		config.GetCacheStatusFaveIDsMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.StatusFaveIDs = &SliceCache[string]{Cache: simple.New[string, []string](
		0,
		cap,
	)}
}

func (c *Caches) initTag() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofTag(), // model in-mem size.
		config.GetCacheTagMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(m1 *gtsmodel.Tag) *gtsmodel.Tag {
		m2 := new(gtsmodel.Tag)
		*m2 = *m1
		return m2
	}

	c.GTS.Tag.Init(structr.Config[*gtsmodel.Tag]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "Name"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initThreadMute() {
	cap := calculateResultCacheMax(
		sizeOfThreadMute(), // model in-mem size.
		config.GetCacheThreadMuteMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(t1 *gtsmodel.ThreadMute) *gtsmodel.ThreadMute {
		t2 := new(gtsmodel.ThreadMute)
		*t2 = *t1
		return t2
	}

	c.GTS.ThreadMute.Init(structr.Config[*gtsmodel.ThreadMute]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "ThreadID", Multiple: true},
			{Fields: "AccountID", Multiple: true},
			{Fields: "ThreadID,AccountID"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initTombstone() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofTombstone(), // model in-mem size.
		config.GetCacheTombstoneMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(t1 *gtsmodel.Tombstone) *gtsmodel.Tombstone {
		t2 := new(gtsmodel.Tombstone)
		*t2 = *t1
		return t2
	}

	c.GTS.Tombstone.Init(structr.Config[*gtsmodel.Tombstone]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "URI"},
		},
		MaxSize:   cap,
		IgnoreErr: ignoreErrors,
		CopyValue: copyF,
	})
}

func (c *Caches) initUser() {
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

	c.GTS.User.Init(structr.Config[*gtsmodel.User]{
		Indices: []structr.IndexConfig{
			{Fields: "ID"},
			{Fields: "AccountID"},
			{Fields: "Email"},
			{Fields: "ConfirmationToken"},
			{Fields: "ExternalID"},
		},
		MaxSize:    cap,
		IgnoreErr:  ignoreErrors,
		CopyValue:  copyF,
		Invalidate: c.OnInvalidateUser,
	})
}

func (c *Caches) initWebfinger() {
	// Calculate maximum cache size.
	cap := calculateCacheMax(
		sizeofURIStr, sizeofURIStr,
		config.GetCacheWebfingerMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.GTS.Webfinger = new(ttl.Cache[string, string])
	c.GTS.Webfinger.Init(
		0,
		cap,
		24*time.Hour,
	)
}
