/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package bundb

import (
	"context"
	"errors"

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type statusFaveDB struct {
	conn  *DBConn
	state *state.State
}

func (s *statusFaveDB) GetStatusFave(ctx context.Context, id string) (*gtsmodel.StatusFave, db.Error) {
	fave := new(gtsmodel.StatusFave)

	err := s.conn.
		NewSelect().
		Model(fave).
		Where("? = ?", bun.Ident("status_fave.ID"), id).
		Scan(ctx)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	fave.Account, err = s.state.DB.GetAccountByID(ctx, fave.AccountID)
	if err != nil {
		log.Errorf(ctx, "error getting status fave account %q: %v", fave.AccountID, err)
	}

	fave.TargetAccount, err = s.state.DB.GetAccountByID(ctx, fave.TargetAccountID)
	if err != nil {
		log.Errorf(ctx, "error getting status fave target account %q: %v", fave.TargetAccountID, err)
	}

	fave.Status, err = s.state.DB.GetStatusByID(ctx, fave.StatusID)
	if err != nil {
		log.Errorf(ctx, "error getting status fave status %q: %v", fave.StatusID, err)
	}

	return fave, nil
}

func (s *statusFaveDB) GetStatusFaveByAccountID(ctx context.Context, accountID string, statusID string) (*gtsmodel.StatusFave, db.Error) {
	var id string

	err := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave")).
		Column("status_fave.id").
		Where("? = ?", bun.Ident("status_fave.account_id"), accountID).
		Where("? = ?", bun.Ident("status_fave.status_id"), statusID).
		Scan(ctx, &id)
	if err != nil {
		return nil, s.conn.ProcessError(err)
	}

	return s.GetStatusFave(ctx, id)
}

func (s *statusFaveDB) GetStatusFaves(ctx context.Context, statusID string) ([]*gtsmodel.StatusFave, db.Error) {
	ids := []string{}

	if err := s.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave")).
		Column("status_fave.id").
		Where("? = ?", bun.Ident("status_fave.status_id"), statusID).
		Scan(ctx, &ids); err != nil {
		return nil, s.conn.ProcessError(err)
	}

	faves := make([]*gtsmodel.StatusFave, 0, len(ids))
	for _, id := range ids {
		fave, err := s.GetStatusFave(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting status fave %q: %v", id, err)
			continue
		}

		faves = append(faves, fave)
	}

	return faves, nil
}

func (s *statusFaveDB) PutStatusFave(ctx context.Context, statusFave *gtsmodel.StatusFave) db.Error {
	_, err := s.conn.
		NewInsert().
		Model(statusFave).
		Exec(ctx)

	return s.conn.ProcessError(err)
}

func (s *statusFaveDB) DeleteStatusFave(ctx context.Context, id string) db.Error {
	_, err := s.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave")).
		Where("? = ?", bun.Ident("status_fave.id"), id).
		Exec(ctx)

	return s.conn.ProcessError(err)
}

func (s *statusFaveDB) DeleteStatusFaves(ctx context.Context, targetAccountID string, originAccountID string) db.Error {
	if targetAccountID == "" && originAccountID == "" {
		return errors.New("DeleteStatusFaves: one of targetAccountID or originAccountID must be set")
	}

	// TODO: Capture fave IDs in a RETURNING
	// statement (when faves have a cache),
	// + use the IDs to invalidate cache entries.

	q := s.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave"))

	if targetAccountID != "" {
		q = q.Where("? = ?", bun.Ident("status_fave.target_account_id"), targetAccountID)
	}

	if originAccountID != "" {
		q = q.Where("? = ?", bun.Ident("status_fave.account_id"), originAccountID)
	}

	if _, err := q.Exec(ctx); err != nil {
		return s.conn.ProcessError(err)
	}

	return nil
}

func (s *statusFaveDB) DeleteStatusFavesForStatus(ctx context.Context, statusID string) db.Error {
	// TODO: Capture fave IDs in a RETURNING
	// statement (when faves have a cache),
	// + use the IDs to invalidate cache entries.

	q := s.conn.
		NewDelete().
		TableExpr("? AS ?", bun.Ident("status_faves"), bun.Ident("status_fave")).
		Where("? = ?", bun.Ident("status_fave.status_id"), statusID)

	if _, err := q.Exec(ctx); err != nil {
		return s.conn.ProcessError(err)
	}

	return nil
}
