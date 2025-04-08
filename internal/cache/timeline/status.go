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
	"sync"
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

const repeatBoostDepth = 40

// StatusMeta contains minimum viable metadata
// about a Status in order to cache a timeline.
type StatusMeta struct {
	ID               string
	AccountID        string
	BoostOfID        string
	BoostOfAccountID string

	// is an internal flag that may be set on
	// a StatusMeta object that will prevent
	// preparation of its apimodel.Status, due
	// to it being a recently repeated boost.
	repeatBoost bool

	// prepared contains prepared frontend API
	// model for the referenced status. This may
	// or may-not be nil depending on whether the
	// status has been "unprepared" since the last
	// call to "prepare" the frontend model.
	prepared *apimodel.Status

	// loaded is a temporary field that may be
	// set for a newly loaded timeline status
	// so that statuses don't need to be loaded
	// from the database twice in succession.
	//
	// i.e. this will only be set if the status
	// was newly inserted into the timeline cache.
	// for existing cache items this will be nil.
	loaded *gtsmodel.Status
}

// StatusTimeline provides a concurrency-safe timeline
// cache of status information. Internally only StatusMeta{}
// objects are stored, and the statuses themselves are loaded
// as-needed, caching prepared frontend representations where
// possible. This is largely wrapping code for our own codebase
// to be able to smoothly interact with structr.Timeline{}.

// ...
type StatusTimeline struct {

	// underlying timeline cache of *StatusMeta{},
	// primary-keyed by ID, with extra indices below.
	cache structr.Timeline[*StatusMeta, string]

	// ...
	preload atomic.Pointer[any]

	// fast-access cache indices.
	idx_ID               *structr.Index //nolint:revive
	idx_AccountID        *structr.Index //nolint:revive
	idx_BoostOfID        *structr.Index //nolint:revive
	idx_BoostOfAccountID *structr.Index //nolint:revive

	// cutoff and maximum item lengths.
	// the timeline is trimmed back to
	// cutoff on each call to Trim(),
	// and maximum len triggers a Trim().
	//
	// the timeline itself does not
	// limit items due to complexities
	// it would introduce, so we apply
	// a 'cut-off' at regular intervals.
	cut, max int
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
			{Fields: "BoostOfID", Multiple: true},
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
				repeatBoost:      s.repeatBoost,
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

	// Set maximum capacity and
	// cutoff threshold we trim to.
	t.cut = int(0.60 * float64(cap))
	t.max = cap
}

func (t *StatusTimeline) startPreload(
	ctx context.Context,
	old *any, // old 'preload' ptr
	loadPage func(page *paging.Page) ([]*gtsmodel.Status, error),
	filter func(*gtsmodel.Status) (bool, error),
) (
	started bool,
	err error,
) {
	// Optimistically setup a
	// new waitgroup to set as
	// the preload waiter.
	var wg sync.WaitGroup
	wg.Add(1)
	defer wg.Done()

	// Wrap waitgroup in
	// 'any' for pointer.
	new := any(&wg)

	// Attempt CAS operation to claim preload start.
	started = t.preload.CompareAndSwap(old, &new)
	if !started {
		return
	}

	// Begin the preload.
	_, err = t.Preload(ctx,
		loadPage,
		filter,
	)
	return
}

func (t *StatusTimeline) checkPreload(
	ctx context.Context,
	loadPage func(page *paging.Page) ([]*gtsmodel.Status, error),
	filter func(*gtsmodel.Status) (bool, error),
) error {
	for {
		// Get preload state.
		p := t.preload.Load()

		if p == nil || *p == false {
			// Timeline needs preloading, start this process.
			ok, err := t.startPreload(ctx, p, loadPage, filter)

			if !ok {
				// Failed to acquire start,
				// other thread beat us to it.
				continue
			}

			// Return
			// result.
			return err
		}

		// Check for a preload currently in progress.
		if wg, _ := (*p).(*sync.WaitGroup); wg != nil {
			wg.Wait()
			continue
		}

		// Anything else means
		// timeline is ready.
		return nil
	}
}

