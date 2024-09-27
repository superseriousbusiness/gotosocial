package maps

import (
	"fmt"
	"reflect"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-kv"
)

// ordered provides a common ordered hashmap base, storing order in a doubly-linked list.
type ordered[K comparable, V any] struct {
	hmap map[K]*elem[K, V]
	list list[K, V]
	pool []*elem[K, V]
	rnly bool
}

// write_check panics if map is not in a safe-state to write to.
func (m *ordered[K, V]) write_check() {
	if m.rnly {
		panic("map write during read loop")
	}
}

// Has returns whether key exists in map.
func (m *ordered[K, V]) Has(key K) bool {
	_, ok := m.hmap[key]
	return ok
}

// Delete will delete given key from map, returns false if not found.
func (m *ordered[K, V]) Delete(key K) bool {
	// Ensure safe
	m.write_check()

	// Look for existing elem
	elem, ok := m.hmap[key]
	if !ok {
		return false
	}

	// Drop from list
	m.list.Unlink(elem)

	// Delete from map
	delete(m.hmap, key)

	// Return to pool
	m.free(elem)

	return true
}

// Range passes given function over the requested range of the map.
func (m *ordered[K, V]) Range(start, length int, fn func(int, K, V)) {
	// Nil check
	if fn == nil {
		panic("nil func")
	}

	// Disallow writes
	m.rnly = true
	defer func() {
		m.rnly = false
	}()

	switch end := start + length; {
	// No loop to iterate
	case length == 0:
		if start < 0 || (m.list.len > 0 && start >= m.list.len) {
			panic("index out of bounds")
		}

	// Step backwards
	case length < 0:
		// Check loop indices are within map bounds
		if end < -1 || start >= m.list.len || m.list.len == 0 {
			panic("index out of bounds")
		}

		// Get starting index elem
		elem := m.list.Index(start)

		for i := start; i > end; i-- {
			fn(i, elem.K, elem.V)
			elem = elem.prev
		}

	// Step forwards
	case length > 0:
		// Check loop indices are within map bounds
		if start < 0 || end > m.list.len || m.list.len == 0 {
			panic("index out of bounds")
		}

		// Get starting index elem
		elem := m.list.Index(start)

		for i := start; i < end; i++ {
			fn(i, elem.K, elem.V)
			elem = elem.next
		}
	}
}

// RangeIf passes given function over the requested range of the map. Returns early on 'fn' -> false.
func (m *ordered[K, V]) RangeIf(start, length int, fn func(int, K, V) bool) {
	// Nil check
	if fn == nil {
		panic("nil func")
	}

	// Disallow writes
	m.rnly = true
	defer func() {
		m.rnly = false
	}()

	switch end := start + length; {
	// No loop to iterate
	case length == 0:
		if start < 0 || (m.list.len > 0 && start >= m.list.len) {
			panic("index out of bounds")
		}

	// Step backwards
	case length < 0:
		// Check loop indices are within map bounds
		if end < -1 || start >= m.list.len || m.list.len == 0 {
			panic("index out of bounds")
		}

		// Get starting index elem
		elem := m.list.Index(start)

		for i := start; i > end; i-- {
			if !fn(i, elem.K, elem.V) {
				return
			}
			elem = elem.prev
		}

	// Step forwards
	case length > 0:
		// Check loop indices are within map bounds
		if start < 0 || end > m.list.len || m.list.len == 0 {
			panic("index out of bounds")
		}

		// Get starting index elem
		elem := m.list.Index(start)

		for i := start; i < end; i++ {
			if !fn(i, elem.K, elem.V) {
				return
			}
			elem = elem.next
		}
	}
}

// Truncate will truncate the map from the back by given amount, passing dropped elements to given function.
func (m *ordered[K, V]) Truncate(sz int, fn func(K, V)) {
	// Check size withing bounds
	if sz > m.list.len {
		panic("index out of bounds")
	}

	// Nil check
	if fn == nil {
		fn = func(K, V) {}
	}

	// Disallow writes
	m.rnly = true
	defer func() {
		m.rnly = false
	}()

	for i := 0; i < sz; i++ {
		// Pop current tail
		elem := m.list.tail
		m.list.Unlink(elem)

		// Delete from map
		delete(m.hmap, elem.K)

		// Pass dropped to func
		fn(elem.K, elem.V)

		// Release to pool
		m.free(elem)
	}
}

// Len returns the current length of the map.
func (m *ordered[K, V]) Len() int {
	return m.list.len
}

// format implements fmt.Formatter, allowing performant string formatting of map.
func (m *ordered[K, V]) format(rtype reflect.Type, state fmt.State, verb rune) {
	var (
		kvbuf byteutil.Buffer
		field kv.Field
		vbose bool
	)

	switch {
	// Only handle 'v' verb
	case verb != 'v':
		panic("invalid verb '" + string(verb) + "' for map")

	// Prefix with type when verbose
	case state.Flag('#'):
		state.Write([]byte(rtype.String()))
	}

	// Disallow writes
	m.rnly = true
	defer func() {
		m.rnly = false
	}()

	// Write map opening brace
	state.Write([]byte{'{'})

	if m.list.len > 0 {
		// Preallocate buffer
		kvbuf.Guarantee(64)

		// Start at index 0
		elem := m.list.head

		for i := 0; i < m.list.len-1; i++ {
			// Append formatted key-val pair to state
			field.K = fmt.Sprint(elem.K)
			field.V = elem.V
			field.AppendFormat(&kvbuf, vbose)
			_, _ = state.Write(kvbuf.B)
			kvbuf.Reset()

			// Prepare buffer with comma separator
			kvbuf.B = append(kvbuf.B, `, `...)

			// Jump to next in list
			elem = elem.next
		}

		// Append formatted key-val pair to state
		field.K = fmt.Sprint(elem.K)
		field.V = elem.V
		field.AppendFormat(&kvbuf, vbose)
		_, _ = state.Write(kvbuf.B)
	}

	// Write map closing brace
	state.Write([]byte{'}'})
}

// Std returns a clone of map's data in the standard library equivalent map type.
func (m *ordered[K, V]) Std() map[K]V {
	std := make(map[K]V, m.list.len)
	for _, elem := range m.hmap {
		std[elem.K] = elem.V
	}
	return std
}

// alloc will acquire list element from pool, or allocate new.
func (m *ordered[K, V]) alloc() *elem[K, V] {
	if len(m.pool) == 0 {
		return &elem[K, V]{}
	}
	idx := len(m.pool) - 1
	elem := m.pool[idx]
	m.pool = m.pool[:idx]
	return elem
}

// free will reset elem fields and place back in pool.
func (m *ordered[K, V]) free(elem *elem[K, V]) {
	var (
		zk K
		zv V
	)
	elem.K = zk
	elem.V = zv
	elem.next = nil
	elem.prev = nil
	m.pool = append(m.pool, elem)
}
