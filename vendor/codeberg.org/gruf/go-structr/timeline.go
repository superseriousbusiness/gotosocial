package structr

import (
	"cmp"
	"os"
	"reflect"
	"slices"
	"sync"
	"unsafe"
)

// Direction defines a direction
// to iterate entries in a Timeline.
type Direction bool

const (
	// Asc = ascending, i.e. bottom-up.
	Asc = Direction(true)

	// Desc = descending, i.e. top-down.
	Desc = Direction(false)
)

// TimelineConfig defines config vars for initializing a Timeline{}.
type TimelineConfig[StructType any, PK cmp.Ordered] struct {

	// Copy provides a means of copying
	// timelined values, to ensure returned values
	// do not share memory with those in timeline.
	Copy func(StructType) StructType

	// Invalidate is called when timelined
	// values are invalidated, either as passed
	// to Insert(), or by calls to Invalidate().
	Invalidate func(StructType)

	// PKey defines the generic parameter StructType's
	// field to use as the primary key for this cache.
	// It must be ordered so that the timeline can
	// maintain correct sorting of inserted values.
	//
	// Field selection logic follows the same path as
	// with IndexConfig{}.Fields. Noting that in this
	// case only a single field is permitted, though
	// it may be nested, and as described above the
	// type must conform to cmp.Ordered.
	PKey IndexConfig

	// Indices defines indices to create
	// in the Timeline for the receiving
	// generic struct type parameter.
	Indices []IndexConfig
}

// Timeline provides an ordered-list like cache of structures,
// with automated indexing and invalidation by any initialization
// defined combination of fields. The list order is maintained
// according to the configured struct primary key.
type Timeline[StructType any, PK cmp.Ordered] struct {

	// hook functions.
	invalid func(StructType)
	copy    func(StructType) StructType

	// main underlying
	// timeline list.
	//
	// where:
	// - head = top = largest
	// - tail = btm = smallest
	list list

	// contains struct field information of
	// the field used as the primary key for
	// this timeline. it can also be found
	// under indices[0]
	pkey pkey_field

	// indices used in storing passed struct
	// types by user defined sets of fields.
	indices []Index

	// protective mutex, guards:
	// - Timeline{}.*
	// - Index{}.data
	mutex sync.Mutex
}

// Init initializes the timeline with given configuration
// including struct fields to index, and necessary fns.
func (t *Timeline[T, PK]) Init(config TimelineConfig[T, PK]) {
	rt := reflect.TypeOf((*T)(nil)).Elem()

	if len(config.Indices) == 0 {
		panic("no indices provided")
	}

	if config.Copy == nil {
		panic("copy function must be provided")
	}

	// Safely copy over
	// provided config.
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// The first index is created from PKey,
	// other indices are created as expected.
	t.indices = make([]Index, len(config.Indices)+1)
	t.indices[0].ptr = unsafe.Pointer(t)
	t.indices[0].init(rt, config.PKey, 0)
	if len(t.indices[0].fields) > 1 {
		panic("primary key must contain only 1 field")
	}
	for i, cfg := range config.Indices {
		t.indices[i+1].ptr = unsafe.Pointer(t)
		t.indices[i+1].init(rt, cfg, 0)
	}

	// Extract pkey details from index.
	field := t.indices[0].fields[0]
	t.pkey = pkey_field{
		rtype:   field.rtype,
		offsets: field.offsets,
		likeptr: field.likeptr,
	}

	// Copy over remaining.
	t.copy = config.Copy
	t.invalid = config.Invalidate
}

// Index selects index with given name from timeline, else panics.
func (t *Timeline[T, PK]) Index(name string) *Index {
	for i, idx := range t.indices {
		if idx.name == name {
			return &(t.indices[i])
		}
	}
	panic("unknown index: " + name)
}

