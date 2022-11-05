package hashmap

import (
	"sync/atomic"
)

// ListElement is an element of a list.
type ListElement[Key comparable, Value any] struct {
	keyHash uintptr

	// deleted marks the item as deleting or deleted
	// this is using uintptr instead of atomic.Bool to avoid using 32 bit int on 64 bit systems
	deleted atomic.Uintptr

	// next points to the next element in the list.
	// it is nil for the last item in the list.
	next atomic.Pointer[ListElement[Key, Value]]

	value atomic.Pointer[Value]

	key Key
}

// Value returns the value of the list item.
func (e *ListElement[Key, Value]) Value() Value {
	return *e.value.Load()
}

// Next returns the item on the right.
func (e *ListElement[Key, Value]) Next() *ListElement[Key, Value] {
	for next := e.next.Load(); next != nil; {
		// if the next item is not deleted, return it
		if next.deleted.Load() == 0 {
			return next
		}

		// point current elements next to the following item
		// after the deleted one until a non deleted or list end is found
		following := next.Next()
		if e.next.CompareAndSwap(next, following) {
			next = following
		} else {
			next = next.Next()
		}
	}
	return nil // end of the list reached
}