// Preload ...
func (t *StatusTimeline) Preload(
	ctx context.Context,

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),

	// filter can be used to perform filtering of returned
	// statuses BEFORE insert into cache. i.e. this will effect
	// what actually gets stored in the timeline cache.
	filter func(each *gtsmodel.Status) (delete bool, err error),
) (int, error) {
	if loadPage == nil {
		panic("nil load page func")
	}

	// Clear timeline
	// before preload.
	t.cache.Clear()

	// Our starting, page at the top
	// of the possible timeline.
	page := new(paging.Page)
	order := paging.OrderDescending
	page.Max.Order = order
	page.Max.Value = plus24hULID()
	page.Min.Order = order
	page.Min.Value = ""
	page.Limit = 100

	// Prepare a slice for gathering status meta.
	metas := make([]*StatusMeta, 0, page.Limit)

	var n int
	for n < t.cut {
		// Load page of timeline statuses.
		statuses, err := loadPage(page)
		if err != nil {
			return n, gtserror.Newf("error loading statuses: %w", err)
		}

		// No more statuses from
		// load function = at end.
		if len(statuses) == 0 {
			break
		}

		// Update our next page cursor from statuses.
		page.Max.Value = statuses[len(statuses)-1].ID

		// Perform any filtering on newly loaded statuses.
		statuses, err = doStatusFilter(statuses, filter)
		if err != nil {
			return n, gtserror.Newf("error filtering statuses: %w", err)
		}

		// After filtering no more
		// statuses remain, retry.
		if len(statuses) == 0 {
			continue
		}

		// Convert statuses to meta and insert.
		metas = toStatusMeta(metas[:0], statuses)
		n = t.cache.Insert(metas...)
	}

	// This is a potentially 100-1000s size map,
	// but still easily manageable memory-wise.
	recentBoosts := make(map[string]int, t.cut)

	// Iterate timeline ascending (i.e. oldest -> newest), marking
	// entry IDs and marking down if boosts have been seen recently.
	for idx, value := range t.cache.RangeUnsafe(structr.Asc) {

		// Store current ID in map.
		recentBoosts[value.ID] = idx

		// If it's a boost, check if the original,
		// or a boost of it has been seen recently.
		if id := value.BoostOfID; id != "" {

			// Check if seen recently.
			last := recentBoosts[id]
			value.repeatBoost = (last < 40)

			// Update last-seen idx.
			recentBoosts[id] = idx
		}
	}

	// Mark timeline as preloaded.
	old := t.preload.Swap(new(any))
	if old != nil {
		switch t := (*old).(type) {
		case *sync.WaitGroup:
		default:
			log.Errorf(ctx, "BUG: invalid timeline preload state: %#v", t)
		}
	}

	return n, nil
}

// Load will load timeline statuses according to given
// page, using provided callbacks to load extra data when
// necessary, and perform fine-grained filtering loaded
// database models before eventual return to the user. The
// returned strings are the lo, hi ID paging values, used
// for generation of next, prev page links in the response.

