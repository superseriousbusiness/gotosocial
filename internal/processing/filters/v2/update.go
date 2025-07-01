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
	"net/http"
	"slices"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/processing/filters/common"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Update an existing filter for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Update(
	ctx context.Context,
	requester *gtsmodel.Account,
	filterID string,
	form *apimodel.FilterUpdateRequestV2,
) (*apimodel.FilterV2, gtserror.WithCode) {
	// Get the filter with given ID, also checking ownership.
	filter, errWithCode := p.c.GetFilter(ctx, requester, filterID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Filter columns that
	// we're going to update.
	cols := make([]string, 0, 6)

	// Check for title change.
	if form.Title != nil {
		cols = append(cols, "title")
		filter.Title = *form.Title
	}

	// Check action type change.
	if form.FilterAction != nil {
		cols = append(cols, "action")

		// Parse filter action from form and set on filter, checking for validity.
		filter.Action = typeutils.APIFilterActionToFilterAction(*form.FilterAction)
		if filter.Action == 0 {
			const text = "invalid filter action"
			return nil, gtserror.NewWithCode(http.StatusBadRequest, text)
		}
	}

	// Check expiry change.
	if form.ExpiresIn != nil {
		cols = append(cols, "expires_at")
		filter.ExpiresAt = time.Time{}

		// Check form for valid
		// expiry and set on filter.
		if *form.ExpiresIn > 0 {
			expiresIn := time.Duration(*form.ExpiresIn) * time.Second
			filter.ExpiresAt = time.Now().Add(expiresIn)
		}
	}

	// Check context change.
	if form.Context != nil {
		cols = append(cols, "contexts")

		// Parse contexts filter applies in from incoming request form data.
		filter.Contexts, errWithCode = common.FromAPIContexts(*form.Context)
		if errWithCode != nil {
			return nil, errWithCode
		}
	}

	// Check for any changes to attached keywords on filter.
	keywordQs, errWithCode := p.updateFilterKeywords(ctx,
		filter, form.Keywords)
	if errWithCode != nil {
		return nil, errWithCode
	} else if len(keywordQs.create) > 0 || len(keywordQs.delete) > 0 {

		// Attached keywords have changed.
		cols = append(cols, "keywords")
	}

	// Check for any changes to attached statuses on filter.
	statusQs, errWithCode := p.updateFilterStatuses(ctx,
		filter, form.Statuses)
	if errWithCode != nil {
		return nil, errWithCode
	} else if len(statusQs.create) > 0 || len(statusQs.delete) > 0 {

		// Attached statuses have changed.
		cols = append(cols, "statuses")
	}

	// Perform all the deferred database queries.
	errWithCode = performTxs(keywordQs, statusQs)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Update the filter model in the database with determined cols.
	switch err := p.state.DB.UpdateFilter(ctx, filter, cols...); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate title"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("error updating filter: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	// Return as converted frontend filter model.
	return typeutils.FilterToAPIFilterV2(filter), nil
}

func (p *Processor) updateFilterKeywords(ctx context.Context, filter *gtsmodel.Filter, form []apimodel.FilterKeywordCreateUpdateDeleteRequest) (deferredQs, gtserror.WithCode) {
	if len(form) == 0 {
		// No keyword changes.
		return deferredQs{}, nil
	}

	var deferred deferredQs
	for _, request := range form {
		if request.ID != nil {
			// Look by ID for keyword attached to filter.
			idx := slices.IndexFunc(filter.Keywords,
				func(f *gtsmodel.FilterKeyword) bool {
					return f.ID == (*request.ID)
				})
			if idx == -1 {
				const text = "filter keyword not found"
				return deferred, gtserror.NewWithCode(http.StatusNotFound, text)
			}

			// If this is a delete, update filter's id list.
			if request.Destroy != nil && *request.Destroy {
				filter.Keywords = slices.Delete(filter.Keywords, idx, idx+1)
				filter.KeywordIDs = slices.Delete(filter.KeywordIDs, idx, idx+1)

				// Append database delete to funcs for later processing by caller.
				deferred.delete = append(deferred.delete, func() gtserror.WithCode {
					if err := p.state.DB.DeleteFilterKeywordsByIDs(ctx, *request.ID); //
					err != nil {
						err := gtserror.Newf("error deleting filter keyword: %w", err)
						return gtserror.NewErrorInternalError(err)
					}
					return nil
				})
				continue
			}

			// Get the filter keyword at index.
			filterKeyword := filter.Keywords[idx]

			// Filter keywords database
			// columns we need to update.
			cols := make([]string, 0, 2)

			// Check for changes to keyword string.
			if val := request.Keyword; val != nil {
				cols = append(cols, "keyword")
				filterKeyword.Keyword = *val
			}

			// Check for changes to wholeword flag.
			if val := request.WholeWord; val != nil {
				cols = append(cols, "whole_word")
				filterKeyword.WholeWord = val
			}

			// Verify that this is valid regular expression.
			if err := filterKeyword.Compile(); err != nil {
				const text = "invalid regular expression"
				err := gtserror.Newf("invalid regular expression: %w", err)
				return deferred, gtserror.NewWithCodeSafe(
					http.StatusBadRequest,
					err, text,
				)
			}

			if len(cols) > 0 {
				// Append database update to funcs for later processing by caller.
				deferred.update = append(deferred.update, func() gtserror.WithCode {
					if err := p.state.DB.UpdateFilterKeyword(ctx, filterKeyword, cols...); //
					err != nil {
						if errors.Is(err, db.ErrAlreadyExists) {
							const text = "duplicate keyword"
							return gtserror.NewWithCode(http.StatusConflict, text)
						}
						err := gtserror.Newf("error updating filter keyword: %w", err)
						return gtserror.NewErrorInternalError(err)
					}
					return nil
				})
			}

			continue
		}

		// Check for valid request.
		if request.Keyword == nil {
			const text = "missing keyword"
			return deferred, gtserror.NewWithCode(http.StatusBadRequest, text)
		}

		// Create new filter keyword for insert.
		filterKeyword := &gtsmodel.FilterKeyword{
			ID:        id.NewULID(),
			FilterID:  filter.ID,
			Keyword:   *request.Keyword,
			WholeWord: request.WholeWord,
		}

		// Verify that this is valid regular expression.
		if err := filterKeyword.Compile(); err != nil {
			const text = "invalid regular expression"
			err := gtserror.Newf("invalid regular expression: %w", err)
			return deferred, gtserror.NewWithCodeSafe(
				http.StatusBadRequest,
				err, text,
			)
		}

		// Append new filter keyword to filter and list of IDs.
		filter.Keywords = append(filter.Keywords, filterKeyword)
		filter.KeywordIDs = append(filter.KeywordIDs, filterKeyword.ID)

		// Append database insert to funcs for later processing by caller.
		deferred.create = append(deferred.create, func() gtserror.WithCode {
			if err := p.state.DB.PutFilterKeyword(ctx, filterKeyword); //
			err != nil {
				if errors.Is(err, db.ErrAlreadyExists) {
					const text = "duplicate keyword"
					return gtserror.NewWithCode(http.StatusConflict, text)
				}
				err := gtserror.Newf("error inserting filter keyword: %w", err)
				return gtserror.NewErrorInternalError(err)
			}
			return nil
		})
	}

	return deferred, nil
}

func (p *Processor) updateFilterStatuses(ctx context.Context, filter *gtsmodel.Filter, form []apimodel.FilterStatusCreateDeleteRequest) (deferredQs, gtserror.WithCode) {
	if len(form) == 0 {
		// No keyword changes.
		return deferredQs{}, nil
	}

	var deferred deferredQs
	for _, request := range form {
		if request.ID != nil {
			// Look by ID for status attached to filter.
			idx := slices.IndexFunc(filter.Statuses,
				func(f *gtsmodel.FilterStatus) bool {
					return f.ID == *request.ID
				})
			if idx == -1 {
				const text = "filter status not found"
				return deferred, gtserror.NewWithCode(http.StatusNotFound, text)
			}

			// If this is a delete, update filter's id list.
			if request.Destroy != nil && *request.Destroy {
				filter.Statuses = slices.Delete(filter.Statuses, idx, idx+1)
				filter.StatusIDs = slices.Delete(filter.StatusIDs, idx, idx+1)

				// Append database delete to funcs for later processing by caller.
				deferred.delete = append(deferred.delete, func() gtserror.WithCode {
					if err := p.state.DB.DeleteFilterStatusesByIDs(ctx, *request.ID); //
					err != nil {
						err := gtserror.Newf("error deleting filter status: %w", err)
						return gtserror.NewErrorInternalError(err)
					}
					return nil
				})
			}
			continue
		}

		// Check for valid request.
		if request.StatusID == nil {
			const text = "missing status"
			return deferred, gtserror.NewWithCode(http.StatusBadRequest, text)
		}

		// Create new filter status for insert.
		filterStatus := &gtsmodel.FilterStatus{
			ID:       id.NewULID(),
			FilterID: filter.ID,
			StatusID: *request.StatusID,
		}

		// Append new filter status to filter and list of IDs.
		filter.Statuses = append(filter.Statuses, filterStatus)
		filter.StatusIDs = append(filter.StatusIDs, filterStatus.ID)

		// Append database insert to funcs for later processing by caller.
		deferred.create = append(deferred.create, func() gtserror.WithCode {
			if err := p.state.DB.PutFilterStatus(ctx, filterStatus); //
			err != nil {
				if errors.Is(err, db.ErrAlreadyExists) {
					const text = "duplicate status"
					return gtserror.NewWithCode(http.StatusConflict, text)
				}
				err := gtserror.Newf("error inserting filter status: %w", err)
				return gtserror.NewErrorInternalError(err)
			}
			return nil
		})
	}

	return deferred, nil
}

// deferredQs stores selection of
// deferred database queries.
type deferredQs struct {
	create []func() gtserror.WithCode
	update []func() gtserror.WithCode
	delete []func() gtserror.WithCode
}

// performTx performs the passed deferredQs functions,
// prioritising create / update operations before deletes.
func performTxs(queries ...deferredQs) gtserror.WithCode {

	// Perform create / update
	// operations before anything.
	for _, q := range queries {
		for _, create := range q.create {
			if errWithCode := create(); errWithCode != nil {
				return errWithCode
			}
		}
		for _, update := range q.update {
			if errWithCode := update(); errWithCode != nil {
				return errWithCode
			}
		}
	}

	// Perform deletes last.
	for _, q := range queries {
		for _, delete := range q.delete {
			if errWithCode := delete(); errWithCode != nil {
				return errWithCode
			}
		}
	}

	return nil
}
