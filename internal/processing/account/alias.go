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
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

const (
	aliasTargetSuspended = "target account %s is suspended from this instance; you will not be able to set alsoKnownAs to that account; if you believe this is an error, please contact your instance admin"
	aliasBadURI          = "invalid also_known_as_uri (%s) provided in account alias request: %w"
)

func (p *Processor) Alias(
	ctx context.Context,
	account *gtsmodel.Account,
	newAKAURIs []string,
) (*apimodel.Account, gtserror.WithCode) {
	var (
		currentLen = len(account.AlsoKnownAsURIs)
		newLen     = len(newAKAURIs)
		// Update by making a
		// call to the database.
		update bool
	)

	switch {
	case newLen == 0 && currentLen == 0:
		// Empty also_known_as_uris,
		// and no aliases set anyway,
		// don't bother with a db call.

	case newLen == 0:
		// Unset existing aliases.
		account.AlsoKnownAsURIs = nil
		account.AlsoKnownAs = nil
		update = true

	default:
		// Prepare to update AlsoKnownAs entries on the account.
		account.AlsoKnownAs = make([]*gtsmodel.Account, 0, newLen)
		account.AlsoKnownAsURIs = make([]string, 0, newLen)
		update = true

		// Use a map to deduplicate desired URIs.
		newAKAsMap := make(map[string]struct{}, len(newAKAURIs))

		// For each entry, ensure it's a valid URI,
		// get and check the target account, and set.
		for _, rawURI := range newAKAURIs {
			alsoKnownAsURI, err := url.Parse(rawURI)
			if err != nil {
				err := gtserror.Newf(aliasBadURI, rawURI, err)
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			// We only deref http or https, so check this.
			if alsoKnownAsURI == nil || (alsoKnownAsURI.Scheme != "https" && alsoKnownAsURI.Scheme != "http") {
				err := gtserror.Newf(aliasBadURI, rawURI, errors.New("uri must not be empty and scheme must be http or https"))
				return nil, gtserror.NewErrorBadRequest(err, err.Error())
			}

			// Use parsed version of this URI from
			// now on to normalize casing etc.
			alsoKnownAsURIStr := alsoKnownAsURI.String()

			// Avoid processing duplicate entries.
			if _, ok := newAKAsMap[alsoKnownAsURIStr]; !ok {
				newAKAsMap[alsoKnownAsURIStr] = struct{}{}
			} else {
				// Already done
				// this one.
				continue
			}

			// Don't let account do anything
			// daft by aliasing to itself.
			if alsoKnownAsURIStr == account.URI {
				continue
			}

			// Ensure we have a valid, up-to-date
			// representation of the target account.
			targetAccount, _, err := p.federator.GetAccountByURI(ctx, account.Username, alsoKnownAsURI)
			if err != nil {
				err := gtserror.Newf("error dereferencing also_known_as_uri (%s) account: %w", rawURI, err)
				return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
			}

			// Alias target must not be suspended.
			if !targetAccount.SuspendedAt.IsZero() {
				err := gtserror.Newf(aliasTargetSuspended, alsoKnownAsURIStr)
				return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
			}

			// Alrighty-roo, looks good, add this one.
			account.AlsoKnownAsURIs = append(account.AlsoKnownAsURIs, alsoKnownAsURIStr)
			account.AlsoKnownAs = append(account.AlsoKnownAs, targetAccount)
		}
	}

	if update {
		err := p.state.DB.UpdateAccount(ctx, account, "also_known_as_uris")
		if err != nil {
			err := gtserror.Newf("db error updating also_known_as_uri: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	acctSensitive, err := p.converter.AccountToAPIAccountSensitive(ctx, account)
	if err != nil {
		err := gtserror.Newf("error converting account: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return acctSensitive, nil
}
