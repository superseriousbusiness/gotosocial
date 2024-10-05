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
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type sinBinStatusDB struct {
	db    *bun.DB
	state *state.State
}

func (s *sinBinStatusDB) GetSinBinStatusByID(ctx context.Context, id string) (*gtsmodel.SinBinStatus, error) {
	return s.getSinBinStatus(
		"ID",
		func(sbStatus *gtsmodel.SinBinStatus) error {
			return s.db.
				NewSelect().
				Model(sbStatus).
				Where("? = ?", bun.Ident("sin_bin_status.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (s *sinBinStatusDB) GetSinBinStatusByURI(ctx context.Context, uri string) (*gtsmodel.SinBinStatus, error) {
	return s.getSinBinStatus(
		"URI",
		func(sbStatus *gtsmodel.SinBinStatus) error {
			return s.db.
				NewSelect().
				Model(sbStatus).
				Where("? = ?", bun.Ident("sin_bin_status.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (s *sinBinStatusDB) getSinBinStatus(
	lookup string,
	dbQuery func(*gtsmodel.SinBinStatus) error,
	keyParts ...any,
) (*gtsmodel.SinBinStatus, error) {
	// Fetch from database cache with loader callback.
	return s.state.Caches.DB.SinBinStatus.LoadOne(lookup, func() (*gtsmodel.SinBinStatus, error) {
		// Not cached! Perform database query.
		sbStatus := new(gtsmodel.SinBinStatus)
		if err := dbQuery(sbStatus); err != nil {
			return nil, err
		}

		return sbStatus, nil
	}, keyParts...)
}

func (s *sinBinStatusDB) PutSinBinStatus(ctx context.Context, sbStatus *gtsmodel.SinBinStatus) error {
	return s.state.Caches.DB.SinBinStatus.Store(sbStatus, func() error {
		_, err := s.db.
			NewInsert().
			Model(sbStatus).
			Exec(ctx)
		return err
	})
}

func (s *sinBinStatusDB) UpdateSinBinStatus(
	ctx context.Context,
	sbStatus *gtsmodel.SinBinStatus,
	columns ...string,
) error {
	sbStatus.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column,
		// ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return s.state.Caches.DB.SinBinStatus.Store(sbStatus, func() error {
		_, err := s.db.
			NewUpdate().
			Model(sbStatus).
			Column(columns...).
			Where("? = ?", bun.Ident("sin_bin_status.id"), sbStatus.ID).
			Exec(ctx)
		return err
	})
}

func (s *sinBinStatusDB) DeleteSinBinStatusByID(ctx context.Context, id string) error {
	// Delete the status from DB.
	if _, err := s.db.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("sin_bin_statuses"), bun.Ident("sin_bin_status")).
		Where("? = ?", bun.Ident("sin_bin_status.id"), id).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate any cached sinbin status model by ID.
	s.state.Caches.DB.SinBinStatus.Invalidate("ID", id)

	return nil
}
