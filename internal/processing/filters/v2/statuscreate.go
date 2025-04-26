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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// StatusCreate adds a filter status to an existing filter for the given account, using the provided parameters.
// These params should have already been validated by the time they reach this function.
func (p *Processor) StatusCreate(ctx context.Context, account *gtsmodel.Account, filterID string, form *apimodel.FilterStatusCreateRequest) (*apimodel.FilterStatus, gtserror.WithCode) {
	// Check that the filter is owned by the given account.
	filter, err := p.state.DB.GetFilterByID(gtscontext.SetBarebones(ctx), filterID)
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

	filterStatus := &gtsmodel.FilterStatus{
		ID:        id.NewULID(),
		AccountID: account.ID,
		FilterID:  filter.ID,
		StatusID:  form.StatusID,
	}

	if err := p.state.DB.PutFilterStatus(ctx, filterStatus); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			err = errors.New("duplicate status")
			return nil, gtserror.NewErrorConflict(err, err.Error())
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return p.converter.FilterStatusToAPIFilterStatus(ctx, filterStatus), nil
}
