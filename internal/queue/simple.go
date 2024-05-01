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

package queue

import (
	"context"
	"sync"

	"codeberg.org/gruf/go-list"
)

// frequency of GC cycles
// per no. unlocks. i.e.
// every 'gcfreq' unlocks.
const gcfreq = 1024

// SimpleQueue provides a simple concurrency safe
// queue using generics and a memory pool of list
// elements to reduce overall memory usage.
type SimpleQueue[T any] struct {
	l list.List[T]
	p elemPool[T]
	w chan struct{}
	m sync.Mutex
	n uint32 // pop counter (safely wraps around)
}

// Push will push given value to the queue.
func (q *SimpleQueue[T]) Push(value T) {
	q.m.Lock()

	// Wrap in element.
	elem := q.p.alloc()
	elem.Value = value

	// Push new elem to queue.
	q.l.PushElemFront(elem)

	if q.w != nil {
		// Notify any goroutines
		// blocking on q.Wait(),
		// or on PopCtx(...).
		close(q.w)
		q.w = nil
	}

	q.m.Unlock()
}

// Pop will attempt to pop value from the queue.
func (q *SimpleQueue[T]) Pop() (value T, ok bool) {
	q.m.Lock()

	// Check for a tail (i.e. not empty).
	if ok = (q.l.Tail != nil); ok {

		// Extract value.
		tail := q.l.Tail
		value = tail.Value

		// Remove tail.
		q.l.Remove(tail)
		q.p.free(tail)

		// Every 'gcfreq' pops perform
		// a garbage collection to keep
		// us squeaky clean :]
		if q.n++; q.n%gcfreq == 0 {
			q.p.GC()
		}
	}

	q.m.Unlock()
	return
}

// PopCtx will attempt to pop value from queue, else blocking on context.
func (q *SimpleQueue[T]) PopCtx(ctx context.Context) (value T, ok bool) {

	// Acquire lock.
	q.m.Lock()

	var elem *list.Elem[T]

	for {
		// Get next elem.
		elem = q.l.Tail
		if ok = (elem != nil); ok {
			break
		}

		if q.w == nil {
			// Create new wait channel.
			q.w = make(chan struct{})
		}

		// Get current
		// ch pointer.
		ch := q.w

		// Done with lock.
		q.m.Unlock()

		select {
		// Context canceled.
		case <-ctx.Done():
			return

		// Pushed!
		case <-ch:
		}

		// Relock queue.
		q.m.Lock()
	}

	// Extract value.
	value = elem.Value

	// Remove element.
	q.l.Remove(elem)
	q.p.free(elem)

	// Every 'gcfreq' pops perform
	// a garbage collection to keep
	// us squeaky clean :]
	if q.n++; q.n%gcfreq == 0 {
		q.p.GC()
	}

	// Done with lock.
	q.m.Unlock()

	return
}

// Len returns the current length of the queue.
func (q *SimpleQueue[T]) Len() int {
	q.m.Lock()
	l := q.l.Len()
	q.m.Unlock()
	return l
}

// elemPool is a very simple
// list.Elem[T] memory pool.
type elemPool[T any] struct {
	current []*list.Elem[T]
	victim  []*list.Elem[T]
}

func (p *elemPool[T]) alloc() *list.Elem[T] {
	// First try the current queue
	if l := len(p.current) - 1; l >= 0 {
		mu := p.current[l]
		p.current = p.current[:l]
		return mu
	}

	// Next try the victim queue.
	if l := len(p.victim) - 1; l >= 0 {
		mu := p.victim[l]
		p.victim = p.victim[:l]
		return mu
	}

	// Lastly, alloc new.
	mu := new(list.Elem[T])
	return mu
}

// free will release given element to pool.
func (p *elemPool[T]) free(elem *list.Elem[T]) {
	var zero T
	elem.Next = nil
	elem.Prev = nil
	elem.Value = zero
	p.current = append(p.current, elem)
}

// GC will clear out unused entries from the elemPool.
func (p *elemPool[T]) GC() {
	current := p.current
	p.current = nil
	p.victim = current
}
