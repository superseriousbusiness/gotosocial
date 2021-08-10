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
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Get(requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Account, error) {
	targetAccount := &gtsmodel.Account{}
	if err := p.db.GetByID(targetAccountID, targetAccount); err != nil {
		if _, ok := err.(db.ErrNoEntries); ok {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("db error: %s", err)
	}

	var blocked bool
	var err error
	if requestingAccount != nil {
		blocked, err = p.db.Blocked(requestingAccount.ID, targetAccountID)
		if err != nil {
			return nil, fmt.Errorf("error checking account block: %s", err)
		}
	}

	var mastoAccount *apimodel.Account
	if blocked {
		mastoAccount, err = p.tc.AccountToMastoBlocked(targetAccount)
		if err != nil {
			return nil, fmt.Errorf("error converting account: %s", err)
		}
		return mastoAccount, nil
	}

	// last-minute check to make sure we have remote account header/avi cached
	if targetAccount.Domain != "" {
		a, err := p.federator.EnrichRemoteAccount(requestingAccount.Username, targetAccount)
		if err == nil {
			targetAccount = a
		}
	}

	if requestingAccount != nil && targetAccount.ID == requestingAccount.ID {
		mastoAccount, err = p.tc.AccountToMastoSensitive(targetAccount)
	} else {
		mastoAccount, err = p.tc.AccountToMastoPublic(targetAccount)
	}
	if err != nil {
		return nil, fmt.Errorf("error converting account: %s", err)
	}
	return mastoAccount, nil
}
