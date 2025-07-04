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

package status

import (
	"context"
	"regexp"
	"slices"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/cache"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// StatusFilterResultsInContext returns status filtering results, limited
// to the given filtering context, about the given status for requester.
// The hide flag is immediately returned if any filters match with the
// HIDE action set, else API model filter results for the WARN action.
func (f *Filter) StatusFilterResultsInContext(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
	context gtsmodel.FilterContext,
) (
	results []apimodel.FilterResult,
	hidden bool,
	err error,
) {
	if context == gtsmodel.FilterContextNone {
		// fast-check any context.
		return nil, false, nil
	}

	// Get cached filter results for status to requester in all contexts.
	allResults, now, err := f.StatusFilterResults(ctx, requester, status)
	if err != nil {
		return nil, false, err
	}

	// Get results applicable to current context.
	var forContext []cache.StatusFilterResult
	switch context {
	case gtsmodel.FilterContextHome:
		forContext = allResults.Results[cache.KeyContextHome]
	case gtsmodel.FilterContextPublic:
		forContext = allResults.Results[cache.KeyContextPublic]
	case gtsmodel.FilterContextNotifications:
		forContext = allResults.Results[cache.KeyContextNotifs]
	case gtsmodel.FilterContextThread:
		forContext = allResults.Results[cache.KeyContextThread]
	case gtsmodel.FilterContextAccount:
		forContext = allResults.Results[cache.KeyContextAccount]
	}

	// Iterate results in context, gathering prepared API models.
	results = make([]apimodel.FilterResult, 0, len(forContext))
	for _, result := range forContext {

		// Check if result expired.
		if result.Expired(now) {
			continue
		}

		// If the result indicates
		// status should just be
		// hidden then return here.
		if result.Result == nil {
			return nil, true, nil
		}

		// Append pre-prepared API model to slice.
		results = append(results, *result.Result)
	}

	return
}

// StatusFilterResults returns status filtering results (in all contexts) about the given status for the given requesting account.
func (f *Filter) StatusFilterResults(ctx context.Context, requester *gtsmodel.Account, status *gtsmodel.Status) (*cache.CachedStatusFilterResults, time.Time, error) {

	// For requester ID use a
	// fallback 'noauth' string
	// by default for lookups.
	requesterID := noauth
	if requester != nil {
		requesterID = requester.ID
	}

	// Get current time.
	now := time.Now()

	// Load status filtering results for this requesting account about status from cache, using load callback function if necessary.
	results, err := f.state.Caches.StatusFilter.LoadOne("RequesterID,StatusID", func() (*cache.CachedStatusFilterResults, error) {

		// Load status filter results for given status.
		results, err := f.getStatusFilterResults(ctx,
			requester,
			status,
			now,
		)
		if err != nil {
			if err == cache.SentinelError {
				// Filter-out our temporary
				// race-condition error.
				return &cache.CachedStatusFilterResults{}, nil
			}

			return nil, err
		}

		// Convert to cacheable results type.
		return &cache.CachedStatusFilterResults{
			StatusID:    status.ID,
			RequesterID: requesterID,
			Results:     results,
		}, nil
	}, requesterID, status.ID)
	if err != nil {
		return nil, now, err
	}

	return results, now, err
}

// getStatusFilterResults loads status filtering results for
// the given status, given the current time (checking expiries).
// this will load results for all possible filtering contexts.
func (f *Filter) getStatusFilterResults(
	ctx context.Context,
	requester *gtsmodel.Account,
	status *gtsmodel.Status,
	now time.Time,
) (
	[5][]cache.StatusFilterResult,
	error,
) {
	var results [5][]cache.StatusFilterResult

	if requester == nil {
		// Without auth, there will be no possible
		// filters to exists, return as 'unfiltered'.
		return results, nil
	}

	// Check if status is boost.
	if status.BoostOfID != "" {
		if status.BoostOf == nil {
			var err error

			// Ensure original status is loaded on boost.
			status.BoostOf, err = f.state.DB.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				status.BoostOfID,
			)
			if err != nil {
				return results, gtserror.Newf("error getting boosted status of %s: %w", status.URI, err)
			}
		}

		// From here look at details
		// for original boosted status.
		status = status.BoostOf
	}

	// For proper status filtering we need all fields populated.
	if err := f.state.DB.PopulateStatus(ctx, status); err != nil {
		return results, gtserror.Newf("error populating status: %w", err)
	}

	// Get the string fields status is
	// filterable on for keyword matching.
	fields := getFilterableFields(status)

	// Get all status filters owned by the requesting account.
	filters, err := f.state.DB.GetFiltersByAccountID(ctx, requester.ID)
	if err != nil {
		return results, gtserror.Newf("error getting account filters: %w", err)
	}

	// Generate result for each filter.
	for _, filter := range filters {

		// Skip already expired.
		if filter.Expired(now) {
			continue
		}

		// Later stored API result, if any.
		// (for the HIDE action, it is unset).
		var apiResult *apimodel.FilterResult

		switch filter.Action {
		case gtsmodel.FilterActionWarn:
			// For filter action WARN get all possible filter matches against status.
			keywordMatches, statusMatches := getFilterMatches(filter, status.ID, fields)
			if len(keywordMatches) == 0 && len(statusMatches) == 0 {
				continue
			}

			// Wrap matches in frontend API model.
			apiResult = &apimodel.FilterResult{
				Filter: toAPIFilterV2(filter),

				KeywordMatches: keywordMatches,
				StatusMatches:  statusMatches,
			}

		// For filter action HIDE quickly
		// look for first possible match
		// against this status, or reloop.
		case gtsmodel.FilterActionHide:
			if !doesFilterMatch(filter, status.ID, fields) {
				continue
			}
		}

		// Wrap the filter result in our cache model.
		// This model simply existing implies this
		// status has been filtered, defaulting to
		// action HIDE, or WARN on a non-nil result.
		result := cache.StatusFilterResult{
			Expiry: filter.ExpiresAt,
			Result: apiResult,
		}

		// Append generated result if
		// applies in 'home' context.
		if filter.Contexts.Home() {
			const key = cache.KeyContextHome
			results[key] = append(results[key], result)
		}

		// Append generated result if
		// applies in 'public' context.
		if filter.Contexts.Public() {
			const key = cache.KeyContextPublic
			results[key] = append(results[key], result)
		}

		// Append generated result if
		// applies in 'notifs' context.
		if filter.Contexts.Notifications() {
			const key = cache.KeyContextNotifs
			results[key] = append(results[key], result)
		}

		// Append generated result if
		// applies in 'thread' context.
		if filter.Contexts.Thread() {
			const key = cache.KeyContextThread
			results[key] = append(results[key], result)
		}

		// Append generated result if
		// applies in 'account' context.
		if filter.Contexts.Account() {
			const key = cache.KeyContextAccount
			results[key] = append(results[key], result)
		}
	}

	// Iterate all filter results.
	for _, key := range [5]int{
		cache.KeyContextHome,
		cache.KeyContextPublic,
		cache.KeyContextNotifs,
		cache.KeyContextThread,
		cache.KeyContextAccount,
	} {
		// Sort the slice of filter results by their expiry, soonest coming first.
		slices.SortFunc(results[key], func(a, b cache.StatusFilterResult) int {
			const k = +1
			switch {
			case a.Expiry.IsZero():
				if b.Expiry.IsZero() {
					return 0
				}
				return +k
			case b.Expiry.IsZero():
				return -k
			case a.Expiry.Before(b.Expiry):
				return -k
			case b.Expiry.Before(a.Expiry):
				return +k
			default:
				return 0
			}
		})
	}

	return results, nil
}

