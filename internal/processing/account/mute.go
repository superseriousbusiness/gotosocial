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
	"time"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// MuteCreate handles the creation or updating of a mute from requestingAccount to targetAccountID.
// The form params should have already been normalized by the time they reach this function.
func (p *Processor) MuteCreate(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetAccountID string,
	form *apimodel.UserMuteCreateUpdateRequest,
) (*apimodel.Relationship, gtserror.WithCode) {
	targetAccount, existingMute, errWithCode := p.getMuteTarget(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingMute != nil &&
		*existingMute.Notifications == *form.Notifications &&
		existingMute.ExpiresAt.IsZero() && form.Duration == nil {
		// Mute already exists and doesn't require updating, nothing to do.
		return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
	}

	// Create a new mute or update an existing one.
	mute := &gtsmodel.UserMute{
		AccountID:       requestingAccount.ID,
		Account:         requestingAccount,
		TargetAccountID: targetAccountID,
		TargetAccount:   targetAccount,
		Notifications:   form.Notifications,
	}
	if existingMute != nil {
		mute.ID = existingMute.ID
	} else {
		mute.ID = id.NewULID()
	}
	if form.Duration != nil {
		mute.ExpiresAt = time.Now().Add(time.Second * time.Duration(*form.Duration))
	}

	if err := p.state.DB.PutMute(ctx, mute); err != nil {
		err = gtserror.Newf("error creating or updating mute in db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

// MuteRemove handles the removal of a mute from requestingAccount to targetAccountID.
func (p *Processor) MuteRemove(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetAccountID string,
) (*apimodel.Relationship, gtserror.WithCode) {
	_, existingMute, errWithCode := p.getMuteTarget(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingMute == nil {
		// Already not muted, nothing to do.
		return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
	}

	// We got a mute, remove it from the db.
	if err := p.state.DB.DeleteMuteByID(ctx, existingMute.ID); err != nil {
		err := gtserror.Newf("error removing mute from db: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.RelationshipGet(ctx, requestingAccount, targetAccountID)
}

// MutesGet retrieves the user's list of muted accounts, with an extra field for mute expiration (if applicable).
func (p *Processor) MutesGet(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	page *paging.Page,
) (*apimodel.PageableResponse, gtserror.WithCode) {
	mutes, err := p.state.DB.GetAccountMutes(ctx,
		requestingAccount.ID,
		page,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("couldn't list account's mutes: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(mutes)
	if len(mutes) == 0 {
		return util.EmptyPageableResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := mutes[count-1].ID
	hi := mutes[0].ID

	items := make([]interface{}, 0, count)

	now := time.Now()
	for _, mute := range mutes {
		// Skip accounts for which the mute has expired.
		if mute.Expired(now) {
			continue
		}

		// Convert target account to frontend API model. (target will never be nil)
		account, err := p.converter.AccountToAPIAccountPublic(ctx, mute.TargetAccount)
		if err != nil {
			log.Errorf(ctx, "error converting account to public api account: %v", err)
			continue
		}
		mutedAccount := &apimodel.MutedAccount{
			Account: *account,
		}
		// Add the mute expiration field (unique to this API).
		if !mute.ExpiresAt.IsZero() {
			mutedAccount.MuteExpiresAt = util.Ptr(util.FormatISO8601(mute.ExpiresAt))
		}

		// Append target to return items.
		items = append(items, mutedAccount)
	}

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/mutes",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

func (p *Processor) getMuteTarget(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetAccountID string,
) (*gtsmodel.Account, *gtsmodel.UserMute, gtserror.WithCode) {
	// Account should not mute or unmute itself.
	if requestingAccount.ID == targetAccountID {
		err := gtserror.Newf("account %s cannot mute or unmute itself", requestingAccount.ID)
		return nil, nil, gtserror.NewErrorNotAcceptable(err, err.Error())
	}

	// Ensure target account retrievable.
	targetAccount, err := p.state.DB.GetAccountByID(ctx, targetAccountID)
	if err != nil {
		if !errors.Is(err, db.ErrNoEntries) {
			// Real db error.
			err = gtserror.Newf("db error looking for target account %s: %w", targetAccountID, err)
			return nil, nil, gtserror.NewErrorInternalError(err)
		}
		// Account not found.
		err = gtserror.Newf("target account %s not found in the db", targetAccountID)
		return nil, nil, gtserror.NewErrorNotFound(err, err.Error())
	}

	// Check if currently muted.
	mute, err := p.state.DB.GetMute(ctx, requestingAccount.ID, targetAccountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error checking existing mute: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	return targetAccount, mute, nil
}
