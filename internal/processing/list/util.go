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

package list

import (
	"context"
	"errors"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// getList is a shortcut to get one list from the database and
// check that it's owned by the given accountID. Will return
// appropriate errors so caller doesn't need to bother.
func (p *Processor) getList(ctx context.Context, accountID string, listID string) (*gtsmodel.List, gtserror.WithCode) {
	list, err := p.state.DB.GetListByID(ctx, listID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting list: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list == nil {
		const text = "list not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	if list.AccountID != accountID {
		const text = "list not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New("list does not belong to account"),
			text,
		)
	}

	return list, nil
}

// apiList is a shortcut to return the API version of the given
// list, or return an appropriate error if conversion fails.
func (p *Processor) apiList(ctx context.Context, list *gtsmodel.List) (*apimodel.List, gtserror.WithCode) {
	apiList, err := p.converter.ListToAPIList(ctx, list)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting list to api: %w", err))
	}

	return apiList, nil
}
