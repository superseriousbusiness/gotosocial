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
	"errors"

	old_gtsmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20241121121623_enum_strings_to_ints"
	new_gtsmodel "github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/util"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Tables with visibility types.
			var visTables = []struct {
				Table   string
				Column  string
				Default *new_gtsmodel.Visibility
			}{
				{Table: "statuses", Column: "visibility"},
				{Table: "sin_bin_statuses", Column: "visibility"},
				{Table: "account_settings", Column: "privacy", Default: util.Ptr(new_gtsmodel.VisibilityDefault)},
				{Table: "account_settings", Column: "web_visibility", Default: util.Ptr(new_gtsmodel.VisibilityDefault)},
			}

			// Visibility type indices.
			var visIndices = []struct {
				name  string
				cols  []string
				order string
			}{
				{
					name:  "statuses_visibility_idx",
					cols:  []string{"visibility"},
					order: "",
				},
				{
					name:  "statuses_profile_web_view_idx",
					cols:  []string{"account_id", "visibility"},
					order: "id DESC",
				},
				{
					name:  "statuses_public_timeline_idx",
					cols:  []string{"visibility"},
					order: "id DESC",
				},
			}

			// Before making changes to the visibility col
			// we must drop all indices that rely on it.
			for _, index := range visIndices {
				if _, err := tx.NewDropIndex().
					Index(index.name).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Convert all visibility tables.
			for _, table := range visTables {
				if err := convertEnums(ctx, tx, table.Table, table.Column,
					map[old_gtsmodel.Visibility]new_gtsmodel.Visibility{
						old_gtsmodel.VisibilityNone:          new_gtsmodel.VisibilityNone,
						old_gtsmodel.VisibilityPublic:        new_gtsmodel.VisibilityPublic,
						old_gtsmodel.VisibilityUnlocked:      new_gtsmodel.VisibilityUnlocked,
						old_gtsmodel.VisibilityFollowersOnly: new_gtsmodel.VisibilityFollowersOnly,
						old_gtsmodel.VisibilityMutualsOnly:   new_gtsmodel.VisibilityMutualsOnly,
						old_gtsmodel.VisibilityDirect:        new_gtsmodel.VisibilityDirect,
					}, table.Default); err != nil {
					return err
				}
			}

			// Recreate the visibility indices.
			for _, index := range visIndices {
				q := tx.NewCreateIndex().
					Table("statuses").
					Index(index.name).
					Column(index.cols...)
				if index.order != "" {
					q = q.ColumnExpr(index.order)
				}
				if _, err := q.Exec(ctx); err != nil {
					return err
				}
			}

			// Migrate over old notifications table column over to new column type.
			if err := convertEnums(ctx, tx, "notifications", "notification_type", //nolint:revive
				map[old_gtsmodel.NotificationType]new_gtsmodel.NotificationType{
					old_gtsmodel.NotificationFollow:        new_gtsmodel.NotificationFollow,
					old_gtsmodel.NotificationFollowRequest: new_gtsmodel.NotificationFollowRequest,
					old_gtsmodel.NotificationMention:       new_gtsmodel.NotificationMention,
					old_gtsmodel.NotificationReblog:        new_gtsmodel.NotificationReblog,
					old_gtsmodel.NotificationFave:          new_gtsmodel.NotificationFave,
					old_gtsmodel.NotificationPoll:          new_gtsmodel.NotificationPoll,
					old_gtsmodel.NotificationStatus:        new_gtsmodel.NotificationStatus,
					old_gtsmodel.NotificationSignup:        new_gtsmodel.NotificationSignup,
					old_gtsmodel.NotificationPendingFave:   new_gtsmodel.NotificationPendingFave,
					old_gtsmodel.NotificationPendingReply:  new_gtsmodel.NotificationPendingReply,
					old_gtsmodel.NotificationPendingReblog: new_gtsmodel.NotificationPendingReblog,
				}, nil); err != nil {
				return err
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

// convertEnums performs a transaction that converts
// a table's column of our old-style enums (strings) to
// more performant and space-saving integer types.
func convertEnums[OldType ~string, NewType ~int](
	ctx context.Context,
	tx bun.Tx,
	table string,
	column string,
	mapping map[OldType]NewType,
	defaultValue *NewType,
) error {
	if len(mapping) == 0 {
		return errors.New("empty mapping")
	}

	// Generate new column name.
	newColumn := column + "_new"

	log.Infof(ctx, "converting %s.%s enums; "+
		"this may take a while, please don't interrupt!",
		table, column,
	)

	// Ensure a default value.
	if defaultValue == nil {
		var zero NewType
		defaultValue = &zero
	}

	// Add new column to database.
	if _, err := tx.NewAddColumn().
		Table(table).
		ColumnExpr("? INTEGER NOT NULL DEFAULT ?",
			bun.Ident(newColumn),
			*defaultValue).
		Exec(ctx); err != nil {
		return err
	}

	// Get a count of all in table.
	total, err := tx.NewSelect().
		Table(table).
		Count(ctx)
	if err != nil {
		return err
	}

	var updated int
	for old, new := range mapping {

		// Update old to new values.
		res, err := tx.NewUpdate().
			Table(table).
			Where("? = ?", bun.Ident(column), old).
			Set("? = ?", bun.Ident(newColumn), new).
			Exec(ctx)
		if err != nil {
			return err
		}

		// Count number items updated.
		n, _ := res.RowsAffected()
		updated += int(n)
	}

	// Check total updated.
	if total != updated {
		log.Warnf(ctx, "total=%d does not match updated=%d", total, updated)
	}

	// Drop the old column from table.
	if _, err := tx.NewDropColumn().
		Table(table).
		ColumnExpr("?", bun.Ident(column)).
		Exec(ctx); err != nil {
		return err
	}

	// Rename new to old name.
	if _, err := tx.NewRaw(
		"ALTER TABLE ? RENAME COLUMN ? TO ?",
		bun.Ident(table),
		bun.Ident(newColumn),
		bun.Ident(column),
	).Exec(ctx); err != nil {
		return err
	}

	return nil
}
