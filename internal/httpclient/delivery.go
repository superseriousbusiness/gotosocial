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

package httpclient

import (
	"context"
	"slices"
	"time"

	"codeberg.org/gruf/go-runners"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/queue"
)

// APDeliveryWorkerPool ...
type APDeliveryWorkerPool struct {
	workers []APDeliveryWorker
}

// Init will initialize the DeliveryWorker{} pool
// with given http client, request queue to pull
// from and number of delivery workers to spawn.
func (p *APDeliveryWorkerPool) Init(
	client *Client,
	queue *queue.StructQueue[*queue.APRequest],
	workers int,
) {
	p.workers = make([]APDeliveryWorker, workers)
	for i := range p.workers {
		p.workers[i] = NewAPDeliveryWorker(
			client,
			queue,
		)
	}
}

// Start will attempt to start all of the contained DeliveryWorker{}s.
// NOTE: this is not safe to call concurrently with .Init().
func (p *APDeliveryWorkerPool) Start() bool {
	if len(p.workers) == 0 {
		return false
	}
	ok := true
	for i := range p.workers {
		ok = p.workers[i].Start() && ok
	}
	return ok
}

// Stop will attempt to stop all of the contained DeliveryWorker{}s.
// NOTE: this is not safe to call concurrently with .Init().
func (p *APDeliveryWorkerPool) Stop() bool {
	if len(p.workers) == 0 {
		return false
	}
	ok := true
	for i := range p.workers {
		ok = p.workers[i].Stop() && ok
	}
	return ok
}

// APDeliveryWorker ...
type APDeliveryWorker struct {
	client  *Client
	queue   *queue.StructQueue[*queue.APRequest]
	backlog []*delivery
	service runners.Service
}

// NewAPDeliveryWorker returns a new APDeliveryWorker that feeds from queue, using given HTTP client.
func NewAPDeliveryWorker(client *Client, queue *queue.StructQueue[*queue.APRequest]) APDeliveryWorker {
	return APDeliveryWorker{
		client:  client,
		queue:   queue,
		backlog: make([]*delivery, 0, 256),
	}
}

// Start will attempt to start the DeliveryWorker{}.
func (w *APDeliveryWorker) Start() bool {
	return w.service.GoRun(w.process)
}

// Stop will attempt to stop the DeliveryWorker{}.
func (w *APDeliveryWorker) Stop() bool {
	return w.service.Stop()
}

// process is the main delivery worker processing routine.
func (w *APDeliveryWorker) process(ctx context.Context) {
	if w.client == nil || w.queue == nil {
		panic("nil delivery worker fields")
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

			// Start backoff sleep timer.
			backoff, cncl := sleepch(d)

			select {
			case <-ctx.Done():
				// Main ctx
				// cancelled.
				cncl()

			case <-w.queue.Wait():
				// A new message was
				// queued, re-add this
				// to backlog + retry.
				w.pushBacklog(dlv)
				cncl()
				continue loop

			case <-backoff:
				// successful
				// backoff!
			}
		}

		dlv.log.Info("performing request")

		// Attempt outoing delivery of request.
		_, retry, err := w.client.do(&dlv.request)
		if err == nil {
			continue loop
		}

		dlv.log.Error(err)

		if !retry || dlv.attempts > maxRetries {
			// Drop deliveries when no retry
			// requested, or we reach max.
			continue loop
		}

		// Determine next delivery attempt.
		dlv.next = time.Now().Add(dlv.BackOff())

		// Push to backlog.
		w.pushBacklog(dlv)
	}
}

// next gets the next available delivery, blocking until available if necessary.
func (w *APDeliveryWorker) next(ctx context.Context) (*delivery, bool) {
	// Try pop next queued.
	msg, ok := w.queue.Pop()

	if !ok {
		// Check the backlog.
		if len(w.backlog) > 0 {

			// Sort by 'next' time.
			sortDeliveries(w.backlog)

			// Pop next delivery.
			dlv := w.popBacklog()

			return dlv, true
		}

		// Backlog is empty, we MUST
		// block until next enqueued.
		msg, ok = w.queue.PopCtx(ctx)
		if !ok {
			return nil, false
		}
	}

	// Wrap msg in delivery type.
	return wrapMsg(ctx, msg), true
}

// popBacklog pops next available from the backlog.
func (w *APDeliveryWorker) popBacklog() *delivery {
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
func (w *APDeliveryWorker) pushBacklog(dlv *delivery) {
	w.backlog = append(w.backlog, dlv)
}

// delivery wraps request{}
// to cache logging fields.
type delivery struct {

	// cached log
	// entry fields.
	log log.Entry

	// next attempt time.
	next time.Time

	// embedded
	// request.
	request
}

// BackOff returns backoff duration to sleep for, calculated
// from the .next attempt field subtracted from current time.
func (d *delivery) BackOff() time.Duration {
	if d.next.IsZero() {
		return 0
	}
	return time.Now().Sub(d.next)
}

// wrapMsg wraps a received queued AP request message in our delivery type.
func wrapMsg(ctx context.Context, msg *queue.APRequest) *delivery {
	dlv := new(delivery)
	dlv.request = wrapRequest(msg.Request)
	dlv.log = requestLog(dlv.req)
	ctx = gtscontext.WithValues(ctx, msg.Request.Context())
	dlv.req = dlv.req.WithContext(ctx)
	return dlv
}

// sortDeliveries sorts deliveries according
// to when is the first requiring re-attempt.
func sortDeliveries(d []*delivery) {
	slices.SortFunc(d, func(a, b *delivery) int {
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