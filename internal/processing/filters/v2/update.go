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

package v2

import (
	"context"
	"errors"
	"fmt"
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/typeutils"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// Update an existing filter for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Update(
	ctx context.Context,
	account *gtsmodel.Account,
	filterID string,
	form *apimodel.FilterUpdateRequestV2,
) (*apimodel.FilterV2, gtserror.WithCode) {
	var errWithCode gtserror.WithCode

	// Get the filter by ID, with existing keywords and statuses.
	filter, err := p.state.DB.GetFilterByID(ctx, filterID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filter.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(
			fmt.Errorf("filter %s doesn't belong to account %s", filter.ID, account.ID),
		)
	}

	// Filter columns that we're going to update.
	filterColumns := []string{}

	// Apply filter changes.
	if form.Title != nil {
		filterColumns = append(filterColumns, "title")
		filter.Title = *form.Title
	}
	if form.FilterAction != nil {
		filterColumns = append(filterColumns, "action")
		filter.Action = typeutils.APIFilterActionToFilterAction(*form.FilterAction)
	}
	// TODO: (Vyr) is it possible to unset a filter expiration with this API?
	if form.ExpiresIn != nil {
		filterColumns = append(filterColumns, "expires_at")
		filter.ExpiresAt = time.Now().Add(time.Second * time.Duration(*form.ExpiresIn))
	}
	if form.Context != nil {
		filterColumns = append(filterColumns,
			"context_home",
			"context_notifications",
			"context_public",
			"context_thread",
			"context_account",
		)
		filter.ContextHome = util.Ptr(false)
		filter.ContextNotifications = util.Ptr(false)
		filter.ContextPublic = util.Ptr(false)
		filter.ContextThread = util.Ptr(false)
		filter.ContextAccount = util.Ptr(false)
		for _, context := range *form.Context {
			switch context {
			case apimodel.FilterContextHome:
				filter.ContextHome = util.Ptr(true)
			case apimodel.FilterContextNotifications:
				filter.ContextNotifications = util.Ptr(true)
			case apimodel.FilterContextPublic:
				filter.ContextPublic = util.Ptr(true)
			case apimodel.FilterContextThread:
				filter.ContextThread = util.Ptr(true)
			case apimodel.FilterContextAccount:
				filter.ContextAccount = util.Ptr(true)
			default:
				return nil, gtserror.NewErrorUnprocessableEntity(
					fmt.Errorf("unsupported filter context '%s'", context),
				)
			}
		}
	}

	filterKeywordColumns, deleteFilterKeywordIDs, errWithCode := applyKeywordChanges(filter, form.Keywords)
	if err != nil {
		return nil, errWithCode
	}

	deleteFilterStatusIDs, errWithCode := applyStatusChanges(filter, form.Statuses)
	if err != nil {
		return nil, errWithCode
	}

	if err := p.state.DB.UpdateFilter(ctx, filter, filterColumns, filterKeywordColumns, deleteFilterKeywordIDs, deleteFilterStatusIDs); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("you already have a filter with this title")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilter, errWithCode := p.apiFilter(ctx, filter)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return apiFilter, nil
}

