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

	"codeberg.org/gruf/go-structr"
)

// StructQueue wraps a structr.Queue{} to
// provide simple index caching by name.
type StructQueue[StructType any] struct {
	queue structr.QueueCtx[StructType]
	index map[string]*structr.Index
}

// Init initializes queue with structr.QueueConfig{}.
func (q *StructQueue[T]) Init(config structr.QueueConfig[T]) {
	q.index = make(map[string]*structr.Index, len(config.Indices))
	// q.queue = structr.QueueCtx[T]{}
	q.queue.Init(config)
	for _, cfg := range config.Indices {
		q.index[cfg.Fields] = q.queue.Index(cfg.Fields)
	}
}

// Pop: see structr.Queue{}.PopFront().
func (q *StructQueue[T]) Pop() (value T, ok bool) {
	values := q.queue.PopFrontN(1)
	if ok = (len(values) > 0); !ok {
		return
	}
	value = values[0]
	return
}

// PopCtx: see structr.QueueCtx{}.PopFront().
func (q *StructQueue[T]) PopCtx(ctx context.Context) (value T, ok bool) {
	return q.queue.PopFront(ctx)
}

// Push: see structr.Queue.PushBack().
func (q *StructQueue[T]) Push(values ...T) {
	q.queue.PushBack(values...)
}

// Delete pops (and drops!) all queued entries under index with key.
func (q *StructQueue[T]) Delete(index string, key ...any) {
	_ = q.queue.Pop(q.index[index], structr.MakeKey(key...))
}

// Len: see structr.Queue{}.Len().
func (q *StructQueue[T]) Len() int {
	return q.queue.Len()
}

// Wait returns current wait channel, which may be
// blocked on to awaken when new value pushed to queue.
func (q *StructQueue[T]) Wait() <-chan struct{} {
	return q.queue.Wait()
}
