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
	"errors"
	"fmt"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

// Get processes the given request for account information.
func (p *Processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Account, gtserror.WithCode) {
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %s", err))
	}

	return p.getFor(ctx, requestingAccount, targetAccount)
}

// GetLocalByUsername processes the given request for account information targeting a local account by username.
func (p *Processor) GetLocalByUsername(ctx context.Context, requestingAccount *gtsmodel.Account, username string) (*apimodel.Account, gtserror.WithCode) {
	targetAccount, err := p.state.DB.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %s", err))
	}

	return p.getFor(ctx, requestingAccount, targetAccount)
}

// GetCustomCSSForUsername returns custom css for the given local username.
func (p *Processor) GetCustomCSSForUsername(ctx context.Context, username string) (string, gtserror.WithCode) {
	customCSS, err := p.state.DB.GetAccountCustomCSSByUsername(ctx, username)
	if err != nil {
		if err == db.ErrNoEntries {
			return "", gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return "", gtserror.NewErrorInternalError(fmt.Errorf("db error: %s", err))
	}

	return customCSS, nil
}

func (p *Processor) getFor(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (*apimodel.Account, gtserror.WithCode) {
	var blocked bool
	var err error
	if requestingAccount != nil {
		blocked, err = p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, targetAccount.ID)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error checking account block: %s", err))
		}
	}

	var apiAccount *apimodel.Account
	if blocked {
		apiAccount, err = p.tc.AccountToAPIAccountBlocked(ctx, targetAccount)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting account: %s", err))
		}
		return apiAccount, nil
	}

	// last-minute check to make sure we have remote account header/avi cached
	if targetAccount.Domain != "" {
		targetAccountURI, err := url.Parse(targetAccount.URI)
		if err != nil {
			return nil, gtserror.NewErrorInternalError(fmt.Errorf("error parsing url %s: %s", targetAccount.URI, err))
		}

		a, err := p.federator.GetAccountByURI(
			transport.WithFastfail(ctx), requestingAccount.Username, targetAccountURI, true,
		)
		if err == nil {
			targetAccount = a
		}
	}

	if requestingAccount != nil && targetAccount.ID == requestingAccount.ID {
		apiAccount, err = p.tc.AccountToAPIAccountSensitive(ctx, targetAccount)
	} else {
		apiAccount, err = p.tc.AccountToAPIAccountPublic(ctx, targetAccount)
	}
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error converting account: %s", err))
	}

	return apiAccount, nil
}
