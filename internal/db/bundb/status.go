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
	"context"
	"errors"
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type statusDB struct {
	db    *bun.DB
	state *state.State
}

func (s *statusDB) GetStatusByID(ctx context.Context, id string) (*gtsmodel.Status, error) {
	return s.getStatus(
		ctx,
		"ID",
		func(status *gtsmodel.Status) error {
			return s.db.NewSelect().Model(status).Where("? = ?", bun.Ident("status.id"), id).Scan(ctx)
		},
		id,
	)
}

func (s *statusDB) GetStatusesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Status, error) {
	// Load all input status IDs via cache loader callback.
	statuses, err := s.state.Caches.DB.Status.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Status, error) {
			// Preallocate expected length of uncached statuses.
			statuses := make([]*gtsmodel.Status, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) status IDs.
			if err := s.db.NewSelect().
				Model(&statuses).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return statuses, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the statuses by their
	// IDs to ensure in correct order.
	getID := func(s *gtsmodel.Status) string { return s.ID }
	xslices.OrderBy(statuses, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return statuses, nil
	}

	// Populate all loaded statuses, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	statuses = slices.DeleteFunc(statuses, func(status *gtsmodel.Status) bool {
		if err := s.PopulateStatus(ctx, status); err != nil {
			log.Errorf(ctx, "error populating status %s: %v", status.ID, err)
			return true
		}
		return false
	})

	return statuses, nil
}

func (s *statusDB) GetStatusByURI(ctx context.Context, uri string) (*gtsmodel.Status, error) {
	return s.getStatus(
		ctx,
		"URI",
		func(status *gtsmodel.Status) error {
			return s.db.NewSelect().Model(status).Where("? = ?", bun.Ident("status.uri"), uri).Scan(ctx)
		},
		uri,
	)
}

func (s *statusDB) GetStatusByURL(ctx context.Context, url string) (*gtsmodel.Status, error) {
	return s.getStatus(
		ctx,
		"URL",
		func(status *gtsmodel.Status) error {
			return s.db.NewSelect().Model(status).Where("? = ?", bun.Ident("status.url"), url).Scan(ctx)
		},
		url,
	)
}

func (s *statusDB) GetStatusByPollID(ctx context.Context, pollID string) (*gtsmodel.Status, error) {
	return s.getStatus(
		ctx,
		"PollID",
		func(status *gtsmodel.Status) error {
			return s.db.NewSelect().Model(status).Where("? = ?", bun.Ident("status.poll_id"), pollID).Scan(ctx)
		},
		pollID,
	)
}

func (s *statusDB) GetStatusBoost(ctx context.Context, boostOfID string, byAccountID string) (*gtsmodel.Status, error) {
	return s.getStatus(
		ctx,
		"BoostOfID,AccountID",
		func(status *gtsmodel.Status) error {
			return s.db.NewSelect().Model(status).
				Where("status.boost_of_id = ?", boostOfID).
				Where("status.account_id = ?", byAccountID).

				// Our old code actually allowed a status to
				// be boosted multiple times by the same author,
				// so limit our query + order to fetch latest.
				Order("status.id DESC"). // our IDs are timestamped
				Limit(1).
				Scan(ctx)
		},
		boostOfID, byAccountID,
	)
}

