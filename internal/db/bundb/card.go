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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type cardDB struct {
	db    *bun.DB
	state *state.State
}

func (c *cardDB) PutCard(ctx context.Context, card *gtsmodel.Card) error {
	return c.state.Caches.DB.Card.Store(card, func() error {
		_, err := c.db.NewInsert().Model(card).Exec(ctx)
		return err
	})
}

func (c *cardDB) GetCardByID(ctx context.Context, id string) (*gtsmodel.Card, error) {
	return c.state.Caches.DB.Card.LoadOne("ID", func() (*gtsmodel.Card, error) {
		var card gtsmodel.Card

		q := c.db.
			NewSelect().
			Model(&card).
			Where("? = ?", bun.Ident("card.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &card, nil
	}, id)
}

func (c *cardDB) UpdateCard(ctx context.Context, card *gtsmodel.Card, columns ...string) error {
	return c.state.Caches.DB.Card.Store(card, func() error {
		return c.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			_, err := tx.NewUpdate().
				Model(card).
				Column(columns...).
				Where("? = ?", bun.Ident("id"), card.ID).
				Exec(ctx)
			return err
		})
	})
}

func (c *cardDB) DeleteCardByID(ctx context.Context, id string) error {
	// Delete card with ID from the database.
	if err := c.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

		// Delete card from database.
		if _, err := tx.NewDelete().
			Table("cards").
			Where("? = ?", bun.Ident("id"), id).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
