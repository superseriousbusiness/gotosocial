/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

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

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func GetParseMentionFunc(dbConn db.DB, federator federation.Federator) gtsmodel.ParseMentionFunc {
	return func(ctx context.Context, targetAccount string, originAccountID string, statusID string) (*gtsmodel.Mention, error) {
		// get the origin account first since we'll need it to create the mention
		originAccount, err := dbConn.GetAccountByID(ctx, originAccountID)
		if err != nil {
			return nil, fmt.Errorf("couldn't get mention origin account with id %s", originAccountID)
		}

		username, domain, err := util.ExtractNamestringParts(targetAccount)
		if err != nil {
			return nil, fmt.Errorf("couldn't extract namestring parts from %s: %s", targetAccount, err)
		}

		var mentionedAccount *gtsmodel.Account
		if domain == "" || domain == config.GetHost() || domain == config.GetAccountDomain() {
			localAccount, err := dbConn.GetAccountByUsernameDomain(ctx, username, "")
			if err != nil {
				return nil, err
			}
			mentionedAccount = localAccount
		} else {
			var requestingUsername string
			if originAccount.Domain == "" {
				requestingUsername = originAccount.Username
			}
			remoteAccount, err := federator.GetAccount(transport.WithFastfail(ctx), dereferencing.GetAccountParams{
				RequestingUsername:    requestingUsername,
				RemoteAccountUsername: username,
				RemoteAccountHost:     domain,
			})
			if err != nil {
				return nil, fmt.Errorf("error dereferencing account: %s", err)
			}

			// we were able to resolve it!
			mentionedAccount = remoteAccount

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
