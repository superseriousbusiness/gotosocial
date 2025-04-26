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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

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

// GetAll returns multiple lists created by the given account, sorted by list ID DESC (newest first).
func (p *Processor) GetAll(ctx context.Context, account *gtsmodel.Account) ([]*apimodel.List, gtserror.WithCode) {
	lists, err := p.state.DB.GetListsByAccountID(

		// Use barebones ctx; no embedded
		// structs necessary for simple GET.
		gtscontext.SetBarebones(ctx),
		account.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
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

// GetListAccounts returns accounts that are in the given list, owned by the given account.
// The additional parameters can be used for paging. Nil page param returns all accounts.
func (p *Processor) GetListAccounts(
	ctx context.Context,
	account *gtsmodel.Account,
	listID string,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Ensure list exists + is owned by requesting account.
	_, errWithCode := p.getList(

		// Use barebones ctx; no embedded
		// structs necessary for this call.
		gtscontext.SetBarebones(ctx),
		account.ID,
		listID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Get all accounts contained within list.
	accounts, err := p.state.DB.GetAccountsInList(ctx,
		listID,
		page,
	)
	if err != nil {
		err := gtserror.Newf("db error getting accounts in list: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for any accounts.
	count := len(accounts)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	var (
		// Preallocate expected frontend items.
		items = make([]interface{}, 0, count)

		// Set paging low / high IDs.
		lo = accounts[count-1].ID
		hi = accounts[0].ID
	)

	// Convert accounts to frontend.
	for _, account := range accounts {
		apiAccount, err := p.converter.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			log.Errorf(ctx, "error converting to api account: %v", err)
			continue
		}
		items = append(items, apiAccount)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/lists/" + listID + "/accounts",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}
