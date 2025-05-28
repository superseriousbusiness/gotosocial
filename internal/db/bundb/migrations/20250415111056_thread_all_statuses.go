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

package migrations

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"slices"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250415111056_thread_all_statuses/new"
	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250415111056_thread_all_statuses/old"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		newType := reflect.TypeOf(&newmodel.Status{})

		// Get the new column definition with not-null thread_id.
		newColDef, err := getBunColumnDef(db, newType, "ThreadID")
		if err != nil {
			return gtserror.Newf("error getting bun column def: %w", err)
		}

		// Update column def to use '${name}_new'.
		newColDef = strings.Replace(newColDef,
			"thread_id", "thread_id_new", 1)

		var sr statusRethreader
		var count int
		var maxID string
		var statuses []*oldmodel.Status

		// Get a total count of all statuses before migration.
		total, err := db.NewSelect().Table("statuses").Count(ctx)
		if err != nil {
			return gtserror.Newf("error getting status table count: %w", err)
		}

		// Start at largest
		// possible ULID value.
		maxID = id.Highest

		log.Warn(ctx, "rethreading top-level statuses, this will take a *long* time")
		for /* TOP LEVEL STATUS LOOP */ {

			// Reset slice.
			clear(statuses)
			statuses = statuses[:0]

			// Select top-level statuses.
			if err := db.NewSelect().
				Model(&statuses).
				Column("id", "thread_id").

				// We specifically use in_reply_to_account_id instead of in_reply_to_id as
				// they should both be set / unset in unison, but we specifically have an
				// index on in_reply_to_account_id with ID ordering, unlike in_reply_to_id.
				Where("? IS NULL", bun.Ident("in_reply_to_account_id")).
				Where("? < ?", bun.Ident("id"), maxID).
				OrderExpr("? DESC", bun.Ident("id")).
				Limit(5000).
				Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gtserror.Newf("error selecting top level statuses: %w", err)
			}

			// Reached end of block.
			if len(statuses) == 0 {
				break
			}

			// Set next maxID value from statuses.
			maxID = statuses[len(statuses)-1].ID

			// Rethread each selected batch of top-level statuses in a transaction.
			if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

				// Rethread each top-level status.
				for _, status := range statuses {
					n, err := sr.rethreadStatus(ctx, tx, status)
					if err != nil {
						return gtserror.Newf("error rethreading status %s: %w", status.URI, err)
					}
					count += n
				}

				return nil
			}); err != nil {
				return err
			}

			log.Infof(ctx, "[approx %d of %d] rethreading statuses (top-level)", count, total)
		}

		// Attempt to merge any sqlite write-ahead-log.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		log.Warn(ctx, "rethreading straggler statuses, this will take a *long* time")
		for /* STRAGGLER STATUS LOOP */ {

			// Reset slice.
			clear(statuses)
			statuses = statuses[:0]

			// Select straggler statuses.
			if err := db.NewSelect().
				Model(&statuses).
				Column("id", "in_reply_to_id", "thread_id").
				Where("? IS NULL", bun.Ident("thread_id")).

				// We select in smaller batches for this part
				// of the migration as there is a chance that
				// we may be fetching statuses that might be
				// part of the same thread, i.e. one call to
				// rethreadStatus() may effect other statuses
				// later in the slice.
				Limit(1000).
				Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gtserror.Newf("error selecting straggler statuses: %w", err)
			}

			// Reached end of block.
			if len(statuses) == 0 {
				break
			}

			// Rethread each selected batch of straggler statuses in a transaction.
			if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

				// Rethread each top-level status.
				for _, status := range statuses {
					n, err := sr.rethreadStatus(ctx, tx, status)
					if err != nil {
						return gtserror.Newf("error rethreading status %s: %w", status.URI, err)
					}
					count += n
				}

				return nil
			}); err != nil {
				return err
			}

			log.Infof(ctx, "[approx %d of %d] rethreading statuses (stragglers)", count, total)
		}

		// Attempt to merge any sqlite write-ahead-log.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		log.Info(ctx, "dropping old thread_to_statuses table")
		if _, err := db.NewDropTable().
			Table("thread_to_statuses").
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping old thread_to_statuses table: %w", err)
		}

		log.Info(ctx, "creating new statuses thread_id column")
		if _, err := db.NewAddColumn().
			Table("statuses").
			ColumnExpr(newColDef).
			Exec(ctx); err != nil {
			return gtserror.Newf("error adding new thread_id column: %w", err)
		}

		log.Info(ctx, "setting thread_id_new = thread_id (this may take a while...)")
		if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			return batchUpdateByID(ctx, tx,
				"statuses",           // table
				"id",                 // batchByCol
				"UPDATE ? SET ? = ?", // updateQuery
				[]any{bun.Ident("statuses"),
					bun.Ident("thread_id_new"),
					bun.Ident("thread_id")},
			)
		}); err != nil {
			return err
		}

		// Attempt to merge any sqlite write-ahead-log.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		log.Info(ctx, "dropping old statuses thread_id index")
		if _, err := db.NewDropIndex().
			Index("statuses_thread_id_idx").
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping old thread_id index: %w", err)
		}

		log.Info(ctx, "dropping old statuses thread_id column")
		if _, err := db.NewDropColumn().
			Table("statuses").
			Column("thread_id").
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping old thread_id column: %w", err)
		}

		log.Info(ctx, "renaming thread_id_new to thread_id")
		if _, err := db.NewRaw(
			"ALTER TABLE ? RENAME COLUMN ? TO ?",
			bun.Ident("statuses"),
			bun.Ident("thread_id_new"),
			bun.Ident("thread_id"),
		).Exec(ctx); err != nil {
			return gtserror.Newf("error renaming new column: %w", err)
		}

		log.Info(ctx, "creating new statuses thread_id index")
		if _, err := db.NewCreateIndex().
			Table("statuses").
			Index("statuses_thread_id_idx").
			Column("thread_id").
			Exec(ctx); err != nil {
			return gtserror.Newf("error creating new thread_id index: %w", err)
		}

		return nil
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}

