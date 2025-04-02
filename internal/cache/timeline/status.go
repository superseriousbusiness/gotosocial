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
	"maps"
	"slices"
	"sync/atomic"

	"codeberg.org/gruf/go-structr"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
)

// StatusMeta contains minimum viable metadata
// about a Status in order to cache a timeline.
type StatusMeta struct {
	ID               string
	AccountID        string
	BoostOfID        string
	BoostOfAccountID string
	Local            bool

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

// isNotLoaded is a small utility func that can fill
// the slices.DeleteFunc() signature requirements.
func (m *StatusMeta) isNotLoaded() bool {
	return m.loaded == nil
}

// StatusTimelines ...
type StatusTimelines struct {
	ptr atomic.Pointer[map[string]*StatusTimeline] // ronly except by CAS
	cap int
}

// Init ...
func (t *StatusTimelines) Init(cap int) { t.cap = cap }

// MustGet ...
func (t *StatusTimelines) MustGet(key string) *StatusTimeline {
	var tt *StatusTimeline

	for {
		// Load current ptr.
		cur := t.ptr.Load()

		// Get timeline map to work on.
		var m map[string]*StatusTimeline

		if cur != nil {
			// Look for existing
			// timeline in cache.
			tt = (*cur)[key]
			if tt != nil {
				return tt
			}

			// Get clone of current
			// before modifications.
			m = maps.Clone(*cur)
		} else {
			// Allocate new timeline map for below.
			m = make(map[string]*StatusTimeline)
		}

		if tt == nil {
			// Allocate new timeline.
			tt = new(StatusTimeline)
			tt.Init(t.cap)
		}

		// Store timeline
		// in new map.
		m[key] = tt

		// Attempt to update the map ptr.
		if !t.ptr.CompareAndSwap(cur, &m) {

			// We failed the
			// CAS, reloop.
			continue
		}

		// Successfully inserted
		// new timeline model.
		return tt
	}
}

// Delete ...
func (t *StatusTimelines) Delete(key string) {
	for {
		// Load current ptr.
		cur := t.ptr.Load()

		// Check for empty map / not in map.
		if cur == nil || (*cur)[key] == nil {
			return
		}

		// Get clone of current
		// before modifications.
		m := maps.Clone(*cur)

		// Delete ID.
		delete(m, key)

		// Attempt to update the map ptr.
		if !t.ptr.CompareAndSwap(cur, &m) {

			// We failed the
			// CAS, reloop.
			continue
		}

		// Successfully
		// deleted ID.
		return
	}
}

// RemoveByStatusIDs ...
func (t *StatusTimelines) RemoveByStatusIDs(statusIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.RemoveByStatusIDs(statusIDs...)
		}
	}
}

// RemoveByAccountIDs ...
func (t *StatusTimelines) RemoveByAccountIDs(accountIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.RemoveByAccountIDs(accountIDs...)
		}
	}
}

// UnprepareByStatusIDs ...
func (t *StatusTimelines) UnprepareByStatusIDs(statusIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.UnprepareByStatusIDs(statusIDs...)
		}
	}
}

// UnprepareByAccountIDs ...
func (t *StatusTimelines) UnprepareByAccountIDs(accountIDs ...string) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.UnprepareByAccountIDs(accountIDs...)
		}
	}
}

// Unprepare ...
func (t *StatusTimelines) Unprepare(key string) {
	if p := t.ptr.Load(); p != nil {
		if tt := (*p)[key]; tt != nil {
			tt.UnprepareAll()
		}
	}
}

// UnprepareAll ...
func (t *StatusTimelines) UnprepareAll() {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.UnprepareAll()
		}
	}
}

// Trim ...
func (t *StatusTimelines) Trim(threshold float64) {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.Trim(threshold)
		}
	}
}

// Clear ...
func (t *StatusTimelines) Clear(key string) {
	if p := t.ptr.Load(); p != nil {
		if tt := (*p)[key]; tt != nil {
			tt.Clear()
		}
	}
}

