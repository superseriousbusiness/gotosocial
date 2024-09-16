package structr

import (
	"sync"
	"unsafe"
)

type indexed_item struct {
	// linked list elem this item
	// is stored in a main list.
	elem list_elem

	// cached data with type.
	data interface{}

	// indexed stores the indices
	// this item is stored under.
	indexed []*index_entry
}

var indexed_item_pool sync.Pool

// new_indexed_item returns a new prepared indexed_item.
func new_indexed_item() *indexed_item {
	v := indexed_item_pool.Get()
	if v == nil {
		v = new(indexed_item)
	}
	item := v.(*indexed_item)
	ptr := unsafe.Pointer(item)
	item.elem.data = ptr
	return item
}

// free_indexed_item releases the indexed_item.
func free_indexed_item(item *indexed_item) {
	item.elem.data = nil
	item.indexed = item.indexed[:0]
	item.data = nil
	indexed_item_pool.Put(item)
}

// drop_index will drop the given index entry from item's indexed.
func (i *indexed_item) drop_index(entry *index_entry) {
	for x := 0; x < len(i.indexed); x++ {
		if i.indexed[x] != entry {
			// Prof. Obiwan:
			// this is not the index
			// we are looking for.
			continue
		}

		// Unset tptr value to
		// ensure GC can take it.
		i.indexed[x] = nil

		// Move all index entries down + reslice.
		_ = copy(i.indexed[x:], i.indexed[x+1:])
		i.indexed = i.indexed[:len(i.indexed)-1]
		break
	}
}
