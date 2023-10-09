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

	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (p *Processor) ScheduleAll(ctx context.Context) error {
	var errs gtserror.MultiError
	var polls []*gtsmodel.Poll

	for _, poll := range polls {
		// Schedule each of the polls and catch any errors.
		if err := p.ScheduleExpiry(ctx, poll); err != nil {
			errs.Append(err)
		}
	}

	return errs.Combine()
}

func (p *Processor) ScheduleExpiry(ctx context.Context, poll *gtsmodel.Poll) error {
	// Ensure poll has a valid expiry time...
	if poll.ExpiresAt.After(time.Now()) {
		return gtserror.Newf("poll %s already expired", poll.ID)
	}

	// Add the given poll to the scheduler.
	ok := p.state.Workers.Scheduler.AddOnce(
		poll.ID,
		poll.ExpiresAt,
		p.onExpiry(poll),
	)

	if !ok {
		// Failed to add the poll to the scheduler, either it was
		// starting / stopping or there already exists a task for poll.
		return gtserror.Newf("failed adding poll %s to scheduler", poll.ID)
	}

	return nil
}

// onExpiry returns a callback function to be used by the scheduler when the given poll expires.
func (p *Processor) onExpiry(poll *gtsmodel.Poll) func(context.Context, time.Time) {
	return func(ctx context.Context, t time.Time) {
		panic("DO SOMETHING FUCKO")
	}
}