// Select allows you to retrieve a slice of values, in order, from the timeline.
// This slice is defined by the minimum and maximum primary key parameters, up to
// a given length in size. The direction in which you select will determine which
// of the min / max primary key values is used as the *cursor* to begin the start
// of the selection, and which is used as the *boundary* to mark the end, if set.
// In either case, the length parameter is always optional.
//
// dir = Asc  : cursors up from 'max' (required), with boundary 'min' (optional).
// dir = Desc : cursors down from 'min' (required), with boundary 'max' (optional).
func (t *Timeline[T, PK]) Select(min, max *PK, length *int, dir Direction) (values []T) {

	// Acquire lock.
	t.mutex.Lock()

	// Check init'd.
	if t.copy == nil {
		t.mutex.Unlock()
		panic("not initialized")
	}

	switch dir {
	case Asc:
		// Verify args.
		if min == nil {
			t.mutex.Unlock()
			panic("min must be provided when selecting asc")
		}

		// Select determined values ASCENDING.
		values = t.select_asc(*min, max, length)

	case Desc:
		// Verify args.
		if max == nil {
			t.mutex.Unlock()
			panic("max must be provided when selecting asc")
		}

		// Select determined values DESCENDING.
		values = t.select_desc(min, *max, length)
	}

	// Done with lock.
	t.mutex.Unlock()

	return values
}

// Insert will insert the given values into the timeline,
// calling any set invalidate hook on each inserted value.
// Returns current list length after performing inserts.
func (t *Timeline[T, PK]) Insert(values ...T) int {

	// Acquire lock.
	t.mutex.Lock()

	// Check init'd.
	if t.copy == nil {
		t.mutex.Unlock()
		panic("not initialized")
	}

	// Allocate a slice of our value wrapping struct type.
	with_keys := make([]value_with_pk[T, PK], len(values))
	if len(with_keys) != len(values) {
		panic(assert("BCE"))
	}

	// Range the provided values.
	for i, value := range values {

		// Create our own copy
		// of value to work with.
		value = t.copy(value)

		// Take ptr to the value copy.
		vptr := unsafe.Pointer(&value)

		// Extract primary key from vptr.
		kptr := extract_pkey(vptr, t.pkey)

		var pkey PK
		if kptr != nil {
			// Cast as PK type.
			pkey = *(*PK)(kptr)
		} else {
			// Use zero value pointer.
			kptr = unsafe.Pointer(&pkey)
		}

		// Append wrapped value to slice with
		// the acquire pointers and primary key.
		with_keys[i] = value_with_pk[T, PK]{
			k: pkey,
			v: value,

			kptr: kptr,
			vptr: vptr,
		}
	}

	var last *list_elem

	// BEFORE inserting the prepared slice of value copies w/ primary
	// keys, sort them by their primary key, ascending. This permits
	// us to re-use the 'last' timeline position as next insert cursor.
	// Otherwise we would have to iterate from 'head' every single time.
	slices.SortFunc(with_keys, func(a, b value_with_pk[T, PK]) int {
		const k = +1
		switch {
		case a.k < b.k:
			return +k
		case b.k < a.k:
			return -k
		default:
			return 0
		}
	})

	// Store each value in the timeline,
	// updating the last used list element
	// each time so we don't have to iter
	// down from head on every single store.
	for _, value := range with_keys {
		last = t.store_one(last, value)
	}

	// Get func ptrs.
	invalid := t.invalid

	// Get length AFTER
	// insert to return.
	len := t.list.len

	// Done with lock.
	t.mutex.Unlock()

	if invalid != nil {
		// Pass all invalidated values
		// to given user hook (if set).
		for _, value := range values {
			invalid(value)
		}
	}

	return len
}

