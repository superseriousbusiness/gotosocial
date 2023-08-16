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
	"fmt"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
	"golang.org/x/exp/slices"
)

type timelineDB struct {
	db    *DB
	state *state.State
}

func (t *timelineDB) GetHomeTimeline(ctx context.Context, accountID string, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		statusIDs   = make([]string, 0, limit)
		frontToBack = true
	)

	q := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id")

	if maxID == "" || maxID >= id.Highest {
		const future = 24 * time.Hour

		var err error

		// don't return statuses more than 24hr in the future
		maxID, err = id.NewULIDFromTime(time.Now().Add(future))
		if err != nil {
			return nil, err
		}
	}

	// return only statuses LOWER (ie., older) than maxID
	q = q.Where("? < ?", bun.Ident("status.id"), maxID)

	if sinceID != "" {
		// return only statuses HIGHER (ie., newer) than sinceID
		q = q.Where("? > ?", bun.Ident("status.id"), sinceID)
	}

	if minID != "" {
		// return only statuses HIGHER (ie., newer) than minID
		q = q.Where("? > ?", bun.Ident("status.id"), minID)

		// page up
		frontToBack = false
	}

	if local {
		// return only statuses posted by local account havers
		q = q.Where("? = ?", bun.Ident("status.local"), local)
	}

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("status.id DESC")
	} else {
		// Page up.
		q = q.Order("status.id ASC")
	}

	// Subquery to select target (followed) account
	// IDs from follows owned by given accountID.
	subQ := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
		Column("follow.target_account_id").
		Where("? = ?", bun.Ident("follow.account_id"), accountID)

	// Use the subquery in a WhereGroup here to specify that we want EITHER
	// - statuses posted by accountID itself OR
	// - statuses posted by accounts that accountID follows
	q = q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Where("? = ?", bun.Ident("status.account_id"), accountID).
			WhereOr("? IN (?)", bun.Ident("status.account_id"), subQ)
	})

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	if len(statusIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want statuses
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(statusIDs)-1; l < r; l, r = l+1, r-1 {
			statusIDs[l], statusIDs[r] = statusIDs[r], statusIDs[l]
		}
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))
	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := t.state.DB.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (t *timelineDB) GetPublicTimeline(ctx context.Context, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	statusIDs := make([]string, 0, limit)

	q := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		Column("status.id").
		// Public only.
		Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
		// Ignore boosts.
		Where("? IS NULL", bun.Ident("status.boost_of_id")).
		Order("status.id DESC")

	if maxID == "" {
		const future = 24 * time.Hour

		var err error

		// don't return statuses more than 24hr in the future
		maxID, err = id.NewULIDFromTime(time.Now().Add(future))
		if err != nil {
			return nil, err
		}
	}

	// return only statuses LOWER (ie., older) than maxID
	q = q.Where("? < ?", bun.Ident("status.id"), maxID)

	if sinceID != "" {
		q = q.Where("? > ?", bun.Ident("status.id"), sinceID)
	}

	if minID != "" {
		q = q.Where("? > ?", bun.Ident("status.id"), minID)
	}

	if local {
		q = q.Where("? = ?", bun.Ident("status.local"), local)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))

	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := t.state.DB.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// TODO optimize this query and the logic here, because it's slow as balls -- it takes like a literal second to return with a limit of 20!
