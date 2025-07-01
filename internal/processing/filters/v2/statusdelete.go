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

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// StatusDelete deletes an existing filter status from a filter.
func (p *Processor) StatusDelete(
	ctx context.Context,
	requester *gtsmodel.Account,
	filterStatusID string,
) gtserror.WithCode {
	// Get filter status with given ID, also checking ownership to requester.
	_, filter, errWithCode := p.c.GetFilterStatus(ctx, requester, filterStatusID)
	if errWithCode != nil {
		return errWithCode
	}

	// Delete this one filter status from the database, now ownership is confirmed.
	if err := p.state.DB.DeleteFilterStatusesByIDs(ctx, filterStatusID); err != nil {
		err := gtserror.Newf("error deleting filter status: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Delete this filter keyword from the slice of IDs attached to filter.
	filter.StatusIDs = slices.DeleteFunc(filter.StatusIDs, func(id string) bool {
		return filterStatusID == id
	})

	// Update filter in the database now the status has been unattached.
	if err := p.state.DB.UpdateFilter(ctx, filter, "statuses"); err != nil {
		err := gtserror.Newf("error updating filter: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	return nil
}
