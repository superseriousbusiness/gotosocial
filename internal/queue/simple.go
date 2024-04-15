package queue

import (
	"sync"

	"codeberg.org/gruf/go-list"
)

// SimpleQueue provides a simple concurrency safe
// queue using generics and a memory pool of list
// elements to reduce overall memory usage.
type SimpleQueue[T any] struct {
	l list.List[T]
	p []*list.Elem[T]
	w chan struct{}
	m sync.Mutex
}

// Push will push given value to the queue.
func (q *SimpleQueue[T]) Push(value T) {
	q.m.Lock()
	elem := q.alloc()
	elem.Value = value
	q.l.PushElemFront(elem)
	q.broadcast()
	q.m.Unlock()
}

// Pop will attempt to pop value from the queue.
func (q *SimpleQueue[T]) Pop() (value T, ok bool) {
	q.m.Lock()
	if ok = (q.l.Tail != nil); ok {
		tail := q.l.Tail
		value = tail.Value
		q.l.Remove(tail)
		q.free(tail)
	}
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

// Wait returns current wait channel, which may be
// blocked on to awaken when new value pushed to queue.
func (q *SimpleQueue[T]) Wait() (ch <-chan struct{}) {
	q.m.Lock()
	if q.w == nil {
		q.w = make(chan struct{})
	}
	ch = q.w
	q.m.Unlock()
	return
}

// alloc will allocate new list element (relying on memory pool).
func (q *SimpleQueue[T]) alloc() *list.Elem[T] {
	if len(q.p) > 0 {
		elem := q.p[len(q.p)-1]
		q.p = q.p[:len(q.p)-1]
		return elem
	}
	return new(list.Elem[T])
}

// free will free list element and release to pool.
func (q *SimpleQueue[T]) free(elem *list.Elem[T]) {
	var zero T
	elem.Next = nil
	elem.Prev = nil
	elem.Value = zero
	q.p = append(q.p, elem)
}

// broadcast safely closes wait channel if
// currently set, releasing waiting goroutines.
func (q *SimpleQueue[T]) broadcast() {
	if q.w != nil {
		close(q.w)
		q.w = nil
	}
}
