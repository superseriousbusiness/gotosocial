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
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

func (ps *postgresService) StatusParents(status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, error) {
	parents := []*gtsmodel.Status{}
	ps.statusParent(status, &parents, onlyDirect)

	return parents, nil
}

func (ps *postgresService) statusParent(status *gtsmodel.Status, foundStatuses *[]*gtsmodel.Status, onlyDirect bool) {
	if status.InReplyToID == "" {
		return
	}

	parentStatus := &gtsmodel.Status{}
	if err := ps.conn.Model(parentStatus).Where("id = ?", status.InReplyToID).Select(); err == nil {
		*foundStatuses = append(*foundStatuses, parentStatus)
	}

	if onlyDirect {
		return
	}
	ps.statusParent(parentStatus, foundStatuses, false)
}

func (ps *postgresService) StatusChildren(status *gtsmodel.Status, onlyDirect bool, minID string) ([]*gtsmodel.Status, error) {
	foundStatuses := &list.List{}
	foundStatuses.PushFront(status)
	ps.statusChildren(status, foundStatuses, onlyDirect, minID)

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

func (ps *postgresService) statusChildren(status *gtsmodel.Status, foundStatuses *list.List, onlyDirect bool, minID string) {
	immediateChildren := []*gtsmodel.Status{}

	q := ps.conn.Model(&immediateChildren).Where("in_reply_to_id = ?", status.ID)
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
		ps.statusChildren(child, foundStatuses, false, minID)
	}
}
