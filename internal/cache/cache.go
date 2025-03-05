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

	"codeberg.org/gruf/go-cache/v3/ttl"
	"github.com/superseriousbusiness/gotosocial/internal/cache/headerfilter"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type Caches struct {

	// DB provides access to the collection of
	// gtsmodel object caches. (used by the database).
	DB DBCaches

	// AllowHeaderFilters provides access to
	// the allow []headerfilter.Filter cache.
	AllowHeaderFilters headerfilter.Cache

	// BlockHeaderFilters provides access to
	// the block []headerfilter.Filter cache.
	BlockHeaderFilters headerfilter.Cache

	// TTL cache of statuses -> filterable text fields.
	// To ensure up-to-date fields, cache is keyed as:
	// `[status.ID][status.UpdatedAt.Unix()]`
	StatusesFilterableFields *ttl.Cache[string, []string]

	// Visibility provides access to the item visibility
	// cache. (used by the visibility filter).
	Visibility VisibilityCache

	// Webfinger provides access to the webfinger URL cache.
	Webfinger *ttl.Cache[string, string] // TTL=24hr, sweep=5min

	// prevent pass-by-value.
	_ nocopy
}

// Init will (re)initialize both the GTS and AP cache collections.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *Caches) Init() {
	log.Infof(nil, "init: %p", c)

	c.initAccount()
	c.initAccountNote()
	c.initAccountSettings()
	c.initAccountStats()
	c.initApplication()
	c.initBlock()
	c.initBlockIDs()
	c.initBoostOfIDs()
	c.initClient()
	c.initConversation()
	c.initConversationLastStatusIDs()
	c.initDomainAllow()
	c.initDomainBlock()
	c.initDomainPermissionDraft()
	c.initDomainPermissionSubscription()
	c.initDomainPermissionExclude()
	c.initEmoji()
	c.initEmojiCategory()
	c.initFilter()
	c.initFilterKeyword()
	c.initFilterStatus()
	c.initFollow()
	c.initFollowIDs()
	c.initFollowRequest()
	c.initFollowRequestIDs()
	c.initFollowingTagIDs()
	c.initInReplyToIDs()
	c.initInstance()
	c.initInteractionRequest()
	c.initList()
	c.initListIDs()
	c.initListedIDs()
	c.initMarker()
	c.initMedia()
	c.initMention()
	c.initMove()
	c.initNotification()
	c.initPoll()
	c.initPollVote()
	c.initPollVoteIDs()
	c.initReport()
	c.initSinBinStatus()
	c.initStatus()
	c.initStatusBookmark()
	c.initStatusBookmarkIDs()
	c.initStatusEdit()
	c.initStatusFave()
	c.initStatusFaveIDs()
	c.initTag()
	c.initThreadMute()
	c.initToken()
	c.initTombstone()
	c.initUser()
	c.initUserMute()
	c.initUserMuteIDs()
	c.initWebfinger()
	c.initWebPushSubscription()
	c.initWebPushSubscriptionIDs()
	c.initVisibility()
	c.initStatusesFilterableFields()
}

// Start will start any caches that require a background
// routine, which usually means any kind of TTL caches.
func (c *Caches) Start() error {
	log.Infof(nil, "start: %p", c)

	if !c.Webfinger.Start(5 * time.Minute) {
		return gtserror.New("could not start webfinger cache")
	}

	if !c.StatusesFilterableFields.Start(5 * time.Minute) {
		return gtserror.New("could not start statusesFilterableFields cache")
	}

	return nil
}

// Stop will stop any caches that require a background
// routine, which usually means any kind of TTL caches.
func (c *Caches) Stop() {
	log.Infof(nil, "stop: %p", c)

	_ = c.Webfinger.Stop()
	_ = c.StatusesFilterableFields.Stop()
}

// Sweep will sweep all the available caches to ensure none
// are above threshold percent full to their total capacity.
//
// This helps with cache performance, as a full cache will
// require an eviction on every single write, which adds
// significant overhead to all cache writes.
func (c *Caches) Sweep(threshold float64) {
	c.DB.Account.Trim(threshold)
	c.DB.AccountNote.Trim(threshold)
	c.DB.AccountSettings.Trim(threshold)
	c.DB.AccountStats.Trim(threshold)
	c.DB.Application.Trim(threshold)
	c.DB.Block.Trim(threshold)
	c.DB.BlockIDs.Trim(threshold)
	c.DB.BoostOfIDs.Trim(threshold)
	c.DB.Client.Trim(threshold)
	c.DB.Conversation.Trim(threshold)
	c.DB.ConversationLastStatusIDs.Trim(threshold)
	c.DB.Emoji.Trim(threshold)
	c.DB.EmojiCategory.Trim(threshold)
	c.DB.Filter.Trim(threshold)
	c.DB.FilterKeyword.Trim(threshold)
	c.DB.FilterStatus.Trim(threshold)
	c.DB.Follow.Trim(threshold)
	c.DB.FollowIDs.Trim(threshold)
	c.DB.FollowRequest.Trim(threshold)
	c.DB.FollowRequestIDs.Trim(threshold)
	c.DB.FollowingTagIDs.Trim(threshold)
	c.DB.InReplyToIDs.Trim(threshold)
	c.DB.Instance.Trim(threshold)
	c.DB.InteractionRequest.Trim(threshold)
	c.DB.List.Trim(threshold)
	c.DB.ListIDs.Trim(threshold)
	c.DB.ListedIDs.Trim(threshold)
	c.DB.Marker.Trim(threshold)
	c.DB.Media.Trim(threshold)
	c.DB.Mention.Trim(threshold)
	c.DB.Move.Trim(threshold)
	c.DB.Notification.Trim(threshold)
	c.DB.Poll.Trim(threshold)
	c.DB.PollVote.Trim(threshold)
	c.DB.PollVoteIDs.Trim(threshold)
	c.DB.Report.Trim(threshold)
	c.DB.SinBinStatus.Trim(threshold)
	c.DB.Status.Trim(threshold)
	c.DB.StatusBookmark.Trim(threshold)
	c.DB.StatusBookmarkIDs.Trim(threshold)
	c.DB.StatusFave.Trim(threshold)
	c.DB.StatusFaveIDs.Trim(threshold)
	c.DB.Tag.Trim(threshold)
	c.DB.ThreadMute.Trim(threshold)
	c.DB.Token.Trim(threshold)
	c.DB.Tombstone.Trim(threshold)
	c.DB.User.Trim(threshold)
	c.DB.UserMute.Trim(threshold)
	c.DB.UserMuteIDs.Trim(threshold)
	c.Visibility.Trim(threshold)
}

func (c *Caches) initStatusesFilterableFields() {
	c.StatusesFilterableFields = new(ttl.Cache[string, []string])
	c.StatusesFilterableFields.Init(
		0,
		512,
		1*time.Hour,
	)
}

func (c *Caches) initWebfinger() {
	// Calculate maximum cache size.
	cap := calculateCacheMax(
		sizeofURIStr, sizeofURIStr,
		config.GetCacheWebfingerMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	c.Webfinger = new(ttl.Cache[string, string])
	c.Webfinger.Init(
		0,
		cap,
		24*time.Hour,
	)
}
