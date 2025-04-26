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
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
)

// getMuteableStatus fetches targetStatusID status and
// ensures that requestingAccount can mute or unmute it.
//
// It checks:
//   - Status exists and is visible to requester.
//   - Status belongs to or mentions requesting account.
//   - Status is not a boost.
//   - Status has a thread ID.
func (p *Processor) getMuteableStatus(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetStatusID string,
) (*gtsmodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if !targetStatus.BelongsToAccount(requestingAccount.ID) &&
		!targetStatus.MentionsAccount(requestingAccount.ID) {
		err := gtserror.Newf("status %s does not belong to or mention account %s", targetStatusID, requestingAccount.ID)
		return nil, gtserror.NewErrorNotFound(err)
	}

	if targetStatus.BoostOfID != "" {
		err := gtserror.New("cannot mute or unmute boosts")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	if targetStatus.ThreadID == "" {
		err := gtserror.New("cannot mute or unmute status with no threadID")
		return nil, gtserror.NewErrorBadRequest(err, err.Error())
	}

	return targetStatus, nil
}

func (p *Processor) MuteCreate(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetStatusID string,
) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.getMuteableStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	var (
		threadID  = targetStatus.ThreadID
		accountID = requestingAccount.ID
	)

	// Check if mute already exists for this thread ID.
	threadMute, err := p.state.DB.GetThreadMutedByAccount(ctx, threadID, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real db error.
		err := gtserror.Newf("db error fetching mute of thread %s for account %s", threadID, accountID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if threadMute != nil {
		// Thread mute already exists.
		// Our job here is done ("but you didn't do anything!").
		return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	}

	// Gotta create a mute.
	if err := p.state.DB.PutThreadMute(ctx, &gtsmodel.ThreadMute{
		ID:        id.NewULID(),
		ThreadID:  threadID,
		AccountID: accountID,
	}); err != nil {
		err := gtserror.Newf("db error putting mute of thread %s for account %s", threadID, accountID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

func (p *Processor) MuteRemove(
	ctx context.Context,
	requestingAccount *gtsmodel.Account,
	targetStatusID string,
) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, errWithCode := p.getMuteableStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	var (
		threadID  = targetStatus.ThreadID
		accountID = requestingAccount.ID
	)

	// Check if mute exists for this thread ID.
	threadMute, err := p.state.DB.GetThreadMutedByAccount(ctx, threadID, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		// Real db error.
		err := gtserror.Newf("db error fetching mute of thread %s for account %s", threadID, accountID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if threadMute == nil {
		// Thread mute doesn't exist.
		// Our job here is done ("but you didn't do anything!").
		return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	}

	// Gotta remove the mute.
	if err := p.state.DB.DeleteThreadMute(ctx, threadMute.ID); err != nil {
		err := gtserror.Newf("db error deleting mute of thread %s for account %s", threadID, accountID)
		return nil, gtserror.NewErrorInternalError(err)
	}

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}
