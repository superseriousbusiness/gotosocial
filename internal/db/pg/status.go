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

package pg

import (
	"container/list"
	"context"
	"errors"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/sirupsen/logrus"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type statusDB struct {
	config *config.Config
	conn   *pg.DB
	log    *logrus.Logger
	cancel context.CancelFunc
}

func (s *statusDB) newStatusQ(status interface{}) *orm.Query {
	return s.conn.Model(status).
		Relation("Attachments").
		Relation("Tags").
		Relation("Mentions").
		Relation("Emojis").
		Relation("Account").
		Relation("InReplyTo").
		Relation("InReplyToAccount").
		Relation("BoostOf").
		Relation("BoostOfAccount").
		Relation("CreatedWithApplication")
}

func (s *statusDB) newFaveQ(faves interface{}) *orm.Query {
	return s.conn.Model(faves).
		Relation("Account").
		Relation("TargetAccount").
		Relation("Status")
}

func (s *statusDB) GetStatusByID(id string) (*gtsmodel.Status, db.DBError) {
	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("status.id = ?", id)

	err := processErrorResponse(q.Select())

	return status, err
}

func (s *statusDB) GetStatusByURI(uri string) (*gtsmodel.Status, db.DBError) {
	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("LOWER(status.uri) = LOWER(?)", uri)

	err := processErrorResponse(q.Select())

	return status, err
}

func (s *statusDB) GetStatusByURL(uri string) (*gtsmodel.Status, db.DBError) {
	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("LOWER(status.url) = LOWER(?)", uri)

	err := processErrorResponse(q.Select())

	return status, err
}

func (s *statusDB) PutStatus(status *gtsmodel.Status) db.DBError {
	transaction := func(tx *pg.Tx) error {
		// create links between this status and any emojis it uses
		for _, i := range status.EmojiIDs {
			if _, err := tx.Model(&gtsmodel.StatusToEmoji{
				StatusID: status.ID,
				EmojiID:  i,
			}).Insert(); err != nil {
				return err
			}
		}

		// create links between this status and any tags it uses
		for _, i := range status.TagIDs {
			if _, err := tx.Model(&gtsmodel.StatusToTag{
				StatusID: status.ID,
				TagID:    i,
			}).Insert(); err != nil {
				return err
			}
		}

		// change the status ID of the media attachments to the new status
		for _, a := range status.Attachments {
			a.StatusID = status.ID
			a.UpdatedAt = time.Now()
			if _, err := s.conn.Model(a).
				Where("id = ?", a.ID).
				Update(); err != nil {
				return err
			}
		}

		_, err := tx.Model(status).Insert()
		return err
	}

	return processErrorResponse(s.conn.RunInTransaction(context.Background(), transaction))
}

func (s *statusDB) GetStatusParents(status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, db.DBError) {
	parents := []*gtsmodel.Status{}
	s.statusParent(status, &parents, onlyDirect)

	return parents, nil
}

func (s *statusDB) statusParent(status *gtsmodel.Status, foundStatuses *[]*gtsmodel.Status, onlyDirect bool) {
	if status.InReplyToID == "" {
		return
	}

	parentStatus, err := s.GetStatusByID(status.InReplyToID)
	if err == nil {
		*foundStatuses = append(*foundStatuses, parentStatus)
	}

	if onlyDirect {
		return
	}

	s.statusParent(parentStatus, foundStatuses, false)
}

func (s *statusDB) GetStatusChildren(status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, db.DBError) {
	foundStatuses := &list.List{}
	foundStatuses.PushFront(status)
	s.statusChildren(status, foundStatuses, onlyDirect, minID)

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

func (s *statusDB) statusChildren(status *gtsmodel.Status, foundStatuses *list.List, onlyDirect bool, minID string) {
	immediateChildren := []*gtsmodel.Status{}

	q := s.conn.Model(&immediateChildren).Where("in_reply_to_id = ?", status.ID)
	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if err := q.Select(); err != nil {
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
		s.statusChildren(child, foundStatuses, false, minID)
	}
}

func (s *statusDB) CountStatusReplies(status *gtsmodel.Status) (int, db.DBError) {
	return s.conn.Model(&gtsmodel.Status{}).Where("in_reply_to_id = ?", status.ID).Count()
}

func (s *statusDB) CountStatusReblogs(status *gtsmodel.Status) (int, db.DBError) {
	return s.conn.Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Count()
}

func (s *statusDB) CountStatusFaves(status *gtsmodel.Status) (int, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Count()
}

func (s *statusDB) IsStatusFavedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) IsStatusRebloggedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) IsStatusMutedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusMute{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) IsStatusBookmarkedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusBookmark{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) GetStatusFaves(status *gtsmodel.Status) ([]*gtsmodel.StatusFave, db.DBError) {
	faves := []*gtsmodel.StatusFave{}

	q := s.newFaveQ(&faves).
		Where("status_id = ?", status.ID)

	err := processErrorResponse(q.Select())

	return faves, err
}

func (s *statusDB) GetStatusReblogs(status *gtsmodel.Status) ([]*gtsmodel.Status, db.DBError) {
	reblogs := []*gtsmodel.Status{}

	q := s.newStatusQ(&reblogs).
		Where("boost_of_id = ?", status.ID)

	err := processErrorResponse(q.Select())

	return reblogs, err
}
