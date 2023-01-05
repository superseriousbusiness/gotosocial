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

package account

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/federation/dereferencing"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/transport"
)

func (p *processor) Get(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Account, gtserror.WithCode) {
	targetAccount, err := p.db.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %s", err))
	}

	return p.getAccountFor(ctx, requestingAccount, targetAccount)
}

func (p *processor) GetLocalByUsername(ctx context.Context, requestingAccount *gtsmodel.Account, username string) (*apimodel.Account, gtserror.WithCode) {
	targetAccount, err := p.db.GetAccountByUsernameDomain(ctx, username, "")
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("db error: %s", err))
	}

	return p.getAccountFor(ctx, requestingAccount, targetAccount)
}

func (p *processor) GetCustomCSSForUsername(ctx context.Context, username string) (string, gtserror.WithCode) {
	customCSS, err := p.db.GetAccountCustomCSSByUsername(ctx, username)
	if err != nil {
		if err == db.ErrNoEntries {
			return "", gtserror.NewErrorNotFound(errors.New("account not found"))
		}
		return "", gtserror.NewErrorInternalError(fmt.Errorf("db error: %s", err))
	}

	return customCSS, nil
}

func (p *processor) getAccountFor(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccount *gtsmodel.Account) (*apimodel.Account, gtserror.WithCode) {
	var blocked bool
	var err error
	if requestingAccount != nil {
		blocked, err = p.db.IsBlocked(ctx, requestingAccount.ID, targetAccount.ID, true)
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

		a, err := p.federator.GetAccount(transport.WithFastfail(ctx), dereferencing.GetAccountParams{
			RequestingUsername:    requestingAccount.Username,
			RemoteAccountID:       targetAccountURI,
			RemoteAccountHost:     targetAccount.Domain,
			RemoteAccountUsername: targetAccount.Username,
			Blocking:              true,
		})
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