// It might be worth serving it through a timeline instead of raw DB queries, like we do for Home feeds.
func (t *timelineDB) GetFavedTimeline(ctx context.Context, accountID string, maxID string, minID string, limit int) ([]*gtsmodel.Status, string, string, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	faves := make([]*gtsmodel.StatusFave, 0, limit)

	fq := t.db.
		NewSelect().
		Model(&faves).
		Where("? = ?", bun.Ident("status_fave.account_id"), accountID).
		Order("status_fave.id DESC")

	if maxID != "" {
		fq = fq.Where("? < ?", bun.Ident("status_fave.id"), maxID)
	}

	if minID != "" {
		fq = fq.Where("? > ?", bun.Ident("status_fave.id"), minID)
	}

	if limit > 0 {
		fq = fq.Limit(limit)
	}

	err := fq.Scan(ctx)
	if err != nil {
		return nil, "", "", err
	}

	if len(faves) == 0 {
		return nil, "", "", db.ErrNoEntries
	}

	// Sort by favourite ID rather than status ID
	slices.SortFunc(faves, func(a, b *gtsmodel.StatusFave) bool {
		return a.ID > b.ID
	})

	statuses := make([]*gtsmodel.Status, 0, len(faves))

	for _, fave := range faves {
		// Fetch status from db for corresponding favourite
		status, err := t.state.DB.GetStatusByID(ctx, fave.StatusID)
		if err != nil {
			log.Errorf(ctx, "error fetching status for fave %q: %v", fave.ID, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	nextMaxID := faves[len(faves)-1].ID
	prevMinID := faves[0].ID
	return statuses, nextMaxID, prevMinID, nil
}

func (t *timelineDB) GetListTimeline(
	ctx context.Context,
	listID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		statusIDs   = make([]string, 0, limit)
		frontToBack = true
	)

	// Fetch all listEntries entries from the database.
	listEntries, err := t.state.DB.GetListEntries(
		// Don't need actual follows
		// for this, just the IDs.
		gtscontext.SetBarebones(ctx),
		listID,
		"", "", "", 0,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting entries for list %s: %w", listID, err)
	}

	// Extract just the IDs of each follow.
	followIDs := make([]string, 0, len(listEntries))
	for _, listEntry := range listEntries {
		followIDs = append(followIDs, listEntry.FollowID)
	}

	// Select target account IDs from follows.
	subQ := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("follows"), bun.Ident("follow")).
		Column("follow.target_account_id").
		Where("? IN (?)", bun.Ident("follow.id"), bun.In(followIDs))

	// Select only status IDs created
	// by one of the followed accounts.
	q := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id").
		Where("? IN (?)", bun.Ident("status.account_id"), subQ)

	if maxID == "" || maxID >= id.Highest {
		const future = 24 * time.Hour

		var err error

		// don't return statuses more than 24hr in the future
		maxID, err = id.NewULIDFromTime(time.Now().Add(future))
		if err != nil {
			return nil, err
		}
	}

	// return only statuses LOWER (ie., older) than maxID
	q = q.Where("? < ?", bun.Ident("status.id"), maxID)

	if sinceID != "" {
		// return only statuses HIGHER (ie., newer) than sinceID
		q = q.Where("? > ?", bun.Ident("status.id"), sinceID)
	}

	if minID != "" {
		// return only statuses HIGHER (ie., newer) than minID
		q = q.Where("? > ?", bun.Ident("status.id"), minID)

		// page up
		frontToBack = false
	}

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("status.id DESC")
	} else {
		// Page up.
		q = q.Order("status.id ASC")
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	if len(statusIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want statuses
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(statusIDs)-1; l < r; l, r = l+1, r-1 {
			statusIDs[l], statusIDs[r] = statusIDs[r], statusIDs[l]
		}
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))
	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := t.state.DB.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}

func (t *timelineDB) GetTagTimeline(
	ctx context.Context,
	tagID string,
	maxID string,
	sinceID string,
	minID string,
	limit int,
) ([]*gtsmodel.Status, error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Make educated guess for slice size
	var (
		statusIDs   = make([]string, 0, limit)
		frontToBack = true
	)

	q := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_to_tags"), bun.Ident("status_to_tag")).
		Column("status_to_tag.status_id").
		// Join with statuses for filtering.
		Join(
			"INNER JOIN ? AS ? ON ? = ?",
			bun.Ident("statuses"), bun.Ident("status"),
			bun.Ident("status.id"), bun.Ident("status_to_tag.status_id"),
		).
		// Public only.
		Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
		// This tag only.
		Where("? = ?", bun.Ident("status_to_tag.tag_id"), tagID)

	if maxID == "" || maxID >= id.Highest {
		const future = 24 * time.Hour

		var err error

		// don't return statuses more than 24hr in the future
		maxID, err = id.NewULIDFromTime(time.Now().Add(future))
		if err != nil {
			return nil, err
		}
	}

	// return only statuses LOWER (ie., older) than maxID
	q = q.Where("? < ?", bun.Ident("status_to_tag.status_id"), maxID)

	if sinceID != "" {
		// return only statuses HIGHER (ie., newer) than sinceID
		q = q.Where("? > ?", bun.Ident("status_to_tag.status_id"), sinceID)
	}

	if minID != "" {
		// return only statuses HIGHER (ie., newer) than minID
		q = q.Where("? > ?", bun.Ident("status_to_tag.status_id"), minID)

		// page up
		frontToBack = false
	}

	if limit > 0 {
		// limit amount of statuses returned
		q = q.Limit(limit)
	}

	if frontToBack {
		// Page down.
		q = q.Order("status_to_tag.status_id DESC")
	} else {
		// Page up.
		q = q.Order("status_to_tag.status_id ASC")
	}

	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	if len(statusIDs) == 0 {
		return nil, nil
	}

	// If we're paging up, we still want statuses
	// to be sorted by ID desc, so reverse ids slice.
	// https://zchee.github.io/golang-wiki/SliceTricks/#reversing
	if !frontToBack {
		for l, r := 0, len(statusIDs)-1; l < r; l, r = l+1, r-1 {
			statusIDs[l], statusIDs[r] = statusIDs[r], statusIDs[l]
		}
	}

	statuses := make([]*gtsmodel.Status, 0, len(statusIDs))
	for _, id := range statusIDs {
		// Fetch status from db for ID
		status, err := t.state.DB.GetStatusByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error fetching status %q: %v", id, err)
			continue
		}

		// Append status to slice
		statuses = append(statuses, status)
	}

	return statuses, nil
}
