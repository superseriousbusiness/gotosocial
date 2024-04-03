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
	"context"
	"log"
	"runtime"

	"codeberg.org/gruf/go-runners"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/messages"
	"github.com/superseriousbusiness/gotosocial/internal/queue"
	"github.com/superseriousbusiness/gotosocial/internal/scheduler"
)

type Workers struct {
	// Main task scheduler instance.
	Scheduler scheduler.Scheduler

	// HTTPClient ...
	HTTPClient httpclient.DeliveryWorkerPool

	// ClientAPI provides a worker pool that handles both
	// incoming client actions, and our own side-effects.
	ClientAPI runners.WorkerPool

	// Federator provides a worker pool that handles both
	// incoming federated actions, and our own side-effects.
	Federator runners.WorkerPool

	// Enqueue functions for clientAPI / federator worker pools,
	// these are pointers to Processor{}.Enqueue___() msg functions.
	// This prevents dependency cycling as Processor depends on Workers.
	EnqueueHTTPClient func(context.Context, ...queue.HTTPRequest)
	EnqueueClientAPI  func(context.Context, ...messages.FromClientAPI)
	EnqueueFediAPI    func(context.Context, ...messages.FromFediAPI)

	// Blocking processing functions for clientAPI / federator.
	// These are pointers to Processor{}.Process___() msg functions.
	// This prevents dependency cycling as Processor depends on Workers.
	//
	// Rather than queueing messages for asynchronous processing, these
	// functions will process immediately and in a blocking manner, and
	// will not use up a worker slot.
	//
	// As such, you should only call them in special cases where something
	// synchronous needs to happen before you can do something else.
	ProcessFromClientAPI func(context.Context, messages.FromClientAPI) error
	ProcessFromFediAPI   func(context.Context, messages.FromFediAPI) error

	// Media manager worker pools.
	Media runners.WorkerPool

	// prevent pass-by-value.
	_ nocopy
}

// Start will start all of the contained worker pools (and global scheduler).
func (w *Workers) Start() {
	// Get currently set GOMAXPROCS.
	maxprocs := runtime.GOMAXPROCS(0)

	tryUntil("starting scheduler", 5, w.Scheduler.Start)

	tryUntil("start http client workerpool", 5, w.HTTPClient.Start)

	tryUntil("starting client API workerpool", 5, func() bool {
		return w.ClientAPI.Start(4*maxprocs, 400*maxprocs)
	})

	tryUntil("starting federator workerpool", 5, func() bool {
		return w.Federator.Start(4*maxprocs, 400*maxprocs)
	})

	tryUntil("starting media workerpool", 5, func() bool {
		return w.Media.Start(8*maxprocs, 80*maxprocs)
	})
}

// Stop will stop all of the contained worker pools (and global scheduler).
func (w *Workers) Stop() {
	tryUntil("stopping scheduler", 5, w.Scheduler.Stop)
	tryUntil("stopping http client workerpool", 5, w.HTTPClient.Stop)
	tryUntil("stopping client API workerpool", 5, w.ClientAPI.Stop)
	tryUntil("stopping federator workerpool", 5, w.Federator.Stop)
	tryUntil("stopping media workerpool", 5, w.Media.Stop)
}

// nocopy when embedded will signal linter to
// error on pass-by-value of parent struct.
type nocopy struct{}

func (*nocopy) Lock() {}

func (*nocopy) Unlock() {}

// tryUntil will attempt to call 'do' for 'count' attempts, before panicking with 'msg'.
func tryUntil(msg string, count int, do func() bool) {
	for i := 0; i < count; i++ {
		if do() {
			return
		}
	}
	log.Panicf("failed %s after %d tries", msg, count)
}
