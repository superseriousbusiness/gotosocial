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

package workers

import (
	"runtime"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/scheduler"
	"github.com/superseriousbusiness/gotosocial/internal/transport/delivery"
)

type Workers struct {
	// Main task scheduler instance.
	Scheduler scheduler.Scheduler

	// Delivery provides a worker pool that
	// handles outgoing ActivityPub deliveries.
	// It contains an embedded (but accessible)
	// indexed queue of Delivery{} objects.
	Delivery delivery.WorkerPool

	// Client provides a worker pool that handles
	// incoming processing jobs from the client API.
	Client MsgWorkerPool[*messages.FromClientAPI]

	// Federator provides a worker pool that handles
	// incoming processing jobs from the fedi API.
	Federator MsgWorkerPool[*messages.FromFediAPI]

	// Dereference provides a worker pool
	// for asynchronous dereferencer jobs.
	Dereference FnWorkerPool

	// Media provides a worker pool for
	// asynchronous media processing jobs.
	Media FnWorkerPool

	// prevent pass-by-value.
	_ nocopy
}

// StartScheduler starts the job scheduler.
func (w *Workers) StartScheduler() {
	_ = w.Scheduler.Start() // false = already running
}

// Start will start contained worker pools.
func (w *Workers) Start() {
	maxprocs := runtime.GOMAXPROCS(0)
	w.Delivery.Start(deliveryWorkers(maxprocs))
	w.Client.Start(4 * maxprocs)
	w.Federator.Start(4 * maxprocs)
	w.Dereference.Start(4 * maxprocs)
	w.Media.Start(8 * maxprocs)
}

// Stop will stop all of the contained worker pools (and global scheduler).
func (w *Workers) Stop() {
	_ = w.Scheduler.Stop() // false = not running
	w.Delivery.Stop()
	w.Client.Stop()
	w.Federator.Stop()
	w.Dereference.Stop()
	w.Media.Stop()
}

// nocopy when embedded will signal linter to
// error on pass-by-value of parent struct.
type nocopy struct{}

func (*nocopy) Lock() {}

func (*nocopy) Unlock() {}

func deliveryWorkers(maxprocs int) int {
	n := config.GetAdvancedSenderMultiplier()
	if n < 1 {
		// clamp to 1
		return 1
	}
	return n * maxprocs
}
