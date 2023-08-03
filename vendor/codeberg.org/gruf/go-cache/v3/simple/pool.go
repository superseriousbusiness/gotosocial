package simple

import "sync"

// entryPool is a global pool for Entry
// objects, regardless of cache type.
var entryPool sync.Pool

// getEntry fetches an Entry from pool, or allocates new.
func getEntry() *Entry {
	v := entryPool.Get()
	if v == nil {
		return new(Entry)
	}
	return v.(*Entry)
}

// putEntry replaces an Entry in the pool.
func putEntry(e *Entry) {
	e.Key = nil
	e.Value = nil
	entryPool.Put(e)
}
