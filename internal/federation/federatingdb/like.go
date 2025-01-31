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

package federatingdb

import (
	"context"
	"errors"
	"net/http"

	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Like(ctx context.Context, likeable vocab.ActivityStreamsLike) error {
	log.DebugKV(ctx, "like", serialize{likeable})

	// Mark activity as handled.
	f.storeActivityID(likeable)

	// Extract relevant values from passed ctx.
	activityContext := getActivityContext(ctx)
	if activityContext.internal {
		return nil // Already processed.
	}

	requesting := activityContext.requestingAcct
	receiving := activityContext.receivingAcct

	if receiving.IsMoving() {
		// A Moving account
		// can't do this.
		return nil
	}

	// Convert received AS like type to internal fave model.
	fave, err := f.converter.ASLikeToFave(ctx, likeable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure fave enacted by correct account.
	if fave.AccountID != requesting.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "requester %s is not expected actor %s",
			requesting.URI, fave.Account.URI)
	}

	// Ensure fave received by correct account.
	if fave.TargetAccountID != receiving.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "receiver %s is not expected object %s",
			receiving.URI, fave.TargetAccount.URI)
	}

	if !*fave.Status.Local {
		// Only process likes of local statuses.
		// TODO: process for remote statuses as well.
		return nil
	}

	// Ensure valid Like target for requester.
	policyResult, err := f.intFilter.StatusLikeable(ctx,
		requesting,
		fave.Status,
	)
	if err != nil {
		return gtserror.Newf("error seeing if status %s is likeable: %w", fave.Status.URI, err)
	}

	if policyResult.Forbidden() {
		return gtserror.NewWithCode(http.StatusForbidden, "requester does not have permission to Like status")
	}

	// Derive pendingApproval
	// and preapproved status.
	var (
		pendingApproval bool
		preApproved     bool
	)

	switch {
	case policyResult.WithApproval():
		// Requester allowed to do
		// this pending approval.
		pendingApproval = true

	case policyResult.MatchedOnCollection():
		// Requester allowed to do this,
		// but matched on collection.
		// Preapprove Like and have the
		// processor send out an Accept.
		pendingApproval = true
		preApproved = true

	case policyResult.Permitted():
		// Requester straight up
		// permitted to do this,
		// no need for Accept.
		pendingApproval = false
	}

	// Set appropriate fields
	// on fave and store it.
	fave.ID = id.NewULID()
	fave.PendingApproval = &pendingApproval
	fave.PreApproved = preApproved

	if err := f.state.DB.PutStatusFave(ctx, fave); err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			// The fave already exists in the
			// database, which means we've already
			// handled side effects. We can just
			// return nil here and be done with it.
			return nil
		}
		return gtserror.Newf("error inserting %s into db: %w", fave.URI, err)
	}

	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityLike,
		APActivityType: ap.ActivityCreate,
		GTSModel:       fave,
		Receiving:      receiving,
		Requesting:     requesting,
	})

	return nil
}
