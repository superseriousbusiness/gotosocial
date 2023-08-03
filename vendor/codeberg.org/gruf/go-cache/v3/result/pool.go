package result

import "sync"

// resultPool is a global pool for result
// objects, regardless of cache type.
var resultPool sync.Pool

// getEntry fetches a result from pool, or allocates new.
func getResult() *result {
	v := resultPool.Get()
	if v == nil {
		return new(result)
	}
	return v.(*result)
}

// putResult replaces a result in the pool.
func putResult(r *result) {
	r.Keys = nil
	r.Value = nil
	r.Error = nil
	resultPool.Put(r)
}
