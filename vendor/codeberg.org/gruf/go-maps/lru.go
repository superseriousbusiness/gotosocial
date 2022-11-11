package maps

import (
	"fmt"
	"reflect"
)

// LRU provides an ordered hashmap implementation that keeps elements ordered according to last recently used (hence, LRU).
type LRUMap[K comparable, V any] struct {
	ordered[K, V]
	size int
}

// NewLRU returns a new instance of LRUMap with given initializing length and maximum capacity.
func NewLRU[K comparable, V any](len, cap int) *LRUMap[K, V] {
	m := new(LRUMap[K, V])
	m.Init(len, cap)
	return m
}

// Init will initialize this map with initial length and maximum capacity.
func (m *LRUMap[K, V]) Init(len, cap int) {
	if cap <= 0 {
		panic("lru cap must be greater than zero")
	} else if m.pool != nil {
		panic("lru map already initialized")
	}
	m.ordered.hmap = make(map[K]*elem[K, V], len)
	m.ordered.pool = allocElems[K, V](len)
	m.size = cap
}

// Get will fetch value for given key from map, in the process pushing it to the front of the map. Returns false if not found.
func (m *LRUMap[K, V]) Get(key K) (V, bool) {
	if elem, ok := m.hmap[key]; ok {
		// Ensure safe
		m.write_check()

		// Unlink elem from list
		m.list.Unlink(elem)

		// Push to front of list
		m.list.PushFront(elem)

		return elem.V, true
	}
	var z V // zero value
	return z, false
}

// Add will add the given key-value pair to the map, pushing them to the front of the map. Returns false if already exists. Evicts old at maximum capacity.
func (m *LRUMap[K, V]) Add(key K, value V) bool {
	return m.AddWithHook(key, value, nil)
}

// AddWithHook performs .Add() but passing any evicted entry to given hook function.
func (m *LRUMap[K, V]) AddWithHook(key K, value V, evict func(K, V)) bool {
	// Ensure safe
	m.write_check()

	// Look for existing elem
	elem, ok := m.hmap[key]
	if ok {
		return false
	}

	if m.list.len >= m.size {
		// We're at capacity, sir!
		// Pop current tail elem
		elem = m.list.PopTail()

		if evict != nil {
			// Pass to evict hook
			evict(elem.K, elem.V)
		}

		// Delete key from map
		delete(m.hmap, elem.K)
	} else {
		// Allocate elem
		elem = m.alloc()
	}

	// Set elem
	elem.K = key
	elem.V = value

	// Add element map entry
	m.hmap[key] = elem

	// Push to front of list
	m.list.PushFront(elem)
	return true
}

// Set will ensure that given key-value pair exists in the map, by either adding new or updating existing, pushing them to the front of the map. Evicts old at maximum capacity.
func (m *LRUMap[K, V]) Set(key K, value V) {
	m.SetWithHook(key, value, nil)
}

// SetWithHook performs .Set() but passing any evicted entry to given hook function.
func (m *LRUMap[K, V]) SetWithHook(key K, value V, evict func(K, V)) {
	// Ensure safe
	m.write_check()

	// Look for existing elem
	elem, ok := m.hmap[key]

	if ok {
		// Unlink elem from list
		m.list.Unlink(elem)

		// Update existing
		elem.V = value
	} else {
		if m.list.len >= m.size {
			// We're at capacity, sir!
			// Pop current tail elem
			elem = m.list.PopTail()

			if evict != nil {
				// Pass to evict hook
				evict(elem.K, elem.V)
			}

			// Delete key from map
			delete(m.hmap, elem.K)
		} else {
			// Allocate elem
			elem = m.alloc()
		}

		// Set elem
		elem.K = key
		elem.V = value

		// Add element map entry
		m.hmap[key] = elem
	}

	// Push to front of list
	m.list.PushFront(elem)
}

// Cap returns the maximum capacity of this LRU map.
func (m *LRUMap[K, V]) Cap() int {
	return m.size
}

// Format implements fmt.Formatter, allowing performant string formatting of map.
func (m *LRUMap[K, V]) Format(state fmt.State, verb rune) {
	m.format(reflect.TypeOf(m), state, verb)
}
