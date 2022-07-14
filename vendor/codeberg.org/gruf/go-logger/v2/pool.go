package logger

import (
	"sync"

	"codeberg.org/gruf/go-logger/v2/entry"
)

// entryPool is an entry.Entry memory pool.
var entryPool = sync.Pool{
	New: func() interface{} {
		return entry.NewEntry(make([]byte, 0, 512), nil)
	},
}

// getEntry will fetch an entry from pool and set to given format.
func getEntry(fmt entry.Formatter) *entry.Entry {
	e := entryPool.Get().(*entry.Entry)
	e.SetFormat(fmt)
	return e
}

// putEntry will take an entry, reset and place in pool.
func putEntry(entry *entry.Entry) {
	const max = int(^uint16(0))
	if entry.Buffer().Cap() > max {
		return // drop large buffer
	}
	entry.Reset()
	entryPool.Put(entry)
}
