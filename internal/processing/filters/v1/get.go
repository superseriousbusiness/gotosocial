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
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Get looks up a filter keyword by ID and returns it as a v1 filter.
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, filterKeywordID string) (*apimodel.FilterV1, gtserror.WithCode) {
	filterKeyword, err := p.state.DB.GetFilterKeywordByID(ctx, filterKeywordID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filterKeyword.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(nil)
	}

	return p.apiFilter(ctx, filterKeyword)
}

// GetAll looks up all filter keywords for the current account and returns them as v1 filters.
func (p *Processor) GetAll(ctx context.Context, account *gtsmodel.Account) ([]*apimodel.FilterV1, gtserror.WithCode) {
	filters, err := p.state.DB.GetFilterKeywordsForAccountID(
		ctx,
		account.ID,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilters := make([]*apimodel.FilterV1, 0, len(filters))
	for _, filter := range filters {
		apiFilter, errWithCode := p.apiFilter(ctx, filter)
		if errWithCode != nil {
			return nil, errWithCode
		}

		apiFilters = append(apiFilters, apiFilter)
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them lexically in a locale-aware manner.
	slices.SortFunc(apiFilters, func(lhs *apimodel.FilterV1, rhs *apimodel.FilterV1) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilters, nil
}
