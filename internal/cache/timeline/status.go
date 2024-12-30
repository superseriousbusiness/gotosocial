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

package timeline

import (
	"context"
	"slices"

	"codeberg.org/gruf/go-structr"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
)

// StatusMeta ...
type StatusMeta struct {

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

	// prepared contains prepared frontend API
	// model for the referenced status. This may
	// or may-not be nil depending on whether the
	// status has been "unprepared" since the last
	// call to "prepare" the frontend model.
	prepared *apimodel.Status

	// Loaded is a temporary field that may be
	// set for a newly loaded timeline status
	// so that statuses don't need to be loaded
	// from the database twice in succession.
	//
	// i.e. this will only be set if the status
	// was newly inserted into the timeline cache.
	// for existing cache items this will be nil.
	loaded *gtsmodel.Status
}

// StatusTimeline ...
type StatusTimeline struct {

	// underlying cache of *StatusMeta{}, primary-keyed by ID string.
	cache structr.Timeline[*StatusMeta, string]

	// fast-access cache indices.
	idx_ID               *structr.Index //nolint:revive
	idx_AccountID        *structr.Index //nolint:revive
	idx_BoostOfID        *structr.Index //nolint:revive
	idx_BoostOfAccountID *structr.Index //nolint:revive
}

// Init ...
func (t *StatusTimeline) Init(cap int) {
	t.cache.Init(structr.TimelineConfig[*StatusMeta, string]{
		PKey: "ID",

		Indices: []structr.IndexConfig{
			// ID as primary key is inherently an index.
			// {Fields: "ID"},
			{Fields: "AccountID", Multiple: true},
			{Fields: "BoostOfStatusID", Multiple: true},
			{Fields: "BoostOfAccountID", Multiple: true},
		},

		Copy: func(s *StatusMeta) *StatusMeta {
			var prepared *apimodel.Status
			if s.prepared != nil {
				prepared = new(apimodel.Status)
				*prepared = *s.prepared
			}
			return &StatusMeta{
				ID:               s.ID,
				AccountID:        s.AccountID,
				BoostOfID:        s.BoostOfID,
				BoostOfAccountID: s.BoostOfAccountID,
				loaded:           nil, // NEVER copied
				prepared:         prepared,
			}
		},
	})

	// Create a fast index lookup ptrs.
	t.idx_ID = t.cache.Index("ID")
	t.idx_AccountID = t.cache.Index("AccountID")
	t.idx_BoostOfID = t.cache.Index("BoostOfID")
	t.idx_BoostOfAccountID = t.cache.Index("BoostOfAccountID")
}

// Load ...
func (t *StatusTimeline) Load(
	ctx context.Context,
	page *paging.Page,

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) ([]*gtsmodel.Status, error),

	// loadIDs should load status models with given IDs.
	loadIDs func([]string) ([]*gtsmodel.Status, error),

	// preFilter can be used to perform filtering of returned
	// statuses BEFORE insert into cache. i.e. this will effect
	// what actually gets stored in the timeline cache.
	preFilter func(*gtsmodel.Status) (bool, error),

	// postFilterFn can be used to perform filtering of returned
	// statuses AFTER insert into cache. i.e. this will not effect
	// what actually gets stored in the timeline cache.
	postFilter func(*StatusMeta) bool,

	// prepareAPI should prepare internal status model to frontend API model.
	prepareAPI func(*gtsmodel.Status) (*apimodel.Status, error),
) (
	[]*apimodel.Status,
	error,
) {
	switch {
	case page == nil:
		panic("nil page")
	case loadPage == nil:
		panic("nil load page func")
	}

	// Get paging details.
	min := page.Min.Value
	max := page.Max.Value
	lim := page.Limit
	ord := page.Order()
	dir := toDirection(ord)

	// Load cached timeline entries for page.
	meta := t.cache.Select(min, max, lim, dir)

	// Perform any timeline post-filtering.
	meta = doPostFilter(meta, postFilter)

	// ...
	if need := len(meta) - lim; need > 0 {

		// Set first page
		// query to load.
		nextPg := page

		// Perform a maximum of 5
		// load attempts fetching
		// statuses to reach limit.
		for i := 0; i < 5; i++ {

			// Load next timeline statuses.
			statuses, err := loadPage(nextPg)
			if err != nil {
				return nil, gtserror.Newf("error loading timeline: %w", err)
			}

			// No more statuses from
			// load function = at end.
			if len(statuses) == 0 {
				break
			}

			// Get the lowest and highest
			// ID values, used for next pg.
			// Done BEFORE status filtering.
			lo := statuses[len(statuses)-1].ID
			hi := statuses[0].ID

			// Perform any status timeline pre-filtering.
			statuses, err = doPreFilter(statuses, preFilter)
			if err != nil {
				return nil, gtserror.Newf("error pre-filtering timeline: %w", err)
			}

			// Convert to our cache type,
			// these will get inserted into
			// the cache in prepare() below.
			m := toStatusMeta(statuses)

			// Perform any post-filtering.
			// and append to main meta slice.
			m = slices.DeleteFunc(m, postFilter)
			meta = append(meta, m...)

			// Check if we reached
			// requested page limit.
			if len(meta) >= lim {
				break
			}

			// Set next paging value.
			nextPg = nextPg.Next(lo, hi)
		}
	}

	// Using meta and given funcs, prepare frontend API models.
	apiStatuses, err := t.prepare(ctx, meta, loadIDs, prepareAPI)
	if err != nil {
		return nil, gtserror.Newf("error preparing api statuses: %w", err)
	}

	// Ensure the returned statuses are ALWAYS in descending order.
	slices.SortFunc(apiStatuses, func(s1, s2 *apimodel.Status) int {
		const k = +1
		switch {
		case s1.ID > s2.ID:
			return +k
		case s1.ID < s2.ID:
			return -k
		default:
			return 0
		}
	})

	return apiStatuses, nil
}

