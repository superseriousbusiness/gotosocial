package timeline

import (
	"fmt"
	"time"
)

func (t *timeline) IndexBefore(statusID string, include bool, amount int) error {
	// 	filtered := []*gtsmodel.Status{}
	// 	offsetStatus := statusID

	// grabloop:
	// 	for len(filtered) < amount {
	// 		statuses, err := t.db.GetStatusesWhereFollowing(t.accountID, amount, offsetStatus, include, true)
	// 		if err != nil {
	// 			if _, ok := err.(db.ErrNoEntries); !ok {
	// 				return fmt.Errorf("IndexBeforeAndIncluding: error getting statuses from db: %s", err)
	// 			}
	// 			break grabloop // we just don't have enough statuses left in the db so index what we've got and then bail
	// 		}

	// 		for _, s := range statuses {
	// 			relevantAccounts, err := t.db.PullRelevantAccountsFromStatus(s)
	// 			if err != nil {
	// 				continue
	// 			}
	// 			visible, err := t.db.StatusVisible(s, t.account, relevantAccounts)
	// 			if err != nil {
	// 				continue
	// 			}
	// 			if visible {
	// 				filtered = append(filtered, s)
	// 			}
	// 			offsetStatus = s.ID
	// 		}
	// 	}

	// 	for _, s := range filtered {
	// 		if err := t.IndexOne(s.CreatedAt, s.ID); err != nil {
	// 			return fmt.Errorf("IndexBeforeAndIncluding: error indexing status with id %s: %s", s.ID, err)
	// 		}
	// 	}

	return nil
}

func (t *timeline) IndexBehind(statusID string, include bool, amount int) error {
	// 	filtered := []*gtsmodel.Status{}
	// 	offsetStatus := statusID

	// grabloop:
	// 	for len(filtered) < amount {
	// 		statuses, err := t.db.GetStatusesWhereFollowing(t.accountID, amount, offsetStatus, include, false)
	// 		if err != nil {
	// 			if _, ok := err.(db.ErrNoEntries); !ok {
	// 				return fmt.Errorf("IndexBehindAndIncluding: error getting statuses from db: %s", err)
	// 			}
	// 			break grabloop // we just don't have enough statuses left in the db so index what we've got and then bail
	// 		}

	// 		for _, s := range statuses {
	// 			relevantAccounts, err := t.db.PullRelevantAccountsFromStatus(s)
	// 			if err != nil {
	// 				continue
	// 			}
	// 			visible, err := t.db.StatusVisible(s, t.account, relevantAccounts)
	// 			if err != nil {
	// 				continue
	// 			}
	// 			if visible {
	// 				filtered = append(filtered, s)
	// 			}
	// 			offsetStatus = s.ID
	// 		}
	// 	}

	// 	for _, s := range filtered {
	// 		if err := t.IndexOne(s.CreatedAt, s.ID); err != nil {
	// 			return fmt.Errorf("IndexBehindAndIncluding: error indexing status with id %s: %s", s.ID, err)
	// 		}
	// 	}

	return nil
}

func (t *timeline) IndexOneByID(statusID string) error {
	return nil
}

func (t *timeline) IndexOne(statusCreatedAt time.Time, statusID string) error {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		createdAt: statusCreatedAt,
		statusID:  statusID,
	}

	return t.postIndex.insertIndexed(postIndexEntry)
}

func (t *timeline) IndexAndPrepareOne(statusCreatedAt time.Time, statusID string) error {
	t.Lock()
	defer t.Unlock()

	postIndexEntry := &postIndexEntry{
		createdAt: statusCreatedAt,
		statusID:  statusID,
	}

	if err := t.postIndex.insertIndexed(postIndexEntry); err != nil {
		return fmt.Errorf("IndexAndPrepareOne: error inserting indexed: %s", err)
	}

	if err := t.prepare(statusID); err != nil {
		return fmt.Errorf("IndexAndPrepareOne: error preparing: %s", err)
	}

	return nil
}
