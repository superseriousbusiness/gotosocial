package timeline

import (
	"container/list"
	"errors"
	"fmt"

	apimodel "github.com/superseriousbusiness/gotosocial/internal/api/model"
)

func (t *timeline) GetXFromTop(amount int) ([]*apimodel.Status, error) {
	// make a slice of statuses with the length we need to return
	statuses := make([]*apimodel.Status, 0, amount)

	// if there are no prepared posts, just return the empty slice
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

	// if there are no prepared posts, just return the empty slice
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
			fmt.Printf("\n\n\n GETXBEHINDID: FOUND BEHINDID %s WITH POSITION %d AND CREATEDAT %s \n\n\n", behindID, position, entry.createdAt.String())
			behindIDMark = e
			break findMarkLoop
		}
	}

	// we didn't find it, so we need to make sure it's indexed and prepared and then try again
	if behindIDMark == nil {
		if err := t.IndexBehind(behindID, true, amount); err != nil {
			return nil, fmt.Errorf("GetXBehindID: error indexing behind and including ID %s", behindID)
		}
		if err := t.PrepareBehind(behindID, true, amount); err != nil {
			return nil, fmt.Errorf("GetXBehindID: error preparing behind and including ID %s", behindID)
		}
		return t.GetXBehindID(amount, behindID)
	}

	// make sure we have enough posts prepared behind it to return what we're being asked for
	if t.preparedPosts.data.Len() < amount+position {
		fmt.Printf("\n\n\n GETXBEHINDID: PREPARED POSTS LENGTH %d WAS LESS THAN AMOUNT %d PLUS POSITION %d", t.preparedPosts.data.Len(), amount, position)
		if err := t.PrepareBehind(behindID, false, amount); err != nil {
			return nil, err
		}
		fmt.Printf("\n\n\n GETXBEHINDID: PREPARED POSTS LENGTH IS NOW %d", t.preparedPosts.data.Len())
	}

	// start serving from the entry right after the mark
	var served int
serveloop:
	for e := behindIDMark.Next(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*preparedPostsEntry)
		if !ok {
			return nil, errors.New("GetXBehindID: could not parse e as a preparedPostsEntry")
		}

		fmt.Printf("\n\n\n GETXBEHINDID: SERVING STATUS ID %s WITH CREATEDAT %s \n\n\n", entry.statusID, entry.createdAt.String())
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

	} else if startFromTop {
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

	// if there are no prepared posts, just return the empty slice
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
		if err := t.PrepareBehind(behindID, false, amount); err != nil {
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
