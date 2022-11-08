package maps

// list is a doubly-linked list containing elemnts with key-value pairs of given generic parameter types.
type list[K comparable, V any] struct {
	head *elem[K, V]
	tail *elem[K, V]
	len  int
}

// Index returns the element at index from list.
func (l *list[K, V]) Index(idx int) *elem[K, V] {
	switch {
	// Idx in first half
	case idx < l.len/2:
		elem := l.head
		for i := 0; i < idx; i++ {
			elem = elem.next
		}
		return elem

	// Index in last half
	default:
		elem := l.tail
		for i := l.len - 1; i > idx; i-- {
			elem = elem.prev
		}
		return elem
	}
}

// PushFront will push the given element to the front of the list.
func (l *list[K, V]) PushFront(elem *elem[K, V]) {
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

// PushBack will push the given element to the back of the list.
func (l *list[K, V]) PushBack(elem *elem[K, V]) {
	if l.len == 0 {
		// Set new tail + head
		l.head = elem
		l.tail = elem

		// Link elem to itself
		elem.next = elem
		elem.prev = elem
	} else {
		oldTail := l.tail

		// Link up to head
		elem.next = l.head
		l.head.prev = elem

		// Link to old tail
		elem.prev = oldTail
		oldTail.next = elem

		// Set new tail
		l.tail = elem
	}

	// Incr count
	l.len++
}

// PopTail will pop the current tail of the list, returns nil if empty.
func (l *list[K, V]) PopTail() *elem[K, V] {
	if l.len == 0 {
		return nil
	}
	elem := l.tail
	l.Unlink(elem)
	return elem
}

// Unlink will unlink the given element from the doubly-linked list chain.
func (l *list[K, V]) Unlink(elem *elem[K, V]) {
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

// elem represents an element in a doubly-linked list.
type elem[K comparable, V any] struct {
	next *elem[K, V]
	prev *elem[K, V]
	K    K
	V    V
}

// allocElems will allocate a slice of empty elements of length.
func allocElems[K comparable, V any](i int) []*elem[K, V] {
	s := make([]*elem[K, V], i)
	for i := range s {
		s[i] = &elem[K, V]{}
	}
	return s
}
