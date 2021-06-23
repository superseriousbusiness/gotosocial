package timeline

import (
	"container/list"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

func (t *timeline) Get(amount int, maxID string, sinceID string, minID string) ([]*apimodel.Status, error) {
	l := t.log.WithFields(logrus.Fields{
		"func":      "Get",
		"accountID": t.accountID,
		"amount":    amount,
		"maxID":     maxID,
		"sinceID":   sinceID,
		"minID":     minID,
	})
	l.Debug("entering get")

	var statuses []*apimodel.Status
	var err error

	// no params are defined to just fetch from the top
	if maxID == "" && sinceID == "" && minID == "" {
		statuses, err = t.GetXFromTop(amount)
		// aysnchronously prepare the next predicted query so it's ready when the user asks for it
		if len(statuses) != 0 {
			nextMaxID := statuses[len(statuses)-1].ID
			go func() {
				if err := t.prepareNextQuery(amount, nextMaxID, "", ""); err != nil {
					l.Errorf("error preparing next query: %s", err)
				}
			}()
		}
	}

	// maxID is defined but sinceID isn't so take from behind
	if maxID != "" && sinceID == "" {
		statuses, err = t.GetXBehindID(amount, maxID)
		// aysnchronously prepare the next predicted query so it's ready when the user asks for it
		if len(statuses) != 0 {
			nextMaxID := statuses[len(statuses)-1].ID
			go func() {
				if err := t.prepareNextQuery(amount, nextMaxID, "", ""); err != nil {
					l.Errorf("error preparing next query: %s", err)
				}
			}()
		}
	}

	// maxID is defined and sinceID || minID are as well, so take a slice between them
	if maxID != "" && sinceID != "" {
		statuses, err = t.GetXBetweenID(amount, maxID, minID)
	}
	if maxID != "" && minID != "" {
		statuses, err = t.GetXBetweenID(amount, maxID, minID)
	}

	// maxID isn't defined, but sinceID || minID are, so take x before
	if maxID == "" && sinceID != "" {
		statuses, err = t.GetXBeforeID(amount, sinceID, true)
	}
	if maxID == "" && minID != "" {
		statuses, err = t.GetXBeforeID(amount, minID, true)
	}

	return statuses, err
}

func (t *timeline) GetXFromTop(amount int) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// make sure we have enough posts prepared to return
	if t.preparedPosts.data.Len() < amount {
		if err := t.PrepareFromTop(amount); err != nil {
			return nil, err
		}
	}

	// work through the prepared posts from the top and return
	var served int
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXFromTop: could not parse e as a preparedPostsEntry")
		}
		statuses = append(statuses, entry.prepared)
		served = served + 1
		if served >= amount {
			break
		}
	}

	return statuses, nil
}

func (t *timeline) GetXBehindID(amount int, behindID string) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for
	var position int
	var behindIDMark *list.Element

findMarkLoop:
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		position = position + 1
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID == behindID {
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it, so we need to make sure it's indexed and prepared and then try again
	if behindIDMark == nil {
		if err := t.IndexBehind(behindID, amount); err != nil {
			return nil, fmt.Errorf("GetXBehindID: error indexing behind and including ID %s", behindID)
		}
		if err := t.PrepareBehind(behindID, amount); err != nil {
			return nil, fmt.Errorf("GetXBehindID: error preparing behind and including ID %s", behindID)
		}
		oldestID, err := t.OldestPreparedPostID()
		if err != nil {
			return nil, err
		}
		if oldestID == "" || oldestID == behindID {
			// there is no oldest prepared post, or the oldest prepared post is still the post we're looking for entries after
			// this means we should just return the empty statuses slice since we don't have any more posts to offer
			return statuses, nil
		}
		return t.GetXBehindID(amount, behindID)
	}

	// make sure we have enough posts prepared behind it to return what we're being asked for
	if t.preparedPosts.data.Len() < amount+position {
		if err := t.PrepareBehind(behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		// serve up to the amount requested
		statuses = append(statuses, entry.prepared)
		served = served + 1
		if served >= amount {
			break serveloop
		}
	}

	return statuses, nil
}

func (t *timeline) GetXBeforeID(amount int, beforeID string, startFromTop bool) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for
	var beforeIDMark *list.Element
findMarkLoop:
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID == beforeID {
			beforeIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it, so we need to make sure it's indexed and prepared and then try again
	if beforeIDMark == nil {
		if err := t.IndexBefore(beforeID, true, amount); err != nil {
			return nil, fmt.Errorf("GetXBeforeID: error indexing before and including ID %s", beforeID)
		}
		if err := t.PrepareBefore(beforeID, true, amount); err != nil {
			return nil, fmt.Errorf("GetXBeforeID: error preparing before and including ID %s", beforeID)
		}
		return t.GetXBeforeID(amount, beforeID, startFromTop)
	}

	var served int

	if startFromTop {
		// start serving from the front/top and keep going until we hit mark or get x amount statuses
	serveloopFromTop:
		for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*preparedPostsEntry)
			if !ok {
				return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
			}

			if entry.statusID == beforeID {
				break serveloopFromTop
			}

			// serve up to the amount requested
			statuses = append(statuses, entry.prepared)
			served = served + 1
			if served >= amount {
				break serveloopFromTop
			}
		}
	} else if !startFromTop {
		// start serving from the entry right before the mark
	serveloopFromBottom:
		for e := beforeIDMark.Prev(); e != nil; e = e.Prev() {
			entry, ok := e.Value.(*preparedPostsEntry)
			if !ok {
				return nil, errors.New("GetXBeforeID: could not parse e as a preparedPostsEntry")
			}

			// serve up to the amount requested
			statuses = append(statuses, entry.prepared)
			served = served + 1
			if served >= amount {
				break serveloopFromBottom
			}
		}
	}

	return statuses, nil
}

func (t *timeline) GetXBetweenID(amount int, behindID string, beforeID string) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
	}

	// iterate through the modified list until we hit the mark we're looking for
	var position int
	var behindIDMark *list.Element
findMarkLoop:
	for e := t.preparedPosts.data.Front(); e != nil; e = e.Next() {
		position = position + 1
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID == behindID {
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it
	if behindIDMark == nil {
		return nil, fmt.Errorf("GetXBetweenID: couldn't find status with ID %s", behindID)
	}

	// make sure we have enough posts prepared behind it to return what we're being asked for
	if t.preparedPosts.data.Len() < amount+position {
		if err := t.PrepareBehind(behindID, amount); err != nil {
			return nil, err
		}
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBetweenID: could not parse e as a preparedPostsEntry")
		}

		if entry.statusID == beforeID {
			break serveloop
		}

		// serve up to the amount requested
		statuses = append(statuses, entry.prepared)
		served = served + 1
		if served >= amount {
			break serveloop
		}
	}

	return statuses, nil
}
