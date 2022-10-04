/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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
	"database/sql"
	"errors"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/cache"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

type statusDB struct {
	conn  *DBConn
	cache *cache.StatusCache

	// TODO: keep method definitions in same place but instead have receiver
	//       all point to one single "db" type, so they can all share methods
	//       and caches where necessary
	accounts *accountDB
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

func (s *statusDB) newFaveQ(faves interface{}) *bun.SelectQuery {
	return s.conn.
		NewSelect().
		Model(faves).
		Relation("Account").
		Relation("TargetAccount").
		Relation("Status")
}

func (s *statusDB) GetStatusByID(ctx context.Context, id string) (*gtsmodel.Status, db.Error) {
	return s.getStatus(
		ctx,
		func() (*gtsmodel.Status, bool) {
			return s.cache.GetByID(id)
		},
		func(status *gtsmodel.Status) error {
			return s.newStatusQ(status).Where("? = ?", bun.Ident("status.id"), id).Scan(ctx)
		},
	)
}

func (s *statusDB) GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, db.Error) {
	return s.getStatus(
		ctx,
		func() (*gtsmodel.Status, bool) {
			return s.cache.GetByURI(uri)
		},
		func(status *gtsmodel.Status) error {
			return s.newStatusQ(status).Where("? = ?", bun.Ident("status.uri"), uri).Scan(ctx)
		},
	)
}

func (s *statusDB) GetStatusByURL(ctx context.Context, url string) (*gtsmodel.Status, db.Error) {
	return s.getStatus(
		ctx,
		func() (*gtsmodel.Status, bool) {
			return s.cache.GetByURL(url)
		},
		func(status *gtsmodel.Status) error {
			return s.newStatusQ(status).Where("? = ?", bun.Ident("status.url"), url).Scan(ctx)
		},
	)
}

func (s *statusDB) getStatus(ctx context.Context, cacheGet func() (*gtsmodel.Status, bool), dbQuery func(*gtsmodel.Status) error) (*gtsmodel.Status, db.Error) {
	// Attempt to fetch cached status
	status, cached := cacheGet()

	if !cached {
		status = &gtsmodel.Status{}

		// Not cached! Perform database query
		if err := dbQuery(status); err != nil {
			return nil, s.conn.ProcessError(err)
		}

		// If there is boosted, fetch from DB also
		if status.BoostOfID != "" {
			boostOf, err := s.GetStatusByID(ctx, status.BoostOfID)
			if err == nil {
				status.BoostOf = boostOf
			}
		}

		// Place in the cache
		s.cache.Put(status)
	}

	// Set the status author account
	author, err := s.accounts.GetAccountByID(ctx, status.AccountID)
	if err != nil {
		return nil, err
	}

	// Return the prepared status
	status.Account = author
	return status, nil
}

func (s *statusDB) PutStatus(ctx context.Context, status *gtsmodel.Status) db.Error {
	err := s.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// create links between this status and any emojis it uses
		for _, i := range status.EmojiIDs {
			if _, err := tx.
				NewInsert().
				Model(&gtsmodel.StatusToEmoji{
					StatusID: status.ID,
					EmojiID:  i,
				}).Exec(ctx); err != nil {
				err = s.conn.errProc(err)
				if !errors.Is(err, db.ErrAlreadyExists) {
					return err
				}
			}
		}

		// create links between this status and any tags it uses
		for _, i := range status.TagIDs {
			if _, err := tx.
				NewInsert().
				Model(&gtsmodel.StatusToTag{
					StatusID: status.ID,
					TagID:    i,
				}).Exec(ctx); err != nil {
				err = s.conn.errProc(err)
				if !errors.Is(err, db.ErrAlreadyExists) {
					return err
				}
			}
		}

		// change the status ID of the media attachments to the new status
		for _, a := range status.Attachments {
			a.StatusID = status.ID
			a.UpdatedAt = time.Now()
			if _, err := tx.
				NewUpdate().
				Model(a).
				Where("? = ?", bun.Ident("media_attachment.id"), a.ID).
				Exec(ctx); err != nil {
				err = s.conn.errProc(err)
				if !errors.Is(err, db.ErrAlreadyExists) {
					return err
				}
			}
		}

		// Finally, insert the status
		if _, err := tx.
			NewInsert().
			Model(status).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		s.conn.ProcessError(err)
	}

	s.cache.Put(status)
	return nil
}

func (s *statusDB) UpdateStatus(ctx context.Context, status *gtsmodel.Status) (*gtsmodel.Status, db.Error) {
	err := s.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// create links between this status and any emojis it uses
		for _, i := range status.EmojiIDs {
			if _, err := tx.
				NewInsert().
				Model(&gtsmodel.StatusToEmoji{
					StatusID: status.ID,
					EmojiID:  i,
				}).Exec(ctx); err != nil {
				err = s.conn.errProc(err)
				if !errors.Is(err, db.ErrAlreadyExists) {
					return err
				}
			}
		}

		// create links between this status and any tags it uses
		for _, i := range status.TagIDs {
			if _, err := tx.
				NewInsert().
				Model(&gtsmodel.StatusToTag{
					StatusID: status.ID,
					TagID:    i,
				}).Exec(ctx); err != nil {
				err = s.conn.errProc(err)
				if !errors.Is(err, db.ErrAlreadyExists) {
					return err
				}
			}
		}

		// change the status ID of the media attachments to this status
		for _, a := range status.Attachments {
			a.StatusID = status.ID
			a.UpdatedAt = time.Now()
			if _, err := tx.
				NewUpdate().
				Model(a).
				Where("? = ?", bun.Ident("media_attachment.id"), a.ID).
				Exec(ctx); err != nil {
				return err
			}
		}

		// Finally, update the status itself
		if _, err := tx.
			NewUpdate().
			Model(status).
			Where("? = ?", bun.Ident("status.id"), status.ID).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	s.cache.Put(status)
	return status, nil
}

