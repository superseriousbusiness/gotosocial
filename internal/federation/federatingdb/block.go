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
	"net/http"

	"codeberg.org/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
)

func (f *federatingDB) Block(ctx context.Context, blockable vocab.ActivityStreamsBlock) error {
	log.DebugKV(ctx, "block", serialize{blockable})

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

	// Convert received AS block type to internal model.
	block, err := f.converter.ASBlockToBlock(ctx, blockable)
	if err != nil {
		err := gtserror.Newf("error converting from AS type: %w", err)
		return gtserror.WrapWithCode(http.StatusBadRequest, err)
	}

	// Ensure block enacted by correct account.
	if block.AccountID != requesting.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "requester %s is not expected actor %s",
			requesting.URI, block.Account.URI)
	}

	// Ensure block received by correct account.
	if block.TargetAccountID != receiving.ID {
		return gtserror.NewfWithCode(http.StatusForbidden, "receiver %s is not expected object %s",
			receiving.URI, block.TargetAccount.URI)
	}

	// Generate new ID for block.
	block.ID = id.NewULID()

	// Insert the new validated block into the database.
	if err := f.state.DB.PutBlock(ctx, block); err != nil {
		return gtserror.Newf("error inserting %s into db: %w", block.URI, err)
	}

	// Push message to worker queue to handle block side-effects.
	f.state.Workers.Federator.Queue.Push(&messages.FromFediAPI{
		APObjectType:   ap.ActivityBlock,
		APActivityType: ap.ActivityCreate,
		GTSModel:       block,
		Receiving:      receiving,
		Requesting:     requesting,
	})

	return nil
}
