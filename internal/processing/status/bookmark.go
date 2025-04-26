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

package status

import (
	"context"
	"errors"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// BookmarkCreate adds a bookmark for the requestingAccount, targeting the given status (no-op if bookmark already exists).
func (p *Processor) BookmarkCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, existing, errWithCode := p.getBookmarkableStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existing != nil {
		// Status is already bookmarked.
		return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	}

	// Create and store a new bookmark.
	gtsBookmark := &gtsmodel.StatusBookmark{
		ID:              id.NewULID(),
		AccountID:       requestingAccount.ID,
		Account:         requestingAccount,
		TargetAccountID: targetStatus.AccountID,
		TargetAccount:   targetStatus.Account,
		StatusID:        targetStatus.ID,
		Status:          targetStatus,
	}

	if err := p.state.DB.PutStatusBookmark(ctx, gtsBookmark); err != nil {
		err = gtserror.Newf("error putting bookmark in database: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.c.InvalidateTimelinedStatus(ctx, requestingAccount.ID, targetStatusID); err != nil {
		err = gtserror.Newf("error invalidating status from timelines: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

// BookmarkRemove removes a bookmark for the requesting account, targeting the given status (no-op if bookmark doesn't exist).
func (p *Processor) BookmarkRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, existing, errWithCode := p.getBookmarkableStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existing == nil {
		// Status isn't bookmarked.
		return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	}

	// We have a bookmark to remove.
	if err := p.state.DB.DeleteStatusBookmarkByID(ctx, existing.ID); err != nil {
		err = gtserror.Newf("error removing status bookmark: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.c.InvalidateTimelinedStatus(ctx, requestingAccount.ID, targetStatusID); err != nil {
		err = gtserror.Newf("error invalidating status from timelines: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

func (p *Processor) getBookmarkableStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	statusID string,
) (
	*gtsmodel.Status,
	*gtsmodel.StatusBookmark,
	gtserror.WithCode,
) {
	target, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requester,
		statusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	bookmark, err := p.state.DB.GetStatusBookmark(
		gtscontext.SetBarebones(ctx),
		requester.ID,
		statusID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err := gtserror.Newf("error getting bookmark: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	return target, bookmark, nil
}
