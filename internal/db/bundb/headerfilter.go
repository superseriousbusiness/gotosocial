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
	"time"
	"unsafe"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/state"
	"github.com/uptrace/bun"
)

type headerFilterDB struct {
	db    *bun.DB
	state *state.State
}

func (h *headerFilterDB) AllowHeaderRegularMatch(ctx context.Context, hdr http.Header) (string, string, error) {
	return h.state.Caches.AllowHeaderFilters.RegularMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetAllowHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) AllowHeaderInverseMatch(ctx context.Context, hdr http.Header) (string, string, error) {
	return h.state.Caches.AllowHeaderFilters.InverseMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetAllowHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) BlockHeaderRegularMatch(ctx context.Context, hdr http.Header) (string, string, error) {
	return h.state.Caches.BlockHeaderFilters.RegularMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetBlockHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) BlockHeaderInverseMatch(ctx context.Context, hdr http.Header) (string, string, error) {
	return h.state.Caches.BlockHeaderFilters.InverseMatch(hdr, func() ([]*gtsmodel.HeaderFilter, error) {
		return h.GetBlockHeaderFilters(ctx)
	})
}

func (h *headerFilterDB) GetAllowHeaderFilter(ctx context.Context, id string) (*gtsmodel.HeaderFilter, error) {
	filter := new(gtsmodel.HeaderFilterAllow)
	if err := h.db.NewSelect().
		Model(filter).
		Where("? = ?", bun.Ident("id"), id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return fromAllowFilter(filter), nil
}

func (h *headerFilterDB) GetBlockHeaderFilter(ctx context.Context, id string) (*gtsmodel.HeaderFilter, error) {
	filter := new(gtsmodel.HeaderFilterBlock)
	if err := h.db.NewSelect().
		Model(filter).
		Where("? = ?", bun.Ident("id"), id).
		Scan(ctx); err != nil {
		return nil, err
	}
	return fromBlockFilter(filter), nil
}

func (h *headerFilterDB) GetAllowHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilter, error) {
	var filters []*gtsmodel.HeaderFilterAllow
	err := h.db.NewSelect().
		Model(&filters).
		Scan(ctx, &filters)
	return fromAllowFilters(filters), err
}

func (h *headerFilterDB) GetBlockHeaderFilters(ctx context.Context) ([]*gtsmodel.HeaderFilter, error) {
	var filters []*gtsmodel.HeaderFilterBlock
	err := h.db.NewSelect().
		Model(&filters).
		Scan(ctx, &filters)
	return fromBlockFilters(filters), err
}

func (h *headerFilterDB) PutAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error {
	if _, err := h.db.NewInsert().
		Model(toAllowFilter(filter)).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) PutBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter) error {
	if _, err := h.db.NewInsert().
		Model(toBlockFilter(filter)).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.BlockHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) UpdateAllowHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter, cols ...string) error {
	filter.UpdatedAt = time.Now()
	if len(cols) > 0 {
		// If we're updating by column,
		// ensure "updated_at" is included.
		cols = append(cols, "updated_at")
	}
	if _, err := h.db.NewUpdate().
		Model(toAllowFilter(filter)).
		Column(cols...).
		Where("? = ?", bun.Ident("id"), filter.ID).
		Exec(ctx); err != nil {
		return err
	}
	h.state.Caches.AllowHeaderFilters.Clear()
	return nil
}

func (h *headerFilterDB) UpdateBlockHeaderFilter(ctx context.Context, filter *gtsmodel.HeaderFilter, cols ...string) error {
	filter.UpdatedAt = time.Now()
	if len(cols) > 0 {
		// If we're updating by column,
		// ensure "updated_at" is included.
		cols = append(cols, "updated_at")
	}
	if _, err := h.db.NewUpdate().
		Model(toBlockFilter(filter)).
		Column(cols...).
		Where("? = ?", bun.Ident("id"), filter.ID).
		Exec(ctx); err != nil {
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

// NOTE:
// all of the below unsafe cast functions
// are only possible because HeaderFilterAllow{},
// HeaderFilterBlock{}, HeaderFilter{} while
// different types in source, have exactly the
// same size and layout in memory. the unsafe
// cast simply changes the type associated with
// that block of memory.

func toAllowFilter(filter *gtsmodel.HeaderFilter) *gtsmodel.HeaderFilterAllow {
	return (*gtsmodel.HeaderFilterAllow)(unsafe.Pointer(filter))
}

func toBlockFilter(filter *gtsmodel.HeaderFilter) *gtsmodel.HeaderFilterBlock {
	return (*gtsmodel.HeaderFilterBlock)(unsafe.Pointer(filter))
}

func fromAllowFilter(filter *gtsmodel.HeaderFilterAllow) *gtsmodel.HeaderFilter {
	return (*gtsmodel.HeaderFilter)(unsafe.Pointer(filter))
}

func fromBlockFilter(filter *gtsmodel.HeaderFilterBlock) *gtsmodel.HeaderFilter {
	return (*gtsmodel.HeaderFilter)(unsafe.Pointer(filter))
}

func fromAllowFilters(filters []*gtsmodel.HeaderFilterAllow) []*gtsmodel.HeaderFilter {
	return *(*[]*gtsmodel.HeaderFilter)(unsafe.Pointer(&filters))
}

func fromBlockFilters(filters []*gtsmodel.HeaderFilterBlock) []*gtsmodel.HeaderFilter {
	return *(*[]*gtsmodel.HeaderFilter)(unsafe.Pointer(&filters))
}
