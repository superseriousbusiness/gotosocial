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

type headerFilterDB struct {
	db    *DB
	state *state.State
}

func (h *headerFilterDB) AllowHeaderRegularMatch(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.AllowHeaderFilters.RegularMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetAllowHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) AllowHeaderInverseMatch(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.AllowHeaderFilters.InverseMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetAllowHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) BlockHeaderRegularMatch(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.AllowHeaderFilters.RegularMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetAllowHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) BlockHeaderInverseMatch(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.AllowHeaderFilters.InverseMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetAllowHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) GetAllowHeaderFilter(ctx context.Context, id string) (*gtsmodel.HeaderFilter, error) {
	var filter gtsmodel.HeaderFilter
	if err := h.db.NewSelect().
		Table("header_filter_allows").
		Where("? = ?", bun.Ident("id"), id).
		Scan(ctx, &filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

func (h *headerFilterDB) GetBlockHeaderFilter(ctx context.Context, id string) (*gtsmodel.HeaderFilter, error) {
	var filter gtsmodel.HeaderFilter
	if err := h.db.NewSelect().
		Table("header_filter_blocks").
		Where("? = ?", bun.Ident("id"), id).
		Scan(ctx, &filter); err != nil {
		return nil, err
	}
	return &filter, nil
}

func (h *headerFilterDB) GetAllowHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilter, error) {
	var filters []*gtsmodel.HeaderFilter
	err := h.db.NewSelect().
		Table("header_filter_allows").
		Scan(ctx, &filters)
	return filters, err
}

func (h *headerFilterDB) GetBlockHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilter, error) {
	var filters []*gtsmodel.HeaderFilter
	err := h.db.NewSelect().
		Table("header_filter_blocks").
		Scan(ctx, &filters)
	return filters, err
}

func (h *headerFilterDB) PutAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error {
	if _, err := h.db.NewInsert().
		Table("header_filter_allows").
		Exec(ctx, filter); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) PutBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error {
	if _, err := h.db.NewInsert().
		Table("header_filter_blocks").
		Exec(ctx, filter); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) UpdateAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter, cols ...string) error {
	if _, err := h.db.NewUpdate().
		Table("header_filter_allows").
		Column(cols...).
		Exec(ctx, filter); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) UpdateBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter, cols ...string) error {
	if _, err := h.db.NewUpdate().
		Table("header_filter_blocks").
		Column(cols...).
		Exec(ctx, filter); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) DeleteAllowHeaderFilter(ctx context.Context, id string) error {
	if _, err := h.db.NewDelete().
		Table("header_filter_allows").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) DeleteBlockHeaderFilter(ctx context.Context, id string) error {
	if _, err := h.db.NewDelete().
		Table("header_filter_blocks").
		Where("? = ?", bun.Ident("id"), id).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}
