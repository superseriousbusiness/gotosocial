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

func (s *statusDB) newStatusQ(status *gtsmodel.Status) *orm.Query {
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

func (s *statusDB) processStatusResponse(status *gtsmodel.Status, err error) (*gtsmodel.Status, db.DBError) {
	switch err {
	case pg.ErrNoRows:
		return nil, db.ErrNoEntries
	case nil:
		return status, nil
	default:
		return nil, err
	}
}

func (s *statusDB) GetStatusByID(id string) (*gtsmodel.Status, db.DBError) {
	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("status.id = ?", id)

	return s.processStatusResponse(status, q.Select())
}

func (s *statusDB) GetStatusByURI(uri string) (*gtsmodel.Status, db.DBError) {
	status := &gtsmodel.Status{}

	q := s.newStatusQ(status).
		Where("LOWER(status.uri) = LOWER(?)", uri)

	return s.processStatusResponse(status, q.Select())
}

func (s *statusDB) StatusParents(status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, db.DBError) {
	parents := []*gtsmodel.Status{}
	s.statusParent(status, &parents, onlyDirect)

	return parents, nil
}

func (s *statusDB) statusParent(status *gtsmodel.Status, foundStatuses *[]*gtsmodel.Status, onlyDirect bool) {
	if status.InReplyToID == "" {
		return
	}

	parentStatus := &gtsmodel.Status{}
	if err := s.conn.Model(parentStatus).Where("id = ?", status.InReplyToID).Select(); err == nil {
		*foundStatuses = append(*foundStatuses, parentStatus)
	}

	if onlyDirect {
		return
	}
	s.statusParent(parentStatus, foundStatuses, false)
}

func (s *statusDB) StatusChildren(status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, db.DBError) {
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

func (s *statusDB) GetReplyCountForStatus(status *gtsmodel.Status) (int, db.DBError) {
	return s.conn.Model(&gtsmodel.Status{}).Where("in_reply_to_id = ?", status.ID).Count()
}

func (s *statusDB) GetReblogCountForStatus(status *gtsmodel.Status) (int, db.DBError) {
	return s.conn.Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Count()
}

func (s *statusDB) GetFaveCountForStatus(status *gtsmodel.Status) (int, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Count()
}

func (s *statusDB) StatusFavedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusFave{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) StatusRebloggedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.Status{}).Where("boost_of_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) StatusMutedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusMute{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) StatusBookmarkedBy(status *gtsmodel.Status, accountID string) (bool, db.DBError) {
	return s.conn.Model(&gtsmodel.StatusBookmark{}).Where("status_id = ?", status.ID).Where("account_id = ?", accountID).Exists()
}

func (s *statusDB) WhoFavedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, db.DBError) {
	accounts := []*gtsmodel.Account{}

	faves := []*gtsmodel.StatusFave{}
	if err := s.conn.Model(&faves).Where("status_id = ?", status.ID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return accounts, nil // no rows just means nobody has faved this status, so that's fine
		}
		return nil, err // an actual error has occurred
	}

	for _, f := range faves {
		acc := &gtsmodel.Account{}
		if err := s.conn.Model(acc).Where("id = ?", f.AccountID).Select(); err != nil {
			if err == pg.ErrNoRows {
				continue // the account doesn't exist for some reason??? but this isn't the place to worry about that so just skip it
			}
			return nil, err // an actual error has occurred
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (s *statusDB) WhoBoostedStatus(status *gtsmodel.Status) ([]*gtsmodel.Account, db.DBError) {
	accounts := []*gtsmodel.Account{}

	boosts := []*gtsmodel.Status{}
	if err := s.conn.Model(&boosts).Where("boost_of_id = ?", status.ID).Select(); err != nil {
		if err == pg.ErrNoRows {
			return accounts, nil // no rows just means nobody has boosted this status, so that's fine
		}
		return nil, err // an actual error has occurred
	}

	for _, f := range boosts {
		acc := &gtsmodel.Account{}
		if err := s.conn.Model(acc).Where("id = ?", f.AccountID).Select(); err != nil {
			if err == pg.ErrNoRows {
				continue // the account doesn't exist for some reason??? but this isn't the place to worry about that so just skip it
			}
			return nil, err // an actual error has occurred
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}
