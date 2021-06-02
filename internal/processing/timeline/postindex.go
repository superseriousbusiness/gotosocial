package timeline

import (
	"container/list"
	"errors"
	"time"
)

type postIndex struct {
	data *list.List
}

type postIndexEntry struct {
	createdAt time.Time
	statusID  string
}

func (p *postIndex) index(i *postIndexEntry) error {

	if p.data == nil {
		p.data = &list.List{}
	}

	// if we have no entries yet, this is both the newest and oldest entry, so just put it in the front
	if p.data.Len() == 0 {
		p.data.PushFront(i)
		return nil
	}

	// we need to iterate through the index to make sure we put this post in the appropriate place according to when it was created
	for e := p.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("Remove: could not parse e as a postIndexEntry")
		}

		// if the post to index is newer than e, insert it before e in the list
		if i.createdAt.After(entry.createdAt) {
			p.data.InsertBefore(i, e)
			return nil
		}
	}

	// if we reach this point it's the oldest post we've seen so put it at the back
	p.data.PushBack(i)
	return nil
}