// ClearAll ...
func (t *StatusTimelines) ClearAll() {
	if p := t.ptr.Load(); p != nil {
		for _, tt := range *p {
			tt.Clear()
		}
	}
}

// StatusTimeline ...
type StatusTimeline struct {

	// underlying timeline cache of *StatusMeta{},
	// primary-keyed by ID, with extra indices below.
	cache structr.Timeline[*StatusMeta, string]

	// fast-access cache indices.
	idx_ID               *structr.Index //nolint:revive
	idx_AccountID        *structr.Index //nolint:revive
	idx_BoostOfID        *structr.Index //nolint:revive
	idx_BoostOfAccountID *structr.Index //nolint:revive

	// last stores the last fetched direction
	// of the timeline, which in turn determines
	// where we will next trim from in keeping the
	// timeline underneath configured 'max'.
	//
	// TODO: this could be more intelligent with
	// a sliding average. a problem for future kim!
	last atomic.Pointer[structr.Direction]

	// defines the 'maximum' count of
	// entries in the timeline that we
	// apply our Trim() call threshold
	// to. the timeline itself does not
	// limit items due to complexities
	// it would introduce, so we apply
	// a 'cut-off' at regular intervals.
	max int
}

// Init will initialize the timeline for usage,
// by preparing internal indices etc. This also
// sets the given max capacity for Trim() operations.
func (t *StatusTimeline) Init(cap int) {
	t.cache.Init(structr.TimelineConfig[*StatusMeta, string]{

		// Timeline item primary key field.
		PKey: structr.IndexConfig{Fields: "ID"},

		// Additional indexed fields.
		Indices: []structr.IndexConfig{
			{Fields: "AccountID", Multiple: true},
			{Fields: "BoostOfAccountID", Multiple: true},

			// By setting multiple=false for BoostOfID, this will prevent
			// timeline entries with matching BoostOfID will not be inserted
			// after the first, which allows us to prevent repeated boosts
			// of the same status from showing up within 'cap' entries.
			{Fields: "BoostOfID", Multiple: false},
		},

		// Timeline item copy function.
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
				loaded:           nil, // NEVER stored
				prepared:         prepared,
			}
		},
	})

	// Get fast index lookup ptrs.
	t.idx_ID = t.cache.Index("ID")
	t.idx_AccountID = t.cache.Index("AccountID")
	t.idx_BoostOfID = t.cache.Index("BoostOfID")
	t.idx_BoostOfAccountID = t.cache.Index("BoostOfAccountID")

	// Set max.
	t.max = cap
}

