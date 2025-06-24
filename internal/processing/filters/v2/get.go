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
	"slices"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
)

// Get looks up a filter by ID and returns it with keywords and statuses.
func (p *Processor) Get(ctx context.Context, requester *gtsmodel.Account, filterID string) (*apimodel.FilterV2, gtserror.WithCode) {
	filter, errWithCode := p.c.GetFilter(ctx, requester, filterID)
	if errWithCode != nil {
		return nil, errWithCode
	}
	return typeutils.FilterToAPIFilterV2(filter), nil
}

// GetAll looks up all filters for the current account and returns them with keywords and statuses.
func (p *Processor) GetAll(ctx context.Context, requester *gtsmodel.Account) ([]*apimodel.FilterV2, gtserror.WithCode) {

	// Get all filters belonging to this requester from the database.
	filters, err := p.state.DB.GetFiltersByAccountID(ctx, requester.ID)
	if err != nil {
		err := gtserror.Newf("error getting account filters: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert all these filters to frontend API models.
	apiFilters := make([]*apimodel.FilterV2, len(filters))
	if len(apiFilters) != len(filters) {
		// bound check eliminiation compiler-hint
		panic(gtserror.New("BCE"))
	}
	for i, filter := range filters {
		apiFilter := typeutils.FilterToAPIFilterV2(filter)
		apiFilters[i] = apiFilter
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them lexically in a locale-aware manner.
	slices.SortFunc(apiFilters, func(lhs *apimodel.FilterV2, rhs *apimodel.FilterV2) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilters, nil
}
