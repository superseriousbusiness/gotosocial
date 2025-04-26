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
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Get looks up a filter by ID and returns it with keywords and statuses.
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, filterID string) (*apimodel.FilterV2, gtserror.WithCode) {
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

	return p.apiFilter(ctx, filter)
}

// GetAll looks up all filters for the current account and returns them with keywords and statuses.
func (p *Processor) GetAll(ctx context.Context, account *gtsmodel.Account) ([]*apimodel.FilterV2, gtserror.WithCode) {
	filters, err := p.state.DB.GetFiltersForAccountID(
		ctx,
		account.ID,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilters := make([]*apimodel.FilterV2, 0, len(filters))
	for _, filter := range filters {
		apiFilter, errWithCode := p.apiFilter(ctx, filter)
		if errWithCode != nil {
			return nil, errWithCode
		}

		apiFilters = append(apiFilters, apiFilter)
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them lexically in a locale-aware manner.
	slices.SortFunc(apiFilters, func(lhs *apimodel.FilterV2, rhs *apimodel.FilterV2) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilters, nil
}
