package timeline

import (
	"container/list"
	"errors"

	"github.com/sirupsen/logrus"
)

func (t *timeline) Remove(statusID string) (int, error) {
	l := t.log.WithFields(logrus.Fields{
		"func": "Remove",
		"accountTimeline": t.accountID,
		"statusID": statusID,
	})
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
				l.Debug("found status in postIndex")
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
				l.Debug("found status in preparedPosts")
				removePrepared = append(removePrepared, e)
			}
		}
	}
	for _, e := range removePrepared {
		t.preparedPosts.data.Remove(e)
		removed = removed + 1
	}

	l.Debugf("removed %d entries", removed)
	return removed, nil
}
