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
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
)

// BookmarkCreate adds a bookmark for the requestingAccount, targeting the given status (no-op if bookmark already exists).
func (p *Processor) BookmarkCreate(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, existingBookmarkID, errWithCode := p.getBookmarkTarget(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingBookmarkID != "" {
		// Status is already bookmarked.
		return p.apiStatus(ctx, targetStatus, requestingAccount)
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

	if err := p.invalidateStatus(ctx, requestingAccount.ID, targetStatusID); err != nil {
		err = gtserror.Newf("error invalidating status from timelines: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiStatus(ctx, targetStatus, requestingAccount)
}

// BookmarkRemove removes a bookmark for the requesting account, targeting the given status (no-op if bookmark doesn't exist).
func (p *Processor) BookmarkRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, existingBookmarkID, errWithCode := p.getBookmarkTarget(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingBookmarkID == "" {
		// Status isn't bookmarked.
		return p.apiStatus(ctx, targetStatus, requestingAccount)
	}

	// We have a bookmark to remove.
	if err := p.state.DB.DeleteStatusBookmark(ctx, existingBookmarkID); err != nil {
		err = gtserror.Newf("error removing status bookmark: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if err := p.invalidateStatus(ctx, requestingAccount.ID, targetStatusID); err != nil {
		err = gtserror.Newf("error invalidating status from timelines: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.apiStatus(ctx, targetStatus, requestingAccount)
}

func (p *Processor) getBookmarkTarget(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*gtsmodel.Status, string, gtserror.WithCode) {
	targetStatus, errWithCode := p.getVisibleStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, "", errWithCode
	}

	bookmarkID, err := p.state.DB.GetStatusBookmarkID(ctx, requestingAccount.ID, targetStatus.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("getBookmarkTarget: error checking existing bookmark: %w", err)
		return nil, "", gtserror.NewErrorInternalError(err)
	}

	return targetStatus, bookmarkID, nil
}
