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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/paging"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type timelineDB struct {
	db    *bun.DB
	state *state.State
}

func (t *timelineDB) GetHomeTimeline(ctx context.Context, accountID string, page *paging.Page) ([]*gtsmodel.Status, error) {
	return loadStatusTimelinePage(ctx, t.db, t.state,

		// Paging
		// params.
		page,

		// The actual meat of the home-timeline query, outside
		// of any paging parameters that selects by followings.
		func(q *bun.SelectQuery) (*bun.SelectQuery, error) {

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

			// Select only statuses authored by
			// accounts with IDs in the slice.
			q = q.Where(
				"? IN (?)",
				bun.Ident("account_id"),
				bun.In(targetAccountIDs),
			)

			// Only include statuses that aren't pending approval.
			q = q.Where("NOT ? = ?", bun.Ident("pending_approval"), true)

			return q, nil
		},
	)
}

func (t *timelineDB) GetPublicTimeline(ctx context.Context, page *paging.Page) ([]*gtsmodel.Status, error) {
	return loadStatusTimelinePage(ctx, t.db, t.state,

		// Paging
		// params.
		page,

		func(q *bun.SelectQuery) (*bun.SelectQuery, error) {
			// Public only.
			q = q.Where("? = ?", bun.Ident("visibility"), gtsmodel.VisibilityPublic)

			// Ignore boosts.
			q = q.Where("? IS NULL", bun.Ident("boost_of_id"))

			// Only include statuses that aren't pending approval.
			q = q.Where("? = ?", bun.Ident("pending_approval"), false)

			return q, nil
		},
	)
}

func (t *timelineDB) GetLocalTimeline(ctx context.Context, page *paging.Page) ([]*gtsmodel.Status, error) {
	return loadStatusTimelinePage(ctx, t.db, t.state,

		// Paging
		// params.
		page,

		func(q *bun.SelectQuery) (*bun.SelectQuery, error) {
			// Local only.
			q = q.Where("? = ?", bun.Ident("local"), true)

			// Public only.
			q = q.Where("? = ?", bun.Ident("visibility"), gtsmodel.VisibilityPublic)

			// Only include statuses that aren't pending approval.
			q = q.Where("? = ?", bun.Ident("pending_approval"), false)

			// Ignore boosts.
			q = q.Where("? IS NULL", bun.Ident("boost_of_id"))

			return q, nil
		},
	)
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

func (t *timelineDB) GetListTimeline(ctx context.Context, listID string, page *paging.Page) ([]*gtsmodel.Status, error) {
	return loadStatusTimelinePage(ctx, t.db, t.state,

		// Paging
		// params.
		page,

		// The actual meat of the list-timeline query, outside
		// of any paging parameters, it selects by list entries.
		func(q *bun.SelectQuery) (*bun.SelectQuery, error) {

			// Fetch all follow IDs contained in list from DB.
			followIDs, err := t.state.DB.GetFollowIDsInList(
				ctx, listID, nil,
			)
			if err != nil {
				return nil, gtserror.Newf("error getting follows in list: %w", err)
			}

			// Select target account
			// IDs from list follows.
			subQ := t.db.NewSelect().
				Table("follows").
				Column("follows.target_account_id").
				Where("? IN (?)", bun.Ident("follows.id"), bun.In(followIDs))
			q = q.Where("? IN (?)", bun.Ident("statuses.account_id"), subQ)

			// Only include statuses that aren't pending approval.
			q = q.Where("NOT ? = ?", bun.Ident("pending_approval"), true)

			return q, nil
		},
	)
}

func (t *timelineDB) GetTagTimeline(ctx context.Context, tagID string, page *paging.Page) ([]*gtsmodel.Status, error) {
	return loadStatusTimelinePage(ctx, t.db, t.state,

		// Paging
		// params.
		page,

		// The actual meat of the list-timeline query, outside of any
		// paging params, selects by status tags with public visibility.
		func(q *bun.SelectQuery) (*bun.SelectQuery, error) {

			// ...
			q = q.Join(
				"INNER JOIN ? ON ? = ?",
				bun.Ident("status_to_tags"),
				bun.Ident("statuses.id"), bun.Ident("status_to_tags.status_id"),
			)

			// This tag only.
			q = q.Where("? = ?", bun.Ident("status_to_tags.tag_id"), tagID)

			// Public only.
			q = q.Where("? = ?", bun.Ident("visibility"), gtsmodel.VisibilityPublic)

			return q, nil
		},
	)
}

func loadStatusTimelinePage(
	ctx context.Context,
	db *bun.DB,
	state *state.State,
	page *paging.Page,
	query func(*bun.SelectQuery) (*bun.SelectQuery, error),
) (
	[]*gtsmodel.Status,
	error,
) {
	// Extract page params.
	minID := page.Min.Value
	maxID := page.Max.Value
	limit := page.Limit
	order := page.Order()

	// Pre-allocate slice of IDs as dest.
	statusIDs := make([]string, 0, limit)

	// Now start building the database query.
	//
	// Select the following:
	// - status ID
	q := db.NewSelect().
		Table("statuses").
		Column("id")

	// Append caller
	// query details.
	q, err := query(q)
	if err != nil {
		return nil, err
	}

	if maxID != "" {
		// Set a maximum ID boundary if was given.
		q = q.Where("? < ?", bun.Ident("id"), maxID)
	}

	if minID != "" {
		// Set a minimum ID boundary if was given.
		q = q.Where("? > ?", bun.Ident("id"), minID)
	}

	// Set query ordering.
	if order.Ascending() {
		q = q.OrderExpr("? ASC", bun.Ident("id"))
	} else /* i.e. descending */ {
		q = q.OrderExpr("? DESC", bun.Ident("id"))
	}

	// A limit should always
	// be supplied for this.
	q = q.Limit(limit)

	// Finally, perform query into status ID slice.
	if err := q.Scan(ctx, &statusIDs); err != nil {
		return nil, err
	}

	// Fetch statuses from DB / cache with given IDs.
	return state.DB.GetStatusesByIDs(ctx, statusIDs)
}