// applyKeywordChanges applies the provided changes to the filter's keywords in place,
// and returns a list of lists of filter columns to update, and a list of filter keyword IDs to delete.
func applyKeywordChanges(filter *gtsmodel.Filter, formKeywords []apimodel.FilterKeywordCreateUpdateDeleteRequest) ([][]string, []string, gtserror.WithCode) {
	if len(formKeywords) == 0 {
		// Detach currently existing keywords from the filter so we don't change them.
		filter.Keywords = nil
		return nil, nil, nil
	}

	deleteFilterKeywordIDs := []string{}
	filterKeywordsByID := map[string]*gtsmodel.FilterKeyword{}
	filterKeywordColumnsByID := map[string][]string{}
	for _, filterKeyword := range filter.Keywords {
		filterKeywordsByID[filterKeyword.ID] = filterKeyword
	}

	for _, formKeyword := range formKeywords {
		if formKeyword.ID != nil {
			id := *formKeyword.ID
			filterKeyword, ok := filterKeywordsByID[id]
			if !ok {
				return nil, nil, gtserror.NewErrorNotFound(
					fmt.Errorf("couldn't find filter keyword '%s' to update or delete", id),
				)
			}

			// Process deletes.
			if *formKeyword.Destroy {
				delete(filterKeywordsByID, id)
				deleteFilterKeywordIDs = append(deleteFilterKeywordIDs, id)
				continue
			}

			// Process updates.
			columns := make([]string, 0, 2)
			if formKeyword.Keyword != nil {
				columns = append(columns, "keyword")
				filterKeyword.Keyword = *formKeyword.Keyword
			}
			if formKeyword.WholeWord != nil {
				columns = append(columns, "whole_word")
				filterKeyword.WholeWord = formKeyword.WholeWord
			}
			filterKeywordColumnsByID[id] = columns
			continue
		}

		// Process creates.
		filterKeyword := &gtsmodel.FilterKeyword{
			ID:        id.NewULID(),
			AccountID: filter.AccountID,
			FilterID:  filter.ID,
			Filter:    filter,
			Keyword:   *formKeyword.Keyword,
			WholeWord: util.Ptr(util.PtrOrValue(formKeyword.WholeWord, false)),
		}
		filterKeywordsByID[filterKeyword.ID] = filterKeyword
		// Don't need to set columns, as we're using all of them.
	}

	// Replace the filter's keywords list with our updated version.
	filterKeywordColumns := [][]string{}
	filter.Keywords = nil
	for id, filterKeyword := range filterKeywordsByID {
		filter.Keywords = append(filter.Keywords, filterKeyword)
		// Okay to use the nil slice zero value for entries being created instead of updated.
		filterKeywordColumns = append(filterKeywordColumns, filterKeywordColumnsByID[id])
	}

	return filterKeywordColumns, deleteFilterKeywordIDs, nil
}

// applyKeywordChanges applies the provided changes to the filter's keywords in place,
// and returns a list of filter status IDs to delete.
func applyStatusChanges(filter *gtsmodel.Filter, formStatuses []apimodel.FilterStatusCreateDeleteRequest) ([]string, gtserror.WithCode) {
	if len(formStatuses) == 0 {
		// Detach currently existing statuses from the filter so we don't change them.
		filter.Statuses = nil
		return nil, nil
	}

	deleteFilterStatusIDs := []string{}
	filterStatusesByID := map[string]*gtsmodel.FilterStatus{}
	for _, filterStatus := range filter.Statuses {
		filterStatusesByID[filterStatus.ID] = filterStatus
	}

	for _, formStatus := range formStatuses {
		if formStatus.ID != nil {
			id := *formStatus.ID
			_, ok := filterStatusesByID[id]
			if !ok {
				return nil, gtserror.NewErrorNotFound(
					fmt.Errorf("couldn't find filter status '%s' to delete", id),
				)
			}

			// Process deletes.
			if *formStatus.Destroy {
				delete(filterStatusesByID, id)
				deleteFilterStatusIDs = append(deleteFilterStatusIDs, id)
				continue
			}

			// Filter statuses don't have updates.
			continue
		}

		// Process creates.
		filterStatus := &gtsmodel.FilterStatus{
			ID:        id.NewULID(),
			AccountID: filter.AccountID,
			FilterID:  filter.ID,
			Filter:    filter,
			StatusID:  *formStatus.StatusID,
		}
		filterStatusesByID[filterStatus.ID] = filterStatus
	}

	// Replace the filter's keywords list with our updated version.
	filter.Statuses = nil
	for _, filterStatus := range filterStatusesByID {
		filter.Statuses = append(filter.Statuses, filterStatus)
	}

	return deleteFilterStatusIDs, nil
}
