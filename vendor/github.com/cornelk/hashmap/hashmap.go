// Package hashmap provides a lock-free and thread-safe HashMap.
package hashmap

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"unsafe"
)

// Map implements a read optimized hash map.
type Map[Key hashable, Value any] struct {
	hasher     func(Key) uintptr
	store      atomic.Pointer[store[Key, Value]] // pointer to a map instance that gets replaced if the map resizes
	linkedList *List[Key, Value]                 // key sorted linked list of elements
	// resizing marks a resizing operation in progress.
	// this is using uintptr instead of atomic.Bool to avoid using 32 bit int on 64 bit systems
	resizing atomic.Uintptr
}

// New returns a new map instance.
func New[Key hashable, Value any]() *Map[Key, Value] {
	return NewSized[Key, Value](defaultSize)
}

// NewSized returns a new map instance with a specific initialization size.
func NewSized[Key hashable, Value any](size uintptr) *Map[Key, Value] {
	m := &Map[Key, Value]{}
	m.allocate(size)
	m.setDefaultHasher()
	return m
}

// SetHasher sets a custom hasher.
func (m *Map[Key, Value]) SetHasher(hasher func(Key) uintptr) {
	m.hasher = hasher
}

// Len returns the number of elements within the map.
func (m *Map[Key, Value]) Len() int {
	return m.linkedList.Len()
}

// Get retrieves an element from the map under given hash key.
func (m *Map[Key, Value]) Get(key Key) (Value, bool) {
	hash := m.hasher(key)

	for element := m.store.Load().item(hash); element != nil; element = element.Next() {
		if element.keyHash == hash && element.key == key {
			return element.Value(), true
		}

		if element.keyHash > hash {
			return *new(Value), false
		}
	}
	return *new(Value), false
}

// GetOrInsert returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The returned bool is true if the value was loaded, false if stored.
func (m *Map[Key, Value]) GetOrInsert(key Key, value Value) (Value, bool) {
	hash := m.hasher(key)
	var newElement *ListElement[Key, Value]

	for {
		for element := m.store.Load().item(hash); element != nil; element = element.Next() {
			if element.keyHash == hash && element.key == key {
				actual := element.Value()
				return actual, true
			}

			if element.keyHash > hash {
				break
			}
		}

		if newElement == nil { // allocate only once
			newElement = &ListElement[Key, Value]{
				key:     key,
				keyHash: hash,
			}
			newElement.value.Store(&value)
		}

		if m.insertElement(newElement, hash, key, value) {
			return value, false
		}
	}
}

// FillRate returns the fill rate of the map as a percentage integer.
func (m *Map[Key, Value]) FillRate() int {
	store := m.store.Load()
	count := int(store.count.Load())
	l := len(store.index)
	return (count * 100) / l
}

// Del deletes the key from the map and returns whether the key was deleted.
func (m *Map[Key, Value]) Del(key Key) bool {
	hash := m.hasher(key)
	store := m.store.Load()
	element := store.item(hash)

	for ; element != nil; element = element.Next() {
		if element.keyHash == hash && element.key == key {
			m.deleteElement(element)
			m.linkedList.Delete(element)
			return true
		}

		if element.keyHash > hash {
			return false
		}
	}
	return false
}

// Insert sets the value under the specified key to the map if it does not exist yet.
// If a resizing operation is happening concurrently while calling Insert, the item might show up in the map
// after the resize operation is finished.
// Returns true if the item was inserted or false if it existed.
func (m *Map[Key, Value]) Insert(key Key, value Value) bool {
	hash := m.hasher(key)
	var (
		existed, inserted bool
		element           *ListElement[Key, Value]
	)

	for {
		store := m.store.Load()
		searchStart := store.item(hash)

		if !inserted { // if retrying after insert during grow, do not add to list again
			element, existed, inserted = m.linkedList.Add(searchStart, hash, key, value)
			if existed {
				return false
			}
			if !inserted {
				continue // a concurrent add did interfere, try again
			}
		}

		count := store.addItem(element)
		currentStore := m.store.Load()
		if store != currentStore { // retry insert in case of insert during grow
			continue
		}

		if m.isResizeNeeded(store, count) && m.resizing.CompareAndSwap(0, 1) {
			go m.grow(0, true)
		}
		return true
	}
}

// Set sets the value under the specified key to the map. An existing item for this key will be overwritten.
// If a resizing operation is happening concurrently while calling Set, the item might show up in the map
// after the resize operation is finished.
func (m *Map[Key, Value]) Set(key Key, value Value) {
	hash := m.hasher(key)

	for {
		store := m.store.Load()
		searchStart := store.item(hash)

		element, added := m.linkedList.AddOrUpdate(searchStart, hash, key, value)
		if !added {
			continue // a concurrent add did interfere, try again
		}

		count := store.addItem(element)
		currentStore := m.store.Load()
		if store != currentStore { // retry insert in case of insert during grow
			continue
		}

		if m.isResizeNeeded(store, count) && m.resizing.CompareAndSwap(0, 1) {
			go m.grow(0, true)
		}
		return
	}
}

