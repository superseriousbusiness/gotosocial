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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
)

// repeatBoostDepth determines the minimum count
// of statuses after which repeat boosts, or boosts
// of the original, may appear. This is may not end
// up *exact*, as small races between insert and the
// repeatBoost calculation may allow 1 or so extra
// to sneak in ahead of time. but it mostly works!
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

// StatusTimeline provides a concurrency-safe sliding-window
// cache of the freshest statuses in a timeline. Internally,
// only StatusMeta{} objects themselves are stored, loading
// the actual statuses when necessary, but caching prepared
// frontend API models where possible.
//
// Notes on design:
//
// Previously, and initially when designing this newer type,
// we had status timeline caches that would dynamically fill
// themselves with statuses on call to Load() with statuses
// at *any* location in the timeline, while simultaneously
// accepting new input of statuses from the background workers.
// This unfortunately can lead to situations where posts need
// to be fetched from the database, but the cache isn't aware
// they exist and instead returns an incomplete selection.
// This problem is best outlined by the follow simple example:
//
// "what if my timeline cache contains posts 0-to-6 and 8-to-12,
// and i make a request for posts between 4-and-10 with no limit,
// how is it to know that it's missing post 7?"
//
// The solution is to unfortunately remove a lot of the caching
// of "older areas" of the timeline, and instead just have it
// be a sliding window of the freshest posts of that timeline.
// It gets preloaded initially on start / first-call, and kept
// up-to-date with new posts by streamed inserts from background
// workers. Any requests for posts outside this we know therefore
// must hit the database, (which we then *don't* cache).
type StatusTimeline struct {

	// underlying timeline cache of *StatusMeta{},
	// primary-keyed by ID, with extra indices below.
	cache structr.Timeline[*StatusMeta, string]

	// preloader synchronizes preload
	// state of the timeline cache.
	preloader preloader

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

// Preload will fill with StatusTimeline{} cache with
// the latest sliding window of status metadata for the
// timeline type returned by database 'loadPage' function.
//
// This function is concurrency-safe and repeated calls to
// it when already preloaded will be no-ops. To trigger a
// preload as being required, call .Clear().
func (t *StatusTimeline) Preload(

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),

	// filter can be used to perform filtering of returned
	// statuses BEFORE insert into cache. i.e. this will effect
	// what actually gets stored in the timeline cache.
	filter func(each *gtsmodel.Status) (delete bool),
) (
	n int,
	err error,
) {
	err = t.preloader.CheckPreload(func() error {
		n, err = t.preload(loadPage, filter)
		return err
	})
	return
}

// preload contains the core logic of
// Preload(), without t.preloader checks.
func (t *StatusTimeline) preload(

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),

	// filter can be used to perform filtering of returned
	// statuses BEFORE insert into cache. i.e. this will effect
	// what actually gets stored in the timeline cache.
	filter func(each *gtsmodel.Status) (delete bool),
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
	page.Max.Value = plus1hULID()
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
		statuses = doStatusFilter(statuses, filter)

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
			last, ok := recentBoosts[id]
			repeat := ok && (idx-last) < 40
			value.repeatBoost = repeat

			// Update last-seen idx.
			recentBoosts[id] = idx
		}
	}

	return n, nil
}

