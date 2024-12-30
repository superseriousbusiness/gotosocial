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
	"codeberg.org/gruf/go-structr"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/cache/timeline"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

type TimelineCaches struct {

	// Home ...
	Home TimelinesCache[*gtsmodel.Status]

	// List ...
	List TimelinesCache[*gtsmodel.Status]

	// Public ...
	Public timeline.StatusTimeline
}

func (c *Caches) initHomeTimelines() {
	cap := 1000

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Home.Init(structr.TimelineConfig[*gtsmodel.Status, string]{
		PKey: "StatusID",
		Indices: []structr.IndexConfig{
			{Fields: "StatusID"},
			{Fields: "AccountID"},
			{Fields: "BoostOfStatusID"},
			{Fields: "BoostOfAccountID"},
		},
		Copy: copyStatus,
	}, cap)
}

func (c *Caches) initListTimelines() {
	cap := 1000

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.List.Init(structr.TimelineConfig[*gtsmodel.Status, string]{
		PKey: "StatusID",
		Indices: []structr.IndexConfig{
			{Fields: "StatusID"},
			{Fields: "AccountID"},
			{Fields: "BoostOfStatusID"},
			{Fields: "BoostOfAccountID"},
		},
		Copy: copyStatus,
	}, cap)
}

func (c *Caches) initPublicTimeline() {
	cap := 1000

	log.Infof(nil, "cache size = %d", cap)

	c.Timelines.Public.Init(cap)
}

type TimelineStatus struct {

	// ID ...
	ID string

	// AccountID ...
	AccountID string

	// BoostOfID ...
	BoostOfID string

	// BoostOfAccountID ...
	BoostOfAccountID string

	// Local ...
	Local bool

	// Loaded is a temporary field that may be
	// set for a newly loaded timeline status
	// so that statuses don't need to be loaded
	// from the database twice in succession.
	//
	// i.e. this will only be set if the status
	// was newly inserted into the timeline cache.
	// for existing cache items this will be nil.
	Loaded *gtsmodel.Status

	// Prepared contains prepared frontend API
	// model for the referenced status. This may
	// or may-not be nil depending on whether the
	// status has been "unprepared" since the last
	// call to "prepare" the frontend model.
	Prepared *apimodel.Status
}

func (s *TimelineStatus) Copy() *TimelineStatus {
	var prepared *apimodel.Status
	if s.Prepared != nil {
		prepared = new(apimodel.Status)
		*prepared = *s.Prepared
	}
	return &TimelineStatus{
		ID:               s.ID,
		AccountID:        s.AccountID,
		BoostOfID:        s.BoostOfID,
		BoostOfAccountID: s.BoostOfAccountID,
		Loaded:           nil, // NEVER set
		Prepared:         prepared,
	}
}
