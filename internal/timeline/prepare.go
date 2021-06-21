package timeline

import (
	"container/list"
	"errors"
	"fmt"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (t *timeline) prepareNextQuery(amount int, maxID string, sinceID string, minID string) error {
	var err error

	// maxID is defined but sinceID isn't so take from behind
	if maxID != "" && sinceID == "" {
		err = t.PrepareBehind(maxID, amount)
	}

	// maxID isn't defined, but sinceID || minID are, so take x before
	if maxID == "" && sinceID != "" {
		err = t.PrepareBefore(sinceID, false, amount)
	}
	if maxID == "" && minID != "" {
		err = t.PrepareBefore(minID, false, amount)
	}

	return err
}

func (t *timeline) PrepareBehind(statusID string, amount int) error {
	t.Lock()
	defer t.Unlock()

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
		t.preparedPosts.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.postIndex.data == nil {
		return nil
	}

	var prepared int
	var preparing bool
prepareloop:
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareBehind: could not parse e as a postIndexEntry")
		}

		if !preparing {
			// we haven't hit the position we need to prepare from yet
			if entry.statusID == statusID {
				preparing = true
			}
		}

		if preparing {
			if err := t.prepare(entry.statusID); err != nil {
				// there's been an error
				if _, ok := err.(db.ErrNoEntries); !ok {
					// it's a real error
					return fmt.Errorf("PrepareBehind: error preparing status with id %s: %s", entry.statusID, err)
				}
				// the status just doesn't exist (anymore) so continue to the next one
				continue
			}
			if prepared == amount {
				// we're done
				break prepareloop
			}
			prepared = prepared + 1
		}
	}

	return nil
}

func (t *timeline) PrepareBefore(statusID string, include bool, amount int) error {
	t.Lock()
	defer t.Unlock()

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
		t.preparedPosts.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.postIndex.data == nil {
		return nil
	}

	var prepared int
	var preparing bool
prepareloop:
	for e := t.postIndex.data.Back(); e != nil; e = e.Prev() {
		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareBefore: could not parse e as a postIndexEntry")
		}

		if !preparing {
			// we haven't hit the position we need to prepare from yet
			if entry.statusID == statusID {
				preparing = true
				if !include {
					continue
				}
			}
		}

		if preparing {
			if err := t.prepare(entry.statusID); err != nil {
				// there's been an error
				if _, ok := err.(db.ErrNoEntries); !ok {
					// it's a real error
					return fmt.Errorf("PrepareBefore: error preparing status with id %s: %s", entry.statusID, err)
				}
				// the status just doesn't exist (anymore) so continue to the next one
				continue
			}
			if prepared == amount {
				// we're done
				break prepareloop
			}
			prepared = prepared + 1
		}
	}

	return nil
}

func (t *timeline) PrepareFromTop(amount int) error {
	t.Lock()
	defer t.Unlock()

	// lazily initialize prepared posts if it hasn't been done already
	if t.preparedPosts.data == nil {
		t.preparedPosts.data = &list.List{}
		t.preparedPosts.data.Init()
	}

	// if the postindex is nil, nothing has been indexed yet so there's nothing to prepare
	if t.postIndex.data == nil {
		return nil
	}

	var prepared int
prepareloop:
	for e := t.postIndex.data.Front(); e != nil; e = e.Next() {
		if e == nil {
			continue
		}

		entry, ok := e.Value.(*postIndexEntry)
		if !ok {
			return errors.New("PrepareFromTop: could not parse e as a postIndexEntry")
		}

		if err := t.prepare(entry.statusID); err != nil {
			// there's been an error
			if _, ok := err.(db.ErrNoEntries); !ok {
				// it's a real error
				return fmt.Errorf("PrepareFromTop: error preparing status with id %s: %s", entry.statusID, err)
			}
			// the status just doesn't exist (anymore) so continue to the next one
			continue
		}

		prepared = prepared + 1
		if prepared == amount {
			// we're done
			break prepareloop
		}
	}

	return nil
}

func (t *timeline) prepare(statusID string) error {

	// start by getting the status out of the database according to its indexed ID
	gtsStatus := &gtsmodel.Status{}
	if err := t.db.GetByID(statusID, gtsStatus); err != nil {
		return err
	}

	// if the account pointer hasn't been set on this timeline already, set it lazily here
	if t.account == nil {
		timelineOwnerAccount := &gtsmodel.Account{}
		if err := t.db.GetByID(t.accountID, timelineOwnerAccount); err != nil {
			return err
		}
		t.account = timelineOwnerAccount
	}

	// serialize the status (or, at least, convert it to a form that's ready to be serialized)
	apiModelStatus, err := t.tc.StatusToMasto(gtsStatus, t.account)
	if err != nil {
		return err
	}

	// shove it in prepared posts as a prepared posts entry
	preparedPostsEntry := &preparedPostsEntry{
		statusID: statusID,
		prepared: apiModelStatus,
	}

	return t.preparedPosts.insertPrepared(preparedPostsEntry)
}

func (t *timeline) OldestPreparedPostID() (string, error) {
	var id string
	if t.preparedPosts == nil || t.preparedPosts.data == nil {
		// return an empty string if prepared posts hasn't been initialized yet
		return id, nil
	}

	e := t.preparedPosts.data.Back()
	if e == nil {
		// return an empty string if there's no back entry (ie., the index list hasn't been initialized yet)
		return id, nil
	}

	entry, ok := e.Value.(*preparedPostsEntry)
	if !ok {
		return id, errors.New("OldestPreparedPostID: could not parse e as a preparedPostsEntry")
	}
	return entry.statusID, nil
}
