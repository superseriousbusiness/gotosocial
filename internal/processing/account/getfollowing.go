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

package account

import (
	"context"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) FollowingGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) ([]apimodel.Account, gtserror.WithCode) {
	if blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, targetAccountID, true); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	} else if blocked {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("block exists between accounts"))
	}

	accounts := []apimodel.Account{}
	follows, err := p.db.GetAccountFollows(ctx, targetAccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			return accounts, nil
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	for _, f := range follows {
		blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, f.AccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		if blocked {
			continue
		}

		if f.TargetAccount == nil {
			a, err := p.db.GetAccountByID(ctx, f.TargetAccountID)
			if err != nil {
				if err == db.ErrNoEntries {
					continue
				}
				return nil, gtserror.NewErrorInternalError(err)
			}
			f.TargetAccount = a
		}

		account, err := p.tc.AccountToAPIAccountPublic(ctx, f.TargetAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(err)
		}
		accounts = append(accounts, *account)
	}
	return accounts, nil
}