// Load will load timeline statuses according to given
// page, using provided callbacks to load extra data when
// necessary, and perform fine-grained filtering loaded
// database models before eventual return to the user. The
// returned strings are the lo, hi ID paging values, used
// for generation of next, prev page links in the response.
func (t *StatusTimeline) Load(
	ctx context.Context,
	page *paging.Page,

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),

	// loadIDs should load status models with given IDs, this is used
	// to load status models of already cached entries in the timeline.
	loadIDs func(ids []string) (statuses []*gtsmodel.Status, err error),

	// filter can be used to perform filtering of returned
	// statuses BEFORE insert into cache. i.e. this will effect
	// what actually gets stored in the timeline cache.
	filter func(each *gtsmodel.Status) (delete bool, err error),

	// prepareAPI should prepare internal status model to frontend API model.
	prepareAPI func(status *gtsmodel.Status) (apiStatus *apimodel.Status, err error),
) (
	[]*apimodel.Status,
	string, // lo
	string, // hi
	error,
) {
	switch {
	case page == nil:
		panic("nil page")
	case loadPage == nil:
		panic("nil load page func")
	}

	// TODO: there's quite a few opportunities for
	// optimization here, with a lot of frequently
	// used slices of the same types. depending on
	// profiles it may be advantageous to pool some.

	// Get paging details.
	lo := page.Min.Value
	hi := page.Max.Value
	limit := page.Limit
	order := page.Order()
	dir := toDirection(order)

	// Use a copy of current page so
	// we can repeatedly update it.
	nextPg := new(paging.Page)
	*nextPg = *page
	nextPg.Min.Value = lo
	nextPg.Max.Value = hi

	// First we attempt to load status
	// metadata entries from the timeline
	// cache, up to given limit.
	metas := t.cache.Select(
		util.PtrIf(lo),
		util.PtrIf(hi),
		util.PtrIf(limit),
		dir,
	)

	// We now reset the lo,hi values to
	// represent the lowest and highest
	// index values of loaded statuses.
	//
	// We continually update these while
	// building up statuses to return, for
	// caller to build next / prev page
	// response values.
	lo, hi = "", ""

	// Preallocate a slice of up-to-limit API models.
	apiStatuses := make([]*apimodel.Status, 0, limit)

	if len(metas) > 0 {
		// Before we can do any filtering, we need
		// to load status models for cached entries.
		err := loadStatuses(metas, loadIDs)
		if err != nil {
			return nil, "", "", gtserror.Newf("error loading statuses: %w", err)
		}

		// Set initial lo, hi values.
		hi = metas[len(metas)-1].ID
		lo = metas[0].ID

		// Update paging parameters used for next database query.
		nextPageParams(nextPg, metas[len(metas)-1].ID, order)

		// Prepare frontend API models for
		// the cached statuses. For now this
		// also does its own extra filtering.
		apiStatuses = prepareStatuses(ctx,
			metas,
			prepareAPI,
			apiStatuses,
			limit,
		)
	}

	// Track all newly loaded status entries
	// after filtering for insert into cache.
	var justLoaded []*StatusMeta

	// Check whether loaded enough from cache.
	if need := limit - len(apiStatuses); need > 0 {

		// Load a little more than
		// limit to reduce db calls.
		nextPg.Limit += 10

		// Perform maximum of 10 load
		// attempts fetching statuses.
		for i := 0; i < 10; i++ {

			// Load next timeline statuses.
			statuses, err := loadPage(nextPg)
			if err != nil {
				return nil, "", "", gtserror.Newf("error loading timeline: %w", err)
			}

			// No more statuses from
			// load function = at end.
			if len(statuses) == 0 {
				break
			}

			if lo == "" {
				// Set min returned paging
				// value if not already set.
				lo = statuses[0].ID
			}

			// Update nextPg cursor parameter for next database query.
			nextPageParams(nextPg, statuses[len(statuses)-1].ID, order)

			// Perform any filtering on newly loaded statuses.
			statuses, err = doStatusFilter(statuses, filter)
			if err != nil {
				return nil, "", "", gtserror.Newf("error filtering statuses: %w", err)
			}

			// After filtering no more
			// statuses remain, retry.
			if len(statuses) == 0 {
				continue
			}

			// Convert to our cache type,
			// these will get inserted into
			// the cache in prepare() below.
			metas := toStatusMeta(statuses)

			// Append to newly loaded for later insert.
			justLoaded = append(justLoaded, metas...)

			// Prepare frontend API models for
			// the loaded statuses. For now this
			// also does its own extra filtering.
			apiStatuses = prepareStatuses(ctx,
				metas,
				prepareAPI,
				apiStatuses,
				limit,
			)

			// If we have anything, return
			// here. Even if below limit.
			if len(apiStatuses) > 0 {

				// Set returned hi status paging value.
				hi = apiStatuses[len(apiStatuses)-1].ID
				break
			}
		}
	}

	if order.Ascending() {
		// The caller always expects the statuses
		// to be returned in DESC order, but we
		// build the status slice in paging order.
		// If paging ASC, we need to reverse the
		// returned statuses and paging values.
		slices.Reverse(apiStatuses)
		lo, hi = hi, lo
	}

	if len(justLoaded) > 0 {
		// Even if we don't return them, insert
		// the excess (post-filtered) into cache.
		t.cache.Insert(justLoaded...)
	}

	return apiStatuses, lo, hi, nil
}

