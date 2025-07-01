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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// KeywordCreate adds a filter keyword to an existing filter for the given account, using the provided parameters.
// These params should have already been normalized and validated by the time they reach this function.
func (p *Processor) KeywordCreate(ctx context.Context, requester *gtsmodel.Account, filterID string, form *apimodel.FilterKeywordCreateUpdateRequest) (*apimodel.FilterKeyword, gtserror.WithCode) {

	// Get the filter with given ID, also checking ownership.
	filter, errWithCode := p.c.GetFilter(ctx, requester, filterID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Create new filter keyword model.
	filterKeyword := &gtsmodel.FilterKeyword{
		ID:        id.NewULID(),
		FilterID:  filter.ID,
		Keyword:   form.Keyword,
		WholeWord: form.WholeWord,
	}

	// Insert the new filter keyword model into the database.
	switch err := p.state.DB.PutFilterKeyword(ctx, filterKeyword); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate keyword"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("error inserting filter keyword: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Now update the filter it is attached to with new keyword.
	filter.KeywordIDs = append(filter.KeywordIDs, filterKeyword.ID)
	filter.Keywords = append(filter.Keywords, filterKeyword)

	// Update the existing filter model in the database (only the needed col).
	if err := p.state.DB.UpdateFilter(ctx, filter, "keywords"); err != nil {
		err := gtserror.Newf("error updating filter: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	return typeutils.FilterKeywordToAPIFilterKeyword(filterKeyword), nil
}
