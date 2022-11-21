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
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type statusDB struct {
	conn  *DBConn
	state *state.State
}

func (s *statusDB) newStatusQ(status interface{}) *bun.SelectQuery {
	return s.conn.
		NewSelect().
		Model(status).
		Relation("Attachments").
		Relation("Tags").
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
		"ID",
		func(status *gtsmodel.Status) error {
			return s.newStatusQ(status).Where("? = ?", bun.Ident("status.id"), id).Scan(ctx)
		},
		id,
	)
}

func (s *statusDB) GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, db.Error) {
	return s.getStatus(
		ctx,
		"URI",
		func(status *gtsmodel.Status) error {
			return s.newStatusQ(status).Where("? = ?", bun.Ident("status.uri"), uri).Scan(ctx)
		},
		uri,
	)
}

func (s *statusDB) GetStatusByURL(ctx context.Context, url string) (*gtsmodel.Status, db.Error) {
	return s.getStatus(
		ctx,
		"URL",
		func(status *gtsmodel.Status) error {
			return s.newStatusQ(status).Where("? = ?", bun.Ident("status.url"), url).Scan(ctx)
		},
		url,
	)
}

func (s *statusDB) getStatus(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Status) error, keyParts ...any) (*gtsmodel.Status, db.Error) {
	// Fetch status from database cache with loader callback
	status, err := s.state.Caches.GTS.Status().Load(lookup, func() (*gtsmodel.Status, error) {
		var status gtsmodel.Status

		// Not cached! Perform database query
		if err := dbQuery(&status); err != nil {
			return nil, s.conn.ProcessError(err)
		}

		if status.InReplyToID != "" {
			// Also load in-reply-to status
			status.InReplyTo = new(gtsmodel.Status)
			err := s.conn.NewSelect().Model(status.InReplyTo).
				Where("? = ?", bun.Ident("status.id"), status.InReplyToID).
				Scan(ctx)
			if err != nil {
				return nil, s.conn.ProcessError(err)
			}
		}

		if status.BoostOfID != "" {
			// Also load original boosted status
			status.BoostOf = new(gtsmodel.Status)
			err := s.conn.NewSelect().Model(status.BoostOf).
				Where("? = ?", bun.Ident("status.id"), status.BoostOfID).
				Scan(ctx)
			if err != nil {
				return nil, s.conn.ProcessError(err)
			}
		}

		return &status, nil
	}, keyParts...)
	if err != nil {
		// error already processed
		return nil, err
	}

	// Set the status author account
	status.Account, err = s.state.DB.GetAccountByID(ctx, status.AccountID)
	if err != nil {
		return nil, fmt.Errorf("error getting status account: %w", err)
	}

	if id := status.BoostOfAccountID; id != "" {
		// Set boost of status' author account
		status.BoostOfAccount, err = s.state.DB.GetAccountByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting boosted status account: %w", err)
		}
	}

	if id := status.InReplyToAccountID; id != "" {
		// Set in-reply-to status' author account
		status.InReplyToAccount, err = s.state.DB.GetAccountByID(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("error getting in reply to status account: %w", err)
		}
	}

	if len(status.EmojiIDs) > 0 {
		// Fetch status emojis
		status.Emojis, err = s.state.DB.GetEmojisByIDs(ctx, status.EmojiIDs)
		if err != nil {
			return nil, fmt.Errorf("error getting status emojis: %w", err)
		}
	}

	if len(status.MentionIDs) > 0 {
		// Fetch status mentions
		status.Mentions, err = s.state.DB.GetMentions(ctx, status.MentionIDs)
		if err != nil {
			return nil, fmt.Errorf("error getting status mentions: %w", err)
		}
	}

	return status, nil
}

func (s *statusDB) PutStatus(ctx context.Context, status *gtsmodel.Status) db.Error {
	return s.state.Caches.GTS.Status().Store(status, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return s.conn.RunInTx(ctx, func(tx bun.Tx) error {
			// create links between this status and any emojis it uses
			for _, i := range status.EmojiIDs {
				if _, err := tx.
					NewInsert().
					Model(&gtsmodel.StatusToEmoji{
						StatusID: status.ID,
						EmojiID:  i,
					}).Exec(ctx); err != nil {
					err = s.conn.ProcessError(err)
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
					err = s.conn.ProcessError(err)
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
					err = s.conn.ProcessError(err)
					if !errors.Is(err, db.ErrAlreadyExists) {
						return err
					}
				}
			}

			// Finally, insert the status
			_, err := tx.NewInsert().Model(status).Exec(ctx)
			return err
		})
	})
}

func (s *statusDB) UpdateStatus(ctx context.Context, status *gtsmodel.Status) db.Error {
	if err := s.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// create links between this status and any emojis it uses
		for _, i := range status.EmojiIDs {
			if _, err := tx.
				NewInsert().
				Model(&gtsmodel.StatusToEmoji{
					StatusID: status.ID,
					EmojiID:  i,
				}).Exec(ctx); err != nil {
				err = s.conn.ProcessError(err)
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
				err = s.conn.ProcessError(err)
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
				err = s.conn.ProcessError(err)
				if !errors.Is(err, db.ErrAlreadyExists) {
					return err
				}
			}
		}

		// Finally, insert the status
		_, err := tx.
			NewUpdate().
			Model(status).
			Where("? = ?", bun.Ident("status.id"), status.ID).
			Exec(ctx)
		return err
	}); err != nil {
		return err
	}

	// Drop any old value from cache by this ID
	s.state.Caches.GTS.Status().Invalidate("ID", status.ID)
	return nil
}

func (s *statusDB) DeleteStatusByID(ctx context.Context, id string) db.Error {
	if err := s.conn.RunInTx(ctx, func(tx bun.Tx) error {
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
	}); err != nil {
		return err
	}

	// Drop any old value from cache by this ID
	s.state.Caches.GTS.Status().Invalidate("ID", id)
	return nil
}

func (s *statusDB) GetStatusParents(ctx context.Context, status *gtsmodel.Status, onlyDirect bool) ([]*gtsmodel.Status, db.Error) {
	if onlyDirect {
		// Only want the direct parent, no further than first level
		parent, err := s.GetStatusByID(ctx, status.InReplyToID)
		if err != nil {
			return nil, err
		}
		return []*gtsmodel.Status{parent}, nil
	}

	var parents []*gtsmodel.Status

	for id := status.InReplyToID; id != ""; {
		parent, err := s.GetStatusByID(ctx, id)
		if err != nil {
			return nil, err
		}

		// Append parent to slice
		parents = append(parents, parent)

		// Set the next parent ID
		id = parent.InReplyToID
	}

	return parents, nil
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
	var childIDs []string

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
