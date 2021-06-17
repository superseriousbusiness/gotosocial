package timeline

import (
	"errors"
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (t *timeline) IndexBefore(statusID string, include bool, amount int) error {
	return nil
}

func (t *timeline) IndexBehind(statusID string, amount int) error {
	filtered := []*gtsmodel.Status{}
	offsetStatus := statusID

grabloop:
	for len(filtered) < amount {
		statuses, err := t.db.GetStatusesWhereFollowing(t.accountID, offsetStatus, "", "", amount, false)
		if err != nil {
			if _, ok := err.(db.ErrNoEntries); ok {
				break grabloop // we just don't have enough statuses left in the db so index what we've got and then bail
			}
			return fmt.Errorf("IndexBehindAndIncluding: error getting statuses from db: %s", err)
		}

		for _, s := range statuses {
			timelineable, err := t.filter.StatusHometimelineable(s, t.account)
			if err != nil {
				continue
			}
			if timelineable {
				filtered = append(filtered, s)
			}
			offsetStatus = s.ID
		}
	}

	for _, s := range filtered {
		if err := t.IndexOne(s.CreatedAt, s.ID, s.BoostOfID); err != nil {
			return fmt.Errorf("IndexBehindAndIncluding: error indexing status with id %s: %s", s.ID, err)
		}
	}

	return nil
}

func (t *timeline) IndexOneByID(statusID string) error {
	return nil
}

func (t *timeline) IndexOne(statusCreatedAt time.Time, statusID string, boostOfID string) error {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		statusID:  statusID,
		boostOfID: boostOfID,
	}

	return t.postIndex.insertIndexed(postIndexEntry)
}

func (t *timeline) IndexAndPrepareOne(statusCreatedAt time.Time, statusID string) error {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		statusID: statusID,
	}

	if err := t.postIndex.insertIndexed(postIndexEntry); err != nil {
		return fmt.Errorf("IndexAndPrepareOne: error inserting indexed: %s", err)
	}

	if err := t.prepare(statusID); err != nil {
		return fmt.Errorf("IndexAndPrepareOne: error preparing: %s", err)
	}

	return nil
}

func (t *timeline) OldestIndexedPostID() (string, error) {
	var id string
	if t.postIndex == nil || t.postIndex.data == nil {
		// return an empty string if postindex hasn't been initialized yet
		return id, nil
	}

	e := t.postIndex.data.Back()

	if e == nil {
		// return an empty string if there's no back entry (ie., the index list hasn't been initialized yet)
		return id, nil
	}

	entry, ok := e.Value.(*postIndexEntry)
	if !ok {
		return id, errors.New("OldestIndexedPostID: could not parse e as a postIndexEntry")
	}
	return entry.statusID, nil
}
