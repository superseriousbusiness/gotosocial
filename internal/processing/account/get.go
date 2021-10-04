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
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Account, error) {
	targetAccount, err := p.db.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, errors.New("account not found")
		}
		return nil, fmt.Errorf("db error: %s", err)
	}

	var blocked bool
	if requestingAccount != nil {
		blocked, err = p.db.IsBlocked(ctx, requestingAccount.ID, targetAccountID, true)
		if err != nil {
			return nil, fmt.Errorf("error checking account block: %s", err)
		}
	}

	var apiAccount *apimodel.Account
	if blocked {
		apiAccount, err = p.tc.AccountToAPIAccountBlocked(ctx, targetAccount)
		if err != nil {
			return nil, fmt.Errorf("error converting account: %s", err)
		}
		return apiAccount, nil
	}

	// last-minute check to make sure we have remote account header/avi cached
	if targetAccount.Domain != "" {
		a, err := p.federator.EnrichRemoteAccount(ctx, requestingAccount.Username, targetAccount)
		if err == nil {
			targetAccount = a
		}
	}

	if requestingAccount != nil && targetAccount.ID == requestingAccount.ID {
		apiAccount, err = p.tc.AccountToAPIAccountSensitive(ctx, targetAccount)
	} else {
		apiAccount, err = p.tc.AccountToAPIAccountPublic(ctx, targetAccount)
	}
	if err != nil {
		return nil, fmt.Errorf("error converting account: %s", err)
	}
	return apiAccount, nil
}
