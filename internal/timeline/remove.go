package timeline

import (
	"container/list"
	"errors"
)

func (t *timeline) Remove(statusID string) (int, error) {
	t.Lock()
	defer t.Unlock()
	var removed int

	// remove entr(ies) from the post index
	removeIndexes := []*list.Element{}
	if t.postIndex != nil && t.postIndex.data != nil {
		for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*postIndexEntry)
			if !ok {
				return removed, errors.New("Remove: could not parse e as a postIndexEntry")
			}
			if entry.statusID == statusID {
				removeIndexes = append(removeIndexes, e)
			}
		}
	}
	for _, e := range removeIndexes {
		t.postIndex.data.Remove(e)
		removed = removed + 1
	}

	// remove entr(ies) from prepared posts
	removePrepared := []*list.Element{}
	if t.preparedPosts != nil && t.preparedPosts.data != nil {
		for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*preparedPostsEntry)
			if !ok {
				return removed, errors.New("Remove: could not parse e as a preparedPostsEntry")
			}
			if entry.statusID == statusID {
				removePrepared = append(removePrepared, e)
			}
		}
	}
	for _, e := range removePrepared {
		t.preparedPosts.data.Remove(e)
		removed = removed + 1
	}

	return removed, nil
}
