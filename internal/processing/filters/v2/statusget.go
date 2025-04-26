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

// StatusGet looks up a filter status by ID.
func (p *Processor) StatusGet(ctx context.Context, account *gtsmodel.Account, filterStatusID string) (*apimodel.FilterStatus, gtserror.WithCode) {
	filterStatus, err := p.state.DB.GetFilterStatusByID(ctx, filterStatusID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}
	if filterStatus.AccountID != account.ID {
		return nil, gtserror.NewErrorNotFound(
			fmt.Errorf("filter status %s doesn't belong to account %s", filterStatus.ID, account.ID),
		)
	}

	return p.converter.FilterStatusToAPIFilterStatus(ctx, filterStatus), nil
}

// StatusesGetForFilterID looks up all filter statuses for the given filter.
func (p *Processor) StatusesGetForFilterID(ctx context.Context, account *gtsmodel.Account, filterID string) ([]*apimodel.FilterStatus, gtserror.WithCode) {
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

	filterStatuses, err := p.state.DB.GetFilterStatusesForFilterID(
		ctx,
		filter.ID,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiFilterStatuses := make([]*apimodel.FilterStatus, 0, len(filterStatuses))
	for _, filterStatus := range filterStatuses {
		apiFilterStatuses = append(apiFilterStatuses, p.converter.FilterStatusToAPIFilterStatus(ctx, filterStatus))
	}

	// Sort them by ID so that they're in a stable order.
	// Clients may opt to sort them by status ID instead.
	slices.SortFunc(apiFilterStatuses, func(lhs *apimodel.FilterStatus, rhs *apimodel.FilterStatus) int {
		return strings.Compare(lhs.ID, rhs.ID)
	})

	return apiFilterStatuses, nil
}
