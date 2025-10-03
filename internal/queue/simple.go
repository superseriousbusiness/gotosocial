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
	"unsafe"

	"codeberg.org/gruf/go-list"
	"codeberg.org/gruf/go-mempool"
)

// SimpleQueue provides a simple concurrency safe
// queue using generics and a memory pool of list
// elements to reduce overall memory usage.
type SimpleQueue[T any] struct {
	l list.List[T]
	p mempool.UnsafeSimplePool
	w chan struct{}
	m sync.Mutex
}

// Push will push given value to the queue.
func (q *SimpleQueue[T]) Push(value T) {
	q.m.Lock()

	// Wrap in element.
	elem := q.acquire()
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
		q.release(tail)
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
	q.release(elem)

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

// acquire will acquire list elem from pool, else alloc new.
func (q *SimpleQueue[T]) acquire() *list.Elem[T] {
	if ptr := q.p.Get(); ptr != nil {
		return (*list.Elem[T])(ptr)
	}
	return new(list.Elem[T])
}

// release will reset list elem and release to pool.
func (q *SimpleQueue[T]) release(e *list.Elem[T]) {
	var zero T
	e.Next = nil
	e.Prev = nil
	e.Value = zero
	ptr := unsafe.Pointer(e)
	q.p.Put(ptr)
}
