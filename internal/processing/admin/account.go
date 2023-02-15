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

func (p *AdminProcessor) AdminAccountAction(ctx context.Context, account *gtsmodel.Account, form *apimodel.AdminAccountActionRequest) gtserror.WithCode {
	targetAccount, err := p.db.GetAccountByID(ctx, form.TargetAccountID)
	if err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	adminAction := &gtsmodel.AdminAccountAction{
		ID:              id.NewULID(),
		AccountID:       account.ID,
		TargetAccountID: targetAccount.ID,
		Text:            form.Text,
	}

	switch form.Type {
	case string(gtsmodel.AdminActionSuspend):
		adminAction.Type = gtsmodel.AdminActionSuspend
		// pass the account delete through the client api channel for processing
		p.clientWorker.Queue(messages.FromClientAPI{
			APObjectType:   ap.ActorPerson,
			APActivityType: ap.ActivityDelete,
			OriginAccount:  account,
			TargetAccount:  targetAccount,
		})
	default:
		return gtserror.NewErrorBadRequest(fmt.Errorf("admin action type %s is not supported for this endpoint", form.Type))
	}

	if err := p.db.Put(ctx, adminAction); err != nil {
		return gtserror.NewErrorInternalError(err)
	}

	return nil
}
