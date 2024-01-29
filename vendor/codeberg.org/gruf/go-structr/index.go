package structr

import (
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"github.com/zeebo/xxh3"
)

// IndexConfig defines config variables
// for initializing a struct index.
type IndexConfig struct {

	// Fields should contain a comma-separated
	// list of struct fields used when generating
	// keys for this index. Nested fields should
	// be specified using periods. An example:
	// "Username,Favorites.Color"
	//
	// Field types supported include:
	// - ~int
	// - ~int8
	// - ~int16
	// - ~int32
	// - ~int64
	// - ~float32
	// - ~float64
	// - ~string
	// - slices of above
	// - ptrs of above
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
// extract struct keys, generate hash checksums for them
// and store struct results by the init defined config.
// This model is exposed to provide faster lookups in the
// case that you would like to manually provide the used
// index via the Cache.___By() series of functions, or
// access the underlying index key generator.
type Index[StructType any] struct {

	// name is the actual name of this
	// index, which is the unparsed
	// string value of contained fields.
	name string

	// backing data store of the index, containing
	// the cached results contained within wrapping
	// index_entry{} which also contains the exact
	// key each result is stored under. the hash map
	// only keys by the xxh3 hash checksum for speed.
	data map[Hash]*list //[*index_entry[StructType]]

	// struct fields encompassed by
	// keys (+ hashes) of this index.
	fields []structfield

	// index flags:
	// - 1 << 0 = unique
	// - 1 << 1 = allow zero
	flags uint8
}

// Key returns the configured fields as key, and hash sum of key.
func (i *Index[T]) Key(value T) ([]any, Hash, bool) {
	h := get_hasher()
	key, sum, ok := index_key(i, h, value)
	hash_pool.Put(h)
	return key, sum, ok
}

func is_unique(f uint8) bool {
	const mask = uint8(1) << 0
	return f&mask != 0
}

func set_is_unique(f *uint8) {
	const mask = uint8(1) << 0
	(*f) |= mask
}

func allow_zero(f uint8) bool {
	const mask = uint8(1) << 1
	return f&mask != 0
}

func set_allow_zero(f *uint8) {
	const mask = uint8(1) << 1
	(*f) |= mask
}

func init_index[T any](i *Index[T], config IndexConfig, max int) {
	// Set name from the raw
	// struct fields string.
	i.name = config.Fields

	// Set struct flags.
	if config.AllowZero {
		set_allow_zero(&i.flags)
	}
	if !config.Multiple {
		set_is_unique(&i.flags)
	}

	// Split to get the containing struct fields.
	fields := strings.Split(config.Fields, ",")

	// Preallocate expected struct field slice.
	i.fields = make([]structfield, len(fields))

	// Get the reflected struct ptr type.
	t := reflect.TypeOf((*T)(nil)).Elem()

	for x, fieldName := range fields {
		// Split name to account for nesting.
		names := strings.Split(fieldName, ".")

		// Look for usable struct field.
		i.fields[x] = find_field(t, names)
	}

	// Initialize index_entry list store.
	i.data = make(map[Hash]*list, max+1)
}

func index_key[T any](i *Index[T], h *xxh3.Hasher, value T) ([]any, Hash, bool) {
	key := extract_fields(value, i.fields)
	sum, zero := hash_sum(i.fields, h, key)
	if zero && !allow_zero(i.flags) {
		var zero Hash
		return nil, zero, false
	}
	return key, sum, true
}

func index_hash[T any](i *Index[T], h *xxh3.Hasher, key []any) (Hash, bool) {
	sum, zero := hash_sum(i.fields, h, key)
	if zero && !allow_zero(i.flags) {
		var zero Hash
		return zero, false
	}
	return sum, true
}

func index_get[T any](i *Index[T], hash Hash, key []any) *list {
	l := i.data[hash]
	if l == nil {
		return nil
	}
	entry := (*index_entry)(l.head.data)
	if !is_equal(entry.key, key) {
		return l
	}
	return l
}

func index_append[T any](c *Cache[T], i *Index[T], hash Hash, key []any, res *result) {
	// Get list at key.
	l := i.data[hash]

	if l == nil {

		// Allocate new list.
		l = list_acquire()
		i.data[hash] = l

	} else if entry := (*index_entry)(l.head.data); //nocollapse
	!is_equal(entry.key, key) {

		// Collision! Drop all.
		delete(i.data, hash)

		// Iterate entries in list.
		for x := 0; x < l.len; x++ {

			// Pop current head.
			list_remove(l, l.head)

			// Extract result.
			res := entry.result

			// Drop index entry from res.
			result_drop_index(res, i)
			if len(res.indexed) == 0 {

				// Old res now unused,
				// release to mem pool.
				result_release(c, res)
			}
		}

		return

	} else if is_unique(i.flags) {

		// Remove current
		// indexed entry.
		list_remove(l, l.head)

		// Get ptr to old
		// entry before we
		// release to pool.
		res := entry.result

		// Drop this index's key from
		// old res now not indexed here.
		result_drop_index(res, i)
		if len(res.indexed) == 0 {

			// Old res now unused,
			// release to mem pool.
			result_release(c, res)
		}
	}

	// Acquire + setup index entry.
	entry := index_entry_acquire()
	entry.index = unsafe.Pointer(i)
	entry.result = res
	entry.key = key
	entry.hash = hash

	// Append to result's indexed entries.
	res.indexed = append(res.indexed, entry)

	// Add index entry to index list.
	list_push_front(l, &entry.elem)
}

func index_delete[T any](c *Cache[T], i *Index[T], hash Hash, key []any, fn func(*result)) {
	if fn == nil {
		panic("nil fn")
	}

	// Get list at hash.
	l := i.data[hash]
	if l == nil {
		return
	}

	entry := (*index_entry)(l.head.data)

	// Check contains expected key for hash.
	if !is_equal(entry.key, key) {
		return
	}

	// Delete data at hash.
	delete(i.data, hash)

	// Iterate entries in list.
	for x := 0; x < l.len; x++ {

		// Pop current head.
		entry := (*index_entry)(l.head.data)
		list_remove(l, l.head)

		// Extract result.
		res := entry.result

		// Call hook.
		fn(res)

		// Drop index entry from res.
		result_drop_index(res, i)
	}

	// Release to pool.
	list_release(l)
}

func index_delete_entry[T any](c *Cache[T], entry *index_entry) {
	// Get from entry.
	i := (*Index[T])(entry.index)

	// Get list at hash sum.
	l := i.data[entry.hash]
	if l == nil {
		return
	}

	// Remove entry from list.
	list_remove(l, &entry.elem)
	if l.len == 0 {

		// Remove list from map.
		delete(i.data, entry.hash)

		// Release to pool.
		list_release(l)
	}

	// Extract result.
	res := entry.result

	// Drop index entry from res.
	result_drop_index(res, i)
}

var entry_pool sync.Pool

type index_entry struct {
	// elem contains the list element
	// appended to each per-hash list
	// within the Index{} type. the
	// contained value is a self-ref.
	elem list_elem

	// index is the Index{} this
	// index_entry{} is stored in.
	index unsafe.Pointer

	// result is the actual
	// underlying result stored
	// within the index. this
	// also contains a ref to
	// this *index_entry in order
	// to track indices each result
	// is currently stored under.
	result *result

	// key contains the actual
	// key this item was stored
	// under, used for collision
	// check.
	key []any

	// hash contains computed
	// hash checksum of .key.
	hash Hash
}

func index_entry_acquire() *index_entry {
	// Acquire from pool.
	v := entry_pool.Get()
	if v == nil {
		v = new(index_entry)
	}

	// Cast index_entry value.
	entry := v.(*index_entry)

	// Set index list elem entry on itself.
	entry.elem.data = unsafe.Pointer(entry)

	return entry
}

func index_entry_release(entry *index_entry) {
	var zero Hash

	// Reset index entry.
	entry.elem.data = nil
	entry.index = nil
	entry.result = nil
	entry.key = nil
	entry.hash = zero

	// Release to pool.
	entry_pool.Put(entry)
}

// is_equal returns whether 2 key slices are equal.
func is_equal(k1, k2 []any) bool {
	if len(k1) != len(k2) {
		return false
	}
	for i := range k1 {
		if k1[i] != k2[i] {
			return false
		}
	}
	return true
}
