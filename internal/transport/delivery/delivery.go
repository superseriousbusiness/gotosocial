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

package delivery

import (
	"context"
	"slices"
	"time"

	"codeberg.org/gruf/go-runners"
	"codeberg.org/gruf/go-structr"
	"github.com/superseriousbusiness/gotosocial/internal/httpclient"
	"github.com/superseriousbusiness/gotosocial/internal/queue"
)

type Delivery struct {
	// ActorID ...
	ActorID string

	// ObjectID ...
	ObjectID string

	// TargetID ...
	TargetID string

	// Request ...
	Request httpclient.Request

	// internal fields.
	next time.Time
}

// BackOff ...
func (dlv *Delivery) BackOff() time.Duration {
	if dlv.next.IsZero() {
		return 0
	}
	return time.Until(dlv.next)
}

// WorkerPool wraps multiple Worker{}s in
// a singular struct for easy multi start/stop.
type WorkerPool struct {

	// Client defines httpclient.Client{}
	// passed to each of delivery pool Worker{}s.
	Client *httpclient.Client

	// Queue is the embedded queue.StructQueue{}
	// passed to each of delivery pool Worker{}s.
	Queue queue.StructQueue[*Delivery]

	// internal fields.
	workers []Worker
}

// Init will initialize the Worker{} pool
// with given http client, request queue to pull
// from and number of delivery workers to spawn.
func (p *WorkerPool) Init(client *httpclient.Client) {
	p.Client = client
	p.Queue.Init(structr.QueueConfig[*Delivery]{
		Indices: []structr.IndexConfig{
			{Fields: "ActorID", Multiple: true},
			{Fields: "ObjectID", Multiple: true},
			{Fields: "TargetID", Multiple: true},
		},
	})
}

// Start will attempt to start 'n' Worker{}s.
func (p *WorkerPool) Start(n int) (ok bool) {
	if ok = (len(p.workers) == 0); ok {
		p.workers = make([]Worker, n)
		for i := range p.workers {
			p.workers[i].Client = p.Client
			p.workers[i].Queue = &p.Queue
			ok = p.workers[i].Start() && ok
		}
	}
	return
}

// Stop will attempt to stop contained Worker{}s.
func (p *WorkerPool) Stop() (ok bool) {
	if ok = (len(p.workers) > 0); ok {
		for i := range p.workers {
			ok = p.workers[i].Stop() && ok
		}
		p.workers = p.workers[:0]
	}
	return
}

// Worker wraps an httpclient.Client{} to feed
// from queue.StructQueue{} for ActivityPub reqs
// to deliver. It does so while prioritizing new
// queued requests over backlogged retries.
type Worker struct {

	// Client is the httpclient.Client{} that
	// delivery worker will use for requests.
	Client *httpclient.Client

	// Queue is the Delivery{} message queue
	// that delivery worker will feed from.
	Queue *queue.StructQueue[*Delivery]

	// internal fields.
	backlog []*Delivery
	service runners.Service
}

// Start will attempt to start the Worker{}.
func (w *Worker) Start() bool {
	return w.service.GoRun(w.process)
}

// Stop will attempt to stop the Worker{}.
func (w *Worker) Stop() bool {
	return w.service.Stop()
}

// process is the main delivery worker processing routine.
func (w *Worker) process(ctx context.Context) {
	if w.Client == nil || w.Queue == nil {
		panic("not yet initialized")
	}

loop:
	for {
		// Get next delivery.
		dlv, ok := w.next(ctx)
		if !ok {
			return
		}

		// Check whether backoff required.
		if d := dlv.BackOff(); d != 0 {

			// Get queue wait ch.
			queue := w.Queue.Wait()

			// Start backoff sleep timer.
			backoff := time.NewTimer(d)

			select {
			case <-ctx.Done():
				// Main ctx
				// cancelled.
				backoff.Stop()

			case <-queue:
				// A new message was
				// queued, re-add this
				// to backlog + retry.
				w.pushBacklog(dlv)
				backoff.Stop()
				continue loop

			case <-backoff.C:
				// success!
			}
		}

		// Attempt delivery of AP request.
		rsp, retry, err := w.Client.DoOnce(
			&dlv.Request,
		)

		if err == nil {
			// Ensure body closed.
			_ = rsp.Body.Close()
			continue loop
		}

		if !retry {
			// Drop deliveries when no
			// retry requested, or they
			// reached max (either).
			continue loop
		}

		// Determine next delivery attempt.
		dlv.next = time.Now().Add(dlv.BackOff())

		// Push to backlog.
		w.pushBacklog(dlv)
	}
}

// next gets the next available delivery, blocking until available if necessary.
func (w *Worker) next(ctx context.Context) (*Delivery, bool) {
loop:
	for {
		// Try pop next queued.
		dlv, ok := w.Queue.Pop()

		if !ok {
			// Check the backlog.
			if len(w.backlog) > 0 {

				// Sort by 'next' time.
				sortDeliveries(w.backlog)

				// Pop next delivery.
				dlv := w.popBacklog()

				return dlv, true
			}

			select {
			// Backlog is empty, we MUST
			// block until next enqueued.
			case <-w.Queue.Wait():
				continue loop

			// Worker was stopped.
			case <-ctx.Done():
				return nil, false
			}
		}

		// Replace request context for worker state canceling.
		dlv.Request.Request = dlv.Request.WithContext(ctx)

		return dlv, true
	}
}

// popBacklog pops next available from the backlog.
func (w *Worker) popBacklog() *Delivery {
	if len(w.backlog) == 0 {
		return nil
	}

	// Pop from backlog.
	dlv := w.backlog[0]

	// Shift backlog down by one.
	copy(w.backlog, w.backlog[1:])
	w.backlog = w.backlog[:len(w.backlog)-1]

	return dlv
}

// pushBacklog pushes the given delivery to backlog.
func (w *Worker) pushBacklog(dlv *Delivery) {
	w.backlog = append(w.backlog, dlv)
}

// sortDeliveries sorts deliveries according
// to when is the first requiring re-attempt.
func sortDeliveries(d []*Delivery) {
	slices.SortFunc(d, func(a, b *Delivery) int {
		const k = +1
		switch {
		case a.next.Before(b.next):
			return +k
		case b.next.Before(a.next):
			return -k
		default:
			return 0
		}
	})
}
