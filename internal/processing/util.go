/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package processing

import (
	"context"
	"fmt"
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

func GetParseMentionFunc(dbConn db.DB, federator federation.Federator) gtsmodel.ParseMentionFunc {
	return func(ctx context.Context, targetAccount string, originAccountID string, statusID string) (*gtsmodel.Mention, error) {
		// get the origin account first since we'll need it to create the mention
		originAccount, err := dbConn.GetAccountByID(ctx, originAccountID)
		if err != nil {
			return nil, fmt.Errorf("couldn't get mention origin account with id %s", originAccountID)
		}

		// A mentioned account looks like "@test@example.org" or just "@test" for a local account
		// -- we can guarantee this from the regex that targetAccounts should have been derived from.
		// But we still need to do a bit of fiddling to get what we need here -- the username and domain (if given).

		// 1.  trim off the first @
		trimmed := strings.TrimPrefix(targetAccount, "@")

		// 2. split the username and domain
		split := strings.Split(trimmed, "@")

		// 3. if it's length 1 it's a local account, length 2 means remote, anything else means something is wrong

		var local bool
		switch len(split) {
		case 1:
			local = true
		case 2:
			local = false
		default:
			return nil, fmt.Errorf("mentioned account format '%s' was not valid", targetAccount)
		}

		var username, domain string
		username = split[0]
		if !local {
			domain = split[1]
		}

		// 4. check we now have a proper username and domain
		if username == "" || (!local && domain == "") {
			return nil, fmt.Errorf("username or domain for '%s' was nil", targetAccount)
		}

		var mentionedAccount *gtsmodel.Account

		if local {
			localAccount, err := dbConn.GetLocalAccountByUsername(ctx, username)
			if err != nil {
				return nil, err
			}
			mentionedAccount = localAccount
		} else {
			remoteAccount := &gtsmodel.Account{}

			where := []db.Where{
				{
					Key:             "username",
					Value:           username,
					CaseInsensitive: true,
				},
				{
					Key:             "domain",
					Value:           domain,
					CaseInsensitive: true,
				},
			}

			err := dbConn.GetWhere(ctx, where, remoteAccount)
			if err == nil {
				// the account was already in the database
				mentionedAccount = remoteAccount
			} else {
				// we couldn't get it from the database
				if err != db.ErrNoEntries {
					// a serious error has happened so bail
					return nil, fmt.Errorf("error getting account with username '%s' and domain '%s': %s", username, domain, err)
				}

				// We just don't have the account, so try webfingering it.
				//
				// If the mention originates from our instance we should use the username of the origin account to do the dereferencing,
				// otherwise we should just use our instance account (that is, provide an empty string), since obviously we can't use
				// a remote account to do remote dereferencing!
				var fingeringUsername string
				if originAccount.Domain == "" {
					fingeringUsername = originAccount.Username
				}

				acctURI, err := federator.FingerRemoteAccount(ctx, fingeringUsername, username, domain)
				if err != nil {
					// something went wrong doing the webfinger lookup so we can't process the request
					return nil, fmt.Errorf("error fingering remote account with username %s and domain %s: %s", username, domain, err)
				}

				resolvedAccount, err := federator.GetRemoteAccount(ctx, fingeringUsername, acctURI, true, true)
				if err != nil {
					return nil, fmt.Errorf("error dereferencing account with uri %s: %s", acctURI.String(), err)
				}

				// we were able to resolve it!
				mentionedAccount = resolvedAccount
			}
		}

		mentionID, err := id.NewRandomULID()
		if err != nil {
			return nil, err
		}

		return &gtsmodel.Mention{
			ID:               mentionID,
			StatusID:         statusID,
			OriginAccountID:  originAccount.ID,
			OriginAccountURI: originAccount.URI,
			TargetAccountID:  mentionedAccount.ID,
			NameString:       targetAccount,
			TargetAccountURI: mentionedAccount.URI,
			TargetAccountURL: mentionedAccount.URL,
			OriginAccount:    mentionedAccount,
		}, nil
	}
}
