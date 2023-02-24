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
	"time"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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

	if targetStatus.BoostOfID != "" {
		err = errors.New("cannot pin boosts")
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Only pin status if it's not pinned already.
	if !targetStatus.PinnedAt.IsZero() {
		// Ensure we have enough slots to pin this.
		pinnedCount, err := p.db.CountAccountPinned(ctx, requestingAccount.ID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking number of pinned statuses: %w", err))
		}

		if pinnedCount >= allowedPinnedCount {
			err = fmt.Errorf("status pin limit exceeded, you've already pinned %d status(es) out of %d", pinnedCount, allowedPinnedCount)
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}

		// We need to mark this status as pinned!
		targetStatus.PinnedAt = time.Now()
		if err := p.db.UpdateStatus(ctx, targetStatus); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error pinning status: %w", err))
		}
	}

	// Return the api representation of the target status.
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

	// Only unpin status if it's pinned.
	if targetStatus.PinnedAt.IsZero() {
		// We need to mark this status as no longer pinned!
		targetStatus.PinnedAt = time.Time{}
		if err := p.db.UpdateStatus(ctx, targetStatus); err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error unpinning status: %w", err))
		}
	}

	// return the api representation of the target status
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %w", targetStatus.ID, err))
	}

	return apiStatus, nil
}
