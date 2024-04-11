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

package admin

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
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

func (p *Processor) AccountApprove(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	accountID string,
) (*apimodel.AdminAccountInfo, gtserror.WithCode) {
	user, err := p.state.DB.GetUserByAccountID(ctx, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("db error getting user for account id %s: %w", accountID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if user == nil {
		err := fmt.Errorf("user for account %s not found", accountID)
		return nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	if !*user.Approved {
		// Mark user as approved.
		user.Approved = util.Ptr(true)
		if err := p.state.DB.UpdateUser(ctx, user, "approved"); err != nil {
			err := gtserror.Newf("db error updating user %s: %w", user.ID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		// Process approval side effects asynschronously.
		p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityAccept,
			GTSModel:       user,
			OriginAccount:  adminAcct,
			TargetAccount:  user.Account,
		})
	}

	apiAccount, err := p.converter.AccountToAdminAPIAccount(ctx, user.Account)
	if err != nil {
		err := gtserror.Newf("error converting account %s to admin api model: %w", accountID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiAccount, nil
}
