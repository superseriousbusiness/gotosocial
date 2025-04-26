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

package migrations

import (
	"context"
	"fmt"
	"slices"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util/xslices"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			var edits []*gtsmodel.StatusEdit

			// Select all status edits that
			// are not actually connected to
			// the status they reference.
			if err := tx.NewSelect().
				Model(&edits).
				Join("JOIN ? AS ? ON ? = ?",
					bun.Ident("statuses"),
					bun.Ident("status"),
					bun.Ident("status.id"),
					bun.Ident("status_edit.status_id"),
				).
				Where("CAST(? AS TEXT) NOT LIKE CONCAT(?, ?, ?)",
					bun.Ident("status.edits"),
					"%", bun.Ident("status_edit.id"), "%",
				).
				Column("id", "status_id", "created_at").
				Scan(ctx, &edits); err != nil {
				return fmt.Errorf("error selecting unlinked edits: %w", err)
			}

			log.Infof(ctx, "relinking %d unlinked status edits", len(edits))

			for _, edit := range edits {
				var status gtsmodel.Status

				// Select the list of edits
				// CURRENTLY attached to the
				// status that edit references.
				if err := tx.NewSelect().
					Model(&status).
					Column("edits").
					Where("? = ?",
						bun.Ident("id"),
						edit.StatusID,
					).
					Scan(ctx); err != nil {
					return fmt.Errorf("error selecting status.edits: %w", err)
				}

				// Select only the ID and creation
				// dates of all the other edits that
				// are attached to referenced status.
				if err := tx.NewSelect().
					Model(&status.Edits).
					Column("id", "created_at").
					Where("? IN (?)",
						bun.Ident("id"),
						bun.In(status.EditIDs),
					).
					Scan(ctx); err != nil {
					return fmt.Errorf("error selecting other status edits: %w", err)
				}

				editID := func(e *gtsmodel.StatusEdit) string { return e.ID }

				// Append this unlinked edit to status' list
				// of edits and then sort edits by creation.
				//
				// On tiny off-chance we end up with dupes,
				// we still deduplicate these status edits.
				status.Edits = append(status.Edits, edit)
				status.Edits = xslices.DeduplicateFunc(status.Edits, editID)
				slices.SortFunc(status.Edits, func(e1, e2 *gtsmodel.StatusEdit) int {
					const k = -1 // oldest at 0th, newest at nth
					switch c1, c2 := e1.CreatedAt, e2.CreatedAt; {
					case c1.Before(c2):
						return +k
					case c2.Before(c1):
						return -k
					default:
						return 0
					}
				})

				// Extract the IDs from edits to update the status edit IDs.
				status.EditIDs = xslices.Gather(nil, status.Edits, editID)

				// Update the relevant status
				// edit IDs column in database.
				if _, err := tx.NewUpdate().
					Model(&status).
					Column("edits").
					Where("? = ?",
						bun.Ident("id"),
						edit.StatusID,
					).
					Exec(ctx); err != nil {
					return fmt.Errorf("error updating status.edits: %w", err)
				}
			}

			return nil
		})
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			return nil
		})
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