type statusRethreader struct {
	// the unique status and thread IDs
	// of all models passed to append().
	// these are later used to update all
	// statuses to a single thread ID, and
	// update all thread related models to
	// use the new updated thread ID.
	statusIDs []string
	threadIDs []string

	// stores the unseen IDs of status
	// InReplyTos newly tracked in append(),
	// which is then used for a SELECT query
	// in getParents(), then promptly reset.
	inReplyToIDs []string

	// statuses simply provides a reusable
	// slice of status models for selects.
	// its contents are ephemeral.
	statuses []*oldmodel.Status

	// seenIDs tracks the unique status and
	// thread IDs we have seen, ensuring we
	// don't append duplicates to statusIDs
	// or threadIDs slices. also helps prevent
	// adding duplicate parents to inReplyToIDs.
	seenIDs map[string]struct{}

	// allThreaded tracks whether every status
	// passed to append() has a thread ID set.
	// together with len(threadIDs) this can
	// determine if already threaded correctly.
	allThreaded bool
}

// rethreadStatus is the main logic handler for statusRethreader{}. this is what gets called from the migration
// in order to trigger a status rethreading operation for the given status, returning total number rethreaded.
func (sr *statusRethreader) rethreadStatus(ctx context.Context, tx bun.Tx, status *oldmodel.Status) (int, error) {

	// Zero slice and
	// map ptr values.
	clear(sr.statusIDs)
	clear(sr.threadIDs)
	clear(sr.statuses)
	clear(sr.seenIDs)

	// Reset slices and values for use.
	sr.statusIDs = sr.statusIDs[:0]
	sr.threadIDs = sr.threadIDs[:0]
	sr.statuses = sr.statuses[:0]
	sr.allThreaded = true

	if sr.seenIDs == nil {
		// Allocate new hash set for status IDs.
		sr.seenIDs = make(map[string]struct{})
	}

	// Ensure the passed status
	// has up-to-date information.
	// This may have changed from
	// the initial batch selection
	// to the rethreadStatus() call.
	if err := tx.NewSelect().
		Model(status).
		Column("in_reply_to_id", "thread_id").
		Where("? = ?", bun.Ident("id"), status.ID).
		Scan(ctx); err != nil {
		return 0, gtserror.Newf("error selecting status: %w", err)
	}

	// status and thread ID cursor
	// index values. these are used
	// to keep track of newly loaded
	// status / thread IDs between
	// loop iterations.
	var statusIdx int
	var threadIdx int

	// Append given status as
	// first to our ID slices.
	sr.append(status)

	for {
		// Fetch parents for newly seen in_reply_tos since last loop.
		if err := sr.getParents(ctx, tx); err != nil {
			return 0, gtserror.Newf("error getting parents: %w", err)
		}

		// Fetch children for newly seen statuses since last loop.
		if err := sr.getChildren(ctx, tx, statusIdx); err != nil {
			return 0, gtserror.Newf("error getting children: %w", err)
		}

		// Check for newly picked-up threads
		// to find stragglers for below. Else
		// we've reached end of what we can do.
		if threadIdx >= len(sr.threadIDs) {
			break
		}

		// Update status IDs cursor.
		statusIdx = len(sr.statusIDs)

		// Fetch any stragglers for newly seen threads since last loop.
		if err := sr.getStragglers(ctx, tx, threadIdx); err != nil {
			return 0, gtserror.Newf("error getting stragglers: %w", err)
		}

		// Check for newly picked-up straggling statuses / replies to
		// find parents / children for. Else we've done all we can do.
		if statusIdx >= len(sr.statusIDs) && len(sr.inReplyToIDs) == 0 {
			break
		}

		// Update thread IDs cursor.
		threadIdx = len(sr.threadIDs)
	}

	// Total number of
	// statuses threaded.
	total := len(sr.statusIDs)

	// Check for the case where the entire
	// batch of statuses is already correctly
	// threaded. Then we have nothing to do!
	if sr.allThreaded && len(sr.threadIDs) == 1 {
		return 0, nil
	}

	// Sort all of the threads and
	// status IDs by age; old -> new.
	slices.Sort(sr.threadIDs)
	slices.Sort(sr.statusIDs)

	var threadID string

	if len(sr.threadIDs) > 0 {
		// Regardless of whether there ended up being
		// multiple threads, we take the oldest value
		// thread ID to use for entire batch of them.
		threadID = sr.threadIDs[0]
		sr.threadIDs = sr.threadIDs[1:]
	}

	if threadID == "" {
		// None of the previous parents were threaded, we instead
		// generate new thread with ID based on oldest creation time.
		createdAt, err := id.TimeFromULID(sr.statusIDs[0])
		if err != nil {
			return 0, gtserror.Newf("error parsing status ulid: %w", err)
		}

		// Generate thread ID from parsed time.
		threadID = id.NewULIDFromTime(createdAt)

		// We need to create a
		// new thread table entry.
		if _, err = tx.NewInsert().
			Model(&newmodel.Thread{ID: threadID}).
			Exec(ctx); err != nil {
			return 0, gtserror.Newf("error creating new thread: %w", err)
		}
	}

	// Update all the statuses to
	// use determined thread_id.
	if _, err := tx.NewUpdate().
		Table("statuses").
		Where("? IN (?)", bun.Ident("id"), bun.In(sr.statusIDs)).
		Set("? = ?", bun.Ident("thread_id"), threadID).
		Exec(ctx); err != nil {
		return 0, gtserror.Newf("error updating status thread ids: %w", err)
	}

	if len(sr.threadIDs) > 0 {
		// Update any existing thread
		// mutes to use latest thread_id.
		if _, err := tx.NewUpdate().
			Table("thread_mutes").
			Where("? IN (?)", bun.Ident("thread_id"), bun.In(sr.threadIDs)).
			Set("? = ?", bun.Ident("thread_id"), threadID).
			Exec(ctx); err != nil {
			return 0, gtserror.Newf("error updating mute thread ids: %w", err)
		}
	}

	return total, nil
}

