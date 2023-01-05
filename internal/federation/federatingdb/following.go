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

package federatingdb

import (
	"context"
	"fmt"
	"net/url"

	"codeberg.org/gruf/go-kv"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

// Following obtains the Following Collection for an actor with the
// given id.
//
// If modified, the library will then call Update.
//
// The library makes this call only after acquiring a lock first.
func (f *federatingDB) Following(ctx context.Context, actorIRI *url.URL) (following vocab.ActivityStreamsCollection, err error) {
	l := log.WithFields(kv.Fields{
		{"id", actorIRI},
	}...)
	l.Debug("entering Following")

	acct, err := f.getAccountForIRI(ctx, actorIRI)
	if err != nil {
		return nil, err
	}

	acctFollowing, err := f.db.GetAccountFollows(ctx, acct.ID)
	if err != nil {
		return nil, fmt.Errorf("Following: db error getting following for account id %s: %s", acct.ID, err)
	}

	iris := []*url.URL{}
	for _, follow := range acctFollowing {
		if follow.TargetAccount == nil {
			a, err := f.db.GetAccountByID(ctx, follow.TargetAccountID)
			if err != nil {
				errWrapped := fmt.Errorf("Following: db error getting account id %s: %s", follow.TargetAccountID, err)
				if err == db.ErrNoEntries {
					// no entry for this account id so it's probably been deleted and we haven't caught up yet
					l.Error(errWrapped)
					continue
				} else {
					// proper error
					return nil, errWrapped
				}
			}
			follow.TargetAccount = a
		}
		u, err := url.Parse(follow.TargetAccount.URI)
		if err != nil {
			return nil, err
		}
		iris = append(iris, u)
	}

	return f.collectIRIs(ctx, iris)
}
