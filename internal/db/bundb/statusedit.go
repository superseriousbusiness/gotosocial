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

type statusEditDB struct {
	db    *bun.DB
	state *state.State
}

func (s *statusEditDB) GetStatusEditByID(ctx context.Context, id string) (*gtsmodel.StatusEdit, error) {
	// Fetch edit from database cache with loader callback.
	edit, err := s.state.Caches.DB.StatusEdit.LoadOne("ID",
		func() (*gtsmodel.StatusEdit, error) {
			var edit gtsmodel.StatusEdit

			// Not cached, load edit
			// from database by its ID.
			if err := s.db.NewSelect().
				Model(&edit).
				Where("? = ?", bun.Ident("id"), id).
				Scan(ctx); err != nil {
				return nil, err
			}

			return &edit, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return edit, nil
	}

	// Further populate the edit fields where applicable.
	if err := s.PopulateStatusEdit(ctx, edit); err != nil {
		return nil, err
	}

	return edit, nil
}

func (s *statusEditDB) GetStatusEditsByIDs(ctx context.Context, ids []string) ([]*gtsmodel.StatusEdit, error) {
	// Load status edits for IDs via cache loader callbacks.
	edits, err := s.state.Caches.DB.StatusEdit.LoadIDs("ID",
		ids,
		func(uncached []string) ([]*gtsmodel.StatusEdit, error) {
			// Preallocate expected length of uncached edits.
			edits := make([]*gtsmodel.StatusEdit, 0, len(uncached))

			// Perform database query scanning
			// the remaining (uncached) edit IDs.
			if err := s.db.NewSelect().
				Model(&edits).
				Where("? IN (?)", bun.Ident("id"), bun.In(uncached)).
				Scan(ctx); err != nil {
				return nil, err
			}

			return edits, nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Reorder the edits by their
	// IDs to ensure in correct order.
	getID := func(e *gtsmodel.StatusEdit) string { return e.ID }
	util.OrderBy(edits, ids, getID)

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return edits, nil
	}

	// Populate all loaded edits, removing those we fail to
	// populate (removes needing so many nil checks everywhere).
	edits = slices.DeleteFunc(edits, func(edit *gtsmodel.StatusEdit) bool {
		if err := s.PopulateStatusEdit(ctx, edit); err != nil {
			log.Errorf(ctx, "error populating edit %s: %v", edit.ID, err)
			return true
		}
		return false
	})

	return edits, nil
}

func (s *statusEditDB) PopulateStatusEdit(ctx context.Context, edit *gtsmodel.StatusEdit) error {
	var err error
	var errs gtserror.MultiError

	// For sub-models we only want
	// barebones versions of them.
	ctx = gtscontext.SetBarebones(ctx)

	if !edit.AttachmentsPopulated() {
		// Fetch all attachments for status edit's IDs.
		edit.Attachments, err = s.state.DB.GetAttachmentsByIDs(
			ctx,
			edit.AttachmentIDs,
		)
		if err != nil {
			errs.Appendf("error populating edit attachments: %w", err)
		}
	}

	return errs.Combine()
}

func (s *statusEditDB) PutStatusEdit(ctx context.Context, edit *gtsmodel.StatusEdit) error {
	return s.state.Caches.DB.StatusEdit.Store(edit, func() error {
		_, err := s.db.NewInsert().Model(edit).Exec(ctx)
		return err
	})
}

func (s *statusEditDB) DeleteStatusEdits(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Gather necessary fields from
	// deleted for cache invalidation.
	var deleted []*gtsmodel.StatusEdit
	deleted = make([]*gtsmodel.StatusEdit, 0, len(ids))

	// Delete all edits with IDs pertaining
	// to given slice, returning status IDs.
	if _, err := s.db.NewDelete().
		Model(&deleted).
		Where("? IN (?)", bun.Ident("id"), bun.In(ids)).
		Returning("?", bun.Ident("status_id")).
		Exec(ctx); err != nil &&
		!errors.Is(err, db.ErrNoEntries) {
		return err
	}

	// Invalidate all the cached status edits with IDs.
	s.state.Caches.DB.StatusEdit.InvalidateIDs("ID", ids)

	// Make sure we only end up calling
	// the invalidate hook for each status
	// once. This should just be the one,
	// but we double check to save cycles.
	m := make(map[string]struct{}, 1)
	for _, edit := range deleted {

		// Check not already called for status.
		if _, ok := m[edit.StatusID]; ok {
			continue
		}

		// Manually call status edit invalidate hook.
		s.state.Caches.OnInvalidateStatusEdit(edit)
		m[edit.StatusID] = struct{}{}
	}

	return nil
}
