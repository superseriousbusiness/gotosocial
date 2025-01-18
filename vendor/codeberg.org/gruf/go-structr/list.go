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
	if list.head != nil ||
		list.tail != nil ||
		list.len != 0 {
		should_not_reach()
		return
	}
	list_pool.Put(list)
}

// push_front will push the given elem to front (head) of list.
func (l *list) push_front(elem *list_elem) {
	// Set new head.
	oldHead := l.head
	l.head = elem

	if oldHead != nil {
		// Link to old head
		elem.next = oldHead
		oldHead.prev = elem
	} else {
		// First in list.
		l.tail = elem
	}

	// Incr count
	l.len++
}

// push_back will push the given elem to back (tail) of list.
func (l *list) push_back(elem *list_elem) {
	// Set new tail.
	oldTail := l.tail
	l.tail = elem

	if oldTail != nil {
		// Link to old tail
		elem.prev = oldTail
		oldTail.next = elem
	} else {
		// First in list.
		l.head = elem
	}

	// Incr count
	l.len++
}

// move_front will move given elem to front (head) of list.
// if it is already at front this call is a no-op.
func (l *list) move_front(elem *list_elem) {
	if elem == l.head {
		return
	}
	l.remove(elem)
	l.push_front(elem)
}

// move_back will move given elem to back (tail) of list,
// if it is already at back this call is a no-op.
func (l *list) move_back(elem *list_elem) {
	if elem == l.tail {
		return
	}
	l.remove(elem)
	l.push_back(elem)
}

// remove will remove given elem from list.
func (l *list) remove(elem *list_elem) {
	// Get linked elems.
	next := elem.next
	prev := elem.prev

	// Unset elem.
	elem.next = nil
	elem.prev = nil

	switch {
	case next == nil:
		if prev == nil {
			// next == nil && prev == nil
			//
			// elem is ONLY one in list.
			l.head = nil
			l.tail = nil
		} else {
			// next == nil && prev != nil
			//
			// elem is last in list.
			l.tail = prev
			prev.next = nil
		}

	case prev == nil:
		// next != nil && prev == nil
		//
		// elem is front in list.
		l.head = next
		next.prev = nil

	// elem in middle of list.
	default:
		next.prev = prev
		prev.next = next
	}

	// Decr count
	l.len--
}
