package structr

import (
	"sync"
	"unsafe"
)

var result_pool sync.Pool

type result struct {
	// linked list elem this result is
	// stored under in Cache.lruList.
	elem list_elem

	// indexed stores the indices
	// this result is stored under.
	indexed []*index_entry

	// cached data (we maintain
	// the type data here using
	// an interface as any one
	// instance can be T / error).
	data interface{}
}

func result_acquire[T any](c *Cache[T]) *result {
	// Acquire from pool.
	v := result_pool.Get()
	if v == nil {
		v = new(result)
	}

	// Cast result value.
	res := v.(*result)

	// Push result elem to front of LRU list.
	list_push_front(&c.lruList, &res.elem)
	res.elem.data = unsafe.Pointer(res)

	return res
}

func result_release[T any](c *Cache[T], res *result) {
	// Remove result elem from LRU list.
	list_remove(&c.lruList, &res.elem)
	res.elem.data = nil

	// Reset all result fields.
	res.indexed = res.indexed[:0]
	res.data = nil

	// Release to pool.
	result_pool.Put(res)
}

func result_drop_index[T any](res *result, index *Index[T]) {
	for i := 0; i < len(res.indexed); i++ {

		if res.indexed[i].index != unsafe.Pointer(index) {
			// Prof. Obiwan:
			// this is not the index
			// we are looking for.
			continue
		}

		// Get index entry ptr.
		entry := res.indexed[i]

		// Move all index entries down + reslice.
		copy(res.indexed[i:], res.indexed[i+1:])
		res.indexed = res.indexed[:len(res.indexed)-1]

		// Release to memory pool.
		index_entry_release(entry)

		return
	}
}
