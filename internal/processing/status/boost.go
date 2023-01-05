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
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *processor) Boost(ctx context.Context, requestingAccount *gtsmodel.Account, application *gtsmodel.Application, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.db.GetStatusByID(ctx, targetStatusID)
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
			b, err := p.db.GetStatusByID(ctx, targetStatus.BoostOfID)
			if err != nil {
				return nil, gtserror.NewErrorNotFound(fmt.Errorf("couldn't fetch boosted status %s", targetStatus.BoostOfID))
			}
			targetStatus.BoostOf = b
		}
		targetStatus = targetStatus.BoostOf
	}

	boostable, err := p.filter.StatusBoostable(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error seeing if status %s is boostable: %s", targetStatus.ID, err))
	}
	if !boostable {
		return nil, gtserror.NewErrorForbidden(errors.New("status is not boostable"))
	}

	// it's visible! it's boostable! so let's boost the FUCK out of it
	boostWrapperStatus, err := p.tc.StatusToBoost(ctx, targetStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	boostWrapperStatus.CreatedWithApplicationID = application.ID
	boostWrapperStatus.BoostOfAccount = targetStatus.Account

	// put the boost in the database
	if err := p.db.PutStatus(ctx, boostWrapperStatus); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// send it back to the processor for async processing
	p.clientWorker.Queue(messages.FromClientAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityCreate,
		GTSModel:       boostWrapperStatus,
		OriginAccount:  requestingAccount,
		TargetAccount:  targetStatus.Account,
	})

	// return the frontend representation of the new status to the submitter
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, boostWrapperStatus, requestingAccount)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting status %s to frontend representation: %s", targetStatus.ID, err))
	}

	return apiStatus, nil
}
