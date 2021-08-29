/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb

import (
	"container/list"
	"context"
	"errors"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type statusDB struct {
	config *config.Config
	conn   *DBConn
	cache  *cache.StatusCache
}

func (s *statusDB) newStatusQ(status interface{}) *bun.SelectQuery {
	return s.conn.
		NewSelect().
		Model(status).
		Relation("Attachments").
		Relation("Tags").
		Relation("Mentions").
		Relation("Emojis").
		Relation("Account").
		Relation("InReplyToAccount").
		Relation("BoostOfAccount").
		Relation("CreatedWithApplication")
}

func (s *statusDB) getAttachedStatuses(ctx context.Context, status *gtsmodel.Status) *gtsmodel.Status {
	if status.InReplyToID != "" && status.InReplyTo == nil {
		// TODO: do we want to keep this possibly recursive strategy?

		if inReplyTo, cached := s.cache.GetByID(status.InReplyToID); cached {
			status.InReplyTo = inReplyTo
		} else if inReplyTo, err := s.GetStatusByID(ctx, status.InReplyToID); err == nil {
			status.InReplyTo = inReplyTo
		}
	}

	if status.BoostOfID != "" && status.BoostOf == nil {
		// TODO: do we want to keep this possibly recursive strategy?

		if boostOf, cached := s.cache.GetByID(status.BoostOfID); cached {
			status.BoostOf = boostOf
		} else if boostOf, err := s.GetStatusByID(ctx, status.BoostOfID); err == nil {
			status.BoostOf = boostOf
		}
	}

	return status
}

func (s *statusDB) newFaveQ(faves interface{}) *bun.SelectQuery {
	return s.conn.
		NewSelect().
		Model(faves).
		Relation("Account").
		Relation("TargetAccount").
		Relation("Status")
}

func (s *statusDB) GetStatusByID(ctx context.Context, id string) (*gtsmodel.Status, db.Error) {
	if status, cached := s.cache.GetByID(id); cached {
		return status, nil
	}

	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("status.id = ?", id)

	err := q.Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	s.cache.Put(status)
	return s.getAttachedStatuses(ctx, status), nil
}

func (s *statusDB) GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, db.Error) {
	if status, cached := s.cache.GetByURI(uri); cached {
		return status, nil
	}

	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("LOWER(status.uri) = LOWER(?)", uri)

	err := q.Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	s.cache.Put(status)
	return s.getAttachedStatuses(ctx, status), nil
}

func (s *statusDB) GetStatusByURL(ctx context.Context, url string) (*gtsmodel.Status, db.Error) {
	if status, cached := s.cache.GetByURL(url); cached {
		return status, nil
	}

	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("LOWER(status.url) = LOWER(?)", url)

	err := q.Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	s.cache.Put(status)
	return s.getAttachedStatuses(ctx, status), nil
}

