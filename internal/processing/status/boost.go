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

	"github.com/superseriousbusiness/gotosocial/internal/ap"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/util"
)

// BoostCreate processes the boost/reblog of target
// status, returning the newly-created boost.
func (p *Processor) BoostCreate(
	ctx context.Context,
	requester *gtsmodel.Account,
	application *gtsmodel.Application,
	targetID string,
) (*apimodel.Status, gtserror.WithCode) {
	// Get target status and ensure it's not a boost.
	target, errWithCode := p.c.GetVisibleTargetStatus(
		ctx,
		requester,
		targetID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Unwrap target in case it is a boost.
	target, errWithCode = p.c.UnwrapIfBoost(
		ctx,
		requester,
		target,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check is viable target.
	if target.BoostOfID != "" {
		err := gtserror.Newf("target status %s is boost wrapper", target.URI)
		return nil, gtserror.NewErrorUnprocessableEntity(err)
	}

	// Ensure valid boost target for requester.
	policyResult, err := p.intFilter.StatusBoostable(ctx,
		requester,
		target,
	)
	if err != nil {
		err := gtserror.Newf("error seeing if status %s is boostable: %w", target.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if policyResult.Forbidden() {
		const errText = "you do not have permission to boost this status"
		err := gtserror.New(errText)
		return nil, gtserror.NewErrorForbidden(err, errText)
	}

	// Status is visible and boostable
	// (though maybe pending approval).
	boost, err := p.converter.StatusToBoost(ctx,
		target,
		requester,
		application.ID,
	)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Derive pendingApproval status.
	var pendingApproval bool
	switch {
	case policyResult.WithApproval():
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

		if *target.Local {
			// If the target is local we don't need
			// to wait for an Accept from remote,
			// we can just preapprove it and have
			// the processor create the Accept.
			boost.PreApproved = true
		}

	case policyResult.Permitted():
		// We're permitted to do this
		// based on another kind of match.
		pendingApproval = false
	}

	boost.PendingApproval = &pendingApproval

	// Store the new boost.
	if err := p.state.DB.PutStatus(ctx, boost); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	// Process side effects asynchronously.
	p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
		APObjectType:   ap.ActivityAnnounce,
		APActivityType: ap.ActivityCreate,
		GTSModel:       boost,
		Origin:         requester,
		Target:         target.Account,
	})

	// If the boost target status replies to a status
	// that we own, and has a pending interaction
	// request, use the boost as an implicit accept.
	implicitlyAccepted, errWithCode := p.implicitlyAccept(ctx,
		requester, target,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// If we ended up implicitly accepting, mark the
	// target status as no longer pending approval so
	// it's serialized properly via the API.
	if implicitlyAccepted {
		target.PendingApproval = util.Ptr(false)
	}

	return p.c.GetAPIStatus(ctx, requester, boost)
}

// BoostRemove processes the unboost/unreblog of
// target status, returning the target status.
func (p *Processor) BoostRemove(
	ctx context.Context,
	requester *gtsmodel.Account,
	application *gtsmodel.Application,
	targetID string,
) (*apimodel.Status, gtserror.WithCode) {
	// Get target status and ensure it's not a boost.
	target, errWithCode := p.c.GetVisibleTargetStatus(
		ctx,
		requester,
		targetID,
		nil, // default freshness
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	target, errWithCode = p.c.UnwrapIfBoost(
		ctx,
		requester,
		target,
	)
	if errWithCode != nil {
		return nil, errWithCode
	}

	// Check whether requester has actually
	// boosted target, by trying to get the boost.
	boost, err := p.state.DB.GetStatusBoost(ctx,
		target.ID,
		requester.ID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		err = gtserror.Newf("db error getting boost of %s: %w", target.ID, err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	if boost != nil {
		// Status was boosted. Process unboost side effects asynchronously.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APObjectType:   ap.ActivityAnnounce,
			APActivityType: ap.ActivityUndo,
			GTSModel:       boost,
			Origin:         requester,
			Target:         target.Account,
		})
	}

	return p.c.GetAPIStatus(ctx, requester, target)
}

// StatusBoostedBy returns a slice of accounts that have boosted the given status, filtered according to privacy settings.
func (p *Processor) StatusBoostedBy(ctx context.Context, requestingAccount *gtsmodel.Account, targetStatusID string) ([]*apimodel.Account, gtserror.WithCode) {
	targetStatus, err := p.state.DB.GetStatusByID(ctx, targetStatusID)
	if err != nil {
		wrapped := fmt.Errorf("BoostedBy: error fetching status %s: %s", targetStatusID, err)
		if !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.NewErrorInternalError(wrapped)
		}
		return nil, gtserror.NewErrorNotFound(wrapped)
	}

	if boostOfID := targetStatus.BoostOfID; boostOfID != "" {
		// the target status is a boost wrapper, redirect this request to the status it boosts
		boostedStatus, err := p.state.DB.GetStatusByID(ctx, boostOfID)
		if err != nil {
			wrapped := fmt.Errorf("BoostedBy: error fetching status %s: %s", boostOfID, err)
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.NewErrorInternalError(wrapped)
			}
			return nil, gtserror.NewErrorNotFound(wrapped)
		}
		targetStatus = boostedStatus
	}

	visible, err := p.visFilter.StatusVisible(ctx, requestingAccount, targetStatus)
	if err != nil {
		err = fmt.Errorf("BoostedBy: error seeing if status %s is visible: %s", targetStatus.ID, err)
		return nil, gtserror.NewErrorNotFound(err)
	}
	if !visible {
		err = errors.New("BoostedBy: status is not visible")
		return nil, gtserror.NewErrorNotFound(err)
	}

	statusBoosts, err := p.state.DB.GetStatusBoosts(ctx, targetStatus.ID)
	if err != nil {
		err = fmt.Errorf("BoostedBy: error seeing who boosted status: %s", err)
		return nil, gtserror.NewErrorNotFound(err)
	}

	// filter account IDs so the user doesn't see accounts they blocked or which blocked them
	accountIDs := make([]string, 0, len(statusBoosts))
	for _, s := range statusBoosts {
		blocked, err := p.state.DB.IsEitherBlocked(ctx, requestingAccount.ID, s.AccountID)
		if err != nil {
			err = fmt.Errorf("BoostedBy: error checking blocks: %s", err)
			return nil, gtserror.NewErrorNotFound(err)
		}
		if !blocked {
			accountIDs = append(accountIDs, s.AccountID)
		}
	}

	// TODO: filter other things here? suspended? muted? silenced?

	// fetch accounts + create their API representations
	apiAccounts := make([]*apimodel.Account, 0, len(accountIDs))
	for _, accountID := range accountIDs {
		account, err := p.state.DB.GetAccountByID(ctx, accountID)
		if err != nil {
			wrapped := fmt.Errorf("BoostedBy: error fetching account %s: %s", accountID, err)
			if !errors.Is(err, db.ErrNoEntries) {
				return nil, gtserror.NewErrorInternalError(wrapped)
			}
			return nil, gtserror.NewErrorNotFound(wrapped)
		}

		apiAccount, err := p.converter.AccountToAPIAccountPublic(ctx, account)
		if err != nil {
			err = fmt.Errorf("BoostedBy: error converting account to api model: %s", err)
			return nil, gtserror.NewErrorInternalError(err)
		}
		apiAccounts = append(apiAccounts, apiAccount)
	}

	return apiAccounts, nil
}
