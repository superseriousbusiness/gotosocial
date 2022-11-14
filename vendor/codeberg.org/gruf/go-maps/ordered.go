package maps

import (
	"fmt"
	"reflect"
)

// OrderedMap provides a hashmap implementation that tracks the order in which keys are added.
type OrderedMap[K comparable, V any] struct {
	ordered[K, V]
}

// NewOrdered returns a new instance of LRUMap with given initializing length and maximum capacity.
func NewOrdered[K comparable, V any](len int) *OrderedMap[K, V] {
	m := new(OrderedMap[K, V])
	m.Init(len)
	return m
}

// Init will initialize this map with initial length.
func (m *OrderedMap[K, V]) Init(len int) {
	if m.pool != nil {
		panic("ordered map already initialized")
	}
	m.ordered.hmap = make(map[K]*elem[K, V], len)
	m.ordered.pool = allocElems[K, V](len)
}

// Get will fetch value for given key from map. Returns false if not found.
func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	if elem, ok := m.hmap[key]; ok {
		return elem.V, true
	}
	var z V // zero value
	return z, false
}

// Add will add the given key-value pair to the map, returns false if already exists.
func (m *OrderedMap[K, V]) Add(key K, value V) bool {
	// Ensure safe
	m.write_check()

	// Look for existing elem
	elem, ok := m.hmap[key]
	if ok {
		return false
	}

	// Allocate elem
	elem = m.alloc()
	elem.K = key
	elem.V = value

	// Add element map entry
	m.hmap[key] = elem

	// Push to back of list
	m.list.PushBack(elem)
	return true
}

// Set will ensure that given key-value pair exists in the map, by either adding new or updating existing.
func (m *OrderedMap[K, V]) Set(key K, value V) {
	// Ensure safe
	m.write_check()

	// Look for existing elem
	elem, ok := m.hmap[key]

	if ok {
		// Update existing
		elem.V = value
	} else {
		// Allocate elem
		elem = m.alloc()
		elem.K = key
		elem.V = value

		// Add element map entry
		m.hmap[key] = elem

		// Push to back of list
		m.list.PushBack(elem)
	}
}

// Index returns the key-value pair at index from map. Returns false if index out of range.
func (m *OrderedMap[K, V]) Index(idx int) (K, V, bool) {
	if idx < 0 || idx >= m.list.len {
		var (
			zk K
			zv V
		) // zero values
		return zk, zv, false
	}
	elem := m.list.Index(idx)
	return elem.K, elem.V, true
}

// Push will insert the given key-value pair at index in the map. Panics if index out of range.
func (m *OrderedMap[K, V]) Push(idx int, key K, value V) {
	// Check index within bounds of map
	if idx < 0 || idx >= m.list.len {
		panic("index out of bounds")
	}

	// Ensure safe
	m.write_check()

	// Get element at index
	next := m.list.Index(idx)

	// Allocate new elem
	elem := m.alloc()
	elem.K = key
	elem.V = value

	// Add element map entry
	m.hmap[key] = elem

	// Move next forward
	elem.next = next
	elem.prev = next.prev

	// Link up elem in chain
	next.prev.next = elem
	next.prev = elem
}

// Pop will remove and return the key-value pair at index in the map. Panics if index out of range.
func (m *OrderedMap[K, V]) Pop(idx int) (K, V) {
	// Check index within bounds of map
	if idx < 0 || idx >= m.list.len {
		panic("index out of bounds")
	}

	// Ensure safe
	m.write_check()

	// Get element at index
	elem := m.list.Index(idx)

	// Unlink elem from list
	m.list.Unlink(elem)

	// Get elem values
	k := elem.K
	v := elem.V

	// Release to pool
	m.free(elem)

	return k, v
}

// Format implements fmt.Formatter, allowing performant string formatting of map.
func (m *OrderedMap[K, V]) Format(state fmt.State, verb rune) {
	m.format(reflect.TypeOf(m), state, verb)
}