func (s *statusDB) PutStatus(ctx context.Context, status *gtsmodel.Status) db.Error {
	transaction := func(ctx context.Context, tx bun.Tx) error {
		// create links between this status and any emojis it uses
		for _, i := range status.EmojiIDs {
			if _, err := tx.NewInsert().Model(&gtsmodel.StatusToEmoji{
				StatusID: status.ID,
				EmojiID:  i,
			}).Exec(ctx); err != nil {
				return err
			}
		}

		// create links between this status and any tags it uses
		for _, i := range status.TagIDs {
			if _, err := tx.NewInsert().Model(&gtsmodel.StatusToTag{
				StatusID: status.ID,
				TagID:    i,
			}).Exec(ctx); err != nil {
				return err
			}
		}

		// change the status ID of the media attachments to the new status
		for _, a := range status.Attachments {
			a.StatusID = status.ID
			a.UpdatedAt = time.Now()
			if _, err := tx.NewUpdate().Model(a).
				Where("id = ?", a.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		_, err := tx.NewInsert().Model(status).Exec(ctx)
		return err
	}
	return s.conn.ProcessError(s.conn.RunInTx(ctx, nil, transaction))
}

func (s *statusDB) GetStatusParents(ctx context.Context, status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, db.Error) {
	parents := []*gtsmodel.Status{}
	s.statusParent(ctx, status, &parents, onlyDirect)
	return parents, nil
}

func (s *statusDB) statusParent(ctx context.Context, status *gtsmodel.Status, foundStatuses *[]*gtsmodel.Status, onlyDirect bool) {
	if status.InReplyToID == "" {
		return
	}

	parentStatus, err := s.GetStatusByID(ctx, status.InReplyToID)
	if err == nil {
		*foundStatuses = append(*foundStatuses, parentStatus)
	}

	if onlyDirect {
		return
	}

	s.statusParent(ctx, parentStatus, foundStatuses, false)
}

func (s *statusDB) GetStatusChildren(ctx context.Context, status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, db.Error) {
	foundStatuses := &list.List{}
	foundStatuses.PushFront(status)
	s.statusChildren(ctx, status, foundStatuses, onlyDirect, minID)

	children := []*gtsmodel.Status{}
	for e := foundStatuses.Front(); e != nil; e = e.Next() {
		entry, ok := e.Value.(*gtsmodel.Status)
		if !ok {
			panic(errors.New("entry in foundStatuses was not a *gtsmodel.Status"))
		}

		// only append children, not the overall parent status
		if entry.ID != status.ID {
			children = append(children, entry)
		}
	}

	return children, nil
}

func (s *statusDB) statusChildren(ctx context.Context, status *gtsmodel.Status, foundStatuses *list.List, onlyDirect bool, minID string) {
	immediateChildren := []*gtsmodel.Status{}

	q := s.conn.
		NewSelect().
		Model(&immediateChildren).
		Where("in_reply_to_id = ?", status.ID)
	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if err := q.Scan(ctx); err != nil {
		return
	}

	for _, child := range immediateChildren {
	insertLoop:
		for e := foundStatuses.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*gtsmodel.Status)
			if !ok {
				panic(errors.New("entry in foundStatuses was not a *gtsmodel.Status"))
			}

			if child.InReplyToAccountID != "" && entry.ID == child.InReplyToID {
				foundStatuses.InsertAfter(child, e)
				break insertLoop
			}
		}

		// only do one loop if we only want direct children
		if onlyDirect {
			return
		}
		s.statusChildren(ctx, child, foundStatuses, false, minID)
	}
}

func (s *statusDB) CountStatusReplies(ctx context.Context, status *gtsmodel.Status) (int, db.Error) {
	return s.conn.NewSelect().Model(&gtsmodel.Status{}).Where("in_reply_to_id = ?", status.ID).Count(ctx)
}

func (s *statusDB) CountStatusReblogs(ctx context.Context, status *gtsmodel.Status) (int, db.Error) {
	return s.conn.NewSelect().Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Count(ctx)
}

func (s *statusDB) CountStatusFaves(ctx context.Context, status *gtsmodel.Status) (int, db.Error) {
	return s.conn.NewSelect().Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Count(ctx)
}

func (s *statusDB) IsStatusFavedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		Model(&gtsmodel.StatusFave{}).
		Where("status_id = ?", status.ID).
		Where("account_id = ?", accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) IsStatusRebloggedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		Model(&gtsmodel.Status{}).
		Where("boost_of_id = ?", status.ID).
		Where("account_id = ?", accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) IsStatusMutedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		Model(&gtsmodel.StatusMute{}).
		Where("status_id = ?", status.ID).
		Where("account_id = ?", accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) IsStatusBookmarkedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		Model(&gtsmodel.StatusBookmark{}).
		Where("status_id = ?", status.ID).
		Where("account_id = ?", accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) GetStatusFaves(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.StatusFave, db.Error) {
	faves := []*gtsmodel.StatusFave{}

	q := s.newFaveQ(&faves).
		Where("status_id = ?", status.ID)

	err := q.Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}
	return faves, nil
}

func (s *statusDB) GetStatusReblogs(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.Status, db.Error) {
	reblogs := []*gtsmodel.Status{}

	q := s.newStatusQ(&reblogs).
		Where("boost_of_id = ?", status.ID)

	err := q.Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}
	return reblogs, nil
}
