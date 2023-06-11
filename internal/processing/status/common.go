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
	"fmt"

	"codeberg.org/gruf/go-kv"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
)

func (p *Processor) apiStatus(ctx context.Context, targetStatus *gtsmodel.Status, requestingAccount *gtsmodel.Account) (*apimodel.Status, gtserror.WithCode) {
	apiStatus, err := p.tc.StatusToAPIStatus(ctx, targetStatus, requestingAccount)
	if err != nil {
		err = gtserror.Newf("error converting status %s to frontend representation: %w", targetStatus.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiStatus, nil
}

func (p *Processor) getVisibleStatus(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*gtsmodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		err = fmt.Errorf("getVisibleStatus: db error fetching status %s: %w", targetStatusID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if requestingAccount != nil {
		// Ensure the status is up-to-date.
		p.federator.RefreshStatusAsync(ctx,
			requestingAccount.Username,
			targetStatus,
			nil,
			false,
		)
	}

	visible, err := p.filter.StatusVisible(ctx, requestingAccount, targetStatus)
	if err != nil {
		err = fmt.Errorf("getVisibleStatus: error seeing if status %s is visible: %w", targetStatus.ID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if !visible {
		err = fmt.Errorf("getVisibleStatus: status %s is not visible to requesting account", targetStatusID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	return targetStatus, nil
}

// invalidateStatus is a shortcut function for invalidating the prepared/cached
// representation one status in the home timeline and all list timelines of the
// given accountID. It should only be called in cases where a status update
// does *not* need to be passed into the processor via the worker queue, since
// such invalidation will, in that case, be handled by the processor instead.
func (p *Processor) invalidateStatus(ctx context.Context, accountID string, statusID string) error {
	// Get lists first + bail if this fails.
	lists, err := p.state.DB.GetListsForAccountID(ctx, accountID)
	if err != nil {
		return gtserror.Newf("db error getting lists for account %s: %w", accountID, err)
	}

	l := log.WithContext(ctx).WithFields(kv.Fields{
		{"accountID", accountID},
		{"statusID", statusID},
	}...)

	// Unprepare item from home + list timelines, just log
	// if something goes wrong since this is not a showstopper.

	if err := p.state.Timelines.Home.UnprepareItem(ctx, accountID, statusID); err != nil {
		l.Errorf("error unpreparing item from home timeline: %v", err)
	}

	for _, list := range lists {
		if err := p.state.Timelines.List.UnprepareItem(ctx, list.ID, statusID); err != nil {
			l.Errorf("error unpreparing item from list timeline %s: %v", list.ID, err)
		}
	}

	return nil
}
