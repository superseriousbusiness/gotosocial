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

package status

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) BoostedBy(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}
	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	visible, err := p.filter.StatusVisible(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	statusReblogs, err := p.db.GetStatusReblogs(ctx, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error seeing who boosted status: %s", err))
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, s := range statusReblogs {
		blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, s.AccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusBoostedBy: error checking blocks: %s", err))
		}
		if !blocked {
			filteredAccounts = append(filteredAccounts, s.Account)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// now we can return the api representation of those accounts
	apiAccounts := []*apimodel.Account{}
	for _, acc := range filteredAccounts {
		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, acc)
		if err != nil {
			return nil, gtserror.NewErrorNotFound(fmt.Errorf("StatusFavedBy: error converting account to api model: %s", err))
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}