// getFilterMatches returns *all* the keyword and status matches of status ID and fields on given filter.
func getFilterMatches(filter *gtsmodel.Filter, statusID string, fields []string) ([]string, []string) {
	keywordMatches := make([]string, 0, len(filter.Keywords))
	for _, keyword := range filter.Keywords {
		if doesKeywordMatch(keyword.Regexp, fields) {
			keywordMatches = append(keywordMatches, keyword.Keyword)
		}
	}
	statusMatches := make([]string, 0, 1)
	for _, status := range filter.Statuses {
		if status.StatusID == statusID {
			statusMatches = append(statusMatches, statusID)
		}
	}
	return keywordMatches, statusMatches
}

// doesFilterMatch returns if any of fields or status ID match on the given filter.
func doesFilterMatch(filter *gtsmodel.Filter, statusID string, fields []string) bool {
	for _, status := range filter.Statuses {
		if status.StatusID == statusID {
			return true
		}
	}
	for _, keyword := range filter.Keywords {
		if doesKeywordMatch(keyword.Regexp, fields) {
			return true
		}
	}
	return false
}

// doesKeywordMatch returns if any of fields match given keyword regex.
func doesKeywordMatch(rgx *regexp.Regexp, fields []string) bool {
	for _, field := range fields {
		if rgx.MatchString(field) {
			return true
		}
	}
	return false
}