// Load will load given page of timeline statuses. First it
// will prioritize fetching statuses from the sliding window
// that is the timeline cache of latest statuses, else it will
// fall back to loading from the database using callback funcs.
// The returned string values are the low / high status ID
// paging values, used in calculating next / prev page links.
func (t *StatusTimeline) Load(
	ctx context.Context,
	page *paging.Page,

	// loadPage should load the timeline of given page for cache hydration.
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),

	// loadIDs should load status models with given IDs, this is used
	// to load status models of already cached entries in the timeline.
	loadIDs func(ids []string) (statuses []*gtsmodel.Status, err error),

	// filter performs filtering of returned statuses.
	filter func(each *gtsmodel.Status) (delete bool),

	// prepareAPI should prepare internal status model to frontend API model.
	prepareAPI func(status *gtsmodel.Status) (apiStatus *apimodel.Status, err error),
) (
	[]*apimodel.Status,
	string, // lo
	string, // hi
	error,
) {
	var err error

	// Get paging details.
	lo := page.Min.Value
	hi := page.Max.Value
	limit := page.Limit
	order := page.Order()
	dir := toDirection(order)
	if limit <= 0 {

		// a page limit MUST be set!
		// this shouldn't be possible
		// but we check anyway to stop
		// chance of limitless db calls!
		panic("invalid page limit")
	}

	// Use a copy of current page so
	// we can repeatedly update it.
	nextPg := new(paging.Page)
	*nextPg = *page
	nextPg.Min.Value = lo
	nextPg.Max.Value = hi

	// Preallocate slice of interstitial models.
	metas := make([]*StatusMeta, 0, limit)

	// Preallocate slice of required status API models.
	apiStatuses := make([]*apimodel.Status, 0, limit)

	// TODO: we can remove this nil
	// check when we've updated all
	// our timeline endpoints to have
	// streamed timeline caches.
	if t != nil {

		// Ensure timeline has been preloaded.
		_, err = t.Preload(loadPage, filter)
		if err != nil {
			return nil, "", "", err
		}

		// Load a little more than limit to
		// reduce chance of db calls below.
		limitPtr := util.Ptr(limit + 10)

		// First we attempt to load status
		// metadata entries from the timeline
		// cache, up to given limit.
		metas = t.cache.Select(
			util.PtrIf(lo),
			util.PtrIf(hi),
			limitPtr,
			dir,
		)

		if len(metas) > 0 {
			// Before we can do any filtering, we need
			// to load status models for cached entries.
			err = loadStatuses(metas, loadIDs)
			if err != nil {
				return nil, "", "", gtserror.Newf("error loading statuses: %w", err)
			}

			// Set returned lo, hi values.
			lo = metas[len(metas)-1].ID
			hi = metas[0].ID

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
	}

	// If not enough cached timeline
	// statuses were found for page,
	// we need to call to database.
	if len(apiStatuses) < limit {

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

// loadStatusTimeline encapsulates the logic of iteratively
// attempting to load a status timeline page from the database,
// that is in the form of given callback functions. these will
// then be prepared to frontend API models for return.
//
// in time it may make sense to move this logic
// into the StatusTimeline{}.Load() function.
func loadStatusTimeline(
	ctx context.Context,
	nextPg *paging.Page,
	metas []*StatusMeta,
	apiStatuses []*apimodel.Status,
	loadPage func(page *paging.Page) (statuses []*gtsmodel.Status, err error),
	filter func(each *gtsmodel.Status) (delete bool),
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

	// Extract paging params, in particular
	// limit is used separate to nextPg to
	// determine the *expected* return limit,
	// not just what we use in db queries.
	returnLimit := nextPg.Limit
	order := nextPg.Order()

	// Perform maximum of 5 load
	// attempts fetching statuses.
	for i := 0; i < 5; i++ {

		// Update page limit to the *remaining*
		// limit of total we're expected to return.
		nextPg.Limit = returnLimit - len(apiStatuses)
		if nextPg.Limit <= 0 {

			// We reached the end! Set lo paging value.
			lo = apiStatuses[len(apiStatuses)-1].ID
			break
		}

		// But load a bit more than
		// limit to reduce db calls.
		nextPg.Limit += 10

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
		statuses = doStatusFilter(statuses, filter)

		// After filtering no more
		// statuses remain, retry.
		if len(statuses) == 0 {
			continue
		}

		// Convert to our interstitial meta type.
		metas = toStatusMeta(metas[:0], statuses)

		// Prepare frontend API models for
		// the loaded statuses. For now this
		// also does its own extra filtering.
		apiStatuses = prepareStatuses(ctx,
			metas,
			prepareAPI,
			apiStatuses,
			returnLimit,
		)
	}

	return apiStatuses, lo, hi, nil
}

// InsertOne allows you to insert a single status into the timeline, with optional prepared API model.
// The return value indicates whether status should be skipped from streams, e.g. if already boosted recently.
func (t *StatusTimeline) InsertOne(status *gtsmodel.Status, prepared *apimodel.Status) (skip bool) {

	// If timeline no preloaded, i.e.
	// no-one using it, don't insert.
	if !t.preloader.Check() {
		return false
	}

	if status.BoostOfID != "" {
		// Check through top $repeatBoostDepth number of items.
		for i, value := range t.cache.RangeUnsafe(structr.Desc) {
			if i >= repeatBoostDepth {
				break
			}

			// We don't care about values that have
			// already been hidden as repeat boosts.
			if value.repeatBoost {
				continue
			}

			// If inserted status has already been boosted, or original was posted
			// within last $repeatBoostDepth, we indicate it as a repeated boost.
			if value.ID == status.BoostOfID || value.BoostOfID == status.BoostOfID {
				skip = true
				break
			}
		}
	}

	// Insert new timeline status.
	t.cache.Insert(&StatusMeta{
		ID:               status.ID,
		AccountID:        status.AccountID,
		BoostOfID:        status.BoostOfID,
		BoostOfAccountID: status.BoostOfAccountID,
		repeatBoost:      skip,
		loaded:           nil,
		prepared:         prepared,
	})

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
func (t *StatusTimeline) Clear() { t.preloader.Clear() }

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
func doStatusFilter(statuses []*gtsmodel.Status, filter func(*gtsmodel.Status) bool) []*gtsmodel.Status {

	// Check for provided
	// filter function.
	if filter == nil {
		return statuses
	}

	// Filter the provided input statuses.
	return slices.DeleteFunc(statuses, filter)
}
