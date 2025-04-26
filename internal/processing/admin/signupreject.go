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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

func (p *Processor) SignupReject(
	ctx context.Context,
	adminAcct *gtsmodel.Account,
	accountID string,
	privateComment string,
	sendEmail bool,
	message string,
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

	// Get a lock on the account URI,
	// since we're going to be deleting
	// it and its associated user.
	unlock := p.state.ProcessingLocks.Lock(user.Account.URI)
	defer unlock()

	// Can't reject an account with a
	// user that's already been approved.
	if *user.Approved {
		err := fmt.Errorf("account %s has already been approved", accountID)
		return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
	}

	// Convert to API account *before* doing the
	// rejection, since the rejection will cause
	// the user and account to be removed.
	apiAccount, err := p.converter.AccountToAdminAPIAccount(ctx, user.Account)
	if err != nil {
		err := gtserror.Newf("error converting account %s to admin api model: %w", accountID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Set approved to false on the API model, to
	// reflect the changes that will occur
	// asynchronously in the processor.
	apiAccount.Approved = false

	// Ensure we an email address.
	var email string
	if user.Email != "" {
		email = user.Email
	} else {
		email = user.UnconfirmedEmail
	}

	// Create a denied user entry for
	// the worker to process + store.
	deniedUser := &gtsmodel.DeniedUser{
		ID:                     user.ID,
		Email:                  email,
		Username:               user.Account.Username,
		SignUpIP:               user.SignUpIP,
		InviteID:               user.InviteID,
		Locale:                 user.Locale,
		CreatedByApplicationID: user.CreatedByApplicationID,
		SignUpReason:           user.Reason,
		PrivateComment:         privateComment,
		SendEmail:              &sendEmail,
		Message:                message,
	}

	// Process rejection side effects asynschronously.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		// Use ap.ObjectProfile here to
		// distinguish this message (user model)
		// from ap.ActorPerson (account model).
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityReject,
		GTSModel:       deniedUser,
		Origin:         adminAcct,
		Target:         user.Account,
	})

	return apiAccount, nil
}
