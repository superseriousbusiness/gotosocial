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

package status

import (
	"context"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

// BoostCreate processes the boost/reblog of a given status, returning the newly-created boost if all is well.
func (p *Processor) BoostCreate(ctx context.Context, requestingAccount *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}
	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	// if targetStatusID refers to a boost, then we should redirect
	// the target to being the status that was boosted; if we don't
	// do this, then we end up in weird situations where people
	// boost boosts, and it looks absolutely bizarre in the UI
	if targetStatus.BoostOfID != "" {
		if targetStatus.BoostOf == nil {
			b, err := p.state.DB.GetStatusByID(ctx, targetStatus.BoostOfID)
			if err != nil {
				return nil, gtserror.NewErrorNotFound(fmt.Errorf("couldn't fetch boosted status %s", targetStatus.BoostOfID))
			}
			targetStatus.BoostOf = b
		}
		targetStatus = targetStatus.BoostOf
	}

	boostable, err := p.filter.StatusBoostable(ctx, requestingAccount, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is boostable: %s", targetStatus.ID, err))
	} else if !boostable {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not boostable"))
	}

	// it's visible! it's boostable! so let's boost the FUCK out of it
	boostWrapperStatus, err := p.converter.StatusToBoost(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	boostWrapperStatus.CreatedWithApplicationID = application.ID
	boostWrapperStatus.BoostOfAccount = targetStatus.Account

	// put the boost in the database
	if err := p.state.DB.PutStatus(ctx, boostWrapperStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it back to the processor for async processing
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityCreate,
		GTSModel:       boostWrapperStatus,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetStatus.Account,
	})

	return p.apiStatus(ctx, boostWrapperStatus, requestingAccount)
}

// BoostRemove processes the unboost/unreblog of a given status, returning the status if all is well.
func (p *Processor) BoostRemove(ctx context.Context, requestingAccount *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}
	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	visible, err := p.filter.StatusVisible(ctx, requestingAccount, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is visible: %s", targetStatus.ID, err))
	}
	if !visible {
		return nil, gtserror.NewErrorNotFound(errors.New("status is not visible"))
	}

	// Check whether the requesting account has boosted the given status ID.
	boost, err := p.state.DB.GetStatusBoost(ctx, targetStatusID, requestingAccount.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error checking status boost %s: %w", targetStatusID, err))
	}

	if boost != nil {
		// pin some stuff onto the boost while we have it out of the db
		boost.Account = requestingAccount
		boost.BoostOf = targetStatus
		boost.BoostOfAccount = targetStatus.Account
		boost.BoostOf.Account = targetStatus.Account

		// send it back to the processor for async processing
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActivityAnnounce,
			APActivityType: ap.ActivityUndo,
			GTSModel:       boost,
			OriginAccount:  requestingAccount,
			TargetAccount:  targetStatus.Account,
		})
	}

	return p.apiStatus(ctx, targetStatus, requestingAccount)
}

// StatusBoostedBy returns a slice of accounts that have boosted the given status, filtered according to privacy settings.
func (p *Processor) StatusBoostedBy(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		wrapped := fmt.Errorf("BoostedBy: error fetching status %s: %s", targetStatusID, err)
		if !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorInternalError(wrapped)
		}
		return nil, gtserror.NewErrorNotFound(wrapped)
	}

	if boostOfID := targetStatus.BoostOfID; boostOfID != "" {
		// the target status is a boost wrapper, redirect this request to the status it boosts
		boostedStatus, err := p.state.DB.GetStatusByID(ctx, boostOfID)
		if err != nil {
			wrapped := fmt.Errorf("BoostedBy: error fetching status %s: %s", boostOfID, err)
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.NewErrorInternalError(wrapped)
			}
			return nil, gtserror.NewErrorNotFound(wrapped)
		}
		targetStatus = boostedStatus
	}

	visible, err := p.filter.StatusVisible(ctx, requestingAccount, targetStatus)
	if err != nil {
		err = fmt.Errorf("BoostedBy: error seeing if status %s is visible: %s", targetStatus.ID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}
	if !visible {
		err = errors.New("BoostedBy: status is not visible")
		return nil, gtserror.NewErrorNotFound(err)
	}

	statusBoosts, err := p.state.DB.GetStatusBoosts(ctx, targetStatus.ID)
	if err != nil {
		err = fmt.Errorf("BoostedBy: error seeing who boosted status: %s", err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// filter account IDs so the user doesn't see accounts they blocked or which blocked them
	accountIDs := make([]string, 0, len(statusBoosts))
	for _, s := range statusBoosts {
		blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, s.AccountID)
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
		account, err := p.state.DB.GetAccountByID(ctx, accountID)
		if err != nil {
			wrapped := fmt.Errorf("BoostedBy: error fetching account %s: %s", accountID, err)
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.NewErrorInternalError(wrapped)
			}
			return nil, gtserror.NewErrorNotFound(wrapped)
		}

		apiAccount, err := p.converter.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			err = fmt.Errorf("BoostedBy: error converting account to api model: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}
