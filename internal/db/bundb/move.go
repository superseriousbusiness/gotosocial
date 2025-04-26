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
	"net/url"
	"time"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtscontext"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type moveDB struct {
	db    *bun.DB
	state *state.State
}

func (m *moveDB) GetMoveByID(
	ctx context.Context,
	id string,
) (*gtsmodel.Move, error) {
	return m.getMove(
		ctx,
		"ID",
		func(move *gtsmodel.Move) error {
			return m.db.
				NewSelect().
				Model(move).
				Where("? = ?", bun.Ident("move.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (m *moveDB) GetMoveByURI(
	ctx context.Context,
	uri string,
) (*gtsmodel.Move, error) {
	return m.getMove(
		ctx,
		"URI",
		func(move *gtsmodel.Move) error {
			return m.db.
				NewSelect().
				Model(move).
				Where("? = ?", bun.Ident("move.uri"), uri).
				Scan(ctx)
		},
		uri,
	)
}

func (m *moveDB) GetMoveByOriginTarget(
	ctx context.Context,
	originURI string,
	targetURI string,
) (*gtsmodel.Move, error) {
	return m.getMove(
		ctx,
		"OriginURI,TargetURI",
		func(move *gtsmodel.Move) error {
			return m.db.
				NewSelect().
				Model(move).
				Where("? = ?", bun.Ident("move.origin_uri"), originURI).
				Where("? = ?", bun.Ident("move.target_uri"), targetURI).
				Scan(ctx)
		},
		originURI, targetURI,
	)
}

func (m *moveDB) GetLatestMoveSuccessInvolvingURIs(
	ctx context.Context,
	uri1 string,
	uri2 string,
) (time.Time, error) {
	// Get at most 1 latest Move
	// involving the provided URIs.
	var moves []*gtsmodel.Move
	err := m.db.
		NewSelect().
		Model(&moves).
		Column("succeeded_at").
		Where("? = ?", bun.Ident("move.origin_uri"), uri1).
		WhereOr("? = ?", bun.Ident("move.origin_uri"), uri2).
		WhereOr("? = ?", bun.Ident("move.target_uri"), uri1).
		WhereOr("? = ?", bun.Ident("move.target_uri"), uri2).
		Order("id DESC").
		Limit(1).
		Scan(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return time.Time{}, err
	}

	if len(moves) != 1 {
		return time.Time{}, nil
	}

	return moves[0].SucceededAt, nil
}

func (m *moveDB) GetLatestMoveAttemptInvolvingURIs(
	ctx context.Context,
	uri1 string,
	uri2 string,
) (time.Time, error) {
	// Get at most 1 latest Move
	// involving the provided URIs.
	var moves []*gtsmodel.Move
	err := m.db.
		NewSelect().
		Model(&moves).
		Column("attempted_at").
		Where("? = ?", bun.Ident("move.origin_uri"), uri1).
		WhereOr("? = ?", bun.Ident("move.origin_uri"), uri2).
		WhereOr("? = ?", bun.Ident("move.target_uri"), uri1).
		WhereOr("? = ?", bun.Ident("move.target_uri"), uri2).
		Order("id DESC").
		Limit(1).
		Scan(ctx)
	if err != nil && !errors.Is(err, db.ErrNoEntries) {
		return time.Time{}, err
	}

	if len(moves) != 1 {
		return time.Time{}, nil
	}

	return moves[0].AttemptedAt, nil
}

func (m *moveDB) getMove(
	ctx context.Context,
	lookup string,
	dbQuery func(*gtsmodel.Move) error,
	keyParts ...any,
) (*gtsmodel.Move, error) {
	move, err := m.state.Caches.DB.Move.LoadOne(lookup, func() (*gtsmodel.Move, error) {
		var move gtsmodel.Move

		// Not cached! Perform database query.
		if err := dbQuery(&move); err != nil {
			return nil, err
		}

		return &move, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		return move, nil
	}

	// Populate the Move by parsing out the URIs.
	if err := m.PopulateMove(ctx, move); err != nil {
		return nil, err
	}

	return move, nil
}

func (m *moveDB) PopulateMove(ctx context.Context, move *gtsmodel.Move) error {
	if move.Origin == nil {
		var err error
		move.Origin, err = url.Parse(move.OriginURI)
		if err != nil {
			return fmt.Errorf("error parsing Move originURI: %w", err)
		}
	}

	if move.Target == nil {
		var err error
		move.Target, err = url.Parse(move.TargetURI)
		if err != nil {
			return fmt.Errorf("error parsing Move targetURI: %w", err)
		}
	}

	return nil
}

func (m *moveDB) PutMove(ctx context.Context, move *gtsmodel.Move) error {
	return m.state.Caches.DB.Move.Store(move, func() error {
		_, err := m.db.
			NewInsert().
			Model(move).
			Exec(ctx)
		return err
	})
}

func (m *moveDB) UpdateMove(ctx context.Context, move *gtsmodel.Move, columns ...string) error {
	move.UpdatedAt = time.Now()
	if len(columns) > 0 {
		// If we're updating by column,
		// ensure "updated_at" is included.
		columns = append(columns, "updated_at")
	}

	return m.state.Caches.DB.Move.Store(move, func() error {
		_, err := m.db.
			NewUpdate().
			Model(move).
			Column(columns...).
			Where("? = ?", bun.Ident("move.id"), move.ID).
			Exec(ctx)
		return err
	})
}

func (m *moveDB) DeleteMoveByID(ctx context.Context, id string) error {
	// Delete move with given ID.
	if _, err := m.db.NewDelete().
		TableExpr("? AS ?", bun.Ident("moves"), bun.Ident("move")).
		Where("? = ?", bun.Ident("move.id"), id).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return nil
	}

	// Invalidate the cached move model with ID.
	m.state.Caches.DB.Move.Invalidate("ID", id)

	return nil
}
