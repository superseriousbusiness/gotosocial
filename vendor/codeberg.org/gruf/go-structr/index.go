package structr

import (
	"reflect"
	"strings"
	"sync"
	"unsafe"

	"codeberg.org/gruf/go-byteutil"

	"github.com/dolthub/swiss"
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
	// Note that nested fields where the nested
	// struct field is a ptr are supported, but
	// nil ptr values in nesting will result in
	// that particular value NOT being indexed.
	// e.g. with "Favorites.Color" if *Favorites
	// is nil then it will not be indexed.
	//
	// Field types supported include any of those
	// supported by the `go-mangler` library.
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
type Index struct {

	// ptr is a pointer to
	// the source Cache/Queue
	// index is attached to.
	ptr unsafe.Pointer

	// name is the actual name of this
	// index, which is the unparsed
	// string value of contained fields.
	name string

	// backing data store of the index, containing
	// the cached results contained within wrapping
	// index_entry{} which also contains the exact
	// key each result is stored under. the hash map
	// only keys by the xxh3 hash checksum for speed.
	data *swiss.Map[string, *list]

	// struct fields encompassed by
	// keys (+ hashes) of this index.
	fields []struct_field

	// index flags:
	// - 1 << 0 = unique
	// - 1 << 1 = allow zero
	flags uint8
}

// Name returns the receiving Index name.
func (i *Index) Name() string {
	return i.name
}

// Key generates Key{} from given parts for
// the type of lookup this Index uses in cache.
// NOTE: panics on incorrect no. parts / types given.
func (i *Index) Key(parts ...any) Key {
	buf := new_buffer()
	key := i.key(buf, parts)
	free_buffer(buf)
	return key
}

// Keys generates []Key{} from given (multiple) parts
// for the type of lookup this Index uses in the cache.
// NOTE: panics on incorrect no. parts / types given.
func (i *Index) Keys(parts ...[]any) []Key {
	keys := make([]Key, 0, len(parts))
	buf := new_buffer()
	for _, parts := range parts {
		key := i.key(buf, parts)
		if key.Zero() {
			continue
		}
		keys = append(keys, key)
	}
	free_buffer(buf)
	return keys
}

// init will initialize the cache with given type, config and capacity.
func (i *Index) init(t reflect.Type, cfg IndexConfig, cap int) {
	switch {
	// The only 2 types we support are
	// structs, and ptrs to a struct.
	case t.Kind() == reflect.Struct:
	case t.Kind() == reflect.Pointer &&
		t.Elem().Kind() == reflect.Struct:
	default:
		panic("index only support struct{} and *struct{}")
	}

	// Set name from the raw
	// struct fields string.
	i.name = cfg.Fields

	// Set struct flags.
	if cfg.AllowZero {
		set_allow_zero(&i.flags)
	}
	if !cfg.Multiple {
		set_is_unique(&i.flags)
	}

	// Split to get containing struct fields.
	fields := strings.Split(cfg.Fields, ",")

	// Preallocate expected struct field slice.
	i.fields = make([]struct_field, len(fields))
	for x, name := range fields {

		// Split name to account for nesting.
		names := strings.Split(name, ".")

		// Look for usable struct field.
		i.fields[x] = find_field(t, names)
	}

	// Initialize index_entry list store.
	i.data = swiss.NewMap[string, *list](uint32(cap))
}

// get_one will fetch one indexed item under key.
func (i *Index) get_one(key Key) *indexed_item {
	// Get list at hash.
	l, _ := i.data.Get(key.key)
	if l == nil {
		return nil
	}

	// Extract entry from first list elem.
	entry := (*index_entry)(l.head.data)

	// Check contains expected key.
	if !entry.key.Equal(key) {
		return nil
	}

	return entry.item
}

// get will fetch all indexed items under key, passing each to hook.
func (i *Index) get(key Key, hook func(*indexed_item)) {
	if hook == nil {
		panic("nil hook")
	}

	// Get list at hash.
	l, _ := i.data.Get(key.key)
	if l == nil {
		return
	}

	// Extract entry from first list elem.
	entry := (*index_entry)(l.head.data)

	// Check contains expected key.
	if !entry.key.Equal(key) {
		return
	}

	// Iterate all entries in list.
	l.rangefn(func(elem *list_elem) {

		// Extract element entry + item.
		entry := (*index_entry)(elem.data)
		item := entry.item

		// Pass to hook.
		hook(item)
	})
}

// key uses hasher to generate Key{} from given raw parts.
func (i *Index) key(buf *byteutil.Buffer, parts []any) Key {
	if len(parts) != len(i.fields) {
		panicf("incorrect number key parts: want=%d received=%d",
			len(i.fields),
			len(parts),
		)
	}
	buf.B = buf.B[:0]
	if !allow_zero(i.flags) {
		for x, field := range i.fields {
			before := len(buf.B)
			buf.B = field.mangle(buf.B, parts[x])
			if string(buf.B[before:]) == field.zerostr {
				return Key{}
			}
			buf.B = append(buf.B, '.')
		}
	} else {
		for x, field := range i.fields {
			buf.B = field.mangle(buf.B, parts[x])
			buf.B = append(buf.B, '.')
		}
	}
	return Key{
		raw: parts,
		key: string(buf.B),
	}
}

// append will append the given index entry to appropriate
// doubly-linked-list in index hashmap. this handles case
// of key collisions and overwriting 'unique' entries.
func (i *Index) append(key Key, item *indexed_item) {
	// Look for existing.
	l, _ := i.data.Get(key.key)

	if l == nil {

		// Allocate new.
		l = new_list()
		i.data.Put(key.key, l)

	} else if is_unique(i.flags) {

		// Remove head.
		elem := l.head
		l.remove(elem)

		// Drop index from inner item.
		e := (*index_entry)(elem.data)
		e.item.drop_index(e)

		// Free unused entry.
		free_index_entry(e)
	}

	// Prepare new index entry.
	entry := new_index_entry()
	entry.item = item
	entry.key = key
	entry.index = i

	// Add ourselves to item's index tracker.
	item.indexed = append(item.indexed, entry)

	// Add entry to index list.
	l.push_front(&entry.elem)
}

// delete will remove all indexed items under key, passing each to hook.
func (i *Index) delete(key Key, hook func(*indexed_item)) {
	if hook == nil {
		panic("nil hook")
	}

	// Get list at hash.
	l, _ := i.data.Get(key.key)
	if l == nil {
		return
	}

	// Extract entry from first list elem.
	entry := (*index_entry)(l.head.data)

	// Check contains expected key.
	if !entry.key.Equal(key) {
		return
	}

	// Delete data at hash.
	i.data.Delete(key.key)

	// Iterate entries in list.
	for x := 0; x < l.len; x++ {

		// Pop list head.
		elem := l.head
		l.remove(elem)

		// Extract element entry + item.
		entry := (*index_entry)(elem.data)
		item := entry.item

		// Drop index from item.
		item.drop_index(entry)

		// Free now-unused entry.
		free_index_entry(entry)

		// Pass to hook.
		hook(item)
	}

	// Release list.
	free_list(l)
}

// delete_entry deletes the given index entry.
func (i *Index) delete_entry(entry *index_entry) {
	// Get list at hash sum.
	l, _ := i.data.Get(entry.key.key)
	if l == nil {
		return
	}

	// Remove list entry.
	l.remove(&entry.elem)

	if l.len == 0 {
		// Remove entry list from map.
		i.data.Delete(entry.key.key)

		// Release list.
		free_list(l)
	}

	// Drop this index from item.
	entry.item.drop_index(entry)
}

// compact will reduce the size of underlying
// index map if the cap vastly exceeds len.
func (i *Index) compact() {

	// Maximum load factor before
	// 'swiss' allocates new hmap:
	// maxLoad = 7 / 8
	//
	// So we apply the inverse/2, once
	// $maxLoad/2 % of hmap is empty we
	// compact the map to drop buckets.
	len := i.data.Count()
	cap := i.data.Capacity()
	if cap-len > (cap*7)/(8*2) {

		// Create a new map only as big as required.
		data := swiss.NewMap[string, *list](uint32(len))
		i.data.Iter(func(k string, v *list) (stop bool) {
			data.Put(k, v)
			return false
		})

		// Set new map.
		i.data = data
	}
}

// index_entry represents a single entry
// in an Index{}, where it will be accessible
// by Key{} pointing to a containing list{}.
type index_entry struct {

	// list elem that entry is stored
	// within, under containing index.
	// elem.data is ptr to index_entry.
	elem list_elem

	// hash checksum
	// + raw key data
	key Key

	// index this is stored in.
	index *Index

	// underlying indexed item.
	item *indexed_item
}

var index_entry_pool sync.Pool

// new_index_entry returns a new prepared index_entry.
func new_index_entry() *index_entry {
	v := index_entry_pool.Get()
	if v == nil {
		v = new(index_entry)
	}
	entry := v.(*index_entry)
	ptr := unsafe.Pointer(entry)
	entry.elem.data = ptr
	return entry
}

// free_index_entry releases the index_entry.
func free_index_entry(entry *index_entry) {
	entry.elem.data = nil
	entry.key = Key{}
	entry.index = nil
	entry.item = nil
	index_entry_pool.Put(entry)
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
