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

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/messages"
	"code.superseriousbusiness.org/gotosocial/internal/scheduler"
	"code.superseriousbusiness.org/gotosocial/internal/transport/delivery"
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

	// Processing provides a worker pool
	// for asynchronous processing jobs,
	// eg., import tasks, admin tasks.
	Processing FnWorkerPool

	// WebPush provides a worker pool for
	// delivering Web Push notifications.
	WebPush FnWorkerPool

	// prevent pass-by-value.
	_ nocopy
}

// StartScheduler starts the job scheduler.
func (w *Workers) StartScheduler() {
	_ = w.Scheduler.Start()
	// false = already running
	log.Info(nil, "started scheduler")
}

// Start will start contained worker pools.
func (w *Workers) Start() {
	var n int

	maxprocs := runtime.GOMAXPROCS(0)

	n = deliveryWorkers(maxprocs)
	w.Delivery.Start(n)
	log.Infof(nil, "started %d delivery workers", n)

	n = 4 * maxprocs
	w.Client.Start(n)
	log.Infof(nil, "started %d client workers", n)

	n = 4 * maxprocs
	w.Federator.Start(n)
	log.Infof(nil, "started %d federator workers", n)

	n = 4 * maxprocs
	w.Dereference.Start(n)
	log.Infof(nil, "started %d dereference workers", n)

	n = maxprocs
	w.Processing.Start(n)
	log.Infof(nil, "started %d processing workers", n)

	n = maxprocs
	w.WebPush.Start(n)
	log.Infof(nil, "started %d Web Push workers", n)
}

// Stop will stop all of the contained
// worker pools (and global scheduler).
func (w *Workers) Stop() {
	_ = w.Scheduler.Stop()
	// false = not running
	log.Info(nil, "stopped scheduler")

	w.Delivery.Stop()
	log.Info(nil, "stopped delivery workers")

	w.Client.Stop()
	log.Info(nil, "stopped client workers")

	w.Federator.Stop()
	log.Info(nil, "stopped federator workers")

	w.Dereference.Stop()
	log.Info(nil, "stopped dereference workers")

	w.Processing.Stop()
	log.Info(nil, "stopped processing workers")

	w.WebPush.Stop()
	log.Info(nil, "stopped WebPush workers")
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
