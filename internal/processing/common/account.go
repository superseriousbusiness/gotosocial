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

package common

import (
	"context"
	"errors"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
)

// GetTargetAccountBy fetches the target account with db load function, given the authorized (or, nil) requester's
// account. This returns an approprate gtserror.WithCode accounting (ha) for not found and visibility to requester.
func (p *Processor) GetTargetAccountBy(
	ctx context.Context,
	requester *gtsmodel.Account,
	getTargetFromDB func() (*gtsmodel.Account, error),
) (
	account *gtsmodel.Account,
	visible bool,
	errWithCode gtserror.WithCode,
) {
	// Fetch the target account from db.
	target, err := getTargetFromDB()
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting from db: %w", err)
		return nil, false, gtserror.NewErrorInternalError(err)
	}

	if target == nil {
		// DB loader could not find account in database.
		const text = "target account not found"
		return nil, false, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	// Check whether target account is visible to requesting account.
	visible, err = p.visFilter.AccountVisible(ctx, requester, target)
	if err != nil {
		err := gtserror.Newf("error checking visibility: %w", err)
		return nil, false, gtserror.NewErrorInternalError(err)
	}

	if requester != nil && visible {
		// Only refresh account if visible to requester,
		// and there is *authorized* requester to prevent
		// a possible DOS vector for unauthorized clients.
		latest, _, err := p.federator.RefreshAccount(ctx,
			requester.Username,
			target,
			nil,
			nil,
		)
		if err != nil {
			log.Errorf(ctx, "error refreshing target %s: %v", target.URI, err)
			return target, visible, nil
		}

		// Set latest.
		target = latest
	}

	return target, visible, nil
}

// GetTargetAccountByID is a call-through to GetTargetAccountBy() using the db GetAccountByID() function.
func (p *Processor) GetTargetAccountByID(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetID string,
) (
	account *gtsmodel.Account,
	visible bool,
	errWithCode gtserror.WithCode,
) {
	return p.GetTargetAccountBy(ctx, requester, func() (*gtsmodel.Account, error) {
		return p.state.DB.GetAccountByID(ctx, targetID)
	})
}

// GetVisibleTargetAccount calls GetTargetAccountByID(),
// but converts a non-visible result to not-found error.
func (p *Processor) GetVisibleTargetAccount(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetID string,
) (
	account *gtsmodel.Account,
	errWithCode gtserror.WithCode,
) {
	// Fetch the target account by ID from the database.
	target, visible, errWithCode := p.GetTargetAccountByID(ctx,
		requester,
		targetID,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if !visible {
		// Pretend account doesn't exist if not visible.
		const text = "target account not found"
		return nil, gtserror.NewErrorNotFound(
			errors.New(text),
			text,
		)
	}

	return target, nil
}

// GetAPIAccount fetches the appropriate API account
// model depending on whether requester = target.
func (p *Processor) GetAPIAccount(
	ctx context.Context,
	requester *gtsmodel.Account,
	target *gtsmodel.Account,
) (
	apiAcc *apimodel.Account,
	errWithCode gtserror.WithCode,
) {
	var err error

	if requester != nil && requester.ID == target.ID {
		// Only return sensitive account model _if_ requester = target.
		apiAcc, err = p.converter.AccountToAPIAccountSensitive(ctx, target)
	} else {
		// Else, fall back to returning the public account model.
		apiAcc, err = p.converter.AccountToAPIAccountPublic(ctx, target)
	}

	if err != nil {
		err := gtserror.Newf("error converting: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return apiAcc, nil
}

// GetAPIAccountBlocked fetches the limited
// "blocked" account model for given target.
func (p *Processor) GetAPIAccountBlocked(
	ctx context.Context,
	targetAcc *gtsmodel.Account,
) (
	apiAcc *apimodel.Account,
	errWithCode gtserror.WithCode,
) {
	apiAccount, err := p.converter.AccountToAPIAccountBlocked(ctx, targetAcc)
	if err != nil {
		err := gtserror.Newf("error converting: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return apiAccount, nil
}

// GetAPIAccountSensitive fetches the "sensitive" account model for the given target.
// *BE CAREFUL!* Only return a sensitive account if targetAcc == account making the request.
func (p *Processor) GetAPIAccountSensitive(
	ctx context.Context,
	targetAcc *gtsmodel.Account,
) (
	apiAcc *apimodel.Account,
	errWithCode gtserror.WithCode,
) {
	apiAccount, err := p.converter.AccountToAPIAccountSensitive(ctx, targetAcc)
	if err != nil {
		err := gtserror.Newf("error converting: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}
	return apiAccount, nil
}

// GetVisibleAPIAccounts converts an array of gtsmodel.Accounts (inputted by next function) into
// public API model accounts, checking first for visibility. Please note that all errors will be
// logged at ERROR level, but will not be returned. Callers are likely to run into show-stopping
// errors in the lead-up to this function, whereas calling this should not be a show-stopper.
func (p *Processor) GetVisibleAPIAccounts(
	ctx context.Context,
	requester *gtsmodel.Account,
	next func(int) *gtsmodel.Account,
	length int,
) []*apimodel.Account {
	return p.getVisibleAPIAccounts(ctx, 3, requester, next, length)
}

// GetVisibleAPIAccountsPaged is functionally equivalent to GetVisibleAPIAccounts(),
// except the accounts are returned as a converted slice of accounts as interface{}.
func (p *Processor) GetVisibleAPIAccountsPaged(
	ctx context.Context,
	requester *gtsmodel.Account,
	next func(int) *gtsmodel.Account,
	length int,
) []interface{} {
	accounts := p.getVisibleAPIAccounts(ctx, 3, requester, next, length)
	items := make([]interface{}, len(accounts))
	for i, account := range accounts {
		items[i] = account
	}
	return items
}

func (p *Processor) getVisibleAPIAccounts(
	ctx context.Context,
	calldepth int, // used to skip wrapping func above these's names
	requester *gtsmodel.Account,
	next func(int) *gtsmodel.Account,
	length int,
) []*apimodel.Account {
	// Start new log entry with
	// the above calling func's name.
	l := log.WithContext(ctx).
		WithField("caller", log.Caller(calldepth+1))

	// Preallocate slice according to expected length.
	accounts := make([]*apimodel.Account, 0, length)

	for i := 0; i < length; i++ {
		// Get next account.
		account := next(i)
		if account == nil {
			continue
		}

		// Check whether this account is visible to requesting account.
		visible, err := p.visFilter.AccountVisible(ctx, requester, account)
		if err != nil {
			l.Errorf("error checking account visibility: %v", err)
			continue
		}

		if !visible {
			// Not visible to requester.
			continue
		}

		// Convert the account to a public API model representation.
		apiAcc, err := p.converter.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			l.Errorf("error converting account: %v", err)
			continue
		}

		// Append API model to return slice.
		accounts = append(accounts, apiAcc)
	}

	return accounts
}
