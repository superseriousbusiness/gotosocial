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
	"slices"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type interactionDB struct {
	db    *bun.DB
	state *state.State
}

func (r *interactionDB) newInteractionApprovalQ(approval interface{}) *bun.SelectQuery {
	return r.db.
		NewSelect().
		Model(approval)
}

func (r *interactionDB) GetInteractionApprovalByID(ctx context.Context, id string) (*gtsmodel.InteractionApproval, error) {
	return r.getInteractionApproval(
		ctx,
		"ID",
		func(approval *gtsmodel.InteractionApproval) error {
			return r.
				newInteractionApprovalQ(approval).
				Where("? = ?", bun.Ident("interaction_approval.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (r *interactionDB) GetInteractionApprovalByURI(ctx context.Context, uri string) (*gtsmodel.InteractionApproval, error) {
	return r.getInteractionApproval(
		ctx,
		"URI",
		func(approval *gtsmodel.InteractionApproval) error {
			return r.
				newInteractionApprovalQ(approval).
				Where("? = ?", bun.Ident("interaction_approval.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (r *interactionDB) getInteractionApproval(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.InteractionApproval) error,
	keyParts ...any,
) (*gtsmodel.InteractionApproval, error) {
	// Fetch approval from database cache with loader callback
	approval, err := r.state.Caches.DB.InteractionApproval.LoadOne(lookup, func() (*gtsmodel.InteractionApproval, error) {
		var approval gtsmodel.InteractionApproval

		// Not cached! Perform database query
		if err := dbQuery(&approval); err != nil {
			return nil, err
		}

		return &approval, nil
	}, keyParts...)
	if err != nil {
		// Error already processed.
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return approval, nil
	}

	if err := r.PopulateInteractionApproval(ctx, approval); err != nil {
		return nil, err
	}

	return approval, nil
}

func (r *interactionDB) PopulateInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error {
	var (
		err  error
		errs = gtserror.NewMultiError(2)
	)

	if approval.Account == nil {
		// Account is not set, fetch from the database.
		approval.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			approval.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionApproval account: %w", err)
		}
	}

	if approval.InteractingAccount == nil {
		// InteractingAccount is not set, fetch from the database.
		approval.InteractingAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			approval.InteractingAccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionApproval interacting account: %w", err)
		}
	}

	return errs.Combine()
}

func (r *interactionDB) PutInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error {
	return r.state.Caches.DB.InteractionApproval.Store(approval, func() error {
		_, err := r.db.NewInsert().Model(approval).Exec(ctx)
		return err
	})
}

func (r *interactionDB) DeleteInteractionApprovalByID(ctx context.Context, id string) error {
	defer r.state.Caches.DB.InteractionApproval.Invalidate("ID", id)

	_, err := r.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("interaction_approvals"), bun.Ident("interaction_approval")).
		Where("? = ?", bun.Ident("interaction_approval.id"), id).
		Exec(ctx)
	return err
}

// Produces something like:
//
//	SELECT * FROM (
//		SELECT "id", 'favourite' AS "interaction_type"
//		FROM "status_faves"
//		WHERE ("pending_approval" = TRUE)
//		AND ("target_account_id" = '01F8MH1H7YV1Z7D2C8K2730QBF')
//		AND ("id" < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ')
//	)
//	UNION
//		SELECT "id", 'reply' AS "interaction_type"
//		FROM "statuses"
//		WHERE ("pending_approval" = TRUE)
//		AND ("in_reply_to_account_id" = '01F8MH1H7YV1Z7D2C8K2730QBF')
//		AND ("id" < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ')
//	UNION
//		SELECT "id", 'reblog' AS "interaction_type"
//		FROM "statuses"
//		WHERE ("pending_approval" = TRUE)
//		AND ("boost_of_account_id" = '01F8MH1H7YV1Z7D2C8K2730QBF')
//		AND ("id" < 'ZZZZZZZZZZZZZZZZZZZZZZZZZZ')
//	ORDER BY "id" DESC
//	LIMIT 20
func (r *interactionDB) GetPendingInteractionsForAcct(
	ctx context.Context,
	acctID string,
	statusID string,
	likes bool,
	replies bool,
	boosts bool,
	page *paging.Page,
) ([]*gtsmodel.PendingInteraction, error) {
	if !likes && !replies && !boosts {
		return nil, gtserror.New("at least one of likes, replies, or boosts must be true")
	}

	var (
		// Get paging params.
		minID = page.GetMin()
		maxID = page.GetMax()
		limit = page.GetLimit()
		order = page.GetOrder()

		// Make educated guess for slice size
		pendingInts = make([]*gtsmodel.PendingInteraction, 0, limit)

		// Slice of subqueries for UNION.
		subQs = make([]*bun.SelectQuery, 0, 3)
	)

	if likes {
		//	SELECT id, 'favourite' AS 'interaction_type'
		//	FROM status_faves
		//	WHERE pending_approval = true
		//	AND target_account_id = '016T5Q3SQKBT337DAKVSKNXXW1'
		likesSubQ := r.db.
			NewSelect().
			Table("status_faves").
			Column("id").
			ColumnExpr("? AS ?", gtsmodel.InteractionLike, bun.Ident("interaction_type")).
			Where("? = ?", bun.Ident("pending_approval"), true).
			Where("? = ?", bun.Ident("target_account_id"), acctID)

		if statusID != "" {
			likesSubQ = likesSubQ.Where("? = ?", bun.Ident("status_id"), statusID)
		}

		subQs = append(subQs, likesSubQ)
	}

	if replies {
		//	SELECT id, 'reply' AS 'interaction_type'
		//	FROM statuses
		//	WHERE pending_approval = true
		//	AND in_reply_to_account_id = '016T5Q3SQKBT337DAKVSKNXXW1'
		repliesSubQ := r.db.
			NewSelect().
			Table("statuses").
			Column("id").
			ColumnExpr("? AS ?", gtsmodel.InteractionReply, bun.Ident("interaction_type")).
			Where("? = ?", bun.Ident("pending_approval"), true).
			Where("? = ?", bun.Ident("in_reply_to_account_id"), acctID)

		if statusID != "" {
			repliesSubQ = repliesSubQ.Where("? = ?", bun.Ident("in_reply_to_id"), statusID)
		}

		subQs = append(subQs, repliesSubQ)
	}

	if boosts {
		//	SELECT id, 'reblog' AS 'interaction_type'
		//	FROM statuses
		//	WHERE pending_approval = true
		//	AND boost_of_account_id = '016T5Q3SQKBT337DAKVSKNXXW1'
		boostsSubQ := r.db.NewSelect().
			Table("statuses").
			Column("id").
			ColumnExpr("? AS ?", gtsmodel.InteractionAnnounce, bun.Ident("interaction_type")).
			Where("? = ?", bun.Ident("pending_approval"), true).
			Where("? = ?", bun.Ident("boost_of_account_id"), acctID)

		if statusID != "" {
			boostsSubQ = boostsSubQ.Where("? = ?", bun.Ident("boost_of_id"), statusID)
		}

		subQs = append(subQs, boostsSubQ)
	}

	var unions string
	for i, subQ := range subQs {
		// Apply max + min ID paging params to
		// subqueries individually; you can't
		// apply them overall in UNION statements.
		if maxID != "" {
			subQ = subQ.Where("? < ?", bun.Ident("id"), maxID)
		}

		if minID != "" {
			subQ = subQ.Where("? > ?", bun.Ident("id"), minID)
		}

		if i == 0 {
			unions += "(" + subQ.String() + ")"
		} else {
			unions += " UNION " + subQ.String()
		}
	}

	q := r.db.
		NewSelect().
		TableExpr("?", bun.Safe(unions))

	if limit > 0 {
		q = q.Limit(limit)
	}

	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr("? ASC", bun.Ident("id"))
	} else {
		// Page down.
		q = q.OrderExpr("? DESC", bun.Ident("id"))
	}

	err := q.Scan(ctx, &pendingInts)
	if err != nil {
		return nil, err
	}

	// Catch case of no interactions early
	if len(pendingInts) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want interactions
	// to be sorted by ID desc, so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(pendingInts)
	}

	// For each pending interaction,
	// populate relevant model from DB.
	for _, pendingInt := range pendingInts {
		switch pendingInt.InteractionType {

		case gtsmodel.InteractionLike:
			pendingInt.Like, err = r.state.DB.GetStatusFaveByID(ctx, pendingInt.ID)

		case gtsmodel.InteractionReply:
			pendingInt.Reply, err = r.state.DB.GetStatusByID(ctx, pendingInt.ID)

		case gtsmodel.InteractionAnnounce:
			pendingInt.Announce, err = r.state.DB.GetStatusByID(ctx, pendingInt.ID)
		}

		if err != nil {
			return nil, err
		}
	}

	return pendingInts, nil
}
