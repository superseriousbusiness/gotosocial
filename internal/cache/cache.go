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

	"github.com/superseriousbusiness/gotosocial/internal/cache/headerfilter"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type Caches struct {

	// GTS provides access to the collection of
	// gtsmodel object caches. (used by the database).
	GTS GTSCaches

	// AllowHeaderFilters provides access to
	// the allow []headerfilter.Filter cache.
	AllowHeaderFilters headerfilter.Cache

	// BlockHeaderFilters provides access to
	// the block []headerfilter.Filter cache.
	BlockHeaderFilters headerfilter.Cache

	// Visibility provides access to the item visibility
	// cache. (used by the visibility filter).
	Visibility VisibilityCache

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
	c.initDomainAllow()
	c.initDomainBlock()
	c.initEmoji()
	c.initEmojiCategory()
	c.initFilter()
	c.initFilterKeyword()
	c.initFilterStatus()
	c.initFollow()
	c.initFollowIDs()
	c.initFollowRequest()
	c.initFollowRequestIDs()
	c.initInReplyToIDs()
	c.initInstance()
	c.initInteractionApproval()
	c.initList()
	c.initListEntry()
	c.initMarker()
	c.initMedia()
	c.initMention()
	c.initMove()
	c.initNotification()
	c.initPoll()
	c.initPollVote()
	c.initPollVoteIDs()
	c.initReport()
	c.initStatus()
	c.initStatusBookmark()
	c.initStatusBookmarkIDs()
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
	c.initVisibility()
}

// Start will start any caches that require a background
// routine, which usually means any kind of TTL caches.
func (c *Caches) Start() {
	log.Infof(nil, "start: %p", c)

	tryUntil("starting webfinger cache", 5, func() bool {
		return c.GTS.Webfinger.Start(5 * time.Minute)
	})
}

// Stop will stop any caches that require a background
// routine, which usually means any kind of TTL caches.
func (c *Caches) Stop() {
	log.Infof(nil, "stop: %p", c)

	tryUntil("stopping webfinger cache", 5, c.GTS.Webfinger.Stop)
}

// Sweep will sweep all the available caches to ensure none
// are above threshold percent full to their total capacity.
//
// This helps with cache performance, as a full cache will
// require an eviction on every single write, which adds
// significant overhead to all cache writes.
func (c *Caches) Sweep(threshold float64) {
	c.GTS.Account.Trim(threshold)
	c.GTS.AccountNote.Trim(threshold)
	c.GTS.AccountSettings.Trim(threshold)
	c.GTS.AccountStats.Trim(threshold)
	c.GTS.Application.Trim(threshold)
	c.GTS.Block.Trim(threshold)
	c.GTS.BlockIDs.Trim(threshold)
	c.GTS.BoostOfIDs.Trim(threshold)
	c.GTS.Client.Trim(threshold)
	c.GTS.Emoji.Trim(threshold)
	c.GTS.EmojiCategory.Trim(threshold)
	c.GTS.Filter.Trim(threshold)
	c.GTS.FilterKeyword.Trim(threshold)
	c.GTS.FilterStatus.Trim(threshold)
	c.GTS.Follow.Trim(threshold)
	c.GTS.FollowIDs.Trim(threshold)
	c.GTS.FollowRequest.Trim(threshold)
	c.GTS.FollowRequestIDs.Trim(threshold)
	c.GTS.InReplyToIDs.Trim(threshold)
	c.GTS.Instance.Trim(threshold)
	c.GTS.InteractionApproval.Trim(threshold)
	c.GTS.List.Trim(threshold)
	c.GTS.ListEntry.Trim(threshold)
	c.GTS.Marker.Trim(threshold)
	c.GTS.Media.Trim(threshold)
	c.GTS.Mention.Trim(threshold)
	c.GTS.Move.Trim(threshold)
	c.GTS.Notification.Trim(threshold)
	c.GTS.Poll.Trim(threshold)
	c.GTS.PollVote.Trim(threshold)
	c.GTS.PollVoteIDs.Trim(threshold)
	c.GTS.Report.Trim(threshold)
	c.GTS.Status.Trim(threshold)
	c.GTS.StatusBookmark.Trim(threshold)
	c.GTS.StatusBookmarkIDs.Trim(threshold)
	c.GTS.StatusFave.Trim(threshold)
	c.GTS.StatusFaveIDs.Trim(threshold)
	c.GTS.Tag.Trim(threshold)
	c.GTS.ThreadMute.Trim(threshold)
	c.GTS.Token.Trim(threshold)
	c.GTS.Tombstone.Trim(threshold)
	c.GTS.User.Trim(threshold)
	c.GTS.UserMute.Trim(threshold)
	c.GTS.UserMuteIDs.Trim(threshold)
	c.Visibility.Trim(threshold)
}