// RemoveByStatusID removes all cached timeline entries pertaining to
// status ID, including those that may be a boost of the given status.
func (t *StatusTimeline) RemoveByStatusIDs(statusIDs ...string) {
	keys := make([]structr.Key, len(statusIDs))

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_ID.Key(id)
	}

	// Invalidate all cached entries with IDs.
	t.cache.Invalidate(t.idx_ID, keys...)

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_BoostOfID.Key(id)
	}

	// Invalidate all cached entries as boost of IDs.
	t.cache.Invalidate(t.idx_BoostOfID, keys...)
}

// RemoveByAccountID removes all cached timeline entries authored by
// account ID, including those that may be boosted by account ID.
func (t *StatusTimeline) RemoveByAccountIDs(accountIDs ...string) {
	keys := make([]structr.Key, len(accountIDs))

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_AccountID.Key(id)
	}

	// Invalidate all cached entries as by IDs.
	t.cache.Invalidate(t.idx_AccountID, keys...)

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_BoostOfAccountID.Key(id)
	}

	// Invalidate all cached entries as boosted by IDs.
	t.cache.Invalidate(t.idx_BoostOfAccountID, keys...)
}

// UnprepareByStatusIDs removes cached frontend API models for all cached
// timeline entries pertaining to status ID, including boosts of given status.
func (t *StatusTimeline) UnprepareByStatusIDs(statusIDs ...string) {
	keys := make([]structr.Key, len(statusIDs))

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_ID.Key(id)
	}

	// TODO: replace below with for-range-function loop when Go1.23.
	t.cache.RangeKeys(t.idx_ID, keys...)(func(meta *StatusMeta) bool {
		meta.prepared = nil
		return true
	})

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_BoostOfID.Key(id)
	}

	// TODO: replace below with for-range-function loop when Go1.23.
	t.cache.RangeKeys(t.idx_BoostOfID, keys...)(func(meta *StatusMeta) bool {
		meta.prepared = nil
		return true
	})
}

// UnprepareByAccountIDs removes cached frontend API models for all cached
// timeline entries authored by account ID, including boosts by account ID.
func (t *StatusTimeline) UnprepareByAccountIDs(accountIDs ...string) {
	keys := make([]structr.Key, len(accountIDs))

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_AccountID.Key(id)
	}

	// TODO: replace below with for-range-function loop when Go1.23.
	t.cache.RangeKeys(t.idx_AccountID, keys...)(func(meta *StatusMeta) bool {
		meta.prepared = nil
		return true
	})

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_BoostOfAccountID.Key(id)
	}

	// TODO: replace below with for-range-function loop when Go1.23.
	t.cache.RangeKeys(t.idx_BoostOfAccountID, keys...)(func(meta *StatusMeta) bool {
		meta.prepared = nil
		return true
	})
}

// Clear will remove all cached entries from timeline.
func (t *StatusTimeline) Clear() { t.cache.Clear() }

