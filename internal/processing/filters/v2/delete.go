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

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Delete an existing filter and all its attached
// keywords and statuses for the given account.
func (p *Processor) Delete(
	ctx context.Context,
	requester *gtsmodel.Account,
	filterID string,
) gtserror.WithCode {

	// Get the filter with given ID, also checking ownership.
	filter, errWithCode := p.c.GetFilter(ctx, requester, filterID)
	if errWithCode != nil {
		return errWithCode
	}

	// Delete filter from the database with all associated models.
	if err := p.state.DB.DeleteFilter(ctx, filter); err != nil {
		err := gtserror.Newf("error deleting filter: %w", err)
		return gtserror.NewErrorInternalError(err)
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	return nil
}
