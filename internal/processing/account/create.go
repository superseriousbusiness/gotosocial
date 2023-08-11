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

package account

import (
	"context"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
	"github.com/superseriousbusiness/oauth2/v4"
)

// Create processes the given form for creating a new account,
// returning an oauth token for that account if successful.
//
// Precondition: the form's fields should have already been validated and normalized by the caller.
func (p *Processor) Create(
	ctx context.Context,
	appToken oauth2.TokenInfo,
	app *gtsmodel.Application,
	form *apimodel.AccountCreateRequest,
) (*apimodel.Token, gtserror.WithCode) {
	emailAvailable, err := p.state.DB.IsEmailAvailable(ctx, form.Email)
	if err != nil {
		err := fmt.Errorf("db error checking email availability: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !emailAvailable {
		err := fmt.Errorf("email address %s is not available", form.Email)
		return nil, gtserror.NewErrorConflict(err, err.Error())
	}

	usernameAvailable, err := p.state.DB.IsUsernameAvailable(ctx, form.Username)
	if err != nil {
		err := fmt.Errorf("db error checking username availability: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	if !usernameAvailable {
		err := fmt.Errorf("username %s is not available", form.Username)
		return nil, gtserror.NewErrorConflict(err, err.Error())
	}

	// Only store reason if one is required.
	var reason string
	if config.GetAccountsReasonRequired() {
		reason = form.Reason
	}

	user, err := p.state.DB.NewSignup(ctx, gtsmodel.NewSignup{
		Username:    form.Username,
		Email:       form.Email,
		Password:    form.Password,
		Reason:      text.SanitizeToPlaintext(reason),
		PreApproved: !config.GetAccountsApprovalRequired(), // Mark as approved if no approval required.
		SignUpIP:    form.IP,
		Locale:      form.Locale,
		AppID:       app.ID,
	})
	if err != nil {
		err := fmt.Errorf("db error creating new signup: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Generate access token *before* doing side effects; we
	// don't want to process side effects if something borks.
	accessToken, err := p.oauthServer.GenerateUserAccessToken(ctx, appToken, app.ClientSecret, user.ID)
	if err != nil {
		err := fmt.Errorf("error creating new access token for user %s: %w", user.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// There are side effects for creating a new account
	// (confirmation emails etc), perform these async.
	p.state.Workers.EnqueueClientAPI(ctx, messages.FromClientAPI{
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityCreate,
		GTSModel:       user.Account,
		OriginAccount:  user.Account,
	})

	return &apimodel.Token{
		AccessToken: accessToken.GetAccess(),
		TokenType:   "Bearer",
		Scope:       accessToken.GetScope(),
		CreatedAt:   accessToken.GetAccessCreateAt().Unix(),
	}, nil
}
