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
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type Caches struct {
	// GTS provides access to the collection of gtsmodel object caches.
	// (used by the database).
	GTS GTSCaches

	// AP provides access to the collection of ActivityPub object caches.
	// (planned to be used by the typeconverter).
	AP APCaches

	// Visibility provides access to the item visibility cache.
	// (used by the visibility filter).
	Visibility VisibilityCache

	// prevent pass-by-value.
	_ nocopy
}

// Init will (re)initialize both the GTS and AP cache collections.
// NOTE: the cache MUST NOT be in use anywhere, this is not thread-safe.
func (c *Caches) Init() {
	log.Infof(nil, "init: %p", c)

	c.GTS.Init()
	c.AP.Init()
	c.Visibility.Init()

	// Setup cache invalidate hooks.
	// !! READ THE METHOD COMMENT
	c.setuphooks()
}

// Start will start both the GTS and AP cache collections.
func (c *Caches) Start() {
	log.Infof(nil, "start: %p", c)

	c.GTS.Start()
	c.AP.Start()
	c.Visibility.Start()
}

// Stop will stop both the GTS and AP cache collections.
func (c *Caches) Stop() {
	log.Infof(nil, "stop: %p", c)

	c.GTS.Stop()
	c.AP.Stop()
	c.Visibility.Stop()
}

// setuphooks sets necessary cache invalidation hooks between caches,
// as an invalidation indicates a database INSERT / UPDATE / DELETE.
// NOTE THEY ARE ONLY CALLED WHEN THE ITEM IS IN THE CACHE, SO FOR
// HOOKS TO BE CALLED ON DELETE YOU MUST FIRST POPULATE IT IN THE CACHE.
func (c *Caches) setuphooks() {
	c.GTS.Account().SetInvalidateCallback(func(account *gtsmodel.Account) {
		// Invalidate account ID cached visibility.
		c.Visibility.Invalidate("ItemID", account.ID)
		c.Visibility.Invalidate("RequesterID", account.ID)

		// Invalidate this account's
		// following / follower lists.
		// (see FollowIDs() comment for details).
		c.GTS.FollowIDs().InvalidateAll(
			">"+account.ID,
			"l>"+account.ID,
			"<"+account.ID,
			"l<"+account.ID,
		)

		// Invalidate this account's
		// follow requesting / request lists.
		// (see FollowRequestIDs() comment for details).
		c.GTS.FollowRequestIDs().InvalidateAll(
			">"+account.ID,
			"<"+account.ID,
		)

		// Invalidate this account's block lists.
		c.GTS.BlockIDs().Invalidate(account.ID)
	})

	c.GTS.Block().SetInvalidateCallback(func(block *gtsmodel.Block) {
		// Invalidate block origin account ID cached visibility.
		c.Visibility.Invalidate("ItemID", block.AccountID)
		c.Visibility.Invalidate("RequesterID", block.AccountID)

		// Invalidate block target account ID cached visibility.
		c.Visibility.Invalidate("ItemID", block.TargetAccountID)
		c.Visibility.Invalidate("RequesterID", block.TargetAccountID)

		// Invalidate source account's block lists.
		c.GTS.BlockIDs().Invalidate(block.AccountID)
	})

	c.GTS.EmojiCategory().SetInvalidateCallback(func(category *gtsmodel.EmojiCategory) {
		// Invalidate any emoji in this category.
		c.GTS.Emoji().Invalidate("CategoryID", category.ID)
	})

	c.GTS.Follow().SetInvalidateCallback(func(follow *gtsmodel.Follow) {
		// Invalidate follow request with this same ID.
		c.GTS.FollowRequest().Invalidate("ID", follow.ID)

		// Invalidate any related list entries.
		c.GTS.ListEntry().Invalidate("FollowID", follow.ID)

		// Invalidate follow origin account ID cached visibility.
		c.Visibility.Invalidate("ItemID", follow.AccountID)
		c.Visibility.Invalidate("RequesterID", follow.AccountID)

		// Invalidate follow target account ID cached visibility.
		c.Visibility.Invalidate("ItemID", follow.TargetAccountID)
		c.Visibility.Invalidate("RequesterID", follow.TargetAccountID)

		// Invalidate source account's following
		// lists, and destination's follwer lists.
		// (see FollowIDs() comment for details).
		c.GTS.FollowIDs().InvalidateAll(
			">"+follow.AccountID,
			"l>"+follow.AccountID,
			"<"+follow.AccountID,
			"l<"+follow.AccountID,
			"<"+follow.TargetAccountID,
			"l<"+follow.TargetAccountID,
			">"+follow.TargetAccountID,
			"l>"+follow.TargetAccountID,
		)
	})

	c.GTS.FollowRequest().SetInvalidateCallback(func(followReq *gtsmodel.FollowRequest) {
		// Invalidate follow with this same ID.
		c.GTS.Follow().Invalidate("ID", followReq.ID)

		// Invalidate source account's followreq
		// lists, and destinations follow req lists.
		// (see FollowRequestIDs() comment for details).
		c.GTS.FollowRequestIDs().InvalidateAll(
			">"+followReq.AccountID,
			"<"+followReq.AccountID,
			">"+followReq.TargetAccountID,
			"<"+followReq.TargetAccountID,
		)
	})

	c.GTS.List().SetInvalidateCallback(func(list *gtsmodel.List) {
		// Invalidate all cached entries of this list.
		c.GTS.ListEntry().Invalidate("ListID", list.ID)
	})

	c.GTS.Media().SetInvalidateCallback(func(media *gtsmodel.MediaAttachment) {
		if *media.Avatar || *media.Header {
			// Invalidate cache of attaching account.
			c.GTS.Account().Invalidate("ID", media.AccountID)
		}

		if media.StatusID != "" {
			// Invalidate cache of attaching status.
			c.GTS.Status().Invalidate("ID", media.StatusID)
		}
	})

	c.GTS.Status().SetInvalidateCallback(func(status *gtsmodel.Status) {
		// Invalidate status ID cached visibility.
		c.Visibility.Invalidate("ItemID", status.ID)

		for _, id := range status.AttachmentIDs {
			// Invalidate each media by the IDs we're aware of.
			// This must be done as the status table is aware of
			// the media IDs in use before the media table is
			// aware of the status ID they are linked to.
			//
			// c.GTS.Media().Invalidate("StatusID") will not work.
			c.GTS.Media().Invalidate("ID", id)
		}

		// if status.BoostOfID != "" {
		// 	// Invalidate boost ID list of the original status.
		// 	c.GTS.BoostIDs().Invalidate(status.BoostOfID)
		// }

		if status.InReplyToID != "" {
			// Invalidate in reply to ID list of original status.
			c.GTS.InReplyToIDs().Invalidate(status.InReplyToID)
		}
	})

	c.GTS.StatusFave().SetInvalidateCallback(func(fave *gtsmodel.StatusFave) {
		// Invalidate status fave ID list for this status.
		c.GTS.StatusFaveIDs().Invalidate(fave.StatusID)
	})

	c.GTS.User().SetInvalidateCallback(func(user *gtsmodel.User) {
		// Invalidate local account ID cached visibility.
		c.Visibility.Invalidate("ItemID", user.AccountID)
		c.Visibility.Invalidate("RequesterID", user.AccountID)
	})
}

