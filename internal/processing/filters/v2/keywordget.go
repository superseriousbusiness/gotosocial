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
	"slices"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// KeywordGet looks up a filter keyword by ID.
func (p *Processor) KeywordGet(ctx context.Context, account *gtsmodel.Account, filterKeywordID string) (*apimodel.FilterKeyword, gtserror.WithCode) {
	filterKeyword, err := p.state.DB.GetFilterKeywordByID(ctx, filterKeywordID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filterKeyword.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(
			fmt.Errorf("filter keyword %s doesn't belong to account %s", filterKeyword.ID, account.ID),
		)
	}

	return p.converter.FilterKeywordToAPIFilterKeyword(ctx, filterKeyword), nil
}

// KeywordsGetForFilterID looks up all filter keywords for the given filter.
func (p *Processor) KeywordsGetForFilterID(ctx context.Context, account *gtsmodel.Account, filterID string) ([]*apimodel.FilterKeyword, gtserror.WithCode) {
	// Check that the filter is owned by the given account.
	filter, err := p.state.DB.GetFilterByID(gtscontext.SetBarebones(ctx), filterID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filter.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(nil)
	}

	filterKeywords, err := p.state.DB.GetFilterKeywordsForFilterID(
		ctx,
		filter.ID,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilterKeywords := make([]*apimodel.FilterKeyword, 0, len(filterKeywords))
	for _, filterKeyword := range filterKeywords {
		apiFilterKeywords = append(apiFilterKeywords, p.converter.FilterKeywordToAPIFilterKeyword(ctx, filterKeyword))
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them lexically in a locale-aware manner.
	slices.SortFunc(apiFilterKeywords, func(lhs *apimodel.FilterKeyword, rhs *apimodel.FilterKeyword) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilterKeywords, nil
}
