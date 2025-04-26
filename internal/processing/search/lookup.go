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
	"fmt"
	"strings"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"codeberg.org/gruf/go-kv"
)

// Lookup does a quick, non-resolving search for accounts that
// match the given query. It expects input that looks like a
// namestring, and will normalize plaintext to look more like
// a namestring. Will only ever return one account, and only on
// an exact match.
//
// This behavior aligns more or less with Mastodon's API.
// See https://docs.joinmastodon.org/methods/accounts/#lookup
func (p *Processor) Lookup(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	query string,
) (*apimodel.Account, gtserror.WithCode) {
	// Include instance accounts in this search.
	//
	// Lookup is for one specific account so we
	// can't return loads of instance accounts by
	// accident.
	const includeInstanceAccounts = true

	// Since lookup is always for a specific
	// account, it's fine to include a blocked
	// account in the results.
	const includeBlockedAccounts = true

	// Validate query.
	query = strings.TrimSpace(query)
	if query == "" {
		err := errors.New("search query was empty string after trimming space")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// Be nice and normalize query by prepending '@'.
	// This will make it easier for accountsByNamestring
	// to pick this up as a valid namestring.
	if query[0] != '@' {
		query = "@" + query
	}

	log.
		WithContext(ctx).
		WithFields(kv.Fields{
			{"query", query},
		}...).
		Debugf("beginning search")

	// See if we have something that looks like a namestring.
	username, domain, err := util.ExtractNamestringParts(query)
	if err != nil {
		err := errors.New("bad search query, must in the form '[username]' or '[username]@[domain]")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	account, err := p.accountByUsernameDomain(
		ctx,
		requestingAccount,
		username,
		domain,
		false, // never resolve!
	)
	if err != nil {
		if gtserror.IsUnretrievable(err) {
			// ErrNotRetrievable is fine, just wrap it in
			// a 404 to indicate we couldn't find anything.
			err := fmt.Errorf("%s not found", query)
			return nil, gtserror.NewErrorNotFound(err, err.Error())
		}

		// Real error has occurred.
		err = gtserror.Newf("error looking up %s as account: %w", query, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// If we reach this point, we found an account. Shortcut
	// using the packageAccounts function to return it. This
	// may cause the account to be filtered out if it's not
	// visible to the caller, so anticipate this.
	accounts, errWithCode := p.packageAccounts(
		ctx,
		requestingAccount,
		[]*gtsmodel.Account{account},
		includeInstanceAccounts,
		includeBlockedAccounts,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if len(accounts) == 0 {
		// Account was not visible to the requesting account.
		err := fmt.Errorf("%s not found", query)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// We got a hit!
	return accounts[0], nil
}
