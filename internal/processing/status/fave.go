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
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
)

// FaveCreate processes the faving of a given status, returning the updated status if the fave goes through.
func (p *Processor) FaveCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
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
	if !*targetStatus.Likeable {
		return nil, gtserror.NewErrorForbidden(errors.New("status is not faveable"))
	}

	// first check if the status is already faved, if so we don't need to do anything
	newFave := true
	gtsFave := &gtsmodel.StatusFave{}
	if err := p.state.DB.GetWhere(ctx, []db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: requestingAccount.ID}}, gtsFave); err == nil {
		// we already have a fave for this status
		newFave = false
	}

	if newFave {
		thisFaveID := id.NewULID()

		// we need to create a new fave in the database
		gtsFave := &gtsmodel.StatusFave{
			ID:              thisFaveID,
			AccountID:       requestingAccount.ID,
			Account:         requestingAccount,
			TargetAccountID: targetStatus.AccountID,
			TargetAccount:   targetStatus.Account,
			StatusID:        targetStatus.ID,
			Status:          targetStatus,
			URI:             uris.GenerateURIForLike(requestingAccount.Username, thisFaveID),
		}

		if err := p.state.DB.Put(ctx, gtsFave); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting fave in database: %s", err))
		}

		// send it back to the processor for async processing
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActivityLike,
			APActivityType: ap.ActivityCreate,
			GTSModel:       gtsFave,
			OriginAccount:  requestingAccount,
			TargetAccount:  targetStatus.Account,
		})
	}

	// return the apidon representation of the target status
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return apiStatus, nil
}

// FaveRemove processes the unfaving of a given status, returning the updated status if the fave goes through.
func (p *Processor) FaveRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
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

	// check if we actually have a fave for this status
	var toUnfave bool

	gtsFave := &gtsmodel.StatusFave{}
	err = p.state.DB.GetWhere(ctx, []db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: requestingAccount.ID}}, gtsFave)
	if err == nil {
		// we have a fave
		toUnfave = true
	}
	if err != nil {
		// something went wrong in the db finding the fave
		if err != db.ErrNoEntries {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error fetching existing fave from database: %s", err))
		}
		// we just don't have a fave
		toUnfave = false
	}

	if toUnfave {
		// we had a fave, so take some action to get rid of it
		if err := p.state.DB.DeleteWhere(ctx, []db.Where{{Key: "status_id", Value: targetStatus.ID}, {Key: "account_id", Value: requestingAccount.ID}}, gtsFave); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error unfaveing status: %s", err))
		}

		// send it back to the processor for async processing
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActivityLike,
			APActivityType: ap.ActivityUndo,
			GTSModel:       gtsFave,
			OriginAccount:  requestingAccount,
			TargetAccount:  targetStatus.Account,
		})
	}

	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return apiStatus, nil
}

// FavedBy returns a slice of accounts that have liked the given status, filtered according to privacy settings.
func (p *Processor) FavedBy(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
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

	statusFaves, err := p.state.DB.GetStatusFaves(ctx, targetStatus)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing who faved status: %s", err))
	}

	// filter the list so the user doesn't see accounts they blocked or which blocked them
	filteredAccounts := []*gtsmodel.Account{}
	for _, fave := range statusFaves {
		blocked, err := p.state.DB.IsBlocked(ctx, requestingAccount.ID, fave.AccountID, true)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking blocks: %s", err))
		}
		if !blocked {
			filteredAccounts = append(filteredAccounts, fave.Account)
		}
	}

	// now we can return the api representation of those accounts
	apiAccounts := []*apimodel.Account{}
	for _, acc := range filteredAccounts {
		apiAccount, err := p.tc.AccountToAPIAccountPublic(ctx, acc)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}
