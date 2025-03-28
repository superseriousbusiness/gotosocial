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
	"github.com/superseriousbusiness/gotosocial/internal/cache/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type TimelineCaches struct {

	// Home ...
	Home timeline.StatusTimelines

	// List ...
	List timeline.StatusTimelines

	// Public ...
	Public timeline.StatusTimeline

	// Local ...
	Local timeline.StatusTimeline
}

func (c *Caches) initHomeTimelines() {
	// Per-user cache
	// so use smaller.
	cap := 400

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Home.Init(cap)
}

func (c *Caches) initListTimelines() {
	// Per-user cache
	// so use smaller.
	cap := 400

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.List.Init(cap)
}

func (c *Caches) initPublicTimeline() {
	// Global cache so
	// allow larger.
	cap := 800

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Public.Init(cap)
}

func (c *Caches) initLocalTimeline() {
	// Global cache so
	// allow larger.
	cap := 800

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Local.Init(cap)
}
