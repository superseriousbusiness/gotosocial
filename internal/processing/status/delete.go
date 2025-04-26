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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

// Delete processes the delete of a given status, returning the deleted status if the delete goes through.
func (p *Processor) Delete(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("error fetching status %s: %s", targetStatusID, err))
	}

	if targetStatus.Account == nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("no status owner for status %s", targetStatusID))
	}

	if targetStatus.AccountID != requestingAccount.ID {
		return nil, gtserror.NewErrorForbidden(errors.New("status doesn't belong to requesting account"))
	}

	// Parse the status to API model BEFORE deleting it.
	apiStatus, errWithCode := p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Replace content warning with raw
	// version if it's available, to make
	// delete + redraft work nicer.
	if targetStatus.ContentWarningText != "" {
		apiStatus.SpoilerText = targetStatus.ContentWarningText
	}

	// Process delete side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ObjectNote,
		APActivityType: ap.ActivityDelete,
		GTSModel:       targetStatus,
		Origin:         requestingAccount,
		Target:         requestingAccount,
	})

	return apiStatus, nil
}
