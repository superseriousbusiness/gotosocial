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
	c.GTS.Init()
	c.AP.Init()
	c.Visibility.Init()

	// Setup cache invalidate hooks.
	// !! READ THE METHOD COMMENT
	c.setuphooks()
}

// Start will start both the GTS and AP cache collections.
func (c *Caches) Start() {
	c.GTS.Start()
	c.AP.Start()
	c.Visibility.Start()
}

// Stop will stop both the GTS and AP cache collections.
func (c *Caches) Stop() {
	c.GTS.Stop()
	c.AP.Stop()
	c.Visibility.Stop()
}

// setuphooks sets necessary cache invalidation hooks between caches,
// as an invalidation indicates a database UPDATE / DELETE. INSERT is
// not handled by invalidation hooks and must be invalidated manually.
func (c *Caches) setuphooks() {
	c.GTS.Account().SetInvalidateCallback(func(account *gtsmodel.Account) {
		// Invalidate account ID cached visibility.
		c.Visibility.Invalidate("ItemID", account.ID)
		c.Visibility.Invalidate("RequesterID", account.ID)
	})

	c.GTS.Block().SetInvalidateCallback(func(block *gtsmodel.Block) {
		// Invalidate block origin account ID cached visibility.
		c.Visibility.Invalidate("ItemID", block.AccountID)
		c.Visibility.Invalidate("RequesterID", block.AccountID)

		// Invalidate block target account ID cached visibility.
		c.Visibility.Invalidate("ItemID", block.TargetAccountID)
		c.Visibility.Invalidate("RequesterID", block.TargetAccountID)
	})

	c.GTS.Follow().SetInvalidateCallback(func(follow *gtsmodel.Follow) {
		// Invalidate follow origin account ID cached visibility.
		c.Visibility.Invalidate("ItemID", follow.AccountID)
		c.Visibility.Invalidate("RequesterID", follow.AccountID)

		// Invalidate follow target account ID cached visibility.
		c.Visibility.Invalidate("ItemID", follow.TargetAccountID)
		c.Visibility.Invalidate("RequesterID", follow.TargetAccountID)
	})

	c.GTS.FollowRequest().SetInvalidateCallback(func(followReq *gtsmodel.FollowRequest) {
		// Invalidate follow request origin account ID cached visibility.
		c.Visibility.Invalidate("ItemID", followReq.AccountID)
		c.Visibility.Invalidate("RequesterID", followReq.AccountID)

		// Invalidate follow request target account ID cached visibility.
		c.Visibility.Invalidate("ItemID", followReq.TargetAccountID)
		c.Visibility.Invalidate("RequesterID", followReq.TargetAccountID)

		// Invalidate any cached follow corresponding to this request.
		c.GTS.Follow().Invalidate("AccountID.TargetAccountID", followReq.AccountID, followReq.TargetAccountID)
	})

	c.GTS.Status().SetInvalidateCallback(func(status *gtsmodel.Status) {
		// Invalidate status ID cached visibility.
		c.Visibility.Invalidate("ItemID", status.ID)
	})

	c.GTS.User().SetInvalidateCallback(func(user *gtsmodel.User) {
		// Invalidate local account ID cached visibility.
		c.Visibility.Invalidate("ItemID", user.AccountID)
		c.Visibility.Invalidate("RequesterID", user.AccountID)
	})
}
