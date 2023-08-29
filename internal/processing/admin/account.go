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
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (p *Processor) AccountAction(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	request *apimodel.AdminActionRequest,
) gtserror.WithCode {
	targetAccount, err := p.state.DB.GetAccountByID(ctx, request.TargetID)
	if err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	actionType := gtsmodel.AdminActionType(request.Type)

	var cMsg messages.FromClientAPI

	switch actionType {
	case gtsmodel.AdminActionSuspend:
		cMsg = messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			OriginAccount:  adminAcct,
			TargetAccount:  targetAccount,
		}

	default:
		supportedTypes := [1]gtsmodel.AdminActionType{
			gtsmodel.AdminActionSuspend,
		}

		err := fmt.Errorf(
			"admin action type %s is not supported for this endpoint, "+
				"currently supported types are: %q",
			request.Type, supportedTypes)

		return gtserror.NewErrorBadRequest(err, err.Error())
	}

	return p.Actions.Run(
		ctx,
		&gtsmodel.AdminAction{
			ID:             id.NewULID(),
			TargetCategory: gtsmodel.AdminActionCategoryAccount,
			TargetID:       targetAccount.ID,
			Target:         targetAccount,
			Type:           actionType,
			AccountID:      adminAcct.ID,
			Text:           request.Text,
		},
		func(ctx context.Context) gtserror.MultiError {
			if err := p.state.Workers.ProcessFromClientAPI(ctx, cMsg); err != nil {
				return gtserror.MultiError{err}
			}

			return nil
		},
	)
}