func (s *statusDB) getStatus(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Status) error, keyParts ...any) (*gtsmodel.Status, error) {
	// Fetch status from database cache with loader callback
	status, err := s.state.Caches.DB.Status.LoadOne(lookup, func() (*gtsmodel.Status, error) {
		var status gtsmodel.Status

		// Not cached! Perform database query.
		if err := dbQuery(&status); err != nil {
			return nil, err
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
	var (
		err  error
		errs = gtserror.NewMultiError(9)
	)

	if status.Account == nil {
		// Status author is not set, fetch from database.
		status.Account, err = s.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			status.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating status author: %w", err)
		}
	}

	if status.InReplyToID != "" {
		if status.InReplyTo == nil {
			// Status parent is not set, fetch from database.
			status.InReplyTo, err = s.GetStatusByID(
				gtscontext.SetBarebones(ctx),
				status.InReplyToID,
			)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				errs.Appendf("error populating status parent: %w", err)
			}
		}

		if status.InReplyToAccount == nil {
			// Status parent author is not set, fetch from database.
			status.InReplyToAccount, err = s.state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				status.InReplyToAccountID,
			)
			if err != nil {
				errs.Appendf("error populating status parent author: %w", err)
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
				errs.Appendf("error populating status boost: %w", err)
			}
		}

		if status.BoostOfAccount == nil {
			// Status boost author is not set, fetch from database.
			status.BoostOfAccount, err = s.state.DB.GetAccountByID(
				gtscontext.SetBarebones(ctx),
				status.BoostOfAccountID,
			)
			if err != nil {
				errs.Appendf("error populating status boost author: %w", err)
			}
		}
	}

	if status.PollID != "" && status.Poll == nil {
		// Status poll is not set, fetch from database.
		status.Poll, err = s.state.DB.GetPollByID(
			gtscontext.SetBarebones(ctx),
			status.PollID,
		)
		if err != nil {
			errs.Appendf("error populating status poll: %w", err)
		}
	}

	if !status.AttachmentsPopulated() {
		// Status attachments are out-of-date with IDs, repopulate.
		status.Attachments, err = s.state.DB.GetAttachmentsByIDs(
			ctx, // these are already barebones
			status.AttachmentIDs,
		)
		if err != nil {
			errs.Appendf("error populating status attachments: %w", err)
		}
	}

	if !status.TagsPopulated() {
		// Status tags are out-of-date with IDs, repopulate.
		status.Tags, err = s.state.DB.GetTags(
			ctx,
			status.TagIDs,
		)
		if err != nil {
			errs.Appendf("error populating status tags: %w", err)
		}
	}

	if !status.MentionsPopulated() {
		// Status mentions are out-of-date with IDs, repopulate.
		status.Mentions, err = s.state.DB.GetMentions(
			ctx, // leave fully populated for now
			status.MentionIDs,
		)
		if err != nil {
			errs.Appendf("error populating status mentions: %w", err)
		}
	}

	if !status.EmojisPopulated() {
		// Status emojis are out-of-date with IDs, repopulate.
		status.Emojis, err = s.state.DB.GetEmojisByIDs(
			ctx, // these are already barebones
			status.EmojiIDs,
		)
		if err != nil {
			errs.Appendf("error populating status emojis: %w", err)
		}
	}

	if status.CreatedWithApplicationID != "" && status.CreatedWithApplication == nil {
		// Populate the status' expected CreatedWithApplication (not always set).
		status.CreatedWithApplication, err = s.state.DB.GetApplicationByID(
			ctx, // these are already barebones
			status.CreatedWithApplicationID,
		)
		if err != nil {
			errs.Appendf("error populating status application: %w", err)
		}
	}

	return errs.Combine()
}

