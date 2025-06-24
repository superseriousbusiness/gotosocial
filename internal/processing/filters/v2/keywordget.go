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
	"slices"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// KeywordGet looks up a filter keyword by ID.
func (p *Processor) KeywordGet(ctx context.Context, requester *gtsmodel.Account, filterKeywordID string) (*apimodel.FilterKeyword, gtserror.WithCode) {
	filterKeyword, _, errWithCode := p.c.GetFilterKeyword(ctx, requester, filterKeywordID)
	if errWithCode != nil {
		return nil, errWithCode
	}
	return typeutils.FilterKeywordToAPIFilterKeyword(filterKeyword), nil
}

// KeywordsGetForFilterID looks up all filter keywords for the given filter.
func (p *Processor) KeywordsGetForFilterID(ctx context.Context, requester *gtsmodel.Account, filterID string) ([]*apimodel.FilterKeyword, gtserror.WithCode) {

	// Get the filter with given ID (but
	// without any sub-models attached).
	filter, errWithCode := p.c.GetFilter(
		gtscontext.SetBarebones(ctx),
		requester,
		filterID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Fetch all associated filter keywords to the determined existent filter.
	filterKeywords, err := p.state.DB.GetFilterKeywordsByIDs(ctx, filter.KeywordIDs)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting filter keywords: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert all of the filter keyword models from internal to frontend form.
	apiFilterKeywords := make([]*apimodel.FilterKeyword, len(filterKeywords))
	if len(apiFilterKeywords) != len(filterKeywords) {
		// bound check eliminiation compiler-hint
		panic(gtserror.New("BCE"))
	}
	for i, filterKeyword := range filterKeywords {
		apiFilterKeywords[i] = typeutils.FilterKeywordToAPIFilterKeyword(filterKeyword)
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them lexically in a locale-aware manner.
	slices.SortFunc(apiFilterKeywords, func(lhs *apimodel.FilterKeyword, rhs *apimodel.FilterKeyword) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilterKeywords, nil
}