// Grow resizes the map to a new size, the size gets rounded up to next power of 2.
// To double the size of the map use newSize 0.
// This function returns immediately, the resize operation is done in a goroutine.
// No resizing is done in case of another resize operation already being in progress.
func (m *Map[Key, Value]) Grow(newSize uintptr) {
	if m.resizing.CompareAndSwap(0, 1) {
		go m.grow(newSize, true)
	}
}

// String returns the map as a string, only hashed keys are printed.
func (m *Map[Key, Value]) String() string {
	buffer := bytes.NewBufferString("")
	buffer.WriteRune('[')

	first := m.linkedList.First()
	item := first

	for item != nil {
		if item != first {
			buffer.WriteRune(',')
		}
		fmt.Fprint(buffer, item.keyHash)
		item = item.Next()
	}
	buffer.WriteRune(']')
	return buffer.String()
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map[Key, Value]) Range(f func(Key, Value) bool) {
	item := m.linkedList.First()

	for item != nil {
		value := item.Value()
		if !f(item.key, value) {
			return
		}
		item = item.Next()
	}
}

func (m *Map[Key, Value]) allocate(newSize uintptr) {
	m.linkedList = NewList[Key, Value]()
	if m.resizing.CompareAndSwap(0, 1) {
		m.grow(newSize, false)
	}
}

func (m *Map[Key, Value]) isResizeNeeded(store *store[Key, Value], count uintptr) bool {
	l := uintptr(len(store.index)) // l can't be 0 as it gets initialized in New()
	fillRate := (count * 100) / l
	return fillRate > maxFillRate
}

func (m *Map[Key, Value]) insertElement(element *ListElement[Key, Value], hash uintptr, key Key, value Value) bool {
	var existed, inserted bool

	for {
		store := m.store.Load()
		searchStart := store.item(element.keyHash)

		if !inserted { // if retrying after insert during grow, do not add to list again
			_, existed, inserted = m.linkedList.Add(searchStart, hash, key, value)
			if existed {
				return false
			}

			if !inserted {
				continue // a concurrent add did interfere, try again
			}
		}

		count := store.addItem(element)
		currentStore := m.store.Load()
		if store != currentStore { // retry insert in case of insert during grow
			continue
		}

		if m.isResizeNeeded(store, count) && m.resizing.CompareAndSwap(0, 1) {
			go m.grow(0, true)
		}
		return true
	}
}

// deleteElement deletes an element from index.
func (m *Map[Key, Value]) deleteElement(element *ListElement[Key, Value]) {
	for {
		store := m.store.Load()
		index := element.keyHash >> store.keyShifts
		ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(store.array) + index*intSizeBytes))

		next := element.Next()
		if next != nil && element.keyHash>>store.keyShifts != index {
			next = nil // do not set index to next item if it's not the same slice index
		}
		atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(element), unsafe.Pointer(next))

		currentStore := m.store.Load()
		if store == currentStore { // check that no resize happened
			break
		}
	}
}

func (m *Map[Key, Value]) grow(newSize uintptr, loop bool) {
	defer m.resizing.CompareAndSwap(1, 0)

	for {
		currentStore := m.store.Load()
		if newSize == 0 {
			newSize = uintptr(len(currentStore.index)) << 1
		} else {
			newSize = roundUpPower2(newSize)
		}

		index := make([]*ListElement[Key, Value], newSize)
		header := (*reflect.SliceHeader)(unsafe.Pointer(&index))

		newStore := &store[Key, Value]{
			keyShifts: strconv.IntSize - log2(newSize),
			array:     unsafe.Pointer(header.Data), // use address of slice data storage
			index:     index,
		}

		m.fillIndexItems(newStore) // initialize new index slice with longer keys

		m.store.Store(newStore)

		m.fillIndexItems(newStore) // make sure that the new index is up-to-date with the current state of the linked list

		if !loop {
			return
		}

		// check if a new resize needs to be done already
		count := uintptr(m.Len())
		if !m.isResizeNeeded(newStore, count) {
			return
		}
		newSize = 0 // 0 means double the current size
	}
}

func (m *Map[Key, Value]) fillIndexItems(store *store[Key, Value]) {
	first := m.linkedList.First()
	item := first
	lastIndex := uintptr(0)

	for item != nil {
		index := item.keyHash >> store.keyShifts
		if item == first || index != lastIndex { // store item with smallest hash key for every index
			store.addItem(item)
			lastIndex = index
		}
		item = item.Next()
	}
}
