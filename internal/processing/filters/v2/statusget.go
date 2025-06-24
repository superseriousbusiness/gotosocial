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

// StatusGet looks up a filter status by ID.
func (p *Processor) StatusGet(ctx context.Context, requester *gtsmodel.Account, filterStatusID string) (*apimodel.FilterStatus, gtserror.WithCode) {
	filterStatus, _, errWithCode := p.c.GetFilterStatus(ctx, requester, filterStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}
	return typeutils.FilterStatusToAPIFilterStatus(filterStatus), nil
}

// StatusesGetForFilterID looks up all filter statuses for the given filter.
func (p *Processor) StatusesGetForFilterID(ctx context.Context, requester *gtsmodel.Account, filterID string) ([]*apimodel.FilterStatus, gtserror.WithCode) {

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

	// Fetch all associated filter statuses to the determined existent filter.
	filterStatuses, err := p.state.DB.GetFilterStatusesByIDs(ctx, filter.StatusIDs)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting filter statuses: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Convert all of the filter status models from internal to frontend form.
	apiFilterStatuses := make([]*apimodel.FilterStatus, len(filterStatuses))
	if len(apiFilterStatuses) != len(filterStatuses) {
		// bound check eliminiation compiler-hint
		panic(gtserror.New("BCE"))
	}
	for i, filterStatus := range filterStatuses {
		apiFilterStatuses[i] = typeutils.FilterStatusToAPIFilterStatus(filterStatus)
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them by status ID instead.
	slices.SortFunc(apiFilterStatuses, func(lhs *apimodel.FilterStatus, rhs *apimodel.FilterStatus) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilterStatuses, nil
}