// Load ...
func (t *StatusTimeline) Load(
	ctx context.Context,
	page *paging.Page,

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),

	// loadIDs should load status models with given IDs, this is used
	// to load status models of already cached entries in the timeline.
	loadIDs func(ids []string) (statuses []*gtsmodel.Status, err error),

	// filter can be used to perform filtering of returned statuses.
	filter func(each *gtsmodel.Status) (delete bool, err error),

	// prepareAPI should prepare internal status model to frontend API model.
	prepareAPI func(status *gtsmodel.Status) (apiStatus *apimodel.Status, err error),
) (
	[]*apimodel.Status,
	string, // lo
	string, // hi
	error,
) {
	// Ensure timeline is loaded.
	if err := t.checkPreload(ctx,
		loadPage,
		filter,
	); err != nil {
		return nil, "", "", err
	}

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

	var apiStatuses []*apimodel.Status

	if len(metas) > 0 {
		// Before we can do any filtering, we need
		// to load status models for cached entries.
		err := loadStatuses(metas, loadIDs)
		if err != nil {
			return nil, "", "", gtserror.Newf("error loading statuses: %w", err)
		}

		// Set initial lo, hi values.
		lo = metas[len(metas)-1].ID
		hi = metas[0].ID

		// Update paging parameters used for next database query.
		nextPageParams(nextPg, metas[len(metas)-1].ID, order)

		// Allocate slice of expected required API models.
		apiStatuses = make([]*apimodel.Status, 0, len(metas))

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

	// If no cached timeline statuses
	// were found for page, we need to
	// call through to the database.
	if len(apiStatuses) == 0 {
		var err error

		// Pass through to main timeline db load function.
		apiStatuses, lo, hi, err = loadStatusTimeline(ctx,
			nextPg,
			metas,
			apiStatuses,
			loadPage,
			filter,
			prepareAPI,
		)
		if err != nil {
			return nil, "", "", err
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

	return apiStatuses, lo, hi, nil
}

// LoadStatusTimeline is a function that may be used to load a timeline
// page in a functionally similar way to StatusTimeline{}.Load(), but without
// actually having access to a StatusTimeline{}. For example, for timelines that
// we want to share code, but without yet implementing a cache for them. Note this
// function may be removed in the future when un-needed.
func LoadStatusTimeline(
	ctx context.Context,
	page *paging.Page,
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),
	filter func(each *gtsmodel.Status) (delete bool, err error),
	prepareAPI func(status *gtsmodel.Status) (apiStatus *apimodel.Status, err error),
) (
	[]*apimodel.Status,
	string, // lo
	string, // hi
	error,
) {
	// Use a copy of current page so
	// we can repeatedly update it.
	nextPg := new(paging.Page)
	*nextPg = *page

	// Pass through to main timeline db load function.
	apiStatuses, lo, hi, err := loadStatusTimeline(ctx,
		nextPg,
		nil,
		nil,
		loadPage,
		filter,
		prepareAPI,
	)
	if err != nil {
		return nil, "", "", err
	}

	if page.Order().Ascending() {
		// The caller always expects the statuses
		// to be returned in DESC order, but we
		// build the status slice in paging order.
		// If paging ASC, we need to reverse the
		// returned statuses and paging values.
		slices.Reverse(apiStatuses)
		lo, hi = hi, lo
	}

	return apiStatuses, lo, hi, nil
}

// loadStatusTimeline encapsulates most of the main
// timeline-load-from-database logic, allowing both
// the temporary LoadStatusTimeline() function AND
// the main StatusTimeline{}.Load() function to share
// as much logic as possible.
//
// TODO: it may be worth moving this into StatusTimeline{}.Load()
// once the temporary function above has been removed. Or it may
// still be worth keeping *some* database logic separate.
func loadStatusTimeline(
	ctx context.Context,
	nextPg *paging.Page,
	metas []*StatusMeta,
	apiStatuses []*apimodel.Status,
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),
	filter func(each *gtsmodel.Status) (delete bool, err error),
	prepareAPI func(status *gtsmodel.Status) (apiStatus *apimodel.Status, err error),
) (
	[]*apimodel.Status,
	string, // lo
	string, // hi
	error,
) {
	if loadPage == nil {
		panic("nil load page func")
	}

	// Lowest and highest ID
	// vals of loaded statuses.
	var lo, hi string

	// Extract paging params.
	order := nextPg.Order()
	limit := nextPg.Limit

	// Load a little more than
	// limit to reduce db calls.
	nextPg.Limit += 10

	// Ensure we have a slice of meta objects to
	// use in later preparation of the API models.
	metas = xslices.GrowJust(metas[:0], nextPg.Limit)

	// Ensure we have a slice of required frontend API models.
	apiStatuses = xslices.GrowJust(apiStatuses[:0], nextPg.Limit)

	// Perform maximum of 5 load
	// attempts fetching statuses.
	for i := 0; i < 5; i++ {

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

		if hi == "" {
			// Set hi returned paging
			// value if not already set.
			hi = statuses[0].ID
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
		metas = toStatusMeta(metas[:0], statuses)

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

			// Set returned lo status paging value.
			lo = apiStatuses[len(apiStatuses)-1].ID
			break
		}
	}

	return apiStatuses, lo, hi, nil
}

// InsertOne allows you to insert a single status into the timeline, with optional prepared API model.
// The return value indicates whether status should be skipped from streams, e.g. if already boosted recently.
func (t *StatusTimeline) InsertOne(status *gtsmodel.Status, prepared *apimodel.Status) (skip bool) {
	if status.BoostOfID != "" {
		// Check through top $repeatBoostDepth number of timeline items.
		for i, value := range t.cache.RangeUnsafe(structr.Desc) {
			if i >= repeatBoostDepth {
				break
			}

			// If inserted status has already been boosted, or original was posted
			// within last $repeatBoostDepth, we indicate it as a repeated boost.
			if value.ID == status.BoostOfID || value.BoostOfID == status.BoostOfID {
				skip = true
				break
			}
		}
	}

	// Insert new status into timeline.
	if t.cache.Insert(&StatusMeta{
		ID:               status.ID,
		AccountID:        status.AccountID,
		BoostOfID:        status.BoostOfID,
		BoostOfAccountID: status.BoostOfAccountID,
		repeatBoost:      skip,
		loaded:           nil,
		prepared:         prepared,
	}) > t.max {

		// If cache reached beyond
		// maximum, perform a trim.
		t.Trim()
	}

	return
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
	for meta := range t.cache.RangeKeysUnsafe(t.idx_ID, keys...) {
		meta.prepared = nil
	}

	// Convert statusIDs to index keys.
	for i, id := range statusIDs {
		keys[i] = t.idx_BoostOfID.Key(id)
	}

	// Unprepare all statuses stored under StatusMeta.BoostOfID.
	for meta := range t.cache.RangeKeysUnsafe(t.idx_BoostOfID, keys...) {
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
	for meta := range t.cache.RangeKeysUnsafe(t.idx_AccountID, keys...) {
		meta.prepared = nil
	}

	// Convert accountIDs to index keys.
	for i, id := range accountIDs {
		keys[i] = t.idx_BoostOfAccountID.Key(id)
	}

	// Unprepare all statuses stored under StatusMeta.BoostOfAccountID.
	for meta := range t.cache.RangeKeysUnsafe(t.idx_BoostOfAccountID, keys...) {
		meta.prepared = nil
	}
}

// UnprepareAll removes cached frontend API
// models for all cached timeline entries.
func (t *StatusTimeline) UnprepareAll() {
	for _, value := range t.cache.RangeUnsafe(structr.Asc) {
		value.prepared = nil
	}
}

// Trim will ensure that receiving timeline is less than or
// equal in length to the given threshold percentage of the
// timeline's preconfigured maximum capacity. This will always
// trim from the bottom-up to prioritize streamed inserts.
func (t *StatusTimeline) Trim() { t.cache.Trim(t.cut, structr.Asc) }

// Clear will mark the entire timeline as requiring preload,
// which will trigger a clear and reload of the entire thing.
func (t *StatusTimeline) Clear() {
	t.preload.Store(func() *any {
		var b bool
		a := any(b)
		return &a
	}())
}

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

		if meta.repeatBoost {
			// This is a repeat boost in
			// short timespan, skip it.
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

		// Append to return slice.
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
func toStatusMeta(in []*StatusMeta, statuses []*gtsmodel.Status) []*StatusMeta {
	return xslices.Gather(in, statuses, func(s *gtsmodel.Status) *StatusMeta {
		return &StatusMeta{
			ID:               s.ID,
			AccountID:        s.AccountID,
			BoostOfID:        s.BoostOfID,
			BoostOfAccountID: s.BoostOfAccountID,
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