// append will append the given status to the internal tracking of statusRethreader{} for
// potential future operations, checking for uniqueness. it tracks the inReplyToID value
// for the next call to getParents(), it tracks the status ID for list of statuses that
// need updating, the thread ID for the list of thread links and mutes that need updating,
// and whether all the statuses all have a provided thread ID (i.e. allThreaded).
func (sr *statusRethreader) append(status *oldmodel.Status) {

	// Check if status already seen before.
	if _, ok := sr.seenIDs[status.ID]; ok {
		return
	}

	if status.InReplyToID != "" {
		// Status has a parent, add any unique parent ID
		// to list of reply IDs that need to be queried.
		if _, ok := sr.seenIDs[status.InReplyToID]; ok {
			sr.inReplyToIDs = append(sr.inReplyToIDs, status.InReplyToID)
		}
	}

	// Add status' ID to list of seen status IDs.
	sr.statusIDs = append(sr.statusIDs, status.ID)

	if status.ThreadID != "" {
		// Status was threaded, add any unique thread
		// ID to our list of known status thread IDs.
		if _, ok := sr.seenIDs[status.ThreadID]; !ok {
			sr.threadIDs = append(sr.threadIDs, status.ThreadID)
		}
	} else {
		// Status was not threaded,
		// we now know not all statuses
		// found were threaded.
		sr.allThreaded = false
	}

	// Add status ID to map of seen IDs.
	sr.seenIDs[status.ID] = struct{}{}
}

