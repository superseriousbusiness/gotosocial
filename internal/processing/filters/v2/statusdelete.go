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
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// StatusDelete deletes an existing filter status from a filter.
func (p *Processor) StatusDelete(
	ctx context.Context,
	account *gtsmodel.Account,
	filterID string,
) gtserror.WithCode {
	// Get the filter status.
	filterStatus, err := p.state.DB.GetFilterStatusByID(ctx, filterID)
	if err != nil {
		return gtserror.NewErrorNotFound(err)
	}

	// Check that the account owns it.
	if filterStatus.AccountID != account.ID {
		return gtserror.NewErrorNotFound(
			fmt.Errorf("filter status %s doesn't belong to account %s", filterStatus.ID, account.ID),
		)
	}

	// Delete the filter status.
	if err := p.state.DB.DeleteFilterStatusByID(ctx, filterStatus.ID); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return nil
}
