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

package status

import (
	"context"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *processor) BoostedBy(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		wrapped := fmt.Errorf("BoostedBy: error fetching status %s: %s", targetStatusID, err)
		if !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorInternalError(wrapped)
		}
		return nil, gtserror.NewErrorNotFound(wrapped)
	}

	if boostOfID := targetStatus.BoostOfID; boostOfID != "" {
		// the target status is a boost wrapper, redirect this request to the status it boosts
		boostedStatus, err := p.db.GetStatusByID(ctx, boostOfID)
		if err != nil {
			wrapped := fmt.Errorf("BoostedBy: error fetching status %s: %s", boostOfID, err)
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.NewErrorInternalError(wrapped)
			}
			return nil, gtserror.NewErrorNotFound(wrapped)
		}
		targetStatus = boostedStatus
	}

	visible, err := p.filter.StatusVisible(ctx, targetStatus, requestingAccount)
	if err != nil {
		err = fmt.Errorf("BoostedBy: error seeing if status %s is visible: %s", targetStatus.ID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}
	if !visible {
		err = errors.New("BoostedBy: status is not visible")
		return nil, gtserror.NewErrorNotFound(err)
	}

	statusReblogs, err := p.db.GetStatusReblogs(ctx, targetStatus)
	if err != nil {
		err = fmt.Errorf("BoostedBy: error seeing who boosted status: %s", err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// filter account IDs so the user doesn't see accounts they blocked or which blocked them
	accountIDs := make([]string, 0, len(statusReblogs))
	for _, s := range statusReblogs {
		blocked, err := p.db.IsBlocked(ctx, requestingAccount.ID, s.AccountID, true)
		if err != nil {
			err = fmt.Errorf("BoostedBy: error checking blocks: %s", err)
			return nil, gtserror.NewErrorNotFound(err)
		}
		if !blocked {
			accountIDs = append(accountIDs, s.AccountID)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// fetch accounts + create their API representations
	apiAccounts := make([]*apimodel.Account, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		account, err := p.db.GetAccountByID(ctx, accountID)
		if err != nil {
			wrapped := fmt.Errorf("BoostedBy: error fetching account %s: %s", accountID, err)
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.NewErrorInternalError(wrapped)
			}
			return nil, gtserror.NewErrorNotFound(wrapped)
		}

		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			err = fmt.Errorf("BoostedBy: error converting account to api model: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}
