package hashmap

import (
	"sync/atomic"
	"unsafe"
)

type store[Key comparable, Value any] struct {
	keyShifts uintptr                    // Pointer size - log2 of array size, to be used as index in the data array
	count     atomic.Uintptr             // count of filled elements in the slice
	array     unsafe.Pointer             // pointer to slice data array
	index     []*ListElement[Key, Value] // storage for the slice for the garbage collector to not clean it up
}

// item returns the item for the given hashed key.
func (s *store[Key, Value]) item(hashedKey uintptr) *ListElement[Key, Value] {
	index := hashedKey >> s.keyShifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(s.array) + index*intSizeBytes))
	item := (*ListElement[Key, Value])(atomic.LoadPointer(ptr))
	return item
}

// adds an item to the index if needed and returns the new item counter if it changed, otherwise 0.
func (s *store[Key, Value]) addItem(item *ListElement[Key, Value]) uintptr {
	index := item.keyHash >> s.keyShifts
	ptr := (*unsafe.Pointer)(unsafe.Pointer(uintptr(s.array) + index*intSizeBytes))

	for { // loop until the smallest key hash is in the index
		element := (*ListElement[Key, Value])(atomic.LoadPointer(ptr)) // get the current item in the index
		if element == nil {                                            // no item yet at this index
			if atomic.CompareAndSwapPointer(ptr, nil, unsafe.Pointer(item)) {
				return s.count.Add(1)
			}
			continue // a new item was inserted concurrently, retry
		}

		if item.keyHash < element.keyHash {
			// the new item is the smallest for this index?
			if !atomic.CompareAndSwapPointer(ptr, unsafe.Pointer(element), unsafe.Pointer(item)) {
				continue // a new item was inserted concurrently, retry
			}
		}
		return 0
	}
}
