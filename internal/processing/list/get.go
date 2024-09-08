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

// GetAllListAccounts returns all accounts that are in the given list,
// owned by the given account. There's no pagination for this endpoint.
//
// See https://docs.joinmastodon.org/methods/lists/#query-parameters:
//
//	Limit: Integer. Maximum number of results. Defaults to 40 accounts.
//	Max 80 accounts. Set to 0 in order to get all accounts without pagination.
func (p *Processor) GetAllListAccounts(
	ctx context.Context,
	account *gtsmodel.Account,
	listID string,
) ([]*apimodel.Account, gtserror.WithCode) {
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

	// Get all entries for this list.
	listEntries, err := p.state.DB.GetListEntries(ctx, listID, "", "", "", 0)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("error getting list entries: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Extract accounts from list entries + add them to response.
	accounts := make([]*apimodel.Account, 0, len(listEntries))
	p.accountsFromListEntries(ctx, listEntries, func(acc *apimodel.Account) {
		accounts = append(accounts, acc)
	})

	return accounts, nil
}

// GetListAccounts returns accounts that are in the given list, owned by the given account.
// The additional parameters can be used for paging.
func (p *Processor) GetListAccounts(
	ctx context.Context,
	account *gtsmodel.Account,
	listID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
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

	// To know which accounts are in the list,
	// we need to first get requested list entries.
	listEntries, err := p.state.DB.GetListEntries(ctx, listID, maxID, sinceID, minID, limit)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("GetListAccounts: error getting list entries: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	count := len(listEntries)
	if count == 0 {
		// No list entries means no accounts.
		return util.EmptyPageableResponse(), nil
	}

	var (
		items = make([]interface{}, 0, count)

		// Set next + prev values before filtering and API
		// converting, so caller can still page properly.
		nextMaxIDValue = listEntries[count-1].ID
		prevMinIDValue = listEntries[0].ID
	)

	// Extract accounts from list entries + add them to response.
	p.accountsFromListEntries(ctx, listEntries, func(acc *apimodel.Account) {
		items = append(items, acc)
	})

	return util.PackagePageableResponse(util.PageableResponseParams{
		Items:          items,
		Path:           "/api/v1/lists/" + listID + "/accounts",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}

func (p *Processor) accountsFromListEntries(
	ctx context.Context,
	listEntries []*gtsmodel.ListEntry,
	appendAcc func(*apimodel.Account),
) {
	// For each list entry, we want the account it points to.
	// To get this, we need to first get the follow that the
	// list entry pertains to, then extract the target account
	// from that follow.
	//
	// We do paging not by account ID, but by list entry ID.
	for _, listEntry := range listEntries {
		if err := p.state.DB.PopulateListEntry(ctx, listEntry); err != nil {
			log.Errorf(ctx, "error populating list entry: %v", err)
			continue
		}

		if err := p.state.DB.PopulateFollow(ctx, listEntry.Follow); err != nil {
			log.Errorf(ctx, "error populating follow: %v", err)
			continue
		}

		apiAccount, err := p.converter.AccountToAPIAccountPublic(ctx, listEntry.Follow.TargetAccount)
		if err != nil {
			log.Errorf(ctx, "error converting to public api account: %v", err)
			continue
		}

		appendAcc(apiAccount)
	}
}
