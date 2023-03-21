// GoToSocial
// Copyright (C) GoToSocial Authors admin@gotosocial.org
// SPDX-License-Identifier: AGPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package bundb

import (
	"container/list"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
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
		Relation("Tags").
		Relation("CreatedWithApplication")
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

func (s *statusDB) GetStatuses(ctx context.Context, ids []string) ([]*gtsmodel.Status, db.Error) {
	statuses := make([]*gtsmodel.Status, 0, len(ids))

	for _, id := range ids {
		// Attempt fetch from DB
		status, err := s.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting status %q: %v", id, err)
			continue
		}

		// Append status
		statuses = append(statuses, status)
	}

	return statuses, nil
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

		// Not cached! Perform database query.
		if err := dbQuery(&status); err != nil {
			return nil, s.conn.ProcessError(err)
		}

		return &status, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return status, nil
	}

	// Further populate the status fields where applicable.
	if err := s.PopulateStatus(ctx, status); err != nil {
		return nil, err
	}

	return status, nil
}

func (s *statusDB) PopulateStatus(ctx context.Context, status *gtsmodel.Status) error {
	var err error

	if status.Account == nil {
		// Status author is not set, fetch from database.
		status.Account, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			status.AccountID,
		)
		if err != nil {
			return fmt.Errorf("error populating status author: %w", err)
		}
	}

	if status.InReplyToID != "" && status.InReplyTo == nil {
		// Status parent is not set, fetch from database.
		status.InReplyTo, err = s.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			status.InReplyToID,
		)
		if err != nil {
			return fmt.Errorf("error populating status parent: %w", err)
		}
	}

	if status.InReplyToID != "" {
		if status.InReplyTo == nil {
			// Status parent is not set, fetch from database.
			status.InReplyTo, err = s.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				status.InReplyToID,
			)
			if err != nil {
				return fmt.Errorf("error populating status parent: %w", err)
			}
		}

		if status.InReplyToAccount == nil {
			// Status parent author is not set, fetch from database.
			status.InReplyToAccount, err = s.state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				status.InReplyToAccountID,
			)
			if err != nil {
				return fmt.Errorf("error populating status parent author: %w", err)
			}
		}
	}

	if status.BoostOfID != "" {
		if status.BoostOf == nil {
			// Status boost is not set, fetch from database.
			status.BoostOf, err = s.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				status.BoostOfID,
			)
			if err != nil {
				return fmt.Errorf("error populating status boost: %w", err)
			}
		}

		if status.BoostOfAccount == nil {
			// Status boost author is not set, fetch from database.
			status.BoostOfAccount, err = s.state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				status.BoostOfAccountID,
			)
			if err != nil {
				return fmt.Errorf("error populating status boost author: %w", err)
			}
		}
	}

	if !status.AttachmentsPopulated() {
		// Status attachments are out-of-date with IDs, repopulate.
		status.Attachments, err = s.state.DB.GetAttachmentsByIDs(
			ctx, // these are already barebones
			status.AttachmentIDs,
		)
		if err != nil {
			return fmt.Errorf("error populating status attachments: %w", err)
		}
	}

	// TODO: once we don't fetch using relations.
	// if !status.TagsPopulated() {
	// }

	if !status.MentionsPopulated() {
		// Status mentions are out-of-date with IDs, repopulate.
		status.Mentions, err = s.state.DB.GetMentions(
			ctx, // leave fully populated for now
			status.MentionIDs,
		)
		if err != nil {
			return fmt.Errorf("error populating status mentions: %w", err)
		}
	}

	if !status.EmojisPopulated() {
		// Status emojis are out-of-date with IDs, repopulate.
		status.Emojis, err = s.state.DB.GetEmojisByIDs(
			ctx, // these are already barebones
			status.EmojiIDs,
		)
		if err != nil {
			return fmt.Errorf("error populating status emojis: %w", err)
		}
	}

	return nil
}

func (s *statusDB) PutStatus(ctx context.Context, status *gtsmodel.Status) db.Error {
	err := s.state.Caches.GTS.Status().Store(status, func() error {
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
					}).
					On("CONFLICT (?, ?) DO NOTHING", bun.Ident("status_id"), bun.Ident("emoji_id")).
					Exec(ctx); err != nil {
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
					}).
					On("CONFLICT (?, ?) DO NOTHING", bun.Ident("status_id"), bun.Ident("tag_id")).
					Exec(ctx); err != nil {
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
	if err != nil {
		return err
	}

	for _, id := range status.AttachmentIDs {
		// Invalidate media attachments from cache.
		//
		// NOTE: this is needed due to the way in which
		// we upload status attachments, and only after
		// update them with a known status ID. This is
		// not the case for header/avatar attachments.
		s.state.Caches.GTS.Media().Invalidate("ID", id)
	}

	return nil
}

func (s *statusDB) UpdateStatus(ctx context.Context, status *gtsmodel.Status, columns ...string) db.Error {
	status.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	if err := s.conn.RunInTx(ctx, func(tx bun.Tx) error {
		// create links between this status and any emojis it uses
		for _, i := range status.EmojiIDs {
			if _, err := tx.
				NewInsert().
				Model(&gtsmodel.StatusToEmoji{
					StatusID: status.ID,
					EmojiID:  i,
				}).
				On("CONFLICT (?, ?) DO NOTHING", bun.Ident("status_id"), bun.Ident("emoji_id")).
				Exec(ctx); err != nil {
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
				}).
				On("CONFLICT (?, ?) DO NOTHING", bun.Ident("status_id"), bun.Ident("tag_id")).
				Exec(ctx); err != nil {
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

		// Finally, update the status
		_, err := tx.
			NewUpdate().
			Model(status).
			Column(columns...).
			Where("? = ?", bun.Ident("status.id"), status.ID).
			Exec(ctx)
		return err
	}); err != nil {
		// already processed
		return err
	}

	// Invalidate status from database lookups.
	s.state.Caches.GTS.Status().Invalidate("ID", status.ID)

	for _, id := range status.AttachmentIDs {
		// Invalidate media attachments from cache.
		//
		// NOTE: this is needed due to the way in which
		// we upload status attachments, and only after
		// update them with a known status ID. This is
		// not the case for header/avatar attachments.
		s.state.Caches.GTS.Media().Invalidate("ID", id)
	}

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

	// Invalidate status from database lookups.
	s.state.Caches.GTS.Status().Invalidate("ID", id)

	// Invalidate status from all visibility lookups.
	s.state.Caches.Visibility.Invalidate("ItemID", id)

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
			log.Panic(ctx, "found status could not be asserted to *gtsmodel.Status")
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
			log.Errorf(ctx, "error getting children for %q: %v", status.ID, err)
		}
		return
	}

	for _, id := range childIDs {
		// Fetch child with ID from database
		child, err := s.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting child status %q: %v", id, err)
			continue
		}

	insertLoop:
		for e := foundStatuses.Front(); e != nil; e = e.Next() {
			entry, ok := e.Value.(*gtsmodel.Status)
			if !ok {
				log.Panic(ctx, "found status could not be asserted to *gtsmodel.Status")
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
