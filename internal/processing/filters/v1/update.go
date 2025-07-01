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

package v1

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/processing/filters/common"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Update an existing filter and filter keyword for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) Update(
	ctx context.Context,
	requester *gtsmodel.Account,
	filterKeywordID string,
	form *apimodel.FilterCreateUpdateRequestV1,
) (*apimodel.FilterV1, gtserror.WithCode) {
	// Get the filter keyword with given ID, and associated filter, also checking ownership.
	filterKeyword, filter, errWithCode := p.c.GetFilterKeyword(ctx, requester, filterKeywordID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	var title string
	var action gtsmodel.FilterAction
	var contexts gtsmodel.FilterContexts
	var expiresAt time.Time
	var wholeword bool

	// Get filter title.
	title = form.Phrase

	if *form.Irreversible {
		// Irreversible = action hide.
		action = gtsmodel.FilterActionHide
	} else {
		// Default action = action warn.
		action = gtsmodel.FilterActionWarn
	}

	// Check form for valid expiry and set on filter.
	if form.ExpiresIn != nil && *form.ExpiresIn > 0 {
		expiresIn := time.Duration(*form.ExpiresIn) * time.Second
		expiresAt = time.Now().Add(expiresIn)
	}

	// Parse contexts filter applies in from incoming form data.
	contexts, errWithCode = common.FromAPIContexts(form.Context)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// v1 filter APIs can't change certain fields for a filter with multiple keywords or any statuses,
	// since it would be an unexpected side effect on filters that, to the v1 API, appear separate.
	// See https://docs.joinmastodon.org/methods/filters/#update-v1
	if len(filter.Keywords) > 1 || len(filter.Statuses) > 0 {
		forbiddenFields := make([]string, 0, 4)
		if title != filter.Title {
			forbiddenFields = append(forbiddenFields, "phrase")
		}
		if action != filter.Action {
			forbiddenFields = append(forbiddenFields, "irreversible")
		}
		if expiresAt != filter.ExpiresAt {
			forbiddenFields = append(forbiddenFields, "expires_in")
		}
		if contexts != filter.Contexts {
			forbiddenFields = append(forbiddenFields, "context")
		}
		if len(forbiddenFields) > 0 {
			return nil, gtserror.NewErrorUnprocessableEntity(
				fmt.Errorf("v1 filter backwards compatibility: can't change these fields for a filter with multiple keywords or any statuses: %s", strings.Join(forbiddenFields, ", ")),
			)
		}
	}

	// Filter columns that
	// we're going to update.
	var filterCols []string
	var keywordCols []string

	// Check for changed filter title / filter keyword phrase.
	if title != filter.Title || title != filterKeyword.Keyword {
		keywordCols = append(keywordCols, "keyword")
		filterCols = append(filterCols, "title")
		filterKeyword.Keyword = title
		filter.Title = title
	}

	// Check for changed action.
	if action != filter.Action {
		filterCols = append(filterCols, "action")
		filter.Action = action
	}

	// Check for changed filter expiry time.
	if !expiresAt.Equal(filter.ExpiresAt) {
		filterCols = append(filterCols, "expires_at")
		filter.ExpiresAt = expiresAt
	}

	// Check for changed filter context.
	if contexts != filter.Contexts {
		filterCols = append(filterCols, "contexts")
		filter.Contexts = contexts
	}

	// Check for changed wholeword flag.
	if form.WholeWord != nil &&
		*form.WholeWord != *filterKeyword.WholeWord {
		keywordCols = append(keywordCols, "whole_word")
		filterKeyword.WholeWord = &wholeword
	}

	// Update filter keyword model in the database with determined changed cols.
	switch err := p.state.DB.UpdateFilterKeyword(ctx, filterKeyword, keywordCols...); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate keyword"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("error updating filter: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Update filter model in the database with determined changed cols.
	switch err := p.state.DB.UpdateFilter(ctx, filter, filterCols...); {
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

	// Return as converted frontend filter keyword model.
	return typeutils.FilterKeywordToAPIFilterV1(filter, filterKeyword), nil
}
