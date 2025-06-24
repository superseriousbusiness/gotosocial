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
	"slices"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Get looks up a filter keyword by ID and returns it as a v1 filter.
func (p *Processor) Get(ctx context.Context, requester *gtsmodel.Account, filterKeywordID string) (*apimodel.FilterV1, gtserror.WithCode) {
	filterKeyword, filter, errWithCode := p.c.GetFilterKeyword(ctx, requester, filterKeywordID)
	if errWithCode != nil {
		return nil, errWithCode
	}
	return typeutils.FilterKeywordToAPIFilterV1(filter, filterKeyword), nil
}

// GetAll looks up all filter keywords for the current account and returns them as v1 filters.
func (p *Processor) GetAll(ctx context.Context, requester *gtsmodel.Account) ([]*apimodel.FilterV1, gtserror.WithCode) {
	var totalKeywords int

	// Get a list of all filters owned by this account,
	// (without any sub-models attached, done later).
	filters, err := p.state.DB.GetFiltersByAccountID(
		gtscontext.SetBarebones(ctx),
		requester.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting filters: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Get a total count of all expected
	// keywords for slice preallocation.
	for _, filter := range filters {
		totalKeywords += len(filter.KeywordIDs)
	}

	// Create a slice to store converted V1 frontend models.
	apiFilters := make([]*apimodel.FilterV1, 0, totalKeywords)

	for _, filter := range filters {
		// For each of the fetched filters, fetch all of their associated keywords.
		keywords, err := p.state.DB.GetFilterKeywordsByIDs(ctx, filter.KeywordIDs)
		if err != nil {
			err := gtserror.Newf("error getting filter keywords: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Convert each keyword to frontend.
		for _, keyword := range keywords {
			apiFilter := typeutils.FilterKeywordToAPIFilterV1(filter, keyword)
			apiFilters = append(apiFilters, apiFilter)
		}
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them lexically in a locale-aware manner.
	slices.SortFunc(apiFilters, func(lhs *apimodel.FilterV1, rhs *apimodel.FilterV1) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilters, nil
}
