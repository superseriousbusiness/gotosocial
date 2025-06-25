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
	"reflect"
	"strings"

	oldmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20241018151036_filter_unique_fix"
	newmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250617122055_filter_improvements"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// Replace 'context_*' and 'action' columns with space-saving enum / bitfields.
		if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			newFilterType := reflect.TypeOf((*newmodel.Filter)(nil))

			// Generate bun definition for new filter table contexts column.
			newColDef, err := getBunColumnDef(tx, newFilterType, "Contexts")
			if err != nil {
				return gtserror.Newf("error getting bun column def: %w", err)
			}

			// Add new column type to table.
			if _, err := tx.NewAddColumn().
				Model((*oldmodel.Filter)(nil)).
				ColumnExpr(newColDef).
				Exec(ctx); err != nil {
				return gtserror.Newf("error adding filter.contexts column: %w", err)
			}

			// Generate bun definition for new filter table action column.
			newColDef, err = getBunColumnDef(tx, newFilterType, "Action")
			if err != nil {
				return gtserror.Newf("error getting bun column def: %w", err)
			}

			// For now, name it as '_new'.
			newColDef = strings.ReplaceAll(
				newColDef,
				"action",
				"action_new",
			)

			// Add new column type to table.
			if _, err := tx.NewAddColumn().
				Model((*oldmodel.Filter)(nil)).
				ColumnExpr(newColDef).
				Exec(ctx); err != nil {
				return gtserror.Newf("error adding filter.contexts column: %w", err)
			}

			var oldFilters []*oldmodel.Filter

			// Select all filters.
			if err := tx.NewSelect().
				Model(&oldFilters).
				Column("id",
					"context_home",
					"context_notifications",
					"context_public",
					"context_thread",
					"context_account",
					"action").
				Scan(ctx); err != nil {
				return gtserror.Newf("error selecting filters: %w", err)
			}

			for _, oldFilter := range oldFilters {
				var newContexts newmodel.FilterContexts
				var newAction newmodel.FilterAction

				// Convert old contexts
				// to new contexts type.
				if *oldFilter.ContextHome {
					newContexts.SetHome()
				}
				if *oldFilter.ContextNotifications {
					newContexts.SetNotifications()
				}
				if *oldFilter.ContextPublic {
					newContexts.SetPublic()
				}
				if *oldFilter.ContextThread {
					newContexts.SetThread()
				}
				if *oldFilter.ContextAccount {
					newContexts.SetAccount()
				}

				// Convert old action
				// to new action type.
				switch oldFilter.Action {
				case oldmodel.FilterActionHide:
					newAction = newmodel.FilterActionHide
				case oldmodel.FilterActionWarn:
					newAction = newmodel.FilterActionWarn
				default:
					return gtserror.Newf("invalid filter action %q for %s", oldFilter.Action, oldFilter.ID)
				}

				// Update filter row with
				// the new contexts value.
				if _, err := tx.NewUpdate().
					Model((*oldmodel.Filter)(nil)).
					Where("? = ?", bun.Ident("id"), oldFilter.ID).
					Set("? = ?", bun.Ident("contexts"), newContexts).
					Set("? = ?", bun.Ident("action_new"), newAction).
					Exec(ctx); err != nil {
					return gtserror.Newf("error updating filter.contexts: %w", err)
				}
			}

			// Drop the old updated columns.
			for _, col := range []string{
				"context_home",
				"context_notifications",
				"context_public",
				"context_thread",
				"context_account",
				"action",
			} {
				if _, err := tx.NewDropColumn().
					Model((*oldmodel.Filter)(nil)).
					Column(col).
					Exec(ctx); err != nil {
					return gtserror.Newf("error dropping filter.%s column: %w", col, err)
				}
			}

			// Rename the new action
			// column to correct name.
			if _, err := tx.NewRaw(
				"ALTER TABLE ? RENAME COLUMN ? TO ?",
				bun.Ident("filters"),
				bun.Ident("action_new"),
				bun.Ident("action"),
			).Exec(ctx); err != nil {
				return gtserror.Newf("error renaming new action column: %w", err)
			}

			return nil
		}); err != nil {
			return err
		}

		// SQLITE: force WAL checkpoint to merge writes.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		// Drop a bunch of (now, and more generally) unused columns from filter tables.
		if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			for model, indices := range map[any][]string{
				(*oldmodel.FilterKeyword)(nil): {"filter_keywords_account_id_idx"},
				(*oldmodel.FilterStatus)(nil):  {"filter_statuses_account_id_idx"},
			} {
				for _, index := range indices {
					if _, err := tx.NewDropIndex().
						Model(model).
						Index(index).
						Exec(ctx); err != nil {
						return gtserror.Newf("error dropping %s index: %w", index, err)
					}
				}
			}
			for model, cols := range map[any][]string{
				(*oldmodel.Filter)(nil):        {"created_at", "updated_at"},
				(*oldmodel.FilterKeyword)(nil): {"created_at", "updated_at", "account_id"},
				(*oldmodel.FilterStatus)(nil):  {"created_at", "updated_at", "account_id"},
			} {
				for _, col := range cols {
					if _, err := tx.NewDropColumn().
						Model(model).
						Column(col).
						Exec(ctx); err != nil {
						return gtserror.Newf("error dropping %T.%s column: %w", model, col, err)
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}

		// SQLITE: force WAL checkpoint
		// to merge writes before return.
		return doWALCheckpoint(ctx, db)
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
