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
	"context"

	log "github.com/dsoprea/go-logging"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
	"golang.org/x/exp/slices"
)

type timelineDB struct {
	conn   *DBConn
	status *statusDB
}

func (t *timelineDB) GetHomeTimeline(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, db.Error) {
	// Make educated guess for slice size
	statusIDs := make([]string, 0, limit)

	q := t.conn.
		NewSelect().
		Table("statuses").

		// Select only IDs from table
		Column("statuses.id").
		// Find out who accountID follows.
		Join("LEFT JOIN follows ON follows.target_account_id = statuses.account_id AND follows.account_id = ?", accountID).
		// Sort by highest ID (newest) to lowest ID (oldest)
		Order("statuses.id DESC")

	if maxID != "" {
		// return only statuses LOWER (ie., older) than maxID
		q = q.Where("statuses.id < ?", maxID)
	}

	if sinceID != "" {
		// return only statuses HIGHER (ie., newer) than sinceID
		q = q.Where("statuses.id > ?", sinceID)
	}

	if minID != "" {
		// return only statuses HIGHER (ie., newer) than minID
		q = q.Where("statuses.id > ?", minID)
	}

	if local {
		// return only statuses posted by local account havers
		q = q.Where("statuses.local = ?", local)
	}

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	// Use a WhereGroup here to specify that we want EITHER statuses posted by accounts that accountID follows,
	// OR statuses posted by accountID itself (since a user should be able to see their own statuses).
	//
	// This is equivalent to something like WHERE ... AND (... OR ...)
	// See: https://bun.uptrace.dev/guide/queries.html#select
	whereGroup := func(*bun.SelectQuery) *bun.SelectQuery {
		return q.
			WhereOr("follows.account_id = ?", accountID).
			WhereOr("statuses.account_id = ?", accountID)
	}

	q = q.WhereGroup(" AND ", whereGroup)

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, t.conn.ProcessError(err)
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))

	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := t.status.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf("GetHomeTimeline: error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (t *timelineDB) GetPublicTimeline(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	statusIDs := make([]string, 0, limit)

	q := t.conn.
		NewSelect().
		Table("statuses").
		Column("statuses.id").
		Where("statuses.visibility = ?", gtsmodel.VisibilityPublic).
		WhereGroup(" AND ", whereEmptyOrNull("statuses.in_reply_to_id")).
		WhereGroup(" AND ", whereEmptyOrNull("statuses.in_reply_to_uri")).
		WhereGroup(" AND ", whereEmptyOrNull("statuses.boost_of_id")).
		Order("statuses.id DESC")

	if maxID != "" {
		q = q.Where("statuses.id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("statuses.id > ?", sinceID)
	}

	if minID != "" {
		q = q.Where("statuses.id > ?", minID)
	}

	if local {
		q = q.Where("statuses.local = ?", local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, t.conn.ProcessError(err)
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))

	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := t.status.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf("GetPublicTimeline: error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// TODO optimize this query and the logic here, because it's slow as balls -- it takes like a literal second to return with a limit of 20!
// It might be worth serving it through a timeline instead of raw DB queries, like we do for Home feeds.
func (t *timelineDB) GetFavedTimeline(ctx context.Context, accountID string, maxID string, minID string, limit int) ([]*gtsmodel.Status, string, string, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	faves := make([]*gtsmodel.StatusFave, 0, limit)

	fq := t.conn.
		NewSelect().
		Model(&faves).
		Where("account_id = ?", accountID).
		Order("id DESC")

	if maxID != "" {
		fq = fq.Where("id < ?", maxID)
	}

	if minID != "" {
		fq = fq.Where("id > ?", minID)
	}

	if limit > 0 {
		fq = fq.Limit(limit)
	}

	err := fq.Scan(ctx)
	if err != nil {
		return nil, "", "", t.conn.ProcessError(err)
	}

	if len(faves) == 0 {
		return nil, "", "", db.ErrNoEntries
	}

	// Sort by favourite ID rather than status ID
	slices.SortFunc(faves, func(a, b *gtsmodel.StatusFave) bool {
		return a.ID < b.ID
	})

	statuses := make([]*gtsmodel.Status, 0, len(faves))

	for _, fave := range faves {
		// Fetch status from db for corresponding favourite
		status, err := t.status.GetStatusByID(ctx, fave.StatusID)
		if err != nil {
			log.Errorf("GetFavedTimeline: error fetching status for fave %q: %v", fave.ID, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	nextMaxID := faves[len(faves)-1].ID
	prevMinID := faves[0].ID
	return statuses, nextMaxID, prevMinID, nil
}
