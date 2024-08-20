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

func (i *interactionDB) newInteractionApprovalQ(approval interface{}) *bun.SelectQuery {
	return i.db.
		NewSelect().
		Model(approval)
}

func (i *interactionDB) GetInteractionApprovalByID(ctx context.Context, id string) (*gtsmodel.InteractionApproval, error) {
	return i.getInteractionApproval(
		ctx,
		"ID",
		func(approval *gtsmodel.InteractionApproval) error {
			return i.
				newInteractionApprovalQ(approval).
				Where("? = ?", bun.Ident("interaction_approval.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (i *interactionDB) GetInteractionApprovalByURI(ctx context.Context, uri string) (*gtsmodel.InteractionApproval, error) {
	return i.getInteractionApproval(
		ctx,
		"URI",
		func(approval *gtsmodel.InteractionApproval) error {
			return i.
				newInteractionApprovalQ(approval).
				Where("? = ?", bun.Ident("interaction_approval.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (i *interactionDB) getInteractionApproval(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.InteractionApproval) error,
	keyParts ...any,
) (*gtsmodel.InteractionApproval, error) {
	// Fetch approval from database cache with loader callback
	approval, err := i.state.Caches.DB.InteractionApproval.LoadOne(lookup, func() (*gtsmodel.InteractionApproval, error) {
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

	if err := i.PopulateInteractionApproval(ctx, approval); err != nil {
		return nil, err
	}

	return approval, nil
}

func (i *interactionDB) PopulateInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error {
	var (
		err  error
		errs = gtserror.NewMultiError(4)
	)

	if approval.Status == nil {
		// Target status is not set, fetch from the database.
		approval.Status, err = i.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			approval.StatusID,
		)
		if err != nil {
			errs.Appendf("error populating interactionApproval target status: %w", err)
		}
	}

	if approval.Account == nil {
		// Account is not set, fetch from the database.
		approval.Account, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			approval.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionApproval account: %w", err)
		}
	}

	if approval.InteractingAccount == nil {
		// InteractingAccount is not set, fetch from the database.
		approval.InteractingAccount, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			approval.InteractingAccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionApproval interacting account: %w", err)
		}
	}

	switch approval.InteractionType {

	case gtsmodel.InteractionLike:
		approval.Like, err = i.state.DB.GetStatusFaveByURI(ctx, approval.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionApproval Like")
		}

	case gtsmodel.InteractionReply:
		approval.Reply, err = i.state.DB.GetStatusByURI(ctx, approval.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionApproval Reply")
		}

	case gtsmodel.InteractionAnnounce:
		approval.Announce, err = i.state.DB.GetStatusByURI(ctx, approval.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionApproval Announce")
		}
	}

	return errs.Combine()
}

func (i *interactionDB) PutInteractionApproval(ctx context.Context, approval *gtsmodel.InteractionApproval) error {
	return i.state.Caches.DB.InteractionApproval.Store(approval, func() error {
		_, err := i.db.NewInsert().Model(approval).Exec(ctx)
		return err
	})
}

func (i *interactionDB) DeleteInteractionApprovalByID(ctx context.Context, id string) error {
	defer i.state.Caches.DB.InteractionApproval.Invalidate("ID", id)

	_, err := i.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("interaction_approvals"), bun.Ident("interaction_approval")).
		Where("? = ?", bun.Ident("interaction_approval.id"), id).
		Exec(ctx)
	return err
}

func (i *interactionDB) newInteractionRejectionQ(rejection interface{}) *bun.SelectQuery {
	return i.db.
		NewSelect().
		Model(rejection)
}

func (i *interactionDB) GetInteractionRejectionByID(ctx context.Context, id string) (*gtsmodel.InteractionRejection, error) {
	return i.getInteractionRejection(
		ctx,
		"ID",
		func(rejection *gtsmodel.InteractionRejection) error {
			return i.
				newInteractionRejectionQ(rejection).
				Where("? = ?", bun.Ident("interaction_rejection.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (i *interactionDB) GetInteractionRejectionByURI(ctx context.Context, uri string) (*gtsmodel.InteractionRejection, error) {
	return i.getInteractionRejection(
		ctx,
		"URI",
		func(rejection *gtsmodel.InteractionRejection) error {
			return i.
				newInteractionRejectionQ(rejection).
				Where("? = ?", bun.Ident("interaction_rejection.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (i *interactionDB) InteractionRejected(ctx context.Context, interactionURI string) (bool, error) {
	ids := []string{}

	err := i.db.
		NewSelect().
		Column("id").
		Table("interaction_rejections").
		Where("? = ?", bun.Ident("interaction_uri"), interactionURI).
		Scan(ctx, &ids)

	return len(ids) != 0, err
}

func (i *interactionDB) getInteractionRejection(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.InteractionRejection) error,
	keyParts ...any,
) (*gtsmodel.InteractionRejection, error) {
	// Fetch rejection from database cache with loader callback
	rejection, err := i.state.Caches.DB.InteractionRejection.LoadOne(lookup, func() (*gtsmodel.InteractionRejection, error) {
		var rejection gtsmodel.InteractionRejection

		// Not cached! Perform database query
		if err := dbQuery(&rejection); err != nil {
			return nil, err
		}

		return &rejection, nil
	}, keyParts...)
	if err != nil {
		// Error already processed.
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return rejection, nil
	}

	if err := i.PopulateInteractionRejection(ctx, rejection); err != nil {
		return nil, err
	}

	return rejection, nil
}

func (i *interactionDB) PopulateInteractionRejection(ctx context.Context, rejection *gtsmodel.InteractionRejection) error {
	var (
		err  error
		errs = gtserror.NewMultiError(4)
	)

	if rejection.Status == nil {
		// Target status is not set, TRY TO fetch from the db,
		// but since we might not keep rejected statuses around,
		// don't fail if this is not available.
		rejection.Status, err = i.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			rejection.StatusID,
		)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Appendf("error populating interactionRejection target status: %w", err)
		}
	}

	if rejection.Account == nil {
		// Account is not set, fetch from the database.
		rejection.Account, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			rejection.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionRejection account: %w", err)
		}
	}

	if rejection.InteractingAccount == nil {
		// InteractingAccount is not set, fetch from the database.
		rejection.InteractingAccount, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			rejection.InteractingAccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionRejection interacting account: %w", err)
		}
	}

	switch rejection.InteractionType {

	case gtsmodel.InteractionLike:
		rejection.Like, err = i.state.DB.GetStatusFaveByURI(ctx, rejection.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionRejection Like")
		}

	case gtsmodel.InteractionReply:
		rejection.Reply, err = i.state.DB.GetStatusByURI(ctx, rejection.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionRejection Reply")
		}

	case gtsmodel.InteractionAnnounce:
		rejection.Announce, err = i.state.DB.GetStatusByURI(ctx, rejection.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionRejection Announce")
		}
	}

	return errs.Combine()
}

func (i *interactionDB) PutInteractionRejection(ctx context.Context, rejection *gtsmodel.InteractionRejection) error {
	return i.state.Caches.DB.InteractionRejection.Store(rejection, func() error {
		_, err := i.db.NewInsert().Model(rejection).Exec(ctx)
		return err
	})
}

func (i *interactionDB) DeleteInteractionRejectionByID(ctx context.Context, id string) error {
	defer i.state.Caches.DB.InteractionRejection.Invalidate("ID", id)

	_, err := i.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("interaction_rejections"), bun.Ident("interaction_rejection")).
		Where("? = ?", bun.Ident("interaction_rejection.id"), id).
		Exec(ctx)
	return err
}

func (i *interactionDB) GetInteractionsRequestsForAcct(
	ctx context.Context,
	acctID string,
	statusID string,
	likes bool,
	replies bool,
	boosts bool,
	page *paging.Page,
) ([]*gtsmodel.InteractionRequest, error) {
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
		reqIDs = make([]string, 0, limit)
	)

	// Create the basic select query.
	q := i.db.
		NewSelect().
		Column("id").
		TableExpr(
			"? AS ?",
			bun.Ident("interaction_requests"),
			bun.Ident("interaction_request"),
		)

	// Select interactions targeting status.
	if statusID != "" {
		q = q.Where("? = ?", bun.Ident("status_id"), statusID)
	}

	// Select interactions targeting account.
	if acctID != "" {
		q = q.Where("? = ?", bun.Ident("target_account_id"), acctID)
	}

	// Figure out which types of interaction are
	// being sought, and add them to the query.
	wantTypes := make([]gtsmodel.InteractionType, 0, 3)
	if likes {
		wantTypes = append(wantTypes, gtsmodel.InteractionLike)
	}
	if replies {
		wantTypes = append(wantTypes, gtsmodel.InteractionReply)
	}
	if boosts {
		wantTypes = append(wantTypes, gtsmodel.InteractionAnnounce)
	}
	q = q.Where("? IN (?)", bun.Ident("interaction_type"), bun.In(wantTypes))

	// Add paging param max ID.
	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("id"), maxID)
	}

	// Add paging param min ID.
	if minID != "" {
		q = q.Where("? > ?", bun.Ident("id"), minID)
	}

	// Add paging param order.
	if order == paging.OrderAscending {
		// Page up.
		q = q.OrderExpr("? ASC", bun.Ident("id"))
	} else {
		// Page down.
		q = q.OrderExpr("? DESC", bun.Ident("id"))
	}

	// Add paging param limit.
	if limit > 0 {
		q = q.Limit(limit)
	}

	// Execute the query and scan into IDs.
	err := q.Scan(ctx, &reqIDs)
	if err != nil {
		return nil, err
	}

	// Catch case of no items early
	if len(reqIDs) == 0 {
		return nil, db.ErrNoEntries
	}

	// If we're paging up, we still want interactions
	// to be sorted by ID desc, so reverse ids slice.
	if order == paging.OrderAscending {
		slices.Reverse(reqIDs)
	}

	// For each interaction request ID,
	// select the interaction request.
	reqs := make([]*gtsmodel.InteractionRequest, 0, len(reqIDs))
	for _, id := range reqIDs {
		req, err := i.GetInteractionRequestByID(ctx, id)
		if err != nil {
			return nil, err
		}

		reqs = append(reqs, req)
	}

	return reqs, nil
}

func (i *interactionDB) newInteractionRequestQ(request interface{}) *bun.SelectQuery {
	return i.db.
		NewSelect().
		Model(request)
}

func (i *interactionDB) GetInteractionRequestByID(ctx context.Context, id string) (*gtsmodel.InteractionRequest, error) {
	return i.getInteractionRequest(
		ctx,
		"ID",
		func(request *gtsmodel.InteractionRequest) error {
			return i.
				newInteractionRequestQ(request).
				Where("? = ?", bun.Ident("interaction_request.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (i *interactionDB) GetInteractionRequestByInteractionURI(ctx context.Context, uri string) (*gtsmodel.InteractionRequest, error) {
	return i.getInteractionRequest(
		ctx,
		"InteractionURI",
		func(request *gtsmodel.InteractionRequest) error {
			return i.
				newInteractionRequestQ(request).
				Where("? = ?", bun.Ident("interaction_request.interaction_uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (i *interactionDB) getInteractionRequest(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.InteractionRequest) error,
	keyParts ...any,
) (*gtsmodel.InteractionRequest, error) {
	// Fetch request from database cache with loader callback
	request, err := i.state.Caches.DB.InteractionRequest.LoadOne(lookup, func() (*gtsmodel.InteractionRequest, error) {
		var request gtsmodel.InteractionRequest

		// Not cached! Perform database query
		if err := dbQuery(&request); err != nil {
			return nil, err
		}

		return &request, nil
	}, keyParts...)
	if err != nil {
		// Error already processed.
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return request, nil
	}

	if err := i.PopulateInteractionRequest(ctx, request); err != nil {
		return nil, err
	}

	return request, nil
}

func (i *interactionDB) PopulateInteractionRequest(ctx context.Context, req *gtsmodel.InteractionRequest) error {
	var (
		err  error
		errs = gtserror.NewMultiError(4)
	)

	if req.Status == nil {
		// Target status is not set, fetch from the database.
		req.Status, err = i.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			req.StatusID,
		)
		if err != nil {
			errs.Appendf("error populating interactionRequest target: %w", err)
		}
	}

	if req.TargetAccount == nil {
		// Target account is not set, fetch from the database.
		req.TargetAccount, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			req.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionRequest target account: %w", err)
		}
	}

	if req.InteractingAccount == nil {
		// InteractingAccount is not set, fetch from the database.
		req.InteractingAccount, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			req.InteractingAccountID,
		)
		if err != nil {
			errs.Appendf("error populating interactionRequest interacting account: %w", err)
		}
	}

	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		req.Like, err = i.state.DB.GetStatusFaveByURI(ctx, req.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionRequest Like")
		}

	case gtsmodel.InteractionReply:
		req.Reply, err = i.state.DB.GetStatusByURI(ctx, req.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionRequest Reply")
		}

	case gtsmodel.InteractionAnnounce:
		req.Announce, err = i.state.DB.GetStatusByURI(ctx, req.InteractionURI)
		if err != nil {
			errs.Appendf("error populating interactionRequest Announce")
		}
	}

	return errs.Combine()
}

func (i *interactionDB) PutInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest) error {
	return i.state.Caches.DB.InteractionRequest.Store(request, func() error {
		_, err := i.db.NewInsert().Model(request).Exec(ctx)
		return err
	})
}

func (i *interactionDB) DeleteInteractionRequestByID(ctx context.Context, id string) error {
	defer i.state.Caches.DB.InteractionRequest.Invalidate("ID", id)

	_, err := i.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("interaction_requests"), bun.Ident("interaction_request")).
		Where("? = ?", bun.Ident("interaction_request.id"), id).
		Exec(ctx)
	return err
}