// prepare will take a slice of cached (or, freshly loaded!) StatusMeta{}
// models, and use given functions to return prepared frontend API models.
func (t *StatusTimeline) prepare(
	ctx context.Context,
	meta []*StatusMeta,
	loadIDs func([]string) ([]*gtsmodel.Status, error),
	prepareAPI func(*gtsmodel.Status) (*apimodel.Status, error),
) (
	[]*apimodel.Status,
	error,
) {
	switch {
	case loadIDs == nil:
		panic("nil load fn")
	case prepareAPI == nil:
		panic("nil prepare fn")
	}

	// Iterate the given StatusMeta objects for pre-prepared frontend
	// models, otherwise storing as unprepared for further processing.
	apiStatuses := make([]*apimodel.Status, len(meta))
	unprepared := make([]*StatusMeta, 0, len(meta))
	for i, meta := range meta {
		apiStatuses[i] = meta.prepared
		if meta.prepared == nil {
			unprepared = append(unprepared, meta)
		}
	}

	// If there were no unprepared
	// StatusMeta objects, then we
	// gathered everything we need!
	if len(unprepared) == 0 {
		return apiStatuses, nil
	}

	// Of the StatusMeta objects missing a prepared
	// frontend model, find those without a recently
	// fetched database model and store their IDs,
	// as well mapping them for faster update below.
	toLoadIDs := make([]string, len(unprepared))
	loadedMap := make(map[string]*StatusMeta, len(unprepared))
	for i, meta := range unprepared {
		if meta.loaded == nil {
			toLoadIDs[i] = meta.ID
			loadedMap[meta.ID] = meta
		}
	}

	// Load statuses with given IDs.
	loaded, err := loadIDs(toLoadIDs)
	if err != nil {
		return nil, gtserror.Newf("error loading statuses: %w", err)
	}

	// Update returned StatusMeta objects
	// with newly loaded statuses by IDs.
	for i := range loaded {
		status := loaded[i]
		meta := loadedMap[status.ID]
		meta.loaded = status
	}

	for i := 0; i < len(unprepared); {
		// Get meta at index.
		meta := unprepared[i]

		if meta.loaded == nil {
			// We failed loading this
			// status, skip preparing.
			continue
		}

		// Prepare the provided status to frontend.
		apiStatus, err := prepareAPI(meta.loaded)
		if err != nil {
			log.Errorf(ctx, "error preparing status %s: %v", meta.loaded.URI, err)
			continue
		}

		if apiStatus != nil {
			// TODO: we won't need nil check when mutes
			// / filters are moved to appropriate funcs.
			apiStatuses = append(apiStatuses, apiStatus)
		}
	}

	// Re-insert all (previously) unprepared
	// status meta types into timeline cache.
	t.cache.Insert(unprepared...)

	return apiStatuses, nil
}

// toStatusMeta converts a slice of database model statuses
// into our cache wrapper type, a slice of []StatusMeta{}.
func toStatusMeta(statuses []*gtsmodel.Status) []*StatusMeta {
	meta := make([]*StatusMeta, len(statuses))
	for i := range statuses {
		status := statuses[i]
		meta[i] = &StatusMeta{
			ID:               status.ID,
			AccountID:        status.AccountID,
			BoostOfID:        status.BoostOfID,
			BoostOfAccountID: status.BoostOfAccountID,
			Local:            *status.Local,
			loaded:           status,
			prepared:         nil,
		}
	}
	return meta
}

// doPreFilter acts similarly to slices.DeleteFunc but it accepts function with error return, or nil, returning early if so.
func doPreFilter(statuses []*gtsmodel.Status, preFilter func(*gtsmodel.Status) (bool, error)) ([]*gtsmodel.Status, error) {
	if preFilter == nil {
		return statuses, nil
	}

	// Iterate through input statuses.
	for i := 0; i < len(statuses); {
		status := statuses[i]

		// Pass through filter func.
		ok, err := preFilter(status)
		if err != nil {
			return nil, err
		}

		if ok {
			// Delete this status from input slice.
			statuses = slices.Delete(statuses, i, i+1)
			continue
		}

		// Iter.
		i++
	}

	return statuses, nil
}

// doPostFilter acts similarly to slices.DeleteFunc but it handles case of a nil function.
func doPostFilter(statuses []*StatusMeta, postFilter func(*StatusMeta) bool) []*StatusMeta {
	if postFilter == nil {
		return statuses
	}
	return slices.DeleteFunc(statuses, postFilter)
}

// toDirection converts page order to timeline direction.
func toDirection(o paging.Order) structr.Direction {
	switch o {
	case paging.OrderAscending:
		return structr.Asc
	case paging.OrderDescending:
		return structr.Desc
	default:
		return false
	}
}
