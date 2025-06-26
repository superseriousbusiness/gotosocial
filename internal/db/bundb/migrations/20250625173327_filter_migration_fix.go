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
	"database/sql"
	"errors"
	"reflect"

	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20241018151036_filter_unique_fix"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250617122055_filter_improvements"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			newFilterType := reflect.TypeOf((*newmodel.Filter)(nil))

			// A SLICE!! SLICE !!!!
			// NOT JUST A STRING!!!
			//
			// silly kim
			var filterIDs []string

			// Select all filter IDs.
			if err := tx.NewSelect().
				Model((*oldmodel.Filter)(nil)).
				Column("id").
				Scan(ctx, &filterIDs); err != nil && !errors.Is(err, sql.ErrNoRows) {
				return gtserror.Newf("error selecting filter ids: %w", err)
			}

			for _, data := range []struct {
				Field string
				Model any
			}{
				{
					Field: "KeywordIDs",
					Model: (*newmodel.FilterKeyword)(nil),
				},
				{
					Field: "StatusIDs",
					Model: (*newmodel.FilterStatus)(nil),
				},
			} {
				// Get the SQL field information from bun for Filter{}.$Field.
				field, table, err := getModelField(tx, newFilterType, data.Field)
				if err != nil {
					return gtserror.Newf("error getting bun model field: %w", err)
				}

				// Check whether this part of the migration
				// has already been run before, if so skip.
				if exists, err := doesColumnExist(ctx, tx,
					table.Name,
					field.Name,
				); err != nil {
					return gtserror.Newf("error checking if column exists: %w", err)
				} else if !exists {

					// Generate bun definition for new filter table field column.
					newColDef, err := getBunColumnDef(tx, newFilterType, data.Field)
					if err != nil {
						return gtserror.Newf("error getting bun column def: %w", err)
					}

					// Add new column type to table.
					if _, err := tx.NewAddColumn().
						Model((*oldmodel.Filter)(nil)).
						ColumnExpr(newColDef).
						Exec(ctx); err != nil {
						return gtserror.Newf("error adding filter.%s column: %w", data.Field, err)
					}
				}

				// Get column name.
				col := field.Name

				var relatedIDs []string
				for _, filterID := range filterIDs {
					// Reset related IDs.
					clear(relatedIDs)
					relatedIDs = relatedIDs[:0]

					// Select $Model IDs that
					// are attached to filterID.
					if err := tx.NewSelect().
						Model(data.Model).
						Column("id").
						Where("? = ?", bun.Ident("filter_id"), filterID).
						Scan(ctx, &relatedIDs); err != nil {
						return gtserror.Newf("error selecting %T ids: %w", data.Model, err)
					}

					// Convert related IDs to bun array
					// type for serialization in query.
					arrIDs := bunArrayType(tx, relatedIDs)

					// Now update the relevant filter
					// row to contain these related IDs.
					if _, err := tx.NewUpdate().
						Model((*newmodel.Filter)(nil)).
						Where("? = ?", bun.Ident("id"), filterID).
						Set("? = ?", bun.Ident(col), arrIDs).
						Exec(ctx); err != nil {
						return gtserror.Newf("error updating filters.%s ids: %w", col, err)
					}
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
