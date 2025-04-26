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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Delete an existing filter keyword and (if empty afterwards) filter for the given account.
func (p *Processor) Delete(
	ctx context.Context,
	account *gtsmodel.Account,
	filterKeywordID string,
) gtserror.WithCode {
	// Get enough of the filter keyword that we can look up its filter ID.
	filterKeyword, err := p.state.DB.GetFilterKeywordByID(gtscontext.SetBarebones(ctx), filterKeywordID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return gtserror.NewErrorNotFound(err)
		}
		return gtserror.NewErrorInternalError(err)
	}
	if filterKeyword.AccountID != account.ID {
		return gtserror.NewErrorNotFound(nil)
	}

	// Get the filter for this keyword.
	filter, err := p.state.DB.GetFilterByID(ctx, filterKeyword.FilterID)
	if err != nil {
		return gtserror.NewErrorNotFound(err)
	}

	if len(filter.Keywords) > 1 || len(filter.Statuses) > 0 {
		// The filter has other keywords or statuses. Delete only the requested filter keyword.
		if err := p.state.DB.DeleteFilterKeywordByID(ctx, filterKeyword.ID); err != nil {
			return gtserror.NewErrorInternalError(err)
		}
	} else {
		// Delete the entire filter.
		if err := p.state.DB.DeleteFilterByID(ctx, filter.ID); err != nil {
			return gtserror.NewErrorInternalError(err)
		}
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return nil
}
