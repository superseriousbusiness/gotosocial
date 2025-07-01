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
	"slices"

	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Delete an existing filter keyword and (if empty
// afterwards) filter for the given account.
func (p *Processor) Delete(
	ctx context.Context,
	requester *gtsmodel.Account,
	filterKeywordID string,
) gtserror.WithCode {
	// Get the filter keyword with given ID, and associated filter, also checking ownership.
	filterKeyword, filter, errWithCode := p.c.GetFilterKeyword(ctx, requester, filterKeywordID)
	if errWithCode != nil {
		return errWithCode
	}

	if len(filter.Keywords) > 1 || len(filter.Statuses) > 0 {
		// The filter has other keywords or statuses, just delete the one filter keyword.
		if err := p.state.DB.DeleteFilterKeywordsByIDs(ctx, filterKeyword.ID); err != nil {
			err := gtserror.Newf("error deleting filter keyword: %w", err)
			return gtserror.NewErrorInternalError(err)
		}

		// Delete this filter keyword from the slice of IDs attached to filter.
		filter.KeywordIDs = slices.DeleteFunc(filter.KeywordIDs, func(id string) bool {
			return filterKeyword.ID == id
		})

		// Update filter in the database now the keyword has been unattached.
		if err := p.state.DB.UpdateFilter(ctx, filter, "keywords"); err != nil {
			err := gtserror.Newf("error updating filter: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
	} else {
		// Delete the filter and this keyword that is attached to it.
		if err := p.state.DB.DeleteFilter(ctx, filter); err != nil {
			err := gtserror.Newf("error deleting filter: %w", err)
			return gtserror.NewErrorInternalError(err)
		}
	}

	// Handle filter change side-effects.
	p.c.OnFilterChanged(ctx, requester)

	return nil
}
