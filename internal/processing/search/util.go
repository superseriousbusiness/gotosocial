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

package search

import (
	"context"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// return true if given queryType should include accounts.
func includeAccounts(queryType string) bool {
	return queryType == queryTypeAny || queryType == queryTypeAccounts
}

// return true if given queryType should include statuses.
func includeStatuses(queryType string) bool {
	return queryType == queryTypeAny || queryType == queryTypeStatuses
}

// packageAccounts is a util function that just
// converts the given accounts into an apimodel
// account slice, or errors appropriately.
func (p *Processor) packageAccounts(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	accounts []*gtsmodel.Account,
) ([]*apimodel.Account, gtserror.WithCode) {
	apiAccounts := make([]*apimodel.Account, 0, len(accounts))

	for _, account := range accounts {
		// Ensure requester can see result account.
		visible, err := p.filter.AccountVisible(ctx, requestingAccount, account)
		if err != nil {
			err = gtserror.Newf("error checking visibility of account %s for account %s: %w", account.ID, requestingAccount.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !visible {
			log.Debugf(ctx, "account %s is not visible to account %s, skipping this result", account.ID, requestingAccount.ID)
			continue
		}

		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			log.Debugf(ctx, "skipping account %s because it couldn't be converted to its api representation: %s", account.ID, err)
			continue
		}

		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}

// packageStatuses is a util function that just
// converts the given statuses into an apimodel
// status slice, or errors appropriately.
func (p *Processor) packageStatuses(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	statuses []*gtsmodel.Status,
) ([]*apimodel.Status, gtserror.WithCode) {
	apiStatuses := make([]*apimodel.Status, 0, len(statuses))

	for _, status := range statuses {
		// Ensure requester can see result status.
		visible, err := p.filter.StatusVisible(ctx, requestingAccount, status)
		if err != nil {
			err = gtserror.Newf("error checking visibility of status %s for account %s: %w", status.ID, requestingAccount.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if !visible {
			log.Debugf(ctx, "status %s is not visible to account %s, skipping this result", status.ID, requestingAccount.ID)
			continue
		}

		apiStatus, err := p.tc.StatusToAPIStatus(ctx, status, requestingAccount)
		if err != nil {
			log.Debugf(ctx, "skipping status %s because it couldn't be converted to its api representation: %s", status.ID, err)
			continue
		}

		apiStatuses = append(apiStatuses, apiStatus)
	}

	return apiStatuses, nil
}

// packageSearchResult wraps up the given accounts
// and statuses into an apimodel SearchResult that
// can be serialized to an API caller as JSON.
func (p *Processor) packageSearchResult(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	accounts []*gtsmodel.Account,
	statuses []*gtsmodel.Status,
) (*apimodel.SearchResult, gtserror.WithCode) {
	apiAccounts, errWithCode := p.packageAccounts(ctx, requestingAccount, accounts)
	if errWithCode != nil {
		return nil, errWithCode
	}

	apiStatuses, errWithCode := p.packageStatuses(ctx, requestingAccount, statuses)
	if errWithCode != nil {
		return nil, errWithCode
	}

	return &apimodel.SearchResult{
		Accounts: apiAccounts,
		Statuses: apiStatuses,
		Hashtags: make([]*apimodel.Tag, 0),
	}, nil
}