// Invalidate invalidates all entries stored in index under given keys.
// Note that if set, this will call the invalidate hook on each value.
func (t *Timeline[T, PK]) Invalidate(index *Index, keys ...Key) {
	if index == nil {
		panic("no index given")
	} else if index.ptr != unsafe.Pointer(t) {
		panic("invalid index for timeline")
	}

	// Acquire lock.
	t.mutex.Lock()

	// Preallocate expected ret slice.
	values := make([]T, 0, len(keys))

	for i := range keys {
		// Delete all items under key from index, collecting
		// value items and dropping them from all their indices.
		index.delete(keys[i].key, func(item *indexed_item) {

			// Cast to *actual* timeline item.
			t_item := to_timeline_item(item)

			if value, ok := item.data.(T); ok {
				// No need to copy, as item
				// being deleted from cache.
				values = append(values, value)
			}

			// Delete item.
			t.delete(t_item)
		})
	}

	// Get func ptrs.
	invalid := t.invalid

	// Done with lock.
	t.mutex.Unlock()

	if invalid != nil {
		// Pass all invalidated values
		// to given user hook (if set).
		for _, value := range values {
			invalid(value)
		}
	}
}

// Range will range over all values in the timeline in given direction.
// dir = Asc  : ranges from the bottom-up.
// dir = Desc : ranges from the top-down.
//
// Please note that the entire Timeline{} will be locked for the duration of the range
// operation, i.e. from the beginning of the first yield call until the end of the last.
func (t *Timeline[T, PK]) Range(dir Direction) func(yield func(index int, value T) bool) {
	return func(yield func(int, T) bool) {
		if t.copy == nil {
			panic("not initialized")
		} else if yield == nil {
			panic("nil func")
		}

		// Acquire lock.
		t.mutex.Lock()
		defer t.mutex.Unlock()

		var i int
		switch dir {

		case Asc:
			// Iterate through linked list from bottom (i.e. tail).
			for prev := t.list.tail; prev != nil; prev = prev.prev {

				// Extract item from list element.
				item := (*timeline_item)(prev.data)

				// Create copy of item value.
				value := t.copy(item.data.(T))

				// Pass to given function.
				if !yield(i, value) {
					break
				}

				// Iter
				i++
			}

		case Desc:
			// Iterate through linked list from top (i.e. head).
			for next := t.list.head; next != nil; next = next.next {

				// Extract item from list element.
				item := (*timeline_item)(next.data)

				// Create copy of item value.
				value := t.copy(item.data.(T))

				// Pass to given function.
				if !yield(i, value) {
					break
				}

				// Iter
				i++
			}
		}
	}
}

// RangeUnsafe is functionally similar to Range(), except it does not pass *copies* of
// data. It allows you to operate on the data directly and modify it. As such it can also
// be more performant to use this function, even for read-write operations.
//
// Please note that the entire Timeline{} will be locked for the duration of the range
// operation, i.e. from the beginning of the first yield call until the end of the last.
func (t *Timeline[T, PK]) RangeUnsafe(dir Direction) func(yield func(index int, value T) bool) {
	return func(yield func(int, T) bool) {
		if t.copy == nil {
			panic("not initialized")
		} else if yield == nil {
			panic("nil func")
		}

		// Acquire lock.
		t.mutex.Lock()
		defer t.mutex.Unlock()

		var i int
		switch dir {

		case Asc:
			// Iterate through linked list from bottom (i.e. tail).
			for prev := t.list.tail; prev != nil; prev = prev.prev {

				// Extract item from list element.
				item := (*timeline_item)(prev.data)

				// Pass to given function.
				if !yield(i, item.data.(T)) {
					break
				}

				// Iter
				i++
			}

		case Desc:
			// Iterate through linked list from top (i.e. head).
			for next := t.list.head; next != nil; next = next.next {

				// Extract item from list element.
				item := (*timeline_item)(next.data)

				// Pass to given function.
				if !yield(i, item.data.(T)) {
					break
				}

				// Iter
				i++
			}
		}
	}
}

