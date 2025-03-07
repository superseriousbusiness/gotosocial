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

package user

import (
	"context"
	"fmt"
	"time"

	"codeberg.org/superseriousbusiness/oauth2/v4"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/text"
)

// Create processes the given form for creating a new user+account.
//
// App should be the app used to create the user+account.
// If nil, the instance app will be used.
//
// Precondition: the form's fields should have already been
// validated and normalized by the caller.
func (p *Processor) Create(
	ctx context.Context,
	app *gtsmodel.Application,
	form *apimodel.AccountCreateRequest,
) (*gtsmodel.User, gtserror.WithCode) {
	var (
		usersPerDay = config.GetAccountsRegistrationDailyLimit()
		regBacklog  = config.GetAccountsRegistrationBacklogLimit()
	)

	// If usersPerDay limit is in place,
	// ensure no more than usersPerDay
	// have registered in the last 24h.
	if usersPerDay > 0 {
		newUsersCount, err := p.state.DB.CountApprovedSignupsSince(ctx, time.Now().Add(-24*time.Hour))
		if err != nil {
			err := fmt.Errorf("db error counting new users: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if newUsersCount >= usersPerDay {
			err := fmt.Errorf("this instance has hit its limit of new sign-ups for today (%d); you can try again tomorrow", usersPerDay)
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}
	}

	// If registration backlog limit is
	// in place, ensure backlog isn't full.
	if regBacklog > 0 {
		backlogLen, err := p.state.DB.CountUnhandledSignups(ctx)
		if err != nil {
			err := fmt.Errorf("db error counting registration backlog length: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}

		if backlogLen >= regBacklog {
			err := fmt.Errorf(
				"this instance's sign-up backlog is currently full (%d sign-ups pending approval); "+
					"you must wait until some pending sign-ups are handled by the admin(s)", regBacklog,
			)
			return nil, gtserror.NewErrorUnprocessableEntity(err, err.Error())
		}
	}

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

	// Use instance app if no app provided.
	if app == nil {
		app, err = p.state.DB.GetInstanceApplication(ctx)
		if err != nil {
			err := fmt.Errorf("db error getting instance app: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
	}

	user, err := p.state.DB.NewSignup(ctx, gtsmodel.NewSignup{
		Username: form.Username,
		Email:    form.Email,
		Password: form.Password,
		Reason:   text.StripHTMLFromText(reason),
		SignUpIP: form.IP,
		Locale:   form.Locale,
		AppID:    app.ID,
	})
	if err != nil {
		err := fmt.Errorf("db error creating new signup: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// There are side effects for creating a new user+account
	// (confirmation emails etc), perform these async.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		// Use ap.ObjectProfile here to
		// distinguish this message (user model)
		// from ap.ActorPerson (account model).
		APObjectType:   ap.ObjectProfile,
		APActivityType: ap.ActivityCreate,
		GTSModel:       user,
		Origin:         user.Account,
	})

	return user, nil
}

// TokenForNewUser generates an OAuth Bearer token
// for a new user (with account) created by Create().
func (p *Processor) TokenForNewUser(
	ctx context.Context,
	appToken oauth2.TokenInfo,
	app *gtsmodel.Application,
	user *gtsmodel.User,
) (*apimodel.Token, gtserror.WithCode) {
	// Generate access token.
	accessToken, err := p.oauthServer.GenerateUserAccessToken(
		ctx,
		appToken,
		app.ClientSecret,
		user.ID,
	)
	if err != nil {
		err := fmt.Errorf("error creating new access token for user %s: %w", user.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return &apimodel.Token{
		AccessToken: accessToken.GetAccess(),
		TokenType:   "Bearer",
		Scope:       accessToken.GetScope(),
		CreatedAt:   accessToken.GetAccessCreateAt().Unix(),
	}, nil
}
