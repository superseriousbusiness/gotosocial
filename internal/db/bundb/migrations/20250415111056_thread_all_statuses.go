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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250415111056_thread_all_statuses/new"
	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250415111056_thread_all_statuses/old"
	"code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250415111056_thread_all_statuses/util"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
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
		var updatedTotal int64
		var maxID string
		var statuses []*oldmodel.Status

		// Create thread_id_new already
		// so we can populate it as we go.
		log.Info(ctx, "creating statuses column thread_id_new")
		if _, err := db.NewAddColumn().
			Table("statuses").
			ColumnExpr(newColDef).
			Exec(ctx); err != nil {
			return gtserror.Newf("error adding statuses column thread_id_new: %w", err)
		}

		// Try to merge the wal so we're
		// not working on the wal file.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		// Get a total count of all statuses before migration.
		total, err := db.NewSelect().Table("statuses").Count(ctx)
		if err != nil {
			return gtserror.Newf("error getting status table count: %w", err)
		}

		// Start at largest
		// possible ULID value.
		maxID = id.Highest

		log.Warnf(ctx, "rethreading %d statuses, this will take a *long* time", total)

		// Open initial transaction.
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		for i := 1; ; i++ {

			// Reset slice.
			clear(statuses)
			statuses = statuses[:0]

			batchStart := time.Now()

			// Select top-level statuses.
			if err := tx.NewSelect().
				Model(&statuses).
				Column("id").
				// We specifically use in_reply_to_account_id instead of in_reply_to_id as
				// they should both be set / unset in unison, but we specifically have an
				// index on in_reply_to_account_id with ID ordering, unlike in_reply_to_id.
				Where("? IS NULL", bun.Ident("in_reply_to_account_id")).
				Where("? < ?", bun.Ident("id"), maxID).
				OrderExpr("? DESC", bun.Ident("id")).
				Limit(500).
				Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gtserror.Newf("error selecting top level statuses: %w", err)
			}

			l := len(statuses)
			if l == 0 {
				// No more statuses!
				//
				// Transaction will be closed
				// after leaving the loop.
				break

			} else if i%200 == 0 {
				// Begin a new transaction every
				// 200 batches (~100,000 statuses),
				// to avoid massive commits.

				// Close existing transaction.
				if err := tx.Commit(); err != nil {
					return err
				}

				// Try to flush the wal
				// to avoid silly wal sizes.
				if err := doWALCheckpoint(ctx, db); err != nil {
					return err
				}

				// Open new transaction.
				tx, err = db.BeginTx(ctx, nil)
				if err != nil {
					return err
				}
			}

			// Set next maxID value from statuses.
			maxID = statuses[len(statuses)-1].ID

			// Rethread using the
			// open transaction.
			var updatedInBatch int64
			for _, status := range statuses {
				n, err := sr.rethreadStatus(ctx, tx, status, false)
				if err != nil {
					return gtserror.Newf("error rethreading status %s: %w", status.URI, err)
				}
				updatedInBatch += n
				updatedTotal += n
			}

			// Show speed for this batch.
			timeTaken := time.Since(batchStart).Milliseconds()
			msPerRow := float64(timeTaken) / float64(updatedInBatch)
			rowsPerMs := float64(1) / float64(msPerRow)
			rowsPerSecond := 1000 * rowsPerMs

			// Show percent migrated overall.
			totalDone := (float64(updatedTotal) / float64(total)) * 100

			log.Infof(
				ctx,
				"[~%.2f%% done; ~%.0f rows/s] migrating threads",
				totalDone, rowsPerSecond,
			)
		}

		// Close transaction.
		if err := tx.Commit(); err != nil {
			return err
		}

		// Create a partial index on thread_id_new to find stragglers.
		// This index will be removed at the end of the migration.
		log.Info(ctx, "creating temporary statuses thread_id_new index")
		if _, err := db.NewCreateIndex().
			Table("statuses").
			Index("statuses_thread_id_new_idx").
			Column("thread_id_new").
			Where("? = ?", bun.Ident("thread_id_new"), id.Lowest).
			Exec(ctx); err != nil {
			return gtserror.Newf("error creating new thread_id index: %w", err)
		}

		for i := 1; ; i++ {

			// Reset slice.
			clear(statuses)
			statuses = statuses[:0]

			batchStart := time.Now()

			// Select straggler statuses.
			if err := db.NewSelect().
				Model(&statuses).
				Column("id").
				Where("? = ?", bun.Ident("thread_id_new"), id.Lowest).

				// We select in smaller batches for this part
				// of the migration as there is a chance that
				// we may be fetching statuses that might be
				// part of the same thread, i.e. one call to
				// rethreadStatus() may effect other statuses
				// later in the slice.
				Limit(250).
				Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gtserror.Newf("error selecting straggler statuses: %w", err)
			}

			// Reached end of block.
			if len(statuses) == 0 {
				break
			}

			// Rethread each selected batch of straggler statuses in a transaction.
			var updatedInBatch int64
			if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
				for _, status := range statuses {
					n, err := sr.rethreadStatus(ctx, tx, status, true)
					if err != nil {
						return gtserror.Newf("error rethreading status %s: %w", status.URI, err)
					}
					updatedInBatch += n
					updatedTotal += n
				}
				return nil
			}); err != nil {
				return err
			}

			// Show speed for this batch.
			timeTaken := time.Since(batchStart).Milliseconds()
			msPerRow := float64(timeTaken) / float64(updatedInBatch)
			rowsPerMs := float64(1) / float64(msPerRow)
			rowsPerSecond := 1000 * rowsPerMs

			// Show percent migrated overall.
			totalDone := (float64(updatedTotal) / float64(total)) * 100

			log.Infof(
				ctx,
				"[~%.2f%% done; ~%.0f rows/s] migrating stragglers",
				totalDone, rowsPerSecond,
			)
		}

		// Attempt to merge any sqlite write-ahead-log.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		log.Info(ctx, "dropping temporary thread_id_new index")
		if _, err := db.NewDropIndex().
			Index("statuses_thread_id_new_idx").
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping temporary thread_id_new index: %w", err)
		}

		log.Info(ctx, "dropping old thread_to_statuses table")
		if _, err := db.NewDropTable().
			Table("thread_to_statuses").
			Exec(ctx); err != nil {
			return gtserror.Newf("error dropping old thread_to_statuses table: %w", err)
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

	// newThreadIDSet is used to track whether
	// statuses in statusIDs have already have
	// thread_id_new set on them.
	newThreadIDSet map[string]struct{}

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
// in order to trigger a status rethreading operation for the given status, returning total number of rows changed.
func (sr *statusRethreader) rethreadStatus(ctx context.Context, tx bun.Tx, status *oldmodel.Status, straggler bool) (int64, error) {

	// Zero slice and
	// map ptr values.
	clear(sr.statusIDs)
	clear(sr.threadIDs)
	clear(sr.statuses)
	clear(sr.newThreadIDSet)
	clear(sr.seenIDs)

	// Reset slices and values for use.
	sr.statusIDs = sr.statusIDs[:0]
	sr.threadIDs = sr.threadIDs[:0]
	sr.statuses = sr.statuses[:0]
	sr.allThreaded = true

	if sr.newThreadIDSet == nil {
		// Allocate new hash set for newThreadIDSet.
		sr.newThreadIDSet = make(map[string]struct{})
	}

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
		Column("in_reply_to_id", "thread_id", "thread_id_new").
		Where("? = ?", bun.Ident("id"), status.ID).
		Scan(ctx); err != nil {
		return 0, gtserror.Newf("error selecting status: %w", err)
	}

	// If we've just threaded this status by setting
	// thread_id_new, then by definition anything we
	// could find from the entire thread must now be
	// threaded, so we can save some database calls
	// by skipping iterating up + down from here.
	if status.ThreadIDNew != id.Lowest {
		log.Debugf(ctx, "skipping just rethreaded status: %s", status.ID)
		return 0, nil
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

	// Check for the case where the entire
	// batch of statuses is already correctly
	// threaded. Then we have nothing to do!
	//
	// Skip this check for straggler statuses
	// that are part of broken threads.
	if !straggler && sr.allThreaded && len(sr.threadIDs) == 1 {
		log.Debug(ctx, "skipping just rethreaded thread")
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

	var (
		res sql.Result
		err error
	)

	if len(sr.statusIDs) == 1 {

		// If we're only updating one status
		// we can use a simple update query.
		res, err = tx.NewUpdate().
			// Update the status model.
			TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
			// Set the new thread ID, which we can use as
			// an indication that we've migrated this batch.
			Set("? = ?", bun.Ident("thread_id_new"), threadID).
			// While we're here, also set old thread_id, as
			// we'll use it for further rethreading purposes.
			Set("? = ?", bun.Ident("thread_id"), threadID).
			Where("? = ?", bun.Ident("status.id"), sr.statusIDs[0]).
			Exec(ctx)

	} else {

		// If we're updating multiple statuses at once,
		// build up a common table expression to update
		// all statuses in this thread to use threadID.
		//
		// This ought to be a little more readable than
		// using an "IN(*)" query, and PG or SQLite *may*
		// be able to optimize it better.
		//
		// See:
		//
		// - https://sqlite.org/lang_with.html
		// - https://www.postgresql.org/docs/current/queries-with.html
		// - https://bun.uptrace.dev/guide/query-update.html#bulk-update
		values := make([]*util.Status, 0, len(sr.statusIDs))
		for _, statusID := range sr.statusIDs {
			// Filter out statusIDs that have already had
			// thread_id_new set, to avoid spurious writes.
			if _, set := sr.newThreadIDSet[statusID]; !set {
				values = append(values, &util.Status{
					ID: statusID,
				})
			}
		}

		// Resulting query will look something like this:
		//
		//  WITH "_data" ("id") AS (
		//    VALUES
		//      ('01JR6PZED0DDR2VZHQ8H87ZW98'),
		//      ('01JR6PZED0J91MJCAFDTCCCG8Q')
		//  )
		//  UPDATE "statuses" AS "status"
		//  SET
		//    "thread_id_new" = '01K6MGKX54BBJ3Y1FBPQY45E5P',
		//    "thread_id"     = '01K6MGKX54BBJ3Y1FBPQY45E5P'
		//  FROM _data
		//  WHERE "status"."id" = "_data"."id"
		res, err = tx.NewUpdate().
			// Update the status model.
			Model((*oldmodel.Status)(nil)).
			// Provide the CTE values as "_data".
			With("_data", tx.NewValues(&values)).
			// Include `FROM _data` statement so we can use
			// `_data` table in SET and WHERE components.
			TableExpr("_data").
			// Set the new thread ID, which we can use as
			// an indication that we've migrated this batch.
			Set("? = ?", bun.Ident("thread_id_new"), threadID).
			// While we're here, also set old thread_id, as
			// we'll use it for further rethreading purposes.
			Set("? = ?", bun.Ident("thread_id"), threadID).
			// "Join" to the CTE on status ID.
			Where("? = ?", bun.Ident("status.id"), bun.Ident("_data.id")).
			Exec(ctx)
	}

	if err != nil {
		return 0, gtserror.Newf("error updating status thread ids: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return 0, gtserror.Newf("error counting rows affected: %w", err)
	}

	if len(sr.threadIDs) > 0 {
		// Update any existing thread
		// mutes to use latest thread_id.

		// Dedupe thread IDs before query
		// to avoid ludicrous "IN" clause.
		threadIDs := sr.threadIDs
		threadIDs = xslices.Deduplicate(threadIDs)
		if _, err := tx.NewUpdate().
			Table("thread_mutes").
			Where("? IN (?)", bun.Ident("thread_id"), bun.In(threadIDs)).
			Set("? = ?", bun.Ident("thread_id"), threadID).
			Exec(ctx); err != nil {
			return 0, gtserror.Newf("error updating mute thread ids: %w", err)
		}
	}

	return rowsAffected, nil
}

// append will append the given status to the internal tracking of statusRethreader{} for
// potential future operations, checking for uniqueness. it tracks the inReplyToID value
// for the next call to getParents(), it tracks the status ID for list of statuses that
// may need updating, whether a new thread ID has been set for each status, the thread ID
// for the list of thread links and mutes that need updating, and whether all the statuses
// all have a provided thread ID (i.e. allThreaded).
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
	mark := struct{}{}
	sr.seenIDs[status.ID] = mark

	// If new thread ID has already been
	// set, add status ID to map of set IDs.
	if status.ThreadIDNew != id.Lowest {
		sr.newThreadIDSet[status.ID] = mark
	}
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
			Column("id", "in_reply_to_id", "thread_id", "thread_id_new").
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
			Column("id", "thread_id", "thread_id_new").
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

	// Dedupe thread IDs before query
	// to avoid ludicrous "IN" clause.
	threadIDs := sr.threadIDs[idx:]
	threadIDs = xslices.Deduplicate(threadIDs)

	// Select stragglers that
	// also have thread IDs.
	if err := tx.NewSelect().
		Model(&sr.statuses).
		Column("id", "thread_id", "in_reply_to_id", "thread_id_new").
		Where("? IN (?) AND ? NOT IN (?)",
			bun.Ident("thread_id"),
			bun.In(threadIDs),
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
