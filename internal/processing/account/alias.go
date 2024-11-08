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
	"net/url"
	"slices"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
)

func (p *Processor) Alias(
	ctx context.Context,
	account *gtsmodel.Account,
	newAKAURIStrs []string,
) (*apimodel.Account, gtserror.WithCode) {
	if slices.Equal(
		newAKAURIStrs,
		account.AlsoKnownAsURIs,
	) {
		// No changes to do
		// here. Return early.
		return p.c.GetAPIAccountSensitive(ctx, account)
	}

	newLen := len(newAKAURIStrs)
	if newLen == 0 {
		// Simply unset existing
		// aliases and return early.
		account.AlsoKnownAsURIs = nil
		account.AlsoKnownAs = nil

		err := p.state.DB.UpdateAccount(ctx, account, "also_known_as_uris")
		if err != nil {
			err := gtserror.Newf("db error updating also_known_as_uri: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		return p.c.GetAPIAccountSensitive(ctx, account)
	}

	// We need to set new AKA URIs!
	//
	// First parse them to URI ptrs and
	// normalized string representations.
	//
	// Use this cheeky type to avoid
	// repeatedly calling uri.String().
	type uri struct {
		uri *url.URL // Parsed URI.
		str string   // uri.String().
	}

	newAKAs := make([]uri, newLen)
	for i, newAKAURIStr := range newAKAURIStrs {
		newAKAURI, err := url.Parse(newAKAURIStr)
		if err != nil {
			err := fmt.Errorf(
				"invalid also_known_as_uri (%s) provided in account alias request: %w",
				newAKAURIStr, err,
			)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		// We only deref http or https, so check this.
		if newAKAURI.Scheme != "https" && newAKAURI.Scheme != "http" {
			err := fmt.Errorf(
				"invalid also_known_as_uri (%s) provided in account alias request: %w",
				newAKAURIStr, errors.New("uri must not be empty and scheme must be http or https"),
			)
			return nil, gtserror.NewErrorBadRequest(err, err.Error())
		}

		newAKAs[i].uri = newAKAURI
		newAKAs[i].str = newAKAURI.String()
	}

	// For each deduped entry, get and
	// check the target account, and set.
	for _, newAKA := range newAKAs {
		// Don't let account do anything
		// daft by aliasing to itself.
		if newAKA.str == account.URI ||
			newAKA.str == account.URL {
			continue
		}

		// Ensure we have account dereferenced.
		targetAccount, _, err := p.federator.GetAccountByURI(ctx,
			account.Username,
			newAKA.uri,
		)
		if err != nil {
			err := fmt.Errorf(
				"error dereferencing also_known_as_uri (%s) account: %w",
				newAKA.str, err,
			)
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Target must not be suspended.
		if !targetAccount.SuspendedAt.IsZero() {
			err := fmt.Errorf(
				"target account %s is suspended from this instance; "+
					"you will not be able to set alsoKnownAs to that account",
				newAKA.str,
			)
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// Alrighty-roo, looks good, add this one.
		account.AlsoKnownAsURIs = append(account.AlsoKnownAsURIs, targetAccount.URI)
		account.AlsoKnownAs = append(account.AlsoKnownAs, targetAccount)
	}

	// Dedupe URIs + accounts, in case someone
	// provided both an account URL and an
	// account URI above, for the same account.
	account.AlsoKnownAsURIs = xslices.Deduplicate(account.AlsoKnownAsURIs)
	account.AlsoKnownAs = xslices.DeduplicateFunc(
		account.AlsoKnownAs,
		func(a *gtsmodel.Account) string {
			return a.URI
		},
	)

	err := p.state.DB.UpdateAccount(ctx, account, "also_known_as_uris")
	if err != nil {
		err := gtserror.Newf("db error updating also_known_as_uri: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.c.GetAPIAccountSensitive(ctx, account)
}
