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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func (r *relationshipDB) GetNote(ctx context.Context, sourceAccountID string, targetAccountID string) (*gtsmodel.AccountNote, error) {
	return r.getNote(
		ctx,
		"AccountID,TargetAccountID",
		func(note *gtsmodel.AccountNote) error {
			return r.db.NewSelect().Model(note).
				Where("? = ?", bun.Ident("account_id"), sourceAccountID).
				Where("? = ?", bun.Ident("target_account_id"), targetAccountID).
				Scan(ctx)
		},
		sourceAccountID,
		targetAccountID,
	)
}

func (r *relationshipDB) getNote(ctx context.Context, lookup string, dbQuery func(*gtsmodel.AccountNote) error, keyParts ...any) (*gtsmodel.AccountNote, error) {
	// Fetch note from cache with loader callback
	note, err := r.state.Caches.DB.AccountNote.LoadOne(lookup, func() (*gtsmodel.AccountNote, error) {
		var note gtsmodel.AccountNote

		// Not cached! Perform database query
		if err := dbQuery(&note); err != nil {
			return nil, err
		}

		return &note, nil
	}, keyParts...)
	if err != nil {
		// already processed
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// Only a barebones model was requested.
		return note, nil
	}

	// Further populate the account fields where applicable.
	if err := r.PopulateNote(ctx, note); err != nil {
		return nil, err
	}

	return note, nil
}

func (r *relationshipDB) PopulateNote(ctx context.Context, note *gtsmodel.AccountNote) error {
	var (
		errs = gtserror.NewMultiError(2)
		err  error
	)

	// Ensure note source account set.
	if note.Account == nil {
		note.Account, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			note.AccountID,
		)
		if err != nil {
			errs.Appendf("error populating note source account: %w", err)
		}
	}

	// Ensure note target account set.
	if note.TargetAccount == nil {
		note.TargetAccount, err = r.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			note.TargetAccountID,
		)
		if err != nil {
			errs.Appendf("error populating note target account: %w", err)
		}
	}

	return errs.Combine()
}

func (r *relationshipDB) PutNote(ctx context.Context, note *gtsmodel.AccountNote) error {
	note.UpdatedAt = time.Now()
	return r.state.Caches.DB.AccountNote.Store(note, func() error {
		_, err := r.db.
			NewInsert().
			Model(note).
			On("CONFLICT (?, ?) DO UPDATE", bun.Ident("account_id"), bun.Ident("target_account_id")).
			Set("? = ?, ? = ?", bun.Ident("updated_at"), note.UpdatedAt, bun.Ident("comment"), note.Comment).
			Exec(ctx)
		return err
	})
}
