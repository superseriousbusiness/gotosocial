package structr

import (
	"os"
	"reflect"
	"strings"
	"unsafe"

	"codeberg.org/gruf/go-byteutil"
	"codeberg.org/gruf/go-mempool"
	"codeberg.org/gruf/go-xunsafe"
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
	// If a nested field encounters a nil pointer
	// along the way, e.g. "Favourites == nil", then
	// a zero value for "Favorites.Color" is used.
	//
	// Field types supported include any of those
	// supported by the `go-mangler/v2` library.
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
	// the source type this
	// index is attached to.
	ptr unsafe.Pointer

	// name is the actual name of this
	// index, which is the unparsed
	// string value of contained fields.
	name string

	// backing data store of the index, containing
	// list{}s of index_entry{}s which each contain
	// the exact key each result is stored under.
	data hashmap

	// struct fields encompassed
	// by keys of this index.
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

// init will initialize the cache with given type, config and capacity.
func (i *Index) init(t xunsafe.TypeIter, cfg IndexConfig, cap int) {
	switch {
	// The only 2 types we support are
	// structs, and ptrs to a struct.
	case t.Type.Kind() == reflect.Struct:
	case t.Type.Kind() == reflect.Pointer &&
		t.Type.Elem().Kind() == reflect.Struct:
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

		// Look for struct field by names.
		i.fields[x], _ = find_field(t, names)
	}

	// Initialize store for
	// index_entry lists.
	i.data.Init(cap)
}

// get_one will fetch one indexed item under key.
func (i *Index) get_one(key Key) *indexed_item {
	// Get list at hash.
	l := i.data.Get(key.key)
	if l == nil {
		return nil
	}

	// Extract entry from first list elem.
	entry := (*index_entry)(l.head.data)

	return entry.item
}

// get will fetch all indexed items under key, passing each to hook.
func (i *Index) get(key string, hook func(*indexed_item)) {
	if hook == nil {
		panic("nil hook")
	}

	// Get list at hash.
	l := i.data.Get(key)
	if l == nil {
		return
	}

	// Iterate the list.
	for elem := l.head; //
	elem != nil;        //
	{
		// Get next before
		// any modification.
		next := elem.next

		// Extract element entry + item.
		entry := (*index_entry)(elem.data)
		item := entry.item

		// Pass to hook.
		hook(item)

		// Set next.
		elem = next
	}
}

// key ...
func (i *Index) key(buf *byteutil.Buffer, parts []unsafe.Pointer) string {
	if len(parts) != len(i.fields) {
		panic(assert("len(parts) = len(i.fields)"))
	}
	buf.B = buf.B[:0]
	if !allow_zero(i.flags) {
		for x, field := range i.fields {
			before := len(buf.B)
			buf.B = field.mangle(buf.B, parts[x])
			if string(buf.B[before:]) == field.zerostr {
				return ""
			}
			buf.B = append(buf.B, '.')
		}
	} else {
		for x, field := range i.fields {
			buf.B = field.mangle(buf.B, parts[x])
			buf.B = append(buf.B, '.')
		}
	}
	return string(buf.B)
}

// add will attempt to add given index entry to appropriate
// doubly-linked-list in index hashmap. in the case of an
// existing entry in a "unique" index, it will return false.
func (i *Index) add(key string, item *indexed_item) bool {
	// Look for existing.
	l := i.data.Get(key)

	if l == nil {

		// Allocate new.
		l = new_list()
		i.data.Put(key, l)

	} else if is_unique(i.flags) {

		// Collision!
		return false
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
	return true
}

// append will append the given index entry to appropriate
// doubly-linked-list in index hashmap. this handles case of
// overwriting "unique" index entries, and removes from given
// outer linked-list in the case that it is no longer indexed.
func (i *Index) append(key string, item *indexed_item) (evicted *indexed_item) {
	// Look for existing.
	l := i.data.Get(key)

	if l == nil {

		// Allocate new.
		l = new_list()
		i.data.Put(key, l)

	} else if is_unique(i.flags) {

		// Remove head.
		elem := l.head
		l.remove(elem)

		// Drop index from inner item,
		// catching the evicted item.
		e := (*index_entry)(elem.data)
		evicted = e.item
		evicted.drop_index(e)

		// Free unused entry.
		free_index_entry(e)

		if len(evicted.indexed) != 0 {
			// Evicted is still stored
			// under index, don't return.
			evicted = nil
		}
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
	return
}

// delete will remove all indexed items under key, passing each to hook.
func (i *Index) delete(key string, hook func(*indexed_item)) {
	if hook == nil {
		panic("nil hook")
	}

	// Get list at hash.
	l := i.data.Get(key)
	if l == nil {
		return
	}

	// Delete at hash.
	i.data.Delete(key)

	// Iterate the list.
	for elem := l.head; //
	elem != nil;        //
	{
		// Get next before
		// any modification.
		next := elem.next

		// Remove elem.
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

		// Set next.
		elem = next
	}

	// Release list.
	free_list(l)
}

// delete_entry deletes the given index entry.
func (i *Index) delete_entry(entry *index_entry) {
	// Get list at hash sum.
	l := i.data.Get(entry.key)
	if l == nil {
		return
	}

	// Remove list entry.
	l.remove(&entry.elem)

	if l.len == 0 {
		// Remove entry from map.
		i.data.Delete(entry.key)

		// Release list.
		free_list(l)
	}

	// Drop this index from item.
	entry.item.drop_index(entry)
}

// index_entry represents a single entry
// in an Index{}, where it will be accessible
// by .key pointing to a containing list{}.
type index_entry struct {

	// list elem that entry is stored
	// within, under containing index.
	// elem.data is ptr to index_entry.
	elem list_elem

	// index this is stored in.
	index *Index

	// underlying indexed item.
	item *indexed_item

	// raw cache key
	// for this entry.
	key string
}

var index_entry_pool mempool.UnsafePool

// new_index_entry returns a new prepared index_entry.
func new_index_entry() *index_entry {
	if ptr := index_entry_pool.Get(); ptr != nil {
		return (*index_entry)(ptr)
	}
	entry := new(index_entry)
	entry.elem.data = unsafe.Pointer(entry)
	return entry
}

// free_index_entry releases the index_entry.
func free_index_entry(entry *index_entry) {
	if entry.elem.next != nil ||
		entry.elem.prev != nil {
		msg := assert("entry not in use")
		os.Stderr.WriteString(msg + "\n")
		return
	}
	entry.key = ""
	entry.index = nil
	entry.item = nil
	ptr := unsafe.Pointer(entry)
	index_entry_pool.Put(ptr)
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
