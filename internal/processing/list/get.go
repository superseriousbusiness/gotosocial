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
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// shortcut to get one list from the db, or error appropriately.
func (p *Processor) getList(ctx context.Context, accountID string, listID string) (*gtsmodel.List, gtserror.WithCode) {
	list, err := p.state.DB.GetListByID(ctx, listID)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	if list.AccountID != accountID {
		err = fmt.Errorf("list with id %s does not belong to account %s", list.ID, accountID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	return list, nil
}

// Get returns the api model of one list with the given ID.
func (p *Processor) Get(ctx context.Context, account *gtsmodel.Account, id string) (*apimodel.List, gtserror.WithCode) {
	list, errWithCode := p.getList(
		// Use barebones ctx; no embedded
		// structs necessary for this call.
		gtscontext.SetBarebones(ctx),
		account.ID,
		id,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return p.apiList(ctx, list)
}

// GetMultiple returns multiple lists created by the given account, sorted by list ID DESC (newest first).
func (p *Processor) GetMultiple(ctx context.Context, account *gtsmodel.Account, id string) ([]*apimodel.List, gtserror.WithCode) {
	lists, err := p.state.DB.GetListsForAccountID(
		// Use barebones ctx; no embedded
		// structs necessary for simple GET.
		gtscontext.SetBarebones(ctx),
		account.ID,
	)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			return nil, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiLists := make([]*apimodel.List, 0, len(lists))
	for _, list := range lists {
		apiList, errWithCode := p.apiList(ctx, list)
		if errWithCode != nil {
			return nil, errWithCode
		}

		apiLists = append(apiLists, apiList)
	}

	return apiLists, nil
}

func (p *Processor) GetListAccounts(ctx context.Context, account *gtsmodel.Account, id string, maxID string, sinceID string, minID string, limit int) (*apimodel.PageableResponse, gtserror.WithCode) {
	if _, errWithCode := p.getList(ctx, account.ID, id); errWithCode != nil {
		return nil, errWithCode
	}

	listEntries, err := p.state.DB.GetListEntries(ctx, id, maxID, sinceID, minID, limit)
	if err != nil {
		err = fmt.Errorf("GetListAccounts: error getting list entries: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(listEntries)
	if count == 0 {
		return util.EmptyPageableResponse(), nil
	}

	var (
		items          = make([]interface{}, count)
		nextMaxIDValue string
		prevMinIDValue string
	)

	for i, listEntry := range listEntries {
		if i == count-1 {
			nextMaxIDValue = listEntry.ID
		}

		if i == 0 {
			prevMinIDValue = listEntry.ID
		}

		if err := p.state.DB.PopulateListEntry(ctx, listEntry); err != nil {
			log.Debugf(ctx, "skipping list entry because of error populating it: %q", err)
			continue
		}

		if err := p.state.DB.PopulateFollow(ctx, listEntry.Follow); err != nil {
			log.Debugf(ctx, "skipping list entry because of error populating follow: %q", err)
			continue
		}

		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, listEntry.Follow.TargetAccount)
		if err != nil {
			log.Debugf(ctx, "skipping list entry because of error converting follow target account: %q", err)
			continue
		}

		items[i] = apiAccount
	}

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "api/v1/lists/" + id + "/accounts",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
