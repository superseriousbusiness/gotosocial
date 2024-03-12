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

package workers

import (
	"context"
	"errors"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/processing/account"
	"github.com/superseriousbusiness/gotosocial/internal/state"
)

// redirectFollowers redirects all local
// followers of originAcct to targetAcct.
//
// Both accounts must be fully dereferenced
// already, and the Move must be valid.
//
// Return bool will be true if all goes OK.
type redirectFollowers func(
	ctx context.Context,
	originAcct *gtsmodel.Account,
	targetAcct *gtsmodel.Account,
) bool

// redirectFollowersF returns a redirectFollowers util function.
func redirectFollowersF(state *state.State, account *account.Processor) redirectFollowers {
	return func(
		ctx context.Context,
		originAcct *gtsmodel.Account,
		targetAcct *gtsmodel.Account,
	) bool {
		// Any local followers of originAcct should
		// send follow requests to targetAcct instead,
		// and have followers of originAcct removed.
		//
		// Select local followers with barebones, since
		// we only need follow.Account and we can get
		// that ourselves.
		followers, err := state.DB.GetAccountLocalFollowers(
			gtscontext.SetBarebones(ctx),
			originAcct.ID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			log.Errorf(ctx,
				"db error getting follows targeting originAcct: %v",
				err,
			)
			return false
		}

		for _, follow := range followers {
			// Fetch the local account that
			// owns the follow targeting originAcct.
			if follow.Account, err = state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				follow.AccountID,
			); err != nil {
				log.Errorf(ctx,
					"db error getting follow account %s: %v",
					follow.AccountID, err,
				)
				return false
			}

			// Use the account processor FollowCreate
			// function to send off the new follow,
			// carrying over the Reblogs and Notify
			// values from the old follow to the new.
			//
			// This will also handle cases where our
			// account has already followed the target
			// account, by just updating the existing
			// follow of target account.
			if _, err := account.FollowCreate(
				ctx,
				follow.Account,
				&apimodel.AccountFollowRequest{
					ID:      targetAcct.ID,
					Reblogs: follow.ShowReblogs,
					Notify:  follow.Notify,
				},
			); err != nil {
				log.Errorf(ctx,
					"error creating new follow for account %s: %v",
					follow.AccountID, err,
				)
				return false
			}

			// New follow is in the process of
			// sending, remove the existing follow.
			// This will send out an Undo Activity for each Follow.
			if _, err := account.FollowRemove(
				ctx,
				follow.Account,
				follow.TargetAccountID,
			); err != nil {
				log.Errorf(ctx,
					"error removing old follow for account %s: %v",
					follow.AccountID, err,
				)
				return false
			}
		}

		return true
	}
}
