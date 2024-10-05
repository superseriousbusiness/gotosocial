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
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type mentionDB struct {
	db    *bun.DB
	state *state.State
}

func (m *mentionDB) GetMention(ctx context.Context, id string) (*gtsmodel.Mention, error) {
	mention, err := m.state.Caches.DB.Mention.LoadOne("ID", func() (*gtsmodel.Mention, error) {
		var mention gtsmodel.Mention

		q := m.db.
			NewSelect().
			Model(&mention).
			Where("? = ?", bun.Ident("mention.id"), id)

		if err := q.Scan(ctx); err != nil {
			return nil, err
		}

		return &mention, nil
	}, id)
	if err != nil {
		return nil, err
	}

	// Further populate the mention fields where applicable.
	if err := m.PopulateMention(ctx, mention); err != nil {
		return nil, err
	}

	return mention, nil
}

func (m *mentionDB) GetMentions(ctx context.Context, ids []string) ([]*gtsmodel.Mention, error) {
	// Load all mention IDs via cache loader callbacks.
	mentions, err := m.state.Caches.DB.Mention.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.Mention, error) {
			// Preallocate expected length of uncached mentions.
			mentions := make([]*gtsmodel.Mention, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) IDs.
			if err := m.db.NewSelect().
				Model(&mentions).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return mentions, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the mentions by their
	// IDs to ensure in correct order.
	getID := func(m *gtsmodel.Mention) string { return m.ID }
	util.OrderBy(mentions, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return mentions, nil
	}

	// Populate all loaded mentions, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	mentions = slices.DeleteFunc(mentions, func(mention *gtsmodel.Mention) bool {
		if err := m.PopulateMention(ctx, mention); err != nil {
			log.Errorf(ctx, "error populating mention %s: %v", mention.ID, err)
			return true
		}
		return false
	})

	return mentions, nil

}

func (m *mentionDB) PopulateMention(ctx context.Context, mention *gtsmodel.Mention) (err error) {
	var errs gtserror.MultiError

	if mention.Status == nil {
		// Set the mention originating status.
		mention.Status, err = m.state.DB.GetStatusByID(
			gtscontext.SetBarebones(ctx),
			mention.StatusID,
		)
		if err != nil {
			return gtserror.Newf("error populating mention status: %w", err)
		}
	}

	if mention.OriginAccount == nil {
		// Set the mention origin account model.
		mention.OriginAccount, err = m.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			mention.OriginAccountID,
		)
		if err != nil {
			return gtserror.Newf("error populating mention origin account: %w", err)
		}
	}

	if mention.TargetAccount == nil {
		// Set the mention target account model.
		mention.TargetAccount, err = m.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			mention.TargetAccountID,
		)
		if err != nil {
			return gtserror.Newf("error populating mention target account: %w", err)
		}
	}

	return errs.Combine()
}

func (m *mentionDB) PutMention(ctx context.Context, mention *gtsmodel.Mention) error {
	return m.state.Caches.DB.Mention.Store(mention, func() error {
		_, err := m.db.NewInsert().Model(mention).Exec(ctx)
		return err
	})
}

func (m *mentionDB) DeleteMentionByID(ctx context.Context, id string) error {
	// Delete mention with given ID,
	// returning the deleted models.
	if _, err := m.db.NewDelete().
		Table("mentions").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate the cached mention with ID.
	m.state.Caches.DB.Mention.Invalidate("ID", id)

	return nil
}
