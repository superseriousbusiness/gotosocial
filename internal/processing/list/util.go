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

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// getList is a shortcut to get one list from the database and
// check that it's owned by the given accountID. Will return
// appropriate errors so caller doesn't need to bother.
func (p *Processor) getList(ctx context.Context, accountID string, listID string) (*gtsmodel.List, gtserror.WithCode) {
	list, err := p.state.DB.GetListByID(ctx, listID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			// List doesn't seem to exist.
			return nil, gtserror.NewErrorNotFound(err)
		}
		// Real database error.
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list.AccountID != accountID {
		err = fmt.Errorf("list with id %s does not belong to account %s", list.ID, accountID)
		return nil, gtserror.NewErrorNotFound(err)
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

// isInList check if thisID is equal to the result of thatID
// for any entry in the given list.
//
// Will return the id of the listEntry if true, empty if false,
// or an error if the result of thatID returns an error.
func isInList(
	list *gtsmodel.List,
	thisID string,
	getThatID func(listEntry *gtsmodel.ListEntry) (string, error),
) (string, error) {
	for _, listEntry := range list.ListEntries {
		thatID, err := getThatID(listEntry)
		if err != nil {
			return "", err
		}

		if thisID == thatID {
			return listEntry.ID, nil
		}
	}
	return "", nil
}