// InsertOne allows you to insert a single status into the timeline, with optional prepared API model.
func (t *StatusTimeline) InsertOne(status *gtsmodel.Status, prepared *apimodel.Status) {
	t.cache.Insert(&StatusMeta{
		ID:               status.ID,
		AccountID:        status.AccountID,
		BoostOfID:        status.BoostOfID,
		BoostOfAccountID: status.BoostOfAccountID,
		Local:            *status.Local,
		loaded:           status,
		prepared:         prepared,
	})
}

// Insert allows you to bulk insert many statuses into the timeline.
func (t *StatusTimeline) Insert(statuses ...*gtsmodel.Status) {
	t.cache.Insert(toStatusMeta(statuses)...)
}

// RemoveByStatusID removes all cached timeline entries pertaining to
// status ID, including those that may be a boost of the given status.
func (t *StatusTimeline) RemoveByStatusIDs(statusIDs ...string) {
	keys := make([]structr.Key, len(statusIDs))

	// Nil check indices outside loops.
	if t.idx_ID == nil ||
		t.idx_BoostOfID == nil {
		panic("indices are nil")
	}

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

	// Nil check indices outside loops.
	if t.idx_AccountID == nil ||
		t.idx_BoostOfAccountID == nil {
		panic("indices are nil")
	}

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

	// Nil check indices outside loops.
	if t.idx_ID == nil ||
		t.idx_BoostOfID == nil {
		panic("indices are nil")
	}

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_ID.Key(id)
	}

	// Unprepare all statuses stored under StatusMeta.ID.
	for meta := range t.cache.RangeKeys(t.idx_ID, keys...) {
		meta.prepared = nil
	}

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_BoostOfID.Key(id)
	}

	// Unprepare all statuses stored under StatusMeta.BoostOfID.
	for meta := range t.cache.RangeKeys(t.idx_BoostOfID, keys...) {
		meta.prepared = nil
	}
}

// UnprepareByAccountIDs removes cached frontend API models for all cached
// timeline entries authored by account ID, including boosts by account ID.
func (t *StatusTimeline) UnprepareByAccountIDs(accountIDs ...string) {
	keys := make([]structr.Key, len(accountIDs))

	// Nil check indices outside loops.
	if t.idx_AccountID == nil ||
		t.idx_BoostOfAccountID == nil {
		panic("indices are nil")
	}

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_AccountID.Key(id)
	}

	// Unprepare all statuses stored under StatusMeta.AccountID.
	for meta := range t.cache.RangeKeys(t.idx_AccountID, keys...) {
		meta.prepared = nil
	}

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_BoostOfAccountID.Key(id)
	}

	// Unprepare all statuses stored under StatusMeta.BoostOfAccountID.
	for meta := range t.cache.RangeKeys(t.idx_BoostOfAccountID, keys...) {
		meta.prepared = nil
	}
}

// UnprepareAll removes cached frontend API
// models for all cached timeline entries.
func (t *StatusTimeline) UnprepareAll() {
	for value := range t.cache.RangeUnsafe(structr.Asc) {
		value.prepared = nil
	}
}

// Trim will ensure that receiving timeline is less than or
// equal in length to the given threshold percentage of the
// timeline's preconfigured maximum capacity. This will trim
// from top / bottom depending on which was recently accessed.
func (t *StatusTimeline) Trim(threshold float64) {

	// Default trim dir.
	dir := structr.Asc

	// Calculate maximum allowed no.
	// items as a percentage of max.
	max := threshold * float64(t.max)

	// Try load the last fetched
	// timeline ordering, getting
	// the inverse value for trimming.
	if p := t.last.Load(); p != nil {
		dir = !(*p)
	}

	// Trim timeline to 'max'.
	t.cache.Trim(int(max), dir)
}

