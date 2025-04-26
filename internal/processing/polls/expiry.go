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

package polls

import (
	"context"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
)

func (p *Processor) ScheduleAll(ctx context.Context) error {
	// Fetch all open polls from the database (barebones models are enough).
	polls, err := p.state.DB.GetOpenPolls(gtscontext.SetBarebones(ctx))
	if err != nil {
		return gtserror.Newf("error getting open polls from db: %w", err)
	}

	var errs gtserror.MultiError

	for _, poll := range polls {
		// Schedule each of the polls and catch any errors.
		if err := p.ScheduleExpiry(ctx, poll); err != nil {
			errs.Append(err)
		}
	}

	return errs.Combine()
}

func (p *Processor) ScheduleExpiry(ctx context.Context, poll *gtsmodel.Poll) error {
	// Ensure has a valid expiry.
	if !poll.ClosedAt.IsZero() {
		return gtserror.Newf("poll %s already expired", poll.ID)
	}

	// Add the given poll to the scheduler.
	ok := p.state.Workers.Scheduler.AddOnce(
		poll.ID,
		poll.ExpiresAt,
		p.onExpiry(poll.ID),
	)

	if !ok {
		// Failed to add the poll to the scheduler, either it was
		// starting / stopping or there already exists a task for poll.
		return gtserror.Newf("failed adding poll %s to scheduler", poll.ID)
	}

	atStr := poll.ExpiresAt.Local().Format("Jan _2 2006 15:04:05")
	log.Infof(ctx, "scheduled poll expiry for %s at '%s'", poll.ID, atStr)
	return nil
}

// onExpiry returns a callback function to be used by the scheduler when the given poll expires.
func (p *Processor) onExpiry(pollID string) func(context.Context, time.Time) {
	return func(ctx context.Context, now time.Time) {
		// Get the latest version of poll from database.
		poll, err := p.state.DB.GetPollByID(ctx, pollID)
		if err != nil {
			log.Errorf(ctx, "error getting poll %s from db: %v", pollID, err)
			return
		}

		if !poll.ClosedAt.IsZero() {
			// Expiry handler has already been run for this poll.
			log.Errorf(ctx, "poll %s already closed", pollID)
			return
		}

		// Extract status and
		// set its Poll field.
		status := poll.Status
		status.Poll = poll

		// Ensure the status is fully populated (we need the account)
		if err := p.state.DB.PopulateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error populating poll %s status: %v", pollID, err)

			if status.Account == nil {
				// cannot continue without
				// status account author.
				return
			}
		}

		// Set "closed" time.
		poll.ClosedAt = now
		poll.Closing = true

		// Update the Poll to mark it as closed in the database.
		if err := p.state.DB.UpdatePoll(ctx, poll, "closed_at"); err != nil {
			log.Errorf(ctx, "error updating poll %s in db: %v", pollID, err)
			return
		}

		// Enqueue a status update operation to the client API worker,
		// this will asynchronously send an update with the Poll close time.
		p.state.Workers.Client.Queue.Push(&messages.FromClientAPI{
			APActivityType: ap.ActivityUpdate,
			APObjectType:   ap.ObjectNote,
			GTSModel:       status,
			Origin:         status.Account,
		})
	}
}
