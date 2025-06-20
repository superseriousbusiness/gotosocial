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

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

func (p *Processor) getFaveableStatus(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetID string,
) (
	*gtsmodel.Status,
	*gtsmodel.StatusFave,
	gtserror.WithCode,
) {
	// Get target status and ensure it's not a boost.
	target, errWithCode := p.c.GetVisibleTargetStatus(
		ctx,
		requester,
		targetID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	target, errWithCode = p.c.UnwrapIfBoost(
		ctx,
		requester,
		target,
	)
	if errWithCode != nil {
		return nil, nil, errWithCode
	}

	fave, err := p.state.DB.GetStatusFave(ctx, requester.ID, target.ID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = fmt.Errorf("getFaveTarget: error checking existing fave: %w", err)
		return nil, nil, gtserror.NewErrorInternalError(err)
	}

	return target, fave, nil
}

// FaveCreate adds a fave for the requestingAccount, targeting the given status (no-op if fave already exists).
func (p *Processor) FaveCreate(
	ctx context.Context,
	requester *gtsmodel.Account,
	targetStatusID string,
) (*apimodel.Status, gtserror.WithCode) {
	status, existingFave, errWithCode := p.getFaveableStatus(ctx, requester, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingFave != nil {
		// Status is already faveed.
		return p.c.GetAPIStatus(ctx, requester, status)
	}

	// Ensure valid fave target for requester.
	policyResult, err := p.intFilter.StatusLikeable(ctx,
		requester,
		status,
	)
	if err != nil {
		err := gtserror.Newf("error seeing if status %s is likeable: %w", status.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if policyResult.Forbidden() {
		const errText = "you do not have permission to fave this status"
		err := gtserror.New(errText)
		return nil, gtserror.NewErrorForbidden(err, errText)
	}

	// Derive pendingApproval
	// and preapproved status.
	var (
		pendingApproval bool
		preApproved     bool
	)

	switch {
	case policyResult.ManualApproval():
		// We're allowed to do
		// this pending approval.
		pendingApproval = true

	case policyResult.MatchedOnCollection():
		// We're permitted to do this, but since
		// we matched due to presence in a followers
		// or following collection, we should mark
		// as pending approval and wait until we can
		// prove it's been Accepted by the target.
		pendingApproval = true

		if *status.Local {
			// If the target is local we don't need
			// to wait for an Accept from remote,
			// we can just preapprove it and have
			// the processor create the Accept.
			preApproved = true
		}

	case policyResult.AutomaticApproval():
		// We're permitted to do this
		// based on another kind of match.
		pendingApproval = false
	}

	// Create a new fave, marking it
	// as pending approval if necessary.
	faveID := id.NewULID()
	gtsFave := &gtsmodel.StatusFave{
		ID:              faveID,
		AccountID:       requester.ID,
		Account:         requester,
		TargetAccountID: status.AccountID,
		TargetAccount:   status.Account,
		StatusID:        status.ID,
		Status:          status,
		URI:             uris.GenerateURIForLike(requester.Username, faveID),
		PreApproved:     preApproved,
		PendingApproval: &pendingApproval,
	}

	if err := p.state.DB.PutStatusFave(ctx, gtsFave); err != nil {
		err = gtserror.Newf("db error putting fave: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// If the fave target status replies to a status
	// that we own, and has a pending interaction
	// request, use the fave as an implicit accept.
	implicitlyAccepted, errWithCode := p.implicitlyAccept(ctx,
		requester, status,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// If we ended up implicitly accepting, mark the
	// target status as no longer pending approval so
	// it's serialized properly via the API.
	if implicitlyAccepted {
		status.PendingApproval = util.Ptr(false)
	}

	// Queue remaining fave side effects
	// (send out fave, update timeline, etc).
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityCreate,
		GTSModel:       gtsFave,
		Origin:         requester,
		Target:         status.Account,
	})

	return p.c.GetAPIStatus(ctx, requester, status)
}

// FaveRemove removes a fave for the requesting account, targeting the given status (no-op if fave doesn't exist).
func (p *Processor) FaveRemove(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) (*apimodel.Status, gtserror.WithCode) {
	targetStatus, existingFave, errWithCode := p.getFaveableStatus(ctx, requestingAccount, targetStatusID)
	if errWithCode != nil {
		return nil, errWithCode
	}

	if existingFave == nil {
		// Status isn't faveed.
		return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
	}

	// We have a fave to remove.
	if err := p.state.DB.DeleteStatusFaveByID(ctx, existingFave.ID); err != nil {
		err = fmt.Errorf("FaveRemove: error removing status fave: %w", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process remove status fave side effects.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityUndo,
		GTSModel:       existingFave,
		Origin:         requestingAccount,
		Target:         targetStatus.Account,
	})

	return p.c.GetAPIStatus(ctx, requestingAccount, targetStatus)
}

// FavedBy returns a slice of accounts that have liked the given status, filtered according to privacy settings.
func (p *Processor) FavedBy(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, errWithCode := p.c.GetVisibleTargetStatus(ctx,
		requestingAccount,
		targetStatusID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	statusFaves, err := p.state.DB.GetStatusFaves(ctx, targetStatus.ID)
	if err != nil {
		return nil, gtserror.NewErrorNotFound(fmt.Errorf("FavedBy: error seeing who faved status: %s", err))
	}

	// For each fave, ensure that we're only showing
	// the requester accounts that they don't block,
	// and which don't block them.
	apiAccounts := make([]*apimodel.Account, 0, len(statusFaves))
	for _, fave := range statusFaves {
		if blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, fave.AccountID); err != nil {
			err = fmt.Errorf("FavedBy: error checking blocks: %w", err)
			return nil, gtserror.NewErrorInternalError(err)
		} else if blocked {
			continue
		}

		if fave.Account == nil {
			// Account isn't set for some reason, just skip.
			log.WithContext(ctx).WithField("fave", fave).Warn("fave had no associated account")
			continue
		}

		apiAccount, err := p.converter.AccountToAPIAccountPublic(ctx, fave.Account)
		if err != nil {
			err = fmt.Errorf("FavedBy: error converting account %s to frontend representation: %w", fave.AccountID, err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}
