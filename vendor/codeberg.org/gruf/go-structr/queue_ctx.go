package structr

import (
	"context"
)

// QueueCtx is a context-aware form of Queue{}.
type QueueCtx[StructType any] struct {
	ch chan struct{}
	Queue[StructType]
}

// PopFront pops the current value at front of the queue, else blocking on ctx.
func (q *QueueCtx[T]) PopFront(ctx context.Context) (T, bool) {
	return q.pop(ctx, func() *list_elem {
		return q.queue.head
	})
}

// PopBack pops the current value at back of the queue, else blocking on ctx.
func (q *QueueCtx[T]) PopBack(ctx context.Context) (T, bool) {
	return q.pop(ctx, func() *list_elem {
		return q.queue.tail
	})
}

// PushFront pushes values to front of queue.
func (q *QueueCtx[T]) PushFront(values ...T) {
	q.mutex.Lock()
	for i := range values {
		item := q.index(values[i])
		q.queue.push_front(&item.elem)
	}
	if q.ch != nil {
		close(q.ch)
		q.ch = nil
	}
	q.mutex.Unlock()
}

// PushBack pushes values to back of queue.
func (q *QueueCtx[T]) PushBack(values ...T) {
	q.mutex.Lock()
	for i := range values {
		item := q.index(values[i])
		q.queue.push_back(&item.elem)
	}
	if q.ch != nil {
		close(q.ch)
		q.ch = nil
	}
	q.mutex.Unlock()
}

// Wait returns a ptr to the current ctx channel,
// this will block until next push to the queue.
func (q *QueueCtx[T]) Wait() <-chan struct{} {
	q.mutex.Lock()
	if q.ch == nil {
		q.ch = make(chan struct{})
	}
	ctx := q.ch
	q.mutex.Unlock()
	return ctx
}

// Debug returns debug stats about queue.
func (q *QueueCtx[T]) Debug() map[string]any {
	m := make(map[string]any)
	q.mutex.Lock()
	m["queue"] = q.queue.len
	indices := make(map[string]any)
	m["indices"] = indices
	for i := range q.indices {
		var n uint64
		for _, l := range q.indices[i].data.m {
			n += uint64(l.len)
		}
		indices[q.indices[i].name] = n
	}
	q.mutex.Unlock()
	return m
}

func (q *QueueCtx[T]) pop(ctx context.Context, next func() *list_elem) (T, bool) {
	if next == nil {
		panic("nil fn")
	} else if ctx == nil {
		panic("nil ctx")
	}

	// Acquire lock.
	q.mutex.Lock()

	var elem *list_elem

	for {
		// Get element.
		elem = next()
		if elem != nil {
			break
		}

		if q.ch == nil {
			// Allocate new ctx channel.
			q.ch = make(chan struct{})
		}

		// Get current
		// ch pointer.
		ch := q.ch

		// Unlock queue.
		q.mutex.Unlock()

		select {
		// Ctx cancelled.
		case <-ctx.Done():
			var z T
			return z, false

		// Pushed!
		case <-ch:
		}

		// Relock queue.
		q.mutex.Lock()
	}

	// Cast the indexed item from elem.
	item := (*indexed_item)(elem.data)

	// Extract item value.
	value := item.data.(T)

	// Delete item.
	q.delete(item)

	// Get func ptrs.
	pop := q.Queue.pop

	// Done with lock.
	q.mutex.Unlock()

	if pop != nil {
		// Pass to
		// user hook.
		pop(value)
	}

	return value, true
}
