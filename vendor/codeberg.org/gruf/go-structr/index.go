package structr

import (
	"strings"
)

// IndexConfig defines config variables
// for initializing a struct index.
type IndexConfig struct {

	// Fields should contain a comma-separated
	// list of struct fields used when generating
	// keys for this index. Nested fields should
	// be specified using periods. An example:
	// "Username,Favorites.Color"
	Fields string

	// Multiple indicates whether to accept multiple
	// possible values for any single index key. The
	// default behaviour is to only accept one value
	// and overwrite existing on any write operation.
	Multiple bool

	// AllowZero indicates whether to accept zero
	// value fields in index keys. i.e. whether to
	// index structs for this set of field values
	// IF any one of those field values is the zero
	// value for that type. The default behaviour
	// is to skip indexing structs for this lookup
	// when any of the indexing fields are zero.
	AllowZero bool
}

// Index is an exposed Cache internal model, used to
// generate keys and store struct results by the init
// defined key generation configuration. This model is
// exposed to provide faster lookups in the case that
// you would like to manually provide the used index
// via the Cache.___By() series of functions, or access
// the underlying index key generator.
type Index[StructType any] struct {

	// name is the actual name of this
	// index, which is the unparsed
	// string value of contained fields.
	name string

	// struct field key serializer.
	keygen KeyGen[StructType]

	// backing in-memory data store of
	// generated index keys to result lists.
	data map[string]*list[*result[StructType]]

	// whether to allow
	// multiple results
	// per index key.
	unique bool
}

// init initializes this index with the given configuration.
func (i *Index[T]) init(config IndexConfig) {
	fields := strings.Split(config.Fields, ",")
	i.name = config.Fields
	i.keygen = NewKeyGen[T](fields, config.AllowZero)
	i.unique = !config.Multiple
	i.data = make(map[string]*list[*result[T]])
}

// KeyGen returns the key generator associated with this index.
func (i *Index[T]) KeyGen() *KeyGen[T] {
	return &i.keygen
}

func index_append[T any](c *Cache[T], i *Index[T], key string, res *result[T]) {
	// Acquire + setup indexkey.
	ikey := indexkey_acquire(c)
	ikey.entry.Value = res
	ikey.key = key
	ikey.index = i

	// Append to result's indexkeys.
	res.keys = append(res.keys, ikey)

	// Get list at key.
	l := i.data[key]

	if l == nil {

		// Allocate new list.
		l = list_acquire(c)
		i.data[key] = l

	} else if i.unique {

		// Remove currently
		// indexed result.
		old := l.head
		l.remove(old)

		// Get ptr to old
		// result before we
		// release to pool.
		res := old.Value

		// Drop this index's key from
		// old res now not indexed here.
		result_dropIndex(c, res, i)
		if len(res.keys) == 0 {

			// Old res now unused,
			// release to mem pool.
			result_release(c, res)
		}
	}

	// Add result indexkey to
	// front of results list.
	l.pushFront(&ikey.entry)
}

func index_deleteOne[T any](c *Cache[T], i *Index[T], ikey *indexkey[T]) {
	// Get list at key.
	l := i.data[ikey.key]
	if l == nil {
		return
	}

	// Remove from list.
	l.remove(&ikey.entry)
	if l.len == 0 {

		// Remove list from map.
		delete(i.data, ikey.key)

		// Release list to pool.
		list_release(c, l)
	}
}

func index_delete[T any](c *Cache[T], i *Index[T], key string, fn func(*result[T])) {
	if fn == nil {
		panic("nil fn")
	}

	// Get list at key.
	l := i.data[key]
	if l == nil {
		return
	}

	// Delete data at key.
	delete(i.data, key)

	// Iterate results in list.
	for x := 0; x < l.len; x++ {

		// Pop current head.
		res := l.head.Value
		l.remove(l.head)

		// Delete index's key
		// from result tracking.
		result_dropIndex(c, res, i)

		// Call hook.
		fn(res)
	}

	// Release list to pool.
	list_release(c, l)
}

type indexkey[T any] struct {
	// linked list entry the related
	// result is stored under in the
	// Index.data[key] linked list.
	entry elem[*result[T]]

	// key is the generated index key
	// the related result is indexed
	// under, in the below index.
	key string

	// index is the index that the
	// related result is indexed in.
	index *Index[T]
}

func indexkey_acquire[T any](c *Cache[T]) *indexkey[T] {
	var ikey *indexkey[T]

	if len(c.keyPool) == 0 {
		// Allocate new key.
		ikey = new(indexkey[T])
	} else {
		// Pop result from pool slice.
		ikey = c.keyPool[len(c.keyPool)-1]
		c.keyPool = c.keyPool[:len(c.keyPool)-1]
	}

	return ikey
}

func indexkey_release[T any](c *Cache[T], ikey *indexkey[T]) {
	// Reset indexkey.
	ikey.entry.Value = nil
	ikey.key = ""
	ikey.index = nil

	// Release indexkey to memory pool.
	c.keyPool = append(c.keyPool, ikey)
}