// RangeKeys will iterate over all values for given keys in the given index.
//
// Please note that the entire Timeline{} will be locked for the duration of the range
// operation, i.e. from the beginning of the first yield call until the end of the last.
func (t *Timeline[T, PK]) RangeKeys(index *Index, keys ...Key) func(yield func(T) bool) {
	return func(yield func(T) bool) {
		if t.copy == nil {
			panic("not initialized")
		} else if index == nil {
			panic("no index given")
		} else if index.ptr != unsafe.Pointer(t) {
			panic("invalid index for timeline")
		} else if yield == nil {
			panic("nil func")
		}

		// Acquire lock.
		t.mutex.Lock()
		defer t.mutex.Unlock()

		for _, key := range keys {
			var done bool

			// Iterate over values in index under key.
			index.get(key.key, func(i *indexed_item) {

				// Cast to timeline_item type.
				item := to_timeline_item(i)

				// Create copy of item value.
				value := t.copy(item.data.(T))

				// Pass val to yield function.
				done = done || !yield(value)
			})

			if done {
				break
			}
		}
	}
}

// RangeKeysUnsafe is functionally similar to RangeKeys(), except it does not pass *copies*
// of data. It allows you to operate on the data directly and modify it. As such it can also
// be more performant to use this function, even for read-write operations.
//
// Please note that the entire Timeline{} will be locked for the duration of the range
// operation, i.e. from the beginning of the first yield call until the end of the last.
func (t *Timeline[T, PK]) RangeKeysUnsafe(index *Index, keys ...Key) func(yield func(T) bool) {
	return func(yield func(T) bool) {
		if t.copy == nil {
			panic("not initialized")
		} else if index == nil {
			panic("no index given")
		} else if index.ptr != unsafe.Pointer(t) {
			panic("invalid index for timeline")
		} else if yield == nil {
			panic("nil func")
		}

		// Acquire lock.
		t.mutex.Lock()
		defer t.mutex.Unlock()

		for _, key := range keys {
			var done bool

			// Iterate over values in index under key.
			index.get(key.key, func(i *indexed_item) {

				// Cast to timeline_item type.
				item := to_timeline_item(i)

				// Pass value data to yield function.
				done = done || !yield(item.data.(T))
			})

			if done {
				break
			}
		}
	}
}

// Trim will remove entries from the timeline in given
// direction, ensuring timeline is no larger than 'max'.
// If 'max' >= t.Len(), this function is a no-op.
// dir = Asc  : trims from the bottom-up.
// dir = Desc : trims from the top-down.
func (t *Timeline[T, PK]) Trim(max int, dir Direction) {
	// Acquire lock.
	t.mutex.Lock()

	// Calculate number to drop.
	diff := t.list.len - int(max)
	if diff <= 0 {

		// Trim not needed.
		t.mutex.Unlock()
		return
	}

	switch dir {
	case Asc:
		// Iterate over 'diff' items
		// from bottom of timeline list.
		for range diff {

			// Get bottom list elem.
			bottom := t.list.tail
			if bottom == nil {

				// reached
				// end.
				break
			}

			// Drop bottom-most item from timeline.
			item := (*timeline_item)(bottom.data)
			t.delete(item)
		}

	case Desc:
		// Iterate over 'diff' items
		// from top of timeline list.
		for range diff {

			// Get top list elem.
			top := t.list.head
			if top == nil {

				// reached
				// end.
				break
			}

			// Drop top-most item from timeline.
			item := (*timeline_item)(top.data)
			t.delete(item)
		}
	}

	// Compact index data stores.
	for _, idx := range t.indices {
		(&idx).data.Compact()
	}

	// Done with lock.
	t.mutex.Unlock()
}

// Clear empties the timeline by calling .TrimBottom(0, Down).
func (t *Timeline[T, PK]) Clear() { t.Trim(0, Desc) }

// Len returns the current length of cache.
func (t *Timeline[T, PK]) Len() int {
	t.mutex.Lock()
	l := t.list.len
	t.mutex.Unlock()
	return l
}

// Debug returns debug stats about cache.
func (t *Timeline[T, PK]) Debug() map[string]any {
	m := make(map[string]any, 2)
	t.mutex.Lock()
	m["list"] = t.list.len
	indices := make(map[string]any, len(t.indices))
	m["indices"] = indices
	for _, idx := range t.indices {
		var n uint64
		for _, l := range idx.data.m {
			n += uint64(l.len)
		}
		indices[idx.name] = n
	}
	t.mutex.Unlock()
	return m
}

