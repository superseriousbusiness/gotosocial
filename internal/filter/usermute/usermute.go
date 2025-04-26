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

package usermute

import (
	"time"

	statusfilter "code.superseriousbusiness.org/gotosocial/internal/filter/status"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

type compiledUserMuteListEntry struct {
	ExpiresAt     time.Time
	Notifications bool
}

func (e *compiledUserMuteListEntry) appliesInContext(filterContext statusfilter.FilterContext) bool {
	switch filterContext {
	case statusfilter.FilterContextHome:
		return true
	case statusfilter.FilterContextNotifications:
		return e.Notifications
	case statusfilter.FilterContextPublic:
		return true
	case statusfilter.FilterContextThread:
		return true
	case statusfilter.FilterContextAccount:
		return false
	}
	return false
}

func (e *compiledUserMuteListEntry) expired(now time.Time) bool {
	return !e.ExpiresAt.IsZero() && !e.ExpiresAt.After(now)
}

type CompiledUserMuteList struct {
	byTargetAccountID map[string]compiledUserMuteListEntry
}

func NewCompiledUserMuteList(mutes []*gtsmodel.UserMute) (c *CompiledUserMuteList) {
	c = &CompiledUserMuteList{byTargetAccountID: make(map[string]compiledUserMuteListEntry, len(mutes))}
	for _, mute := range mutes {
		c.byTargetAccountID[mute.TargetAccountID] = compiledUserMuteListEntry{
			ExpiresAt:     mute.ExpiresAt,
			Notifications: *mute.Notifications,
		}
	}
	return
}

func (c *CompiledUserMuteList) Len() int {
	if c == nil {
		return 0
	}
	return len(c.byTargetAccountID)
}

func (c *CompiledUserMuteList) Matches(accountID string, filterContext statusfilter.FilterContext, now time.Time) bool {
	if c == nil {
		return false
	}
	e, found := c.byTargetAccountID[accountID]
	return found && e.appliesInContext(filterContext) && !e.expired(now)
}