func (sr *statusRethreader) getParents(ctx context.Context, tx bun.Tx) error {
	var parent oldmodel.Status

	// Iteratively query parent for each stored
	// reply ID. Note this is safe to do as slice
	// loop since 'seenIDs' prevents duplicates.
	for i := 0; i < len(sr.inReplyToIDs); i++ {

		// Get next status ID.
		id := sr.statusIDs[i]

		// Select next parent status.
		if err := tx.NewSelect().
			Model(&parent).
			Column("id", "in_reply_to_id", "thread_id").
			Where("? = ?", bun.Ident("id"), id).
			Scan(ctx); err != nil && err != db.ErrNoEntries {
			return err
		}

		// Parent was missing.
		if parent.ID == "" {
			continue
		}

		// Add to slices.
		sr.append(&parent)
	}

	// Reset reply slice.
	clear(sr.inReplyToIDs)
	sr.inReplyToIDs = sr.inReplyToIDs[:0]

	return nil
}

func (sr *statusRethreader) getChildren(ctx context.Context, tx bun.Tx, idx int) error {
	// Iteratively query all children for each
	// of fetched parent statuses. Note this is
	// safe to do as a slice loop since 'seenIDs'
	// ensures it only ever contains unique IDs.
	for i := idx; i < len(sr.statusIDs); i++ {

		// Get next status ID.
		id := sr.statusIDs[i]

		// Reset child slice.
		clear(sr.statuses)
		sr.statuses = sr.statuses[:0]

		// Select children of ID.
		if err := tx.NewSelect().
			Model(&sr.statuses).
			Column("id", "thread_id").
			Where("? = ?", bun.Ident("in_reply_to_id"), id).
			Scan(ctx); err != nil && err != db.ErrNoEntries {
			return err
		}

		// Append child status IDs to slices.
		for _, child := range sr.statuses {
			sr.append(child)
		}
	}

	return nil
}

func (sr *statusRethreader) getStragglers(ctx context.Context, tx bun.Tx, idx int) error {
	// Check for threads to query.
	if idx >= len(sr.threadIDs) {
		return nil
	}

	// Reset status slice.
	clear(sr.statuses)
	sr.statuses = sr.statuses[:0]

	// Select stragglers that
	// also have thread IDs.
	if err := tx.NewSelect().
		Model(&sr.statuses).
		Column("id", "thread_id", "in_reply_to_id").
		Where("? IN (?) AND ? NOT IN (?)",
			bun.Ident("thread_id"),
			bun.In(sr.threadIDs[idx:]),
			bun.Ident("id"),
			bun.In(sr.statusIDs),
		).
		Scan(ctx); err != nil && err != db.ErrNoEntries {
		return err
	}

	// Append status IDs to slices.
	for _, status := range sr.statuses {
		sr.append(status)
	}

	return nil
}
