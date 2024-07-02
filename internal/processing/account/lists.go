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

package account

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

var noLists = make([]*apimodel.List, 0)

// ListsGet returns all lists owned by requestingAccount, which contain a follow for targetAccountID.
func (p *Processor) ListsGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) ([]*apimodel.List, gtserror.WithCode) {
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %w", err))
	}

	visible, err := p.visFilter.AccountVisible(ctx, requestingAccount, targetAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %w", err))
	}

	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("account not found"))
	}

	// Requester has to follow targetAccount
	// for them to be in any of their lists.
	follow, err := p.state.DB.GetFollow(
		// Don't populate follow.
		gtscontext.SetBarebones(ctx),
		requestingAccount.ID,
		targetAccountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %w", err))
	}

	if follow == nil {
		return noLists, nil // by definition we know they're in no lists
	}

	listEntries, err := p.state.DB.GetListEntriesForFollowID(
		// Don't populate entries.
		gtscontext.SetBarebones(ctx),
		follow.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %w", err))
	}

	count := len(listEntries)
	if count == 0 {
		return noLists, nil
	}

	apiLists := make([]*apimodel.List, 0, count)
	for _, listEntry := range listEntries {
		list, err := p.state.DB.GetListByID(
			// Don't populate list.
			gtscontext.SetBarebones(ctx),
			listEntry.ListID,
		)

		if err != nil {
			log.Debugf(ctx, "skipping list %s due to error %q", listEntry.ListID, err)
			continue
		}

		apiList, err := p.converter.ListToAPIList(ctx, list)
		if err != nil {
			log.Debugf(ctx, "skipping list %s due to error %q", listEntry.ListID, err)
			continue
		}

		apiLists = append(apiLists, apiList)
	}

	return apiLists, nil
}
