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
	"net/http"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type hdrFilterDB struct {
	db    *DB
	state *state.State
}

func (h *hdrFilterDB) HeaderAllow(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.AllowHeaderFilters.Allow(hdr, func() ([]gtsmodel.HeaderFilter, error) {
		var filters []gtsmodel.HeaderFilterAllow
		if err := h.db.NewSelect().
			Model(&filters).
			Scan(ctx); err != nil {
			return nil, err
		}
		base := make([]gtsmodel.HeaderFilter, len(filters))
		for i := range base {
			base[i] = filters[i].HeaderFilter
		}
		return base, nil
	})
}

func (h *hdrFilterDB) HeaderBlock(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.BlockHeaderFilters.Block(hdr, func() ([]gtsmodel.HeaderFilter, error) {
		var filters []gtsmodel.HeaderFilterBlock
		if err := h.db.NewSelect().
			Model(&filters).
			Scan(ctx); err != nil {
			return nil, err
		}
		base := make([]gtsmodel.HeaderFilter, len(filters))
		for i := range base {
			base[i] = filters[i].HeaderFilter
		}
		return base, nil
	})
}

func (h *hdrFilterDB) PutAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow) error {
	if _, err := h.db.NewInsert().
		Model(filter).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *hdrFilterDB) PutBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow) error {
	if _, err := h.db.NewInsert().
		Model(filter).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}

func (h *hdrFilterDB) UpdateAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow, cols ...string) error {
	if _, err := h.db.NewUpdate().
		Model(filter).
		Column(cols...).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *hdrFilterDB) UpdateBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilterAllow, cols ...string) error {
	if _, err := h.db.NewUpdate().
		Model(filter).
		Column(cols...).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}

func (h *hdrFilterDB) DeleteAllowHeaderFilter(ctx context.Context, id string) error {
	if _, err := h.db.NewDelete().
		Table("allow_header_filters").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *hdrFilterDB) DeleteBlockHeaderFilter(ctx context.Context, id string) error {
	if _, err := h.db.NewDelete().
		Table("block_header_filters").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}
