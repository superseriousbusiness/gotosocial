package structr

import (
	"os"
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
		i := new(indexed_item)
		i.elem.data = unsafe.Pointer(i)
		v = i
	}
	item := v.(*indexed_item)
	return item
}

// free_indexed_item releases the indexed_item.
func free_indexed_item(item *indexed_item) {
	if len(item.indexed) > 0 ||
		item.elem.next != nil ||
		item.elem.prev != nil {
		msg := assert("item not in use")
		os.Stderr.WriteString(msg + "\n")
		return
	}
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

		// Reslice index entries minus 'x'.
		_ = copy(i.indexed[x:], i.indexed[x+1:])
		i.indexed[len(i.indexed)-1] = nil
		i.indexed = i.indexed[:len(i.indexed)-1]
		break
	}
}
