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
	"code.superseriousbusiness.org/gotosocial/internal/cache/timeline"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

type TimelineCaches struct {
	// Home provides a concurrency-safe map of status timeline
	// caches for home timelines, keyed by home's account ID.
	Home timeline.StatusTimelines

	// List provides a concurrency-safe map of status
	// timeline caches for lists, keyed by list ID.
	List timeline.StatusTimelines
}

func (c *Caches) initHomeTimelines() {
	// TODO: configurable
	cap := 800

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Home.Init(cap)
}

func (c *Caches) initListTimelines() {
	// TODO: configurable
	cap := 800

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.List.Init(cap)
}
