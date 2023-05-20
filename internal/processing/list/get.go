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

// GetMultiple returns multiple lists created by the given account, sorted by list ID DESC (newest first).
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
	if _, errWithCode := p.getList(ctx, account.ID, listID); errWithCode != nil {
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
		items          = make([]interface{}, count)
		nextMaxIDValue string
		prevMinIDValue string
	)

	// For each list entry, we want the account it points to.
	// To get this, we need to first get the follow that the
	// list entry pertains to, then extract the target account
	// from that follow.
	//
	// We do paging not by account ID, but by list entry ID.
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
		Path:           "api/v1/lists/" + listID + "/accounts",
		NextMaxIDValue: nextMaxIDValue,
		PrevMinIDValue: prevMinIDValue,
		Limit:          limit,
	})
}
