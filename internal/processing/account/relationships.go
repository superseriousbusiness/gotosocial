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

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
)

// FollowersGet fetches a list of the target account's followers.
func (p *Processor) FollowersGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, page *paging.Page) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch target account to check it exists, and visibility of requester->target.
	targetAccount, errWithCode := p.c.GetVisibleTargetAccount(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if targetAccount.IsInstance() {
		// Instance accounts can't follow/be followed.
		return paging.EmptyResponse(), nil
	}

	// If account isn't requesting its own followers list,
	// but instead the list for a local account that has
	// hide_followers set, just return an empty array.
	if targetAccountID != requestingAccount.ID &&
		targetAccount.IsLocal() &&
		*targetAccount.Settings.HideCollections {
		return paging.EmptyResponse(), nil
	}

	follows, err := p.state.DB.GetAccountFollowers(ctx, targetAccountID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(follows)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := follows[count-1].ID
	hi := follows[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return follows[i].Account
	}

	// Get a filtered slice of public API account models.
	items := p.c.GetVisibleAPIAccountsPaged(ctx,
		requestingAccount,
		getIdx,
		len(follows),
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/accounts/" + targetAccountID + "/followers",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// FollowingGet fetches a list of the accounts that target account is following.
func (p *Processor) FollowingGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string, page *paging.Page) (*apimodel.PageableResponse, gtserror.WithCode) {
	// Fetch target account to check it exists, and visibility of requester->target.
	targetAccount, errWithCode := p.c.GetVisibleTargetAccount(ctx, requestingAccount, targetAccountID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if targetAccount.IsInstance() {
		// Instance accounts can't follow/be followed.
		return paging.EmptyResponse(), nil
	}

	// If account isn't requesting its own following list,
	// but instead the list for a local account that has
	// hide_followers set, just return an empty array.
	if targetAccountID != requestingAccount.ID &&
		targetAccount.IsLocal() &&
		*targetAccount.Settings.HideCollections {
		return paging.EmptyResponse(), nil
	}

	// Fetch known accounts that follow given target account ID.
	follows, err := p.state.DB.GetAccountFollows(ctx, targetAccountID, page)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting followers: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Check for empty response.
	count := len(follows)
	if count == 0 {
		return paging.EmptyResponse(), nil
	}

	// Get the lowest and highest
	// ID values, used for paging.
	lo := follows[count-1].ID
	hi := follows[0].ID

	// Func to fetch follow source at index.
	getIdx := func(i int) *gtsmodel.Account {
		return follows[i].TargetAccount
	}

	// Get a filtered slice of public API account models.
	items := p.c.GetVisibleAPIAccountsPaged(ctx,
		requestingAccount,
		getIdx,
		len(follows),
	)

	return paging.PackageResponse(paging.ResponseParams{
		Items: items,
		Path:  "/api/v1/accounts/" + targetAccountID + "/following",
		Next:  page.Next(lo, hi),
		Prev:  page.Prev(lo, hi),
	}), nil
}

// RelationshipGet returns a relationship model describing the relationship of the targetAccount to the Authed account.
func (p *Processor) RelationshipGet(ctx context.Context, requestingAccount *gtsmodel.Account, targetAccountID string) (*apimodel.Relationship, gtserror.WithCode) {
	if requestingAccount == nil {
		return nil, gtserror.NewErrorForbidden(gtserror.New("not authed"))
	}

	gtsR, err := p.state.DB.GetRelationship(ctx, requestingAccount.ID, targetAccountID)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(gtserror.Newf("error getting relationship: %s", err))
	}

	r, err := p.converter.RelationshipToAPIRelationship(ctx, gtsR)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(gtserror.Newf("error converting relationship: %s", err))
	}

	return r, nil
}