func (s *statusDB) PutStatus(ctx context.Context, status *gtsmodel.Status) error {
	return s.state.Caches.DB.Status.Store(status, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
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
					Column("status_id", "updated_at").
					Where("? = ?", bun.Ident("media_attachment.id"), a.ID).
					Exec(ctx); err != nil {
					if !errors.Is(err, db.ErrAlreadyExists) {
						return err
					}
				}
			}

			// If the status is threaded, create
			// link between thread and status.
			if status.ThreadID != "" {
				if _, err := tx.
					NewInsert().
					Model(&gtsmodel.ThreadToStatus{
						ThreadID: status.ThreadID,
						StatusID: status.ID,
					}).
					On("CONFLICT (?, ?) DO NOTHING", bun.Ident("thread_id"), bun.Ident("status_id")).
					Exec(ctx); err != nil {
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

func (s *statusDB) UpdateStatus(ctx context.Context, status *gtsmodel.Status, columns ...string) error {
	status.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column, ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return s.state.Caches.DB.Status.Store(status, func() error {
		// It is safe to run this database transaction within cache.Store
		// as the cache does not attempt a mutex lock until AFTER hook.
		//
		return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
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
					if !errors.Is(err, db.ErrAlreadyExists) {
						return err
					}
				}
			}

			// If the status is threaded, create
			// link between thread and status.
			if status.ThreadID != "" {
				if _, err := tx.
					NewInsert().
					Model(&gtsmodel.ThreadToStatus{
						ThreadID: status.ThreadID,
						StatusID: status.ID,
					}).
					On("CONFLICT (?, ?) DO NOTHING", bun.Ident("thread_id"), bun.Ident("status_id")).
					Exec(ctx); err != nil {
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
		})
	})
}

func (s *statusDB) DeleteStatusByID(ctx context.Context, id string) error {
	// Gather necessary fields from
	// deleted for cache invaliation.
	var deleted gtsmodel.Status
	deleted.ID = id

	// Delete status from database and any related links in a transaction.
	if err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
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

		// Delete links between this status
		// and any threads it was a part of.
		if _, err := tx.
			NewDelete().
			TableExpr("? AS ?", bun.Ident("thread_to_statuses"), bun.Ident("thread_to_status")).
			Where("? = ?", bun.Ident("thread_to_status.status_id"), id).
			Exec(ctx); err != nil {
			return err
		}

		// delete the status itself
		if _, err := tx.
			NewDelete().
			Model(&deleted).
			Where("? = ?", bun.Ident("id"), id).
			Returning("?, ?, ?, ?, ?",
				bun.Ident("account_id"),
				bun.Ident("boost_of_id"),
				bun.Ident("in_reply_to_id"),
				bun.Ident("attachments"),
				bun.Ident("poll_id"),
			).
			Exec(ctx); err != nil &&
			!errors.Is(err, db.ErrNoEntries) {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Invalidate cached status by its ID, manually
	// call the invalidate hook in case not cached.
	s.state.Caches.DB.Status.Invalidate("ID", id)
	s.state.Caches.OnInvalidateStatus(&deleted)

	return nil
}

func (s *statusDB) GetStatusesUsingEmoji(ctx context.Context, emojiID string) ([]*gtsmodel.Status, error) {
	var statusIDs []string

	// SELECT all statuses using this emoji,
	// using a relational table for improved perf.
	if _, err := s.db.NewSelect().
		Table("status_to_emojis").
		Column("status_id").
		Where("? = ?", bun.Ident("emoji_id"), emojiID).
		Exec(ctx, &statusIDs); err != nil {
		return nil, err
	}

	// Convert status IDs into status objects.
	return s.GetStatusesByIDs(ctx, statusIDs)
}

func (s *statusDB) GetStatusParents(ctx context.Context, status *gtsmodel.Status) ([]*gtsmodel.Status, error) {
	var parents []*gtsmodel.Status

	for id := status.InReplyToID; id != ""; {
		parent, err := s.GetStatusByID(ctx, id)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, err
		}

		if parent == nil {
			// Parent status not found (e.g. deleted)
			break
		}

		// Append parent status to slice
		parents = append(parents, parent)

		// Set the next parent ID
		id = parent.InReplyToID
	}

	return parents, nil
}

func (s *statusDB) GetStatusChildren(ctx context.Context, statusID string) ([]*gtsmodel.Status, error) {
	// Get all replies for the currently set status.
	replies, err := s.GetStatusReplies(ctx, statusID)
	if err != nil {
		return nil, err
	}

	// Make estimated preallocation based on direct replies.
	children := make([]*gtsmodel.Status, 0, len(replies)*2)

	for _, status := range replies {
		// Append status to children.
		children = append(children, status)

		// Further, recursively get all children for this reply.
		grandChildren, err := s.GetStatusChildren(ctx, status.ID)
		if err != nil {
			return nil, err
		}

		// Append all sub children after status.
		children = append(children, grandChildren...)
	}

	return children, nil
}

func (s *statusDB) GetStatusReplies(ctx context.Context, statusID string) ([]*gtsmodel.Status, error) {
	statusIDs, err := s.getStatusReplyIDs(ctx, statusID)
	if err != nil {
		return nil, err
	}
	return s.GetStatusesByIDs(ctx, statusIDs)
}

func (s *statusDB) CountStatusReplies(ctx context.Context, statusID string) (int, error) {
	statusIDs, err := s.getStatusReplyIDs(ctx, statusID)
	return len(statusIDs), err
}

func (s *statusDB) getStatusReplyIDs(ctx context.Context, statusID string) ([]string, error) {
	return s.state.Caches.DB.InReplyToIDs.Load(statusID, func() ([]string, error) {
		var statusIDs []string

		// Status reply IDs not in cache, perform DB query!
		if err := s.db.
			NewSelect().
			Table("statuses").
			Column("id").
			Where("? = ?", bun.Ident("in_reply_to_id"), statusID).
			Order("id DESC").
			Scan(ctx, &statusIDs); err != nil {
			return nil, err
		}

		return statusIDs, nil
	})
}

func (s *statusDB) GetStatusBoosts(ctx context.Context, statusID string) ([]*gtsmodel.Status, error) {
	statusIDs, err := s.getStatusBoostIDs(ctx, statusID)
	if err != nil {
		return nil, err
	}
	return s.GetStatusesByIDs(ctx, statusIDs)
}

func (s *statusDB) IsStatusBoostedBy(ctx context.Context, statusID string, accountID string) (bool, error) {
	boost, err := s.GetStatusBoost(
		gtscontext.SetBarebones(ctx),
		statusID,
		accountID,
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, err
	}
	return (boost != nil), nil
}

func (s *statusDB) CountStatusBoosts(ctx context.Context, statusID string) (int, error) {
	statusIDs, err := s.getStatusBoostIDs(ctx, statusID)
	return len(statusIDs), err
}

func (s *statusDB) getStatusBoostIDs(ctx context.Context, statusID string) ([]string, error) {
	return s.state.Caches.DB.BoostOfIDs.Load(statusID, func() ([]string, error) {
		var statusIDs []string

		// Status boost IDs not in cache, perform DB query!
		if err := s.db.
			NewSelect().
			Table("statuses").
			Column("id").
			Where("? = ?", bun.Ident("boost_of_id"), statusID).
			Order("id DESC").
			Scan(ctx, &statusIDs); err != nil {
			return nil, err
		}

		return statusIDs, nil
	})
}

func (s *statusDB) MaxDirectStatusID(ctx context.Context) (string, error) {
	maxID := ""
	if err := s.db.
		NewSelect().
		Model((*gtsmodel.Status)(nil)).
		ColumnExpr("COALESCE(MAX(?), '')", bun.Ident("id")).
		Where("? = ?", bun.Ident("visibility"), gtsmodel.VisibilityDirect).
		Scan(ctx, &maxID); // nocollapse
	err != nil {
		return "", err
	}
	return maxID, nil
}

func (s *statusDB) GetDirectStatusIDsBatch(ctx context.Context, minID string, maxIDInclusive string, count int) ([]string, error) {
	var statusIDs []string
	if err := s.db.
		NewSelect().
		Model((*gtsmodel.Status)(nil)).
		Column("id").
		Where("? = ?", bun.Ident("visibility"), gtsmodel.VisibilityDirect).
		Where("? > ?", bun.Ident("id"), minID).
		Where("? <= ?", bun.Ident("id"), maxIDInclusive).
		Order("id ASC").
		Limit(count).
		Scan(ctx, &statusIDs); // nocollapse
	err != nil {
		return nil, err
	}
	return statusIDs, nil
}
