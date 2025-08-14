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
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/processing/filters/common"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Create a new filter for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Create(ctx context.Context, requester *gtsmodel.Account, form *apimodel.FilterCreateRequestV2) (*apimodel.FilterV2, gtserror.WithCode) {
	var errWithCode gtserror.WithCode

	// Create new filter model.
	filter := &gtsmodel.Filter{
		ID:        id.NewULID(),
		AccountID: requester.ID,
		Title:     form.Title,
	}

	// Parse filter action from form and set on filter, checking for validity.
	filter.Action = typeutils.APIFilterActionToFilterAction(*form.FilterAction)
	if filter.Action == 0 {
		const text = "invalid filter action"
		return nil, gtserror.NewWithCode(http.StatusBadRequest, text)
	}

	// Parse contexts filter applies in from incoming request form data.
	filter.Contexts, errWithCode = common.FromAPIContexts(form.Context)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check form for valid expiry and set on filter.
	if form.ExpiresIn != nil && *form.ExpiresIn > 0 {
		expiresIn := time.Duration(*form.ExpiresIn) * time.Second
		filter.ExpiresAt = time.Now().Add(expiresIn)
	}

	// Create new attached filter keywords.
	keywordQueries, errWithCode := p.createFilterKeywords(ctx,
		filter, form.Keywords)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Create new attached filter statuses.
	statusQueries, errWithCode := p.createFilterStatuses(ctx,
		filter, form.Statuses)
	if errWithCode != nil {
		return nil, errWithCode
	}

	for _, keywordCreate := range keywordQueries {
		if errWithCode := keywordCreate(); errWithCode != nil {
			return nil, errWithCode
		}
	}

	for _, statusCreate := range statusQueries {
		if errWithCode := statusCreate(); errWithCode != nil {
			return nil, errWithCode
		}
	}

	// Insert the new filter model into the database.
	switch err := p.state.DB.PutFilter(ctx, filter); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate title, keyword or status"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("error inserting filter: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	// Return as converted frontend filter model.
	return typeutils.FilterToAPIFilterV2(filter), nil
}

func (p *Processor) createFilterKeywords(ctx context.Context, filter *gtsmodel.Filter, form []apimodel.FilterKeywordCreateUpdateRequest) ([]func() gtserror.WithCode, gtserror.WithCode) {
	if len(form) == 0 {
		// No keywords created.
		return nil, nil
	}

	var deferred []func() gtserror.WithCode

	// Create filter keywords in the database.
	for _, request := range form {
		// Check for valid request.
		if request.Keyword == "" {
			const text = "missing keyword"
			return deferred, gtserror.NewWithCode(http.StatusBadRequest, text)
		}

		// Create new filter keyword for insert.
		filterKeyword := &gtsmodel.FilterKeyword{
			ID:        id.NewULID(),
			FilterID:  filter.ID,
			Keyword:   request.Keyword,
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
		deferred = append(deferred, func() gtserror.WithCode {
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

func (p *Processor) createFilterStatuses(ctx context.Context, filter *gtsmodel.Filter, form []apimodel.FilterStatusCreateRequest) ([]func() gtserror.WithCode, gtserror.WithCode) {
	if len(form) == 0 {
		// No statuses added.
		return nil, nil
	}

	var deferred []func() gtserror.WithCode

	// Create filter statuses in the database.
	for _, request := range form {
		// Check for valid request.
		if request.StatusID == "" {
			const text = "missing status"
			return deferred, gtserror.NewWithCode(http.StatusBadRequest, text)
		}

		// Create new filter status for insert.
		filterStatus := &gtsmodel.FilterStatus{
			ID:       id.NewULID(),
			FilterID: filter.ID,
			StatusID: request.StatusID,
		}

		// Append new filter status to filter and list of IDs.
		filter.Statuses = append(filter.Statuses, filterStatus)
		filter.StatusIDs = append(filter.StatusIDs, filterStatus.ID)

		// Append database insert to funcs for later processing by caller.
		deferred = append(deferred, func() gtserror.WithCode {
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