func (t *Timeline[T, PK]) select_asc(min PK, max *PK, length *int) (values []T) {
	// Iterate through linked list
	// from bottom (i.e. tail), asc.
	prev := t.list.tail

	// Iterate from 'prev' up, skipping all
	// entries with pkey below cursor 'min'.
	for ; prev != nil; prev = prev.prev {
		item := (*timeline_item)(prev.data)
		pkey := *(*PK)(item.pk)

		// Check below min.
		if pkey <= min {
			continue
		}

		// Reached
		// cursor.
		break
	}

	if prev == nil {
		// No values
		// remaining.
		return
	}

	// Optimized switch case to handle
	// each set of argument combinations
	// separately, in order to minimize
	// number of checks during loops.
	switch {

	case length != nil && max != nil:
		// Deref arguments.
		length := *length
		max := *max

		// Optimistic preallocate slice.
		values = make([]T, 0, length)

		// Both a length and maximum were given,
		// select from cursor until either reached.
		for ; prev != nil; prev = prev.prev {
			item := (*timeline_item)(prev.data)
			pkey := *(*PK)(item.pk)

			// Check above max.
			if pkey >= max {
				break
			}

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)

			// Check if length reached.
			if len(values) >= length {
				break
			}
		}

	case length != nil:
		// Deref length.
		length := *length

		// Optimistic preallocate slice.
		values = make([]T, 0, length)

		// Only a length was given, select
		// from cursor until length reached.
		for ; prev != nil; prev = prev.prev {
			item := (*timeline_item)(prev.data)

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)

			// Check if length reached.
			if len(values) >= length {
				break
			}
		}

	case max != nil:
		// Deref min.
		max := *max

		// Only a maximum was given, select
		// from cursor until max is reached.
		for ; prev != nil; prev = prev.prev {
			item := (*timeline_item)(prev.data)
			pkey := *(*PK)(item.pk)

			// Check above max.
			if pkey >= max {
				break
			}

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)
		}

	default:
		// No maximum or length were given,
		// ALL from cursor need selecting.
		for ; prev != nil; prev = prev.prev {
			item := (*timeline_item)(prev.data)

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)
		}
	}

	return
}

func (t *Timeline[T, PK]) select_desc(min *PK, max PK, length *int) (values []T) {
	// Iterate through linked list
	// from top (i.e. head), desc.
	next := t.list.head

	// Iterate from 'next' down, skipping
	// all entries with pkey above cursor 'max'.
	for ; next != nil; next = next.next {
		item := (*timeline_item)(next.data)
		pkey := *(*PK)(item.pk)

		// Check above max.
		if pkey >= max {
			continue
		}

		// Reached
		// cursor.
		break
	}

	if next == nil {
		// No values
		// remaining.
		return
	}

	// Optimized switch case to handle
	// each set of argument combinations
	// separately, in order to minimize
	// number of checks during loops.
	switch {

	case length != nil && min != nil:
		// Deref arguments.
		length := *length
		min := *min

		// Optimistic preallocate slice.
		values = make([]T, 0, length)

		// Both a length and minimum were given,
		// select from cursor until either reached.
		for ; next != nil; next = next.next {
			item := (*timeline_item)(next.data)
			pkey := *(*PK)(item.pk)

			// Check below min.
			if pkey <= min {
				break
			}

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)

			// Check if length reached.
			if len(values) >= length {
				break
			}
		}

	case length != nil:
		// Deref length.
		length := *length

		// Optimistic preallocate slice.
		values = make([]T, 0, length)

		// Only a length was given, select
		// from cursor until length reached.
		for ; next != nil; next = next.next {
			item := (*timeline_item)(next.data)

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)

			// Check if length reached.
			if len(values) >= length {
				break
			}
		}

	case min != nil:
		// Deref min.
		min := *min

		// Only a minimum was given, select
		// from cursor until minimum reached.
		for ; next != nil; next = next.next {
			item := (*timeline_item)(next.data)
			pkey := *(*PK)(item.pk)

			// Check below min.
			if pkey <= min {
				break
			}

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)
		}

	default:
		// No minimum or length were given,
		// ALL from cursor need selecting.
		for ; next != nil; next = next.next {
			item := (*timeline_item)(next.data)

			// Append value copy.
			value := item.data.(T)
			value = t.copy(value)
			values = append(values, value)
		}
	}

	return
}

