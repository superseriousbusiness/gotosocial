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
	"errors"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-kv"
)

// Accounts does a partial search for accounts that
// match the given query. It expects input that looks
// like a namestring, and will normalize plaintext to look
// more like a namestring. For queries that include domain,
// it will only return one match at most. For namestrings
// that exclude domain, multiple matches may be returned.
//
// This behavior aligns more or less with Mastodon's API.
// See https://docs.joinmastodon.org/methods/accounts/#search.
func (p *Processor) Accounts(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
	limit int,
	offset int,
	resolve bool,
	following bool,
) ([]*apimodel.Account, gtserror.WithCode) {
	// Don't include instance accounts in this search.
	//
	// We don't want someone to start typing '@mastodon'
	// and then get a million instance service accounts
	// in their search results.
	const includeInstanceAccounts = false

	// We *might* want to include blocked accounts
	// in this search, but only if it's a search
	// for a specific account.
	includeBlockedAccounts := false

	var (
		foundAccounts = make([]*gtsmodel.Account, 0, limit)
		appendAccount = func(foundAccount *gtsmodel.Account) { foundAccounts = append(foundAccounts, foundAccount) }
	)

	// Validate query.
	query = strings.TrimSpace(query)
	if query == "" {
		err := gtserror.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"limit", limit},
			{"offset", offset},
			{"query", query},
			{"resolve", resolve},
			{"following", following},
		}...).
		Debugf("beginning search")

	// todo: Currently we don't support offset for paging;
	// if caller supplied an offset greater than 0, return
	// nothing as though there were no additional results.
	if offset > 0 {
		return p.packageAccounts(
			ctx,
			requestingAccount,
			foundAccounts,
			includeInstanceAccounts,
			includeBlockedAccounts,
		)
	}

	// See if we have something that looks like a namestring.
	username, domain, err := util.ExtractNamestringParts(query)
	if err == nil {
		if domain != "" {
			// Search was an exact namestring;
			// we can safely assume caller is
			// searching for a specific account,
			// and show it to them even if they
			// have it blocked.
			includeBlockedAccounts = true
		}

		// Get all accounts we can find
		// that match the provided query.
		if err := p.accountsByUsernameDomain(
			ctx,
			requestingAccount,
			id.Highest,
			id.Lowest,
			limit,
			offset,
			username,
			domain,
			resolve,
			following,
			appendAccount,
		); err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("error searching by namestring: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	} else {
		// Query Doesn't look like a
		// namestring, use text search.
		if err := p.accountsByText(
			ctx,
			requestingAccount.ID,
			id.Highest,
			id.Lowest,
			limit,
			offset,
			query,
			following,
			appendAccount,
		); err != nil && !errors.Is(err, db.ErrNoEntries) {
			err = gtserror.Newf("error searching by text: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	// Return whatever we got (if anything).
	return p.packageAccounts(
		ctx,
		requestingAccount,
		foundAccounts,
		includeInstanceAccounts,
		includeBlockedAccounts,
	)
}
