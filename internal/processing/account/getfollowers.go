/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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

package account

import (
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) FollowersGet(requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	blocked, err := p.db.Blocked(requestingAccount.ID, targetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts"))
	}

	followers := []gtsmodel.Follow{}
	accounts := []apimodel.Account{}
	if err := p.db.GetFollowersByAccountID(targetAccountID, &followers, false); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return accounts, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, f := range followers {
		blocked, err := p.db.Blocked(requestingAccount.ID, f.AccountID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		if blocked {
			continue
		}

		a := &gtsmodel.Account{}
		if err := p.db.GetByID(f.AccountID, a); err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				continue
			}
			return nil, gtserror.NewErrorInternalError(err)
		}

		// derefence account fields in case we haven't done it already
		if err := p.federator.DereferenceAccountFields(a, requestingAccount.Username, false); err != nil {
			// don't bail if we can't fetch them, we'll try another time
			p.log.WithField("func", "AccountFollowersGet").Debugf("error dereferencing account fields: %s", err)
		}

		account, err := p.tc.AccountToMastoPublic(a)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		accounts = append(accounts, *account)
	}
	return accounts, nil
}