// Sweep will sweep all the available caches to ensure none
// are above threshold percent full to their total capacity.
//
// This helps with cache performance, as a full cache will
// require an eviction on every single write, which adds
// significant overhead to all cache writes.
func (c *Caches) Sweep(threshold float64) {
	c.GTS.Account().Trim(threshold)
	c.GTS.AccountNote().Trim(threshold)
	c.GTS.Block().Trim(threshold)
	c.GTS.BlockIDs().Trim(threshold)
	c.GTS.Emoji().Trim(threshold)
	c.GTS.EmojiCategory().Trim(threshold)
	c.GTS.Follow().Trim(threshold)
	c.GTS.FollowIDs().Trim(threshold)
	c.GTS.FollowRequest().Trim(threshold)
	c.GTS.FollowRequestIDs().Trim(threshold)
	c.GTS.Instance().Trim(threshold)
	c.GTS.List().Trim(threshold)
	c.GTS.ListEntry().Trim(threshold)
	c.GTS.Marker().Trim(threshold)
	c.GTS.Media().Trim(threshold)
	c.GTS.Mention().Trim(threshold)
	c.GTS.Notification().Trim(threshold)
	c.GTS.Report().Trim(threshold)
	c.GTS.Status().Trim(threshold)
	c.GTS.StatusFave().Trim(threshold)
	c.GTS.Tag().Trim(threshold)
	c.GTS.Tombstone().Trim(threshold)
	c.GTS.User().Trim(threshold)
	c.Visibility.Trim(threshold)
}
