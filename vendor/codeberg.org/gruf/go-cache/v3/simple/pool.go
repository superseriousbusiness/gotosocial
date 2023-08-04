package simple

import "sync"

// entryPool is a global pool for Entry
// objects, regardless of cache type.
var entryPool sync.Pool

// GetEntry fetches an Entry from pool, or allocates new.
func GetEntry() *Entry {
	v := entryPool.Get()
	if v == nil {
		return new(Entry)
	}
	return v.(*Entry)
}

// PutEntry replaces an Entry in the pool.
func PutEntry(e *Entry) {
	e.Key = nil
	e.Value = nil
	entryPool.Put(e)
}
