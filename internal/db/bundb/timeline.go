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

package bundb

import (
	"context"
	"database/sql"
	"sort"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type timelineDB struct {
	config *config.Config
	conn   *DBConn
}

func (t *timelineDB) GetHomeTimeline(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	statuses := make([]*gtsmodel.Status, 0, limit)

	q := t.conn.
		NewSelect().
		Model(&statuses)

	q = q.ColumnExpr("status.*").
		// Find out who accountID follows.
		Join("LEFT JOIN follows AS f ON f.target_account_id = status.account_id").
		// Sort by highest ID (newest) to lowest ID (oldest)
		Order("status.id DESC")

	if maxID != "" {
		// return only statuses LOWER (ie., older) than maxID
		q = q.Where("status.id < ?", maxID)
	}

	if sinceID != "" {
		// return only statuses HIGHER (ie., newer) than sinceID
		q = q.Where("status.id > ?", sinceID)
	}

	if minID != "" {
		// return only statuses HIGHER (ie., newer) than minID
		q = q.Where("status.id > ?", minID)
	}

	if local {
		// return only statuses posted by local account havers
		q = q.Where("status.local = ?", local)
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
			WhereOr("f.account_id = ?", accountID).
			WhereOr("status.account_id = ?", accountID)
	}

	q = q.WhereGroup(" AND ", whereGroup)

	err := q.Scan(ctx)
	if err != nil {
		return nil, t.conn.ProcessError(err)
	}
	return statuses, nil
}

func (t *timelineDB) GetPublicTimeline(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	statuses := make([]*gtsmodel.Status, 0, limit)

	q := t.conn.
		NewSelect().
		Model(&statuses).
		Where("visibility = ?", gtsmodel.VisibilityPublic).
		WhereGroup(" AND ", whereEmptyOrNull("in_reply_to_id")).
		WhereGroup(" AND ", whereEmptyOrNull("in_reply_to_uri")).
		WhereGroup(" AND ", whereEmptyOrNull("boost_of_id")).
		Order("status.id DESC")

	if maxID != "" {
		q = q.Where("status.id < ?", maxID)
	}

	if sinceID != "" {
		q = q.Where("status.id > ?", sinceID)
	}

	if minID != "" {
		q = q.Where("status.id > ?", minID)
	}

	if local {
		q = q.Where("status.local = ?", local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, t.conn.ProcessError(err)
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
		if err == sql.ErrNoRows {
			return nil, "", "", db.ErrNoEntries
		}
		return nil, "", "", err
	}

	if len(faves) == 0 {
		return nil, "", "", db.ErrNoEntries
	}

	// map[statusID]faveID -- we need this to sort statuses by fave ID rather than status ID
	statusesFavesMap := make(map[string]string, len(faves))
	statusIDs := make([]string, 0, len(faves))
	for _, f := range faves {
		statusesFavesMap[f.StatusID] = f.ID
		statusIDs = append(statusIDs, f.StatusID)
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))

	err = t.conn.
		NewSelect().
		Model(&statuses).
		Where("id IN (?)", bun.In(statusIDs)).
		Scan(ctx)
	if err != nil {
		return nil, "", "", t.conn.ProcessError(err)
	}

	if len(statuses) == 0 {
		return nil, "", "", db.ErrNoEntries
	}

	// arrange statuses by fave ID
	sort.Slice(statuses, func(i int, j int) bool {
		statusI := statuses[i]
		statusJ := statuses[j]
		return statusesFavesMap[statusI.ID] < statusesFavesMap[statusJ.ID]
	})

	nextMaxID := faves[len(faves)-1].ID
	prevMinID := faves[0].ID
	return statuses, nextMaxID, prevMinID, nil
}
