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

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
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
