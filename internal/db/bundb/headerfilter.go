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

func (h *hdrFilterDB) HeaderMatchPositive(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.PositiveHdrFilters.MatchPositive(hdr, func() ([]gtsmodel.HeaderFilter, error) {
		return h.selectFilters(ctx, gtsmodel.HeaderFilterTypePositive)
	})
}

func (h *hdrFilterDB) HeaderMatchNegative(ctx context.Context, hdr http.Header) (bool, error) {
	return h.state.Caches.NegativeHdrFilters.MatchNegative(hdr, func() ([]gtsmodel.HeaderFilter, error) {
		return h.selectFilters(ctx, gtsmodel.HeaderFilterTypeNegative)
	})
}

func (h *hdrFilterDB) selectFilters(ctx context.Context, filterType uint8) ([]gtsmodel.HeaderFilter, error) {
	var filters []gtsmodel.HeaderFilter
	if err := h.db.NewSelect().
		Model(&filters).
		Where("? = ?", bun.Ident("type"), filterType).
		Scan(ctx); err != nil {
		return nil, err
	}
	return filters, nil
}

func (h *hdrFilterDB) PutHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error {
	if _, err := h.db.NewInsert().
		Model(filter).
		Exec(ctx); err != nil {
		return err
	}
	switch filter.Type {
	case gtsmodel.HeaderFilterTypePositive:
		h.state.Caches.PositiveHdrFilters.Clear()
	case gtsmodel.HeaderFilterTypeNegative:
		h.state.Caches.NegativeHdrFilters.Clear()
	}
	return nil
}

func (h *hdrFilterDB) DeleteHeaderFilterByID(ctx context.Context, id string) error {
	var filterType uint8
	if _, err := h.db.NewDelete().
		Table("header_filters").
		Where("? = ?", bun.Ident("id"), id).
		Returning("?", bun.Ident("type")).
		Exec(ctx, &filterType); err != nil {
		return err
	}
	switch filterType {
	case gtsmodel.HeaderFilterTypePositive:
		h.state.Caches.PositiveHdrFilters.Clear()
	case gtsmodel.HeaderFilterTypeNegative:
		h.state.Caches.NegativeHdrFilters.Clear()
	default: // i.e. zero value, does nothing
	}
	return nil
}