func (s *statusDB) DeleteStatusByID(ctx context.Context, id string) db.Error {
	err := s.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// delete links between this status and any emojis it uses
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("status_to_emojis"), bun.Ident("status_to_emoji")).
			Where("? = ?", bun.Ident("status_to_emoji.status_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// delete links between this status and any tags it uses
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("status_to_tags"), bun.Ident("status_to_tag")).
			Where("? = ?", bun.Ident("status_to_tag.status_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// delete the status itself
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
			Where("? = ?", bun.Ident("status.id"), id).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return s.conn.ProcessError(err)
	}

	s.cache.Invalidate(id)
	return nil
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
		// only append children, not the overall parent status
		entry, ok := e.Value.(*gtsmodel.Status)
		if !ok {
			log.Panic("GetStatusChildren: found status could not be asserted to *gtsmodel.Status")
		}

		if entry.ID != status.ID {
			children = append(children, entry)
		}
	}

	return children, nil
}

func (s *statusDB) statusChildren(ctx context.Context, status *gtsmodel.Status, foundStatuses *list.List, onlyDirect bool, minID string) {
	childIDs := []string{}

	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.id").
		Where("? = ?", bun.Ident("status.in_reply_to_id"), status.ID)
	if minID != "" {
		q = q.Where("? > ?", bun.Ident("status.id"), minID)
	}

	if err := q.Scan(ctx, &childIDs); err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("statusChildren: error getting children for %q: %v", status.ID, err)
		}
		return
	}

	for _, id := range childIDs {
		// Fetch child with ID from database
		child, err := s.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf("statusChildren: error getting child status %q: %v", id, err)
			continue
		}

	insertLoop:
		for e := foundStatuses.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*gtsmodel.Status)
			if !ok {
				log.Panic("statusChildren: found status could not be asserted to *gtsmodel.Status")
			}

			if child.InReplyToAccountID != "" && entry.ID == child.InReplyToID {
				foundStatuses.InsertAfter(child, e)
				break insertLoop
			}
		}

		// if we're not only looking for direct children of status, then do the same children-finding
		// operation for the found child status too.
		if !onlyDirect {
			s.statusChildren(ctx, child, foundStatuses, false, minID)
		}
	}
}

func (s *statusDB) CountStatusReplies(ctx context.Context, status *gtsmodel.Status) (int, db.Error) {
	return s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Where("? = ?", bun.Ident("status.in_reply_to_id"), status.ID).
		Count(ctx)
}

func (s *statusDB) CountStatusReblogs(ctx context.Context, status *gtsmodel.Status) (int, db.Error) {
	return s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Where("? = ?", bun.Ident("status.boost_of_id"), status.ID).
		Count(ctx)
}

func (s *statusDB) CountStatusFaves(ctx context.Context, status *gtsmodel.Status) (int, db.Error) {
	return s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave")).
		Where("? = ?", bun.Ident("status_fave.status_id"), status.ID).
		Count(ctx)
}

func (s *statusDB) IsStatusFavedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave")).
		Where("? = ?", bun.Ident("status_fave.status_id"), status.ID).
		Where("? = ?", bun.Ident("status_fave.account_id"), accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) IsStatusRebloggedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Where("? = ?", bun.Ident("status.boost_of_id"), status.ID).
		Where("? = ?", bun.Ident("status.account_id"), accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) IsStatusMutedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_mutes"), bun.Ident("status_mute")).
		Where("? = ?", bun.Ident("status_mute.status_id"), status.ID).
		Where("? = ?", bun.Ident("status_mute.account_id"), accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) IsStatusBookmarkedBy(ctx context.Context, status *gtsmodel.Status, accountID string) (bool, db.Error) {
	q := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_bookmarks"), bun.Ident("status_bookmark")).
		Where("? = ?", bun.Ident("status_bookmark.status_id"), status.ID).
		Where("? = ?", bun.Ident("status_bookmark.account_id"), accountID)

	return s.conn.Exists(ctx, q)
}

func (s *statusDB) GetStatusFaves(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.StatusFave, db.Error) {
	faves := []*gtsmodel.StatusFave{}

	q := s.
		newFaveQ(&faves).
		Where("? = ?", bun.Ident("status_fave.status_id"), status.ID)

	if err := q.Scan(ctx); err != nil {
		return nil, s.conn.ProcessError(err)
	}
	return faves, nil
}

func (s *statusDB) GetStatusReblogs(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.Status, db.Error) {
	reblogs := []*gtsmodel.Status{}

	q := s.
		newStatusQ(&reblogs).
		Where("? = ?", bun.Ident("status.boost_of_id"), status.ID)

	if err := q.Scan(ctx); err != nil {
		return nil, s.conn.ProcessError(err)
	}
	return reblogs, nil
}
