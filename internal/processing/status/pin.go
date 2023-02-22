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
)

const allowedPinnedCount = 10

// PinCreate pins the target status to the top of requestingAccount's profile, if possible.
func (p *Processor) PinCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %w", targetStatusID, err))
	}

	if targetStatus.AccountID != requestingAccount.ID {
		err = fmt.Errorf("status %s does not belong to account %s", targetStatusID, requestingAccount.ID)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	if targetStatus.Visibility == gtsmodel.VisibilityDirect {
		err = errors.New("cannot pin direct messages")
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// First check if the status is already pinned;
	// if it is then we don't need to do anything.
	pinned, err := p.db.IsStatusPinnedBy(ctx, targetStatus, requestingAccount.ID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking if status is pinned: %w", err))
	}

	if !pinned {
		// ensure we have enough slots to pin this
		pinnedCount, err := p.db.CountAccountPinned(ctx, requestingAccount.ID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking number of pinned statuses: %w", err))
		}

		if pinnedCount >= allowedPinnedCount {
			err = fmt.Errorf("status pin limit exceeded, you've already pinned %d status(es) out of %d", pinnedCount, allowedPinnedCount)
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// We need to create a new pin in the database!
		pin := &gtsmodel.StatusPin{
			ID:        id.NewULID(),
			AccountID: requestingAccount.ID,
			Account:   requestingAccount,
			StatusID:  targetStatus.ID,
			Status:    targetStatus,
		}

		if err := p.db.Put(ctx, pin); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error putting pin in database: %w", err))
		}

		// put the pinned status in the worker channel
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ObjectCollection,
			APActivityType: ap.ActivityAdd,
			GTSModel:       targetStatus,
			OriginAccount:  requestingAccount,
		})
	}

	// return the api representation of the target status
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %w", targetStatus.ID, err))
	}

	return apiStatus, nil
}

// PinRemove unpins the target status from the top of requestingAccount's profile, if possible.
func (p *Processor) PinRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %w", targetStatusID, err))
	}

	if targetStatus.AccountID != requestingAccount.ID {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("status %s does not belong to account %s", targetStatusID, requestingAccount.ID))
	}

	if targetStatus.Visibility == gtsmodel.VisibilityDirect {
		err := errors.New("cannot pin direct messages")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	// First check if the status is actually pinned;
	// if is not then we don't need to do anything.
	pinned, err := p.db.IsStatusPinnedBy(ctx, targetStatus, requestingAccount.ID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking if status is pinned: %w", err))
	}

	if pinned {
		// remove the pin
		if err := p.db.DeleteWhere(ctx, []db.Where{{Key: "account_id", Value: requestingAccount.ID}, {Key: "status_id", Value: targetStatus.ID}}, &gtsmodel.StatusPin{}); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error deleting pin from database: %w", err))
		}

		// put the unpinned status in the worker channel
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ObjectCollection,
			APActivityType: ap.ActivityRemove,
			GTSModel:       targetStatus,
			OriginAccount:  requestingAccount,
		})
	}

	// return the api representation of the target status
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %w", targetStatus.ID, err))
	}

	return apiStatus, nil
}