// Clear will remove all cached entries from underlying timeline.
func (t *StatusTimeline) Clear() { t.cache.Trim(0, structr.Desc) }

// prepareStatuses takes a slice of cached (or, freshly loaded!) StatusMeta{}
// models, and use given function to return prepared frontend API models.
func prepareStatuses(
	ctx context.Context,
	meta []*StatusMeta,
	prepareAPI func(*gtsmodel.Status) (*apimodel.Status, error),
	apiStatuses []*apimodel.Status,
	limit int,
) []*apimodel.Status {
	switch { //nolint:gocritic
	case prepareAPI == nil:
		panic("nil prepare fn")
	}

	// Iterate the given StatusMeta objects for pre-prepared
	// frontend models, otherwise attempting to prepare them.
	for _, meta := range meta {

		// Check if we have prepared enough
		// API statuses for caller to return.
		if len(apiStatuses) >= limit {
			break
		}

		if meta.loaded == nil {
			// We failed loading this
			// status, skip preparing.
			continue
		}

		if meta.prepared == nil {
			var err error

			// Prepare the provided status to frontend.
			meta.prepared, err = prepareAPI(meta.loaded)
			if err != nil {
				log.Errorf(ctx, "error preparing status %s: %v", meta.loaded.URI, err)
				continue
			}
		}

		if meta.prepared != nil {
			apiStatuses = append(apiStatuses, meta.prepared)
		}
	}

	return apiStatuses
}

// loadStatuses loads statuses using provided callback
// for the statuses in meta slice that aren't loaded.
// the amount very much depends on whether meta objects
// are yet-to-be-cached (i.e. newly loaded, with status),
// or are from the timeline cache (unloaded status).
func loadStatuses(
	metas []*StatusMeta,
	loadIDs func([]string) ([]*gtsmodel.Status, error),
) error {

	// Determine which of our passed status
	// meta objects still need statuses loading.
	toLoadIDs := make([]string, len(metas))
	loadedMap := make(map[string]*StatusMeta, len(metas))
	for i, meta := range metas {
		if meta.loaded == nil {
			toLoadIDs[i] = meta.ID
			loadedMap[meta.ID] = meta
		}
	}

	// Load statuses with given IDs.
	loaded, err := loadIDs(toLoadIDs)
	if err != nil {
		return gtserror.Newf("error loading statuses: %w", err)
	}

	// Update returned StatusMeta objects
	// with newly loaded statuses by IDs.
	for i := range loaded {
		status := loaded[i]
		meta := loadedMap[status.ID]
		meta.loaded = status
	}

	return nil
}

// toStatusMeta converts a slice of database model statuses
// into our cache wrapper type, a slice of []StatusMeta{}.
func toStatusMeta(statuses []*gtsmodel.Status) []*StatusMeta {
	return xslices.Gather(nil, statuses, func(s *gtsmodel.Status) *StatusMeta {
		return &StatusMeta{
			ID:               s.ID,
			AccountID:        s.AccountID,
			BoostOfID:        s.BoostOfID,
			BoostOfAccountID: s.BoostOfAccountID,
			Local:            *s.Local,
			loaded:           s,
			prepared:         nil,
		}
	})
}

// doStatusFilter performs given filter function on provided statuses,
// returning early if an error is returned. returns filtered statuses.
func doStatusFilter(statuses []*gtsmodel.Status, filter func(*gtsmodel.Status) (bool, error)) ([]*gtsmodel.Status, error) {

	// Check for provided
	// filter function.
	if filter == nil {
		return statuses, nil
	}

	// Iterate through input statuses.
	for i := 0; i < len(statuses); {
		status := statuses[i]

		// Pass through filter func.
		ok, err := filter(status)
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
