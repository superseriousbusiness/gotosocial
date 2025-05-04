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

package processing

import (
	"context"
	"errors"
	"fmt"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/federation"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// GetParseMentionFunc returns a new ParseMentionFunc using the provided state and federator.
// State is used for doing local database lookups; federator is used for remote account lookups (if necessary).
func GetParseMentionFunc(state *state.State, federator *federation.Federator) gtsmodel.ParseMentionFunc {
	return func(ctx context.Context, namestring string, originAccountID string, statusID string) (*gtsmodel.Mention, error) {
		// Get the origin account first since
		// we'll need it to create the mention.
		originAcct, err := state.DB.GetAccountByID(ctx, originAccountID)
		if err != nil {
			return nil, fmt.Errorf(
				"db error getting mention origin account %s: %w",
				originAccountID, err,
			)
		}

		// Parse target components from the
		// "@someone@example.org" namestring.
		targetUsername, targetHost, err := util.ExtractNamestringParts(namestring)
		if err != nil {
			return nil, fmt.Errorf(
				"error extracting mention target: %w",
				err,
			)
		}

		// It's a "local" mention if namestring
		// looks like one of the following:
		//
		//   - "@someone" with no host component.
		//   - "@someone@gts.example.org" and we're host "gts.example.org".
		//   - "@someone@example.org" and we're account-domain "example.org".
		local := targetHost == "" ||
			targetHost == config.GetHost() ||
			targetHost == config.GetAccountDomain()

		// Either a local or remote
		// target for the mention.
		var targetAcct *gtsmodel.Account
		if local {
			// Lookup local target accounts in the db only.
			targetAcct, err = state.DB.GetAccountByUsernameDomain(ctx, targetUsername, "")
			if err != nil {
				return nil, fmt.Errorf(
					"db error getting mention local target account %s: %w",
					targetUsername, err,
				)
			}
		} else {
			// If origin account is local, use
			// it to do potential dereference.
			// Else fallback to empty string,
			// which uses instance account.
			var requestUser string
			if originAcct.IsLocal() {
				requestUser = originAcct.Username
			}

			targetAcct, _, err = federator.GetAccountByUsernameDomain(
				gtscontext.SetFastFail(ctx),
				requestUser,
				targetUsername,
				targetHost,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"error fetching mention remote target account: %w",
					err,
				)
			}
		}

		// Check if the mention was
		// in the database already.
		if statusID != "" {
			mention, err := state.DB.GetMentionByTargetAcctStatus(ctx, targetAcct.ID, statusID)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return nil, fmt.Errorf(
					"db error checking for existing mention: %w",
					err,
				)
			}

			if mention != nil {
				// We had it, return this rather
				// than creating a new one.
				mention.NameString = namestring
				mention.OriginAccountURI = originAcct.URI
				mention.TargetAccountURI = targetAcct.URI
				mention.TargetAccountURL = targetAcct.URL
				return mention, nil
			}
		}

		// Return new mention with useful populated fields,
		// but *don't* store it in the database; that's
		// up to the calling function to do, if they want.
		return &gtsmodel.Mention{
			ID:               id.NewULID(),
			StatusID:         statusID,
			OriginAccountID:  originAcct.ID,
			OriginAccountURI: originAcct.URI,
			OriginAccount:    originAcct,
			TargetAccountID:  targetAcct.ID,
			TargetAccountURI: targetAcct.URI,
			TargetAccountURL: targetAcct.URL,
			TargetAccount:    targetAcct,
			NameString:       namestring,

			// Mention wasn't
			// stored in the db.
			IsNew: true,
		}, nil
	}
}
