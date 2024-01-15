package structr

// elem represents an element
// in a doubly-linked list.
type elem[T any] struct {
	next  *elem[T]
	prev  *elem[T]
	Value T
}

// list implements a doubly-linked list, where:
// - head = index 0   (i.e. the front)
// - tail = index n-1 (i.e. the back)
type list[T any] struct {
	head *elem[T]
	tail *elem[T]
	len  int
}

func list_acquire[T any](c *Cache[T]) *list[*result[T]] {
	var l *list[*result[T]]

	if len(c.llsPool) == 0 {
		// Allocate new list.
		l = new(list[*result[T]])
	} else {
		// Pop list from pool slice.
		l = c.llsPool[len(c.llsPool)-1]
		c.llsPool = c.llsPool[:len(c.llsPool)-1]
	}

	return l
}

func list_release[T any](c *Cache[T], l *list[*result[T]]) {
	// Reset list.
	l.head = nil
	l.tail = nil
	l.len = 0

	// Release list to memory pool.
	c.llsPool = append(c.llsPool, l)
}

// pushFront pushes new 'elem' to front of list.
func (l *list[T]) pushFront(elem *elem[T]) {
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

// moveFront calls remove() on elem, followed by pushFront().
func (l *list[T]) moveFront(elem *elem[T]) {
	l.remove(elem)
	l.pushFront(elem)
}

// remove removes the 'elem' from the list.
func (l *list[T]) remove(elem *elem[T]) {
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

// rangefn ranges all the elements in the list, passing each to fn.
func (l *list[T]) rangefn(fn func(*elem[T])) {
	if fn == nil {
		panic("nil fn")
	}
	elem := l.head
	for i := 0; i < l.len; i++ {
		fn(elem)
		elem = elem.next
	}
}
