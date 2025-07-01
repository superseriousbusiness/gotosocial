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
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// KeywordUpdate updates an existing filter keyword for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) KeywordUpdate(
	ctx context.Context,
	requester *gtsmodel.Account,
	filterKeywordID string,
	form *apimodel.FilterKeywordCreateUpdateRequest,
) (*apimodel.FilterKeyword, gtserror.WithCode) {

	// Get the filter keyword with given ID, also checking ownership to requester.
	filterKeyword, _, errWithCode := p.c.GetFilterKeyword(ctx, requester, filterKeywordID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Update the keyword model fields.
	filterKeyword.Keyword = form.Keyword
	filterKeyword.WholeWord = form.WholeWord

	// Update existing filter keyword model in the database, (only necessary cols).
	switch err := p.state.DB.UpdateFilterKeyword(ctx, filterKeyword, []string{
		"keyword", "whole_word"}...); {
	case err == nil:
		// no issue

	case errors.Is(err, db.ErrAlreadyExists):
		const text = "duplicate keyword"
		return nil, gtserror.NewWithCode(http.StatusConflict, text)

	default:
		err := gtserror.Newf("error inserting filter keyword: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	return typeutils.FilterKeywordToAPIFilterKeyword(filterKeyword), nil
}
