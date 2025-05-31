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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-structr"
)

type MutesCache struct {
	StructCache[*CachedMute]
}

func (c *Caches) initMutes() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofMute(), // model in-mem size.
		config.GetCacheMutesMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(m1 *CachedMute) *CachedMute {
		m2 := new(CachedMute)
		*m2 = *m1
		return m2
	}

	c.Mutes.Init(structr.CacheConfig[*CachedMute]{
		Indices: []structr.IndexConfig{
			{Fields: "RequesterID,StatusID"},
			{Fields: "RequesterID,ThreadID", Multiple: true},
			{Fields: "StatusID", Multiple: true},
			{Fields: "ThreadID", Multiple: true},
			{Fields: "RequesterID", Multiple: true},
		},
		MaxSize: cap,
		IgnoreErr: func(err error) bool {
			// don't cache any errors,
			// it gets a little too tricky
			// otherwise with ensuring
			// errors are cleared out
			return true
		},
		Copy: copyF,
	})
}

// CachedMute contains the details
// of a cached mute lookup.
type CachedMute struct {

	// StatusID is the ID of the
	// status this is a result for.
	StatusID string

	// ThreadID is the ID of the
	// thread status is a part of.
	ThreadID string

	// RequesterID is the ID of the requesting
	// account for this user mute lookup.
	RequesterID string

	// Mute indicates whether ItemID
	// is muted by RequesterID.
	Mute bool

	// MuteExpiry stores the time at which
	// (if any) the stored mute value expires.
	MuteExpiry time.Time

	// Notifications indicates whether
	// this mute should prevent notifications
	// being shown for ItemID to RequesterID.
	Notifications bool

	// NotificationExpiry stores the time at which
	// (if any) the stored notification value expires.
	NotificationExpiry time.Time
}

// MuteExpired returns whether the mute value has expired.
func (m *CachedMute) MuteExpired(now time.Time) bool {
	return !m.MuteExpiry.IsZero() && !m.MuteExpiry.After(now)
}

// NotificationExpired returns whether the notification mute value has expired.
func (m *CachedMute) NotificationExpired(now time.Time) bool {
	return !m.NotificationExpiry.IsZero() && !m.NotificationExpiry.After(now)
}