// value_with_pk wraps an incoming value type, with
// its extracted primary key, and pointers to both.
// this encompasses all arguments related to a value
// required by store_one(), simplifying some logic.
//
// with all the primary keys extracted, it also
// makes it much easier to sort input before insert.
type value_with_pk[T any, PK comparable] struct {
	k PK // primary key value
	v T  // value copy

	kptr unsafe.Pointer // primary key ptr
	vptr unsafe.Pointer // value copy ptr
}

func (t *Timeline[T, PK]) store_one(last *list_elem, value value_with_pk[T, PK]) *list_elem {
	// NOTE: the value passed here should
	// already be a copy of the original.

	// Alloc new index item.
	t_item := new_timeline_item()
	if cap(t_item.indexed) < len(t.indices) {

		// Preallocate item indices slice to prevent Go auto
		// allocating overlying large slices we don't need.
		t_item.indexed = make([]*index_entry, 0, len(t.indices))
	}

	// Set item value data.
	t_item.data = value.v
	t_item.pk = value.kptr

	// Get zero'th index, i.e.
	// the primary key index.
	idx0 := (&t.indices[0])

	// Acquire key buf.
	buf := new_buffer()

	// Calculate index key from already extracted
	// primary key, checking for zero return value.
	partptrs := []unsafe.Pointer{value.kptr}
	key := idx0.key(buf, partptrs)
	if key == "" { // i.e. (!allow_zero && pkey == zero)
		free_timeline_item(t_item)
		free_buffer(buf)
		return last
	}

	// Convert to indexed_item pointer.
	i_item := from_timeline_item(t_item)

	if last == nil {
		// No previous element was provided, this is
		// first insert, we need to work from head.

		// Check for emtpy head.
		if t.list.head == nil {

			// The easiest case, this will
			// be the first item in list.
			t.list.push_front(&t_item.elem)
			last = t.list.head // return value
			goto indexing
		}

		// Extract head item and its primary key.
		headItem := (*timeline_item)(t.list.head.data)
		headPK := *(*PK)(headItem.pk)
		if value.k > headPK {

			// Another easier case, this also
			// will be the first item in list.
			t.list.push_front(&t_item.elem)
			last = t.list.head // return value
			goto indexing
		}

		// Check (and drop) if pkey is a collision!
		if value.k == headPK && is_unique(idx0.flags) {
			free_timeline_item(t_item)
			free_buffer(buf)
			return t.list.head
		}

		// Set last = head.next
		// as next to work from.
		last = t.list.head.next
	}

	// Iterate through list from head
	// to find location. Optimized into two
	// cases to minimize loop CPU cycles.
	if is_unique(idx0.flags) {
		for next := last; //
		next != nil; next = next.next {

			// Extract item and it's primary key.
			nextItem := (*timeline_item)(next.data)
			nextPK := *(*PK)(nextItem.pk)

			// If pkey smaller than
			// cursor's, keep going.
			if value.k < nextPK {
				continue
			}

			// Check (and drop) if
			// pkey is a collision!
			if value.k == nextPK {
				free_timeline_item(t_item)
				free_buffer(buf)
				return next
			}

			// New pkey is larger than cursor,
			// insert into list just before it.
			t.list.insert(&t_item.elem, next.prev)
			last = next // return value
			goto indexing
		}
	} else {
		for next := last; //
		next != nil; next = next.next {

			// Extract item and it's primary key.
			nextItem := (*timeline_item)(next.data)
			nextPK := *(*PK)(nextItem.pk)

			// If pkey smaller than
			// cursor's, keep going.
			if value.k < nextPK {
				continue
			}

			// New pkey is larger than cursor,
			// insert into list just before it.
			t.list.insert(&t_item.elem, next.prev)
			last = next // return value
			goto indexing
		}
	}

	// We reached the end of the
	// list, insert at tail pos.
	t.list.push_back(&t_item.elem)
	last = t.list.tail // return value
	goto indexing

indexing:
	// Append already-extracted
	// primary key to 0th index.
	_ = idx0.add(key, i_item)

	// Insert item into each of indices.
	for i := 1; i < len(t.indices); i++ {

		// Get current index ptr.
		idx := (&t.indices[i])

		// Extract fields comprising index key from value.
		parts := extract_fields(value.vptr, idx.fields)

		// Calculate this index key,
		// checking for zero values.
		key := idx.key(buf, parts)
		if key == "" {
			continue
		}

		// Add this item to index,
		// checking for collisions.
		if !idx.add(key, i_item) {

			// This key already appears
			// in this unique index. So
			// drop new timeline item.
			t.delete(t_item)
			free_buffer(buf)
			return last
		}
	}

	// Done with bufs.
	free_buffer(buf)
	return last
}

