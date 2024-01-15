package structr

type result[T any] struct {
	// linked list entry this result is
	// stored under in Cache.lruList.
	entry elem[*result[T]]

	// keys tracks the indices
	// result is stored under.
	keys []*indexkey[T]

	// cached value.
	value T

	// cached error.
	err error
}

func result_acquire[T any](c *Cache[T]) *result[T] {
	var res *result[T]

	if len(c.resPool) == 0 {
		// Allocate new result.
		res = new(result[T])
	} else {
		// Pop result from pool slice.
		res = c.resPool[len(c.resPool)-1]
		c.resPool = c.resPool[:len(c.resPool)-1]
	}

	// Push to front of LRU list.
	c.lruList.pushFront(&res.entry)
	res.entry.Value = res

	return res
}

func result_release[T any](c *Cache[T], res *result[T]) {
	// Remove from the LRU list.
	c.lruList.remove(&res.entry)
	res.entry.Value = nil

	var zero T

	// Reset all result fields.
	res.keys = res.keys[:0]
	res.value = zero
	res.err = nil

	// Release result to memory pool.
	c.resPool = append(c.resPool, res)
}

func result_dropIndex[T any](c *Cache[T], res *result[T], index *Index[T]) {
	for i := 0; i < len(res.keys); i++ {

		if res.keys[i].index != index {
			// Prof. Obiwan:
			// this is not the index
			// we are looking for.
			continue
		}

		// Get index key ptr.
		ikey := res.keys[i]

		// Move all index keys down + reslice.
		copy(res.keys[i:], res.keys[i+1:])
		res.keys = res.keys[:len(res.keys)-1]

		// Release ikey to memory pool.
		indexkey_release(c, ikey)

		return
	}
}
