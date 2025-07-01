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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"codeberg.org/gruf/go-structr"
)

func (c *Caches) initStatusFilter() {
	// Calculate maximum cache size.
	cap := calculateResultCacheMax(
		sizeofStatusFilterResults(), // model in-mem size.
		config.GetCacheStatusFilterMemRatio(),
	)

	log.Infof(nil, "cache size = %d", cap)

	copyF := func(r1 *CachedStatusFilterResults) *CachedStatusFilterResults {
		r2 := new(CachedStatusFilterResults)
		*r2 = *r1
		return r2
	}

	c.StatusFilter.Init(structr.CacheConfig[*CachedStatusFilterResults]{
		Indices: []structr.IndexConfig{
			{Fields: "RequesterID,StatusID"},
			{Fields: "StatusID", Multiple: true},
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

const (
	KeyContextHome = iota
	KeyContextPublic
	KeyContextNotifs
	KeyContextThread
	KeyContextAccount
	keysLen // must always be last in list
)

// CachedStatusFilterResults contains the
// results of a cached status filter lookup.
type CachedStatusFilterResults struct {

	// StatusID is the ID of the
	// status this is a result for.
	StatusID string

	// RequesterID is the ID of the requesting
	// account for this status filter lookup.
	RequesterID string

	// Results is a map (int-key-array) of status filter
	// result slices in all possible filtering contexts.
	Results [keysLen][]StatusFilterResult
}

// StatusFilterResult stores a single (positive,
// i.e. match) filter result for a status by a filter.
type StatusFilterResult struct {

	// Expiry stores the time at which
	// (if any) the filter result expires.
	Expiry time.Time

	// Result stores any generated filter result for
	// this match intended to be shown at the frontend.
	// This can be used to determine the filter action:
	// - value => gtsmodel.FilterActionWarn
	// - nil => gtsmodel.FilterActionHide
	Result *apimodel.FilterResult
}

// Expired returns whether the filter result has expired.
func (r *StatusFilterResult) Expired(now time.Time) bool {
	return !r.Expiry.IsZero() && !r.Expiry.After(now)
}