func (t *Timeline[T, PK]) delete(i *timeline_item) {
	for len(i.indexed) != 0 {
		// Pop last indexed entry from list.
		entry := i.indexed[len(i.indexed)-1]
		i.indexed[len(i.indexed)-1] = nil
		i.indexed = i.indexed[:len(i.indexed)-1]

		// Get entry's index.
		index := entry.index

		// Drop this index_entry.
		index.delete_entry(entry)
	}

	// Drop from main list.
	t.list.remove(&i.elem)

	// Free unused item.
	free_timeline_item(i)
}

type timeline_item struct {
	indexed_item

	// retains fast ptr access
	// to primary key value of
	// above indexed_item{}.data
	pk unsafe.Pointer

	// check bits always all set
	// to 1. used to ensure cast
	// from indexed_item to this
	// type was originally a
	// timeline_item to begin with.
	ck uint
}

func init() {
	// ensure the embedded indexed_item struct is ALWAYS at zero offset.
	// we rely on this to allow a ptr to one to be a ptr to either of them.
	const off = unsafe.Offsetof(timeline_item{}.indexed_item)
	if off != 0 {
		panic(assert("offset_of(timeline_item{}.indexed_item) = 0"))
	}
}

// from_timeline_item converts a timeline_item ptr to indexed_item, given the above init() guarantee.
func from_timeline_item(item *timeline_item) *indexed_item {
	return (*indexed_item)(unsafe.Pointer(item))
}

// to_timeline_item converts an indexed_item ptr to timeline_item, given the above init() guarantee.
// NOTE THIS MUST BE AN indexed_item THAT WAS INITIALLY CONVERTED WITH from_timeline_item().
func to_timeline_item(item *indexed_item) *timeline_item {
	to := (*timeline_item)(unsafe.Pointer(item))
	if to.ck != ^uint(0) {
		// ensure check bits are set indicating
		// it was a timeline_item originally.
		panic(assert("check bits are set"))
	}
	return to
}

var timeline_item_pool sync.Pool

// new_timeline_item returns a new prepared timeline_item.
func new_timeline_item() *timeline_item {
	v := timeline_item_pool.Get()
	if v == nil {
		i := new(timeline_item)
		i.elem.data = unsafe.Pointer(i)
		i.ck = ^uint(0)
		v = i
	}
	item := v.(*timeline_item)
	return item
}

// free_timeline_item releases the timeline_item.
func free_timeline_item(item *timeline_item) {
	if len(item.indexed) > 0 ||
		item.elem.next != nil ||
		item.elem.prev != nil {
		msg := assert("item not in use")
		os.Stderr.WriteString(msg + "\n")
		return
	}
	item.data = nil
	item.pk = nil
	timeline_item_pool.Put(item)
}
