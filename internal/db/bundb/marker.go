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
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type markerDB struct {
	db    *bun.DB
	state *state.State
}

/*
	MARKER FUNCTIONS
*/

func (m *markerDB) GetMarker(ctx context.Context, accountID string, name gtsmodel.MarkerName) (*gtsmodel.Marker, error) {
	marker, err := m.state.Caches.DB.Marker.LoadOne(
		"AccountID,Name",
		func() (*gtsmodel.Marker, error) {
			var marker gtsmodel.Marker

			if err := m.db.NewSelect().
				Model(&marker).
				Where("? = ? AND ? = ?", bun.Ident("account_id"), accountID, bun.Ident("name"), name).
				Scan(ctx); err != nil {
				return nil, err
			}

			return &marker, nil
		}, accountID, name,
	)
	if err != nil {
		return nil, err // already processed
	}

	return marker, nil
}

func (m *markerDB) UpdateMarker(ctx context.Context, marker *gtsmodel.Marker) error {
	prevMarker, err := m.GetMarker(ctx, marker.AccountID, marker.Name)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return fmt.Errorf("UpdateMarker: error fetching previous version of marker: %w", err)
	}

	marker.UpdatedAt = time.Now()
	if prevMarker != nil {
		marker.Version = prevMarker.Version + 1
	}

	return m.state.Caches.DB.Marker.Store(marker, func() error {
		if prevMarker == nil {
			if _, err := m.db.NewInsert().
				Model(marker).
				Exec(ctx); err != nil {
				return err
			}
			return nil
		}

		// Optimistic concurrency control: start a transaction, try to update a row with a previously retrieved version.
		// If the update in the transaction fails to actually change anything, another update happened concurrently, and
		// this update should be retried by the caller, which in this case involves sending HTTP 409 to the API client.
		return m.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			result, err := tx.NewUpdate().
				Model(marker).
				WherePK().
				Where("? = ?", bun.Ident("version"), prevMarker.Version).
				Exec(ctx)
			if err != nil {
				return err
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				return err
			}
			if rowsAffected == 0 {
				// Will trigger a rollback, although there should be no changes to roll back.
				return db.ErrAlreadyExists
			} else if rowsAffected > 1 {
				// This shouldn't happen.
				return db.ErrNoEntries
			}

			return nil
		})
	})
}
