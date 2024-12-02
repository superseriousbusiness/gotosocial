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
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/paging"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

type interactionDB struct {
	db    *bun.DB
	state *state.State
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

func (i *interactionDB) GetInteractionRequestByURI(ctx context.Context, uri string) (*gtsmodel.InteractionRequest, error) {
	return i.getInteractionRequest(
		ctx,
		"URI",
		func(request *gtsmodel.InteractionRequest) error {
			return i.
				newInteractionRequestQ(request).
				Where("? = ?", bun.Ident("interaction_request.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (i *interactionDB) GetInteractionRequestsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.InteractionRequest, error) {
	// Load all interaction request IDs via cache loader callbacks.
	requests, err := i.state.Caches.DB.InteractionRequest.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.InteractionRequest, error) {
			// Preallocate expected length of uncached interaction requests.
			requests := make([]*gtsmodel.InteractionRequest, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := i.db.NewSelect().
				Model(&requests).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return requests, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the requests by their
	// IDs to ensure in correct order.
	getID := func(r *gtsmodel.InteractionRequest) string { return r.ID }
	xslices.OrderBy(requests, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return requests, nil
	}

	// Populate all loaded interaction requests, removing those we
	// fail to populate (removes needing so many nil checks everywhere).
	requests = slices.DeleteFunc(requests, func(request *gtsmodel.InteractionRequest) bool {
		if err := i.PopulateInteractionRequest(ctx, request); err != nil {
			log.Errorf(ctx, "error populating %s: %v", request.ID, err)
			return true
		}
		return false
	})

	return requests, nil
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

	// Depending on the interaction type, *try* to populate
	// the related model, but don't error if this is not
	// possible, as it may have just already been deleted
	// by its owner and we haven't cleaned up yet.
	switch req.InteractionType {

	case gtsmodel.InteractionLike:
		req.Like, err = i.state.DB.GetStatusFaveByURI(ctx, req.InteractionURI)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Appendf("error populating interactionRequest Like: %w", err)
		}

	case gtsmodel.InteractionReply:
		req.Reply, err = i.state.DB.GetStatusByURI(ctx, req.InteractionURI)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Appendf("error populating interactionRequest Reply: %w", err)
		}

	case gtsmodel.InteractionAnnounce:
		req.Announce, err = i.state.DB.GetStatusByURI(ctx, req.InteractionURI)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			errs.Appendf("error populating interactionRequest Announce: %w", err)
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

func (i *interactionDB) UpdateInteractionRequest(ctx context.Context, request *gtsmodel.InteractionRequest, columns ...string) error {
	return i.state.Caches.DB.InteractionRequest.Store(request, func() error {
		_, err := i.db.
			NewUpdate().
			Model(request).
			Where("? = ?", bun.Ident("interaction_request.id"), request.ID).
			Column(columns...).
			Exec(ctx)
		return err
	})
}

func (i *interactionDB) DeleteInteractionRequestByID(ctx context.Context, id string) error {
	// Delete interaction request by ID.
	if _, err := i.db.NewDelete().
		Table("interaction_requests").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}

	// Invalidate cached interaction request with ID.
	i.state.Caches.DB.InteractionRequest.Invalidate("ID", id)

	return nil
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
		).
		// Select only interaction requests that
		// are neither accepted or rejected yet.
		Where("? IS NULL", bun.Ident("accepted_at")).
		Where("? IS NULL", bun.Ident("rejected_at"))

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

	// Load all interaction requests by their IDs.
	return i.GetInteractionRequestsByIDs(ctx, reqIDs)
}

func (i *interactionDB) IsInteractionRejected(ctx context.Context, interactionURI string) (bool, error) {
	req, err := i.GetInteractionRequestByInteractionURI(ctx, interactionURI)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return false, gtserror.Newf("db error getting interaction request: %w", err)
	}

	if req == nil {
		// No interaction req at all with this
		// interactionURI so it can't be rejected.
		return false, nil
	}

	return req.IsRejected(), nil
}
