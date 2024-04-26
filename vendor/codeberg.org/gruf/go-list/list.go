package list

// Elem represents an element in a doubly-linked list.
type Elem[T any] struct {
	Next  *Elem[T]
	Prev  *Elem[T]
	Value T
}

// List implements a doubly-linked list, where:
// - Head = index 0   (i.e. the front)
// - Tail = index n-1 (i.e. the back)
type List[T any] struct {
	Head *Elem[T]
	Tail *Elem[T]
	len  int
}

// Len returns the current list length.
func (l *List[T]) Len() int {
	return l.len
}

// PushFront adds 'v' to the beginning of the list.
func (l *List[T]) PushFront(v T) *Elem[T] {
	elem := &Elem[T]{Value: v}
	l.PushElemFront(elem)
	return elem
}

// PushBack adds 'v' to the end of the list.
func (l *List[T]) PushBack(v T) *Elem[T] {
	elem := &Elem[T]{Value: v}
	l.PushElemBack(elem)
	return elem
}

// InsertBefore adds 'v' into the list before 'at'.
func (l *List[T]) InsertBefore(v T, at *Elem[T]) *Elem[T] {
	elem := &Elem[T]{Value: v}
	l.InsertElemBefore(elem, at)
	return elem
}

// InsertAfter adds 'v' into the list after 'at'.
func (l *List[T]) InsertAfter(v T, at *Elem[T]) *Elem[T] {
	elem := &Elem[T]{Value: v}
	l.InsertElemAfter(elem, at)
	return elem
}

// PushFrontNode adds 'elem' to the front of the list.
func (l *List[T]) PushElemFront(elem *Elem[T]) {
	if elem == l.Head {
		return
	}

	// Set new head.
	oldHead := l.Head
	l.Head = elem

	if oldHead != nil {
		// Link to old head
		elem.Next = oldHead
		oldHead.Prev = elem
	} else {
		// First in list.
		l.Tail = elem
	}

	// Incr count
	l.len++
}

// PushBackNode adds 'elem' to the back of the list.
func (l *List[T]) PushElemBack(elem *Elem[T]) {
	if elem == l.Tail {
		return
	}

	// Set new tail.
	oldTail := l.Tail
	l.Tail = elem

	if oldTail != nil {
		// Link to old tail
		elem.Prev = oldTail
		oldTail.Next = elem
	} else {
		// First in list.
		l.Head = elem
	}

	// Incr count
	l.len++
}

// InsertElemAfter adds 'elem' into the list after 'at' (i.e. at.Next = elem).
func (l *List[T]) InsertElemAfter(elem *Elem[T], at *Elem[T]) {
	if elem == at {
		return
	}

	// Set new 'next'.
	oldNext := at.Next
	at.Next = elem

	// Link to 'at'.
	elem.Prev = at

	if oldNext == nil {
		// Set new tail
		l.Tail = elem
	} else {
		// Link to 'prev'.
		oldNext.Prev = elem
		elem.Next = oldNext
	}

	// Incr count
	l.len++
}

// InsertElemBefore adds 'elem' into the list before 'at' (i.e. at.Prev = elem).
func (l *List[T]) InsertElemBefore(elem *Elem[T], at *Elem[T]) {
	if elem == at {
		return
	}

	// Set new 'prev'.
	oldPrev := at.Prev
	at.Prev = elem

	// Link to 'at'.
	elem.Next = at

	if oldPrev == nil {
		// Set new head
		l.Head = elem
	} else {
		// Link to 'next'.
		oldPrev.Next = elem
		elem.Prev = oldPrev
	}

	// Incr count
	l.len++
}

// Remove removes the 'elem' from the list.
func (l *List[T]) Remove(elem *Elem[T]) {
	// Get linked elems.
	next := elem.Next
	prev := elem.Prev

	// Unset elem.
	elem.Next = nil
	elem.Prev = nil

	switch {
	// elem is ONLY one in list.
	case next == nil && prev == nil:
		l.Head = nil
		l.Tail = nil

	// elem is front in list.
	case next != nil && prev == nil:
		l.Head = next
		next.Prev = nil

	// elem is last in list.
	case prev != nil && next == nil:
		l.Tail = prev
		prev.Next = nil

	// elem in middle of list.
	default:
		next.Prev = prev
		prev.Next = next
	}

	// Decr count
	l.len--
}

// Range calls 'fn' on every element from head forward in list.
func (l *List[T]) Range(fn func(*Elem[T])) {
	if fn == nil {
		panic("nil function")
	}
	for elem := l.Head; elem != nil; elem = elem.Next {
		fn(elem)
	}
}

// RangeReverse calls 'fn' on every element from tail backward in list.
func (l *List[T]) RangeReverse(fn func(*Elem[T])) {
	if fn == nil {
		panic("nil function")
	}
	for elem := l.Tail; elem != nil; elem = elem.Prev {
		fn(elem)
	}
}
