package structr

import (
	"sync"
	"unsafe"
)

var list_pool sync.Pool

// elem represents an elem
// in a doubly-linked list.
type list_elem struct {
	next *list_elem
	prev *list_elem

	// data is a ptr to the
	// value this linked list
	// element is embedded-in.
	data unsafe.Pointer
}

// list implements a doubly-linked list, where:
// - head = index 0   (i.e. the front)
// - tail = index n-1 (i.e. the back)
type list struct {
	head *list_elem
	tail *list_elem
	len  int
}

func list_acquire() *list {
	// Acquire from pool.
	v := list_pool.Get()
	if v == nil {
		v = new(list)
	}

	// Cast list value.
	return v.(*list)
}

func list_release(l *list) {
	// Reset list.
	l.head = nil
	l.tail = nil
	l.len = 0

	// Release to pool.
	list_pool.Put(l)
}

func list_push_front(l *list, elem *list_elem) {
	if l.len == 0 {
		// Set new tail + head
		l.head = elem
		l.tail = elem

		// Link elem to itself
		elem.next = elem
		elem.prev = elem
	} else {
		oldHead := l.head

		// Link to old head
		elem.next = oldHead
		oldHead.prev = elem

		// Link up to tail
		elem.prev = l.tail
		l.tail.next = elem

		// Set new head
		l.head = elem
	}

	// Incr count
	l.len++
}

func list_move_front(l *list, elem *list_elem) {
	list_remove(l, elem)
	list_push_front(l, elem)
}

func list_remove(l *list, elem *list_elem) {
	if l.len <= 1 {
		// Drop elem's links
		elem.next = nil
		elem.prev = nil

		// Only elem in list
		l.head = nil
		l.tail = nil
		l.len = 0
		return
	}

	// Get surrounding elems
	next := elem.next
	prev := elem.prev

	// Relink chain
	next.prev = prev
	prev.next = next

	switch elem {
	// Set new head
	case l.head:
		l.head = next

	// Set new tail
	case l.tail:
		l.tail = prev
	}

	// Drop elem's links
	elem.next = nil
	elem.prev = nil

	// Decr count
	l.len--
}

func list_rangefn(l *list, fn func(*list_elem)) {
	if fn == nil {
		panic("nil fn")
	}
	elem := l.head
	for i := 0; i < l.len; i++ {
		fn(elem)
		elem = elem.next
	}
}
