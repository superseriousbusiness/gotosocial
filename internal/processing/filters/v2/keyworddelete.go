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

// KeywordDelete deletes an existing filter keyword from a filter.
func (p *Processor) KeywordDelete(
	ctx context.Context,
	account *gtsmodel.Account,
	filterID string,
) gtserror.WithCode {
	// Get the filter keyword.
	filterKeyword, err := p.state.DB.GetFilterKeywordByID(ctx, filterID)
	if err != nil {
		return gtserror.NewErrorNotFound(err)
	}

	// Check that the account owns it.
	if filterKeyword.AccountID != account.ID {
		return gtserror.NewErrorNotFound(
			fmt.Errorf("filter keyword %s doesn't belong to account %s", filterKeyword.ID, account.ID),
		)
	}

	// Delete the filter keyword.
	if err := p.state.DB.DeleteFilterKeywordByID(ctx, filterKeyword.ID); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	// Send a filters changed event.
	p.stream.FiltersChanged(ctx, account)

	return nil
}
