package structr

import (
	"sync"
	"unsafe"
)

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

var list_pool sync.Pool

// new_list returns a new prepared list.
func new_list() *list {
	v := list_pool.Get()
	if v == nil {
		v = new(list)
	}
	list := v.(*list)
	return list
}

// free_list releases the list.
func free_list(list *list) {
	list.head = nil
	list.tail = nil
	list.len = 0
	list_pool.Put(list)
}

// push_front will push the given elem to front (head) of list.
func (l *list) push_front(elem *list_elem) {
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

// push_back will push the given elem to back (tail) of list.
func (l *list) push_back(elem *list_elem) {
	if l.len == 0 {
		// Set new tail + head
		l.head = elem
		l.tail = elem

		// Link elem to itself
		elem.next = elem
		elem.prev = elem
	} else {
		oldTail := l.tail

		// Link to old tail
		elem.prev = oldTail
		oldTail.next = elem

		// Link up to head
		elem.next = l.head
		l.head.prev = elem

		// Set new tail
		l.tail = elem
	}

	// Incr count
	l.len++
}

// move_front will move given elem to front (head) of list.
func (l *list) move_front(elem *list_elem) {
	l.remove(elem)
	l.push_front(elem)
}

// move_back will move given elem to back (tail) of list.
func (l *list) move_back(elem *list_elem) {
	l.remove(elem)
	l.push_back(elem)
}

// remove will remove given elem from list.
func (l *list) remove(elem *list_elem) {
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

// rangefn will range all elems in list, passing each to fn.
func (l *list) rangefn(fn func(*list_elem)) {
	if fn == nil {
		panic("nil fn")
	}
	elem := l.head
	for i := 0; i < l.len; i++ {
		fn(elem)
		elem = elem.next
	}
}
