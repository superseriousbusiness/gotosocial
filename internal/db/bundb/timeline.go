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
	"fmt"
	"slices"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type timelineDB struct {
	db    *bun.DB
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

	// As this is the home timeline, it should be
	// populated by statuses from accounts followed
	// by accountID, and posts from accountID itself.
	//
	// So, begin by seeing who accountID follows.
	// It should be a little cheaper to do this in
	// a separate query like this, rather than using
	// a join, since followIDs are cached in memory.
	follows, err := t.state.DB.GetAccountFollows(
		gtscontext.SetBarebones(ctx),
		accountID,
		nil, // select all
	)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("db error getting follows for account %s: %w", accountID, err)
	}

	// To take account of exclusive lists, get all of
	// this account's lists, so we can filter out follows
	// that are in contained in exclusive lists.
	lists, err := t.state.DB.GetListsByAccountID(ctx, accountID)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return nil, gtserror.Newf("db error getting lists for account %s: %w", accountID, err)
	}

	// Index all follow IDs that fall in exclusive lists.
	ignoreFollowIDs := make(map[string]struct{})
	for _, list := range lists {
		if !*list.Exclusive {
			// Not exclusive,
			// we don't care.
			continue
		}

		// Fetch all follow IDs of the entries ccontained in this list.
		listFollowIDs, err := t.state.DB.GetFollowIDsInList(ctx, list.ID, nil)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return nil, gtserror.Newf("db error getting list entry follow ids: %w", err)
		}

		// Exclusive list, index all its follow IDs.
		for _, followID := range listFollowIDs {
			ignoreFollowIDs[followID] = struct{}{}
		}
	}

	// Extract just the accountID from each follow,
	// ignoring follows that are in exclusive lists.
	targetAccountIDs := make([]string, 0, len(follows)+1)
	for _, f := range follows {
		_, ignore := ignoreFollowIDs[f.ID]
		if !ignore {
			targetAccountIDs = append(
				targetAccountIDs,
				f.TargetAccountID,
			)
		}
	}

	// Add accountID itself as a pseudo follow so that
	// accountID can see its own posts in the timeline.
	targetAccountIDs = append(targetAccountIDs, accountID)

	// Now start building the database query.
	q := t.db.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status")).
		// Select only IDs from table
		Column("status.id")

	if maxID == "" || maxID >= id.Highest {
		const future = 24 * time.Hour

		// don't return statuses more than 24hr in the future
		maxID = id.NewULIDFromTime(time.Now().Add(future))
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

	// Select only statuses authored by
	// accounts with IDs in the slice.
	q = q.Where(
		"? IN (?)",
		bun.Ident("status.account_id"),
		bun.In(targetAccountIDs),
	)

	// Only include statuses that aren't pending approval.
	q = q.Where("NOT ? = ?", bun.Ident("status.pending_approval"), true)

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

	// Return status IDs loaded from cache + db.
	return t.state.DB.GetStatusesByIDs(ctx, statusIDs)
}

func (t *timelineDB) GetPublicTimeline(ctx context.Context, maxID string, sinceID string, minID string, limit int, local bool) ([]*gtsmodel.Status, error) {
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
		// Public only.
		Where("? = ?", bun.Ident("status.visibility"), gtsmodel.VisibilityPublic).
		// Ignore boosts.
		Where("? IS NULL", bun.Ident("status.boost_of_id")).
		// Select only IDs from table
		Column("status.id")

	if maxID == "" || maxID >= id.Highest {
		const future = 24 * time.Hour

		// don't return statuses more than 24hr in the future
		maxID = id.NewULIDFromTime(time.Now().Add(future))
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

	// Only include statuses that aren't pending approval.
	q = q.Where("NOT ? = ?", bun.Ident("status.pending_approval"), true)

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

	// Return status IDs loaded from cache + db.
	return t.state.DB.GetStatusesByIDs(ctx, statusIDs)
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
	slices.SortFunc(faves, func(a, b *gtsmodel.StatusFave) int {
		const k = -1
		switch {
		case a.ID > b.ID:
			return +k
		case a.ID < b.ID:
			return -k
		default:
			return 0
		}
	})

	// Convert fave IDs to status IDs.
	statusIDs := make([]string, len(faves))
	for i, fave := range faves {
		statusIDs[i] = fave.StatusID
	}

	statuses, err := t.state.DB.GetStatusesByIDs(ctx, statusIDs)
	if err != nil {
		return nil, "", "", err
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

	// Fetch all follow IDs contained in list from DB.
	followIDs, err := t.state.DB.GetFollowIDsInList(
		ctx, listID, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting follows in list: %w", err)
	}

	// If there's no list follows we can't
	// possibly return anything for this list.
	if len(followIDs) == 0 {
		return make([]*gtsmodel.Status, 0), nil
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

		// don't return statuses more than 24hr in the future
		maxID = id.NewULIDFromTime(time.Now().Add(future))
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

	// Only include statuses that aren't pending approval.
	q = q.Where("NOT ? = ?", bun.Ident("status.pending_approval"), true)

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

	// Return status IDs loaded from cache + db.
	return t.state.DB.GetStatusesByIDs(ctx, statusIDs)
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

		// don't return statuses more than 24hr in the future
		maxID = id.NewULIDFromTime(time.Now().Add(future))
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

	// Only include statuses that aren't pending approval.
	q = q.Where("NOT ? = ?", bun.Ident("status.pending_approval"), true)

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

	// Return status IDs loaded from cache + db.
	return t.state.DB.GetStatusesByIDs(ctx, statusIDs)
}
