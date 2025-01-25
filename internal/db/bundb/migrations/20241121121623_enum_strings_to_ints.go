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
			log.Info(ctx, "dropping old visibility indexes...")
			for _, index := range visIndices {
				log.Info(ctx, "dropping old index %s...", index.name)
				if _, err := tx.NewDropIndex().
					Index(index.name).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Get the mapping of old enum string values to new integer values.
			visibilityMapping := visibilityEnumMapping[old_gtsmodel.Visibility]()

			// Convert all visibility tables.
			for _, table := range visTables {
				if err := convertEnums(ctx, tx, table.Table, table.Column,
					visibilityMapping, table.Default); err != nil {
					return err
				}
			}

			// Recreate the visibility indices.
			log.Info(ctx, "creating new visibility indexes...")
			for _, index := range visIndices {
				log.Info(ctx, "creating new index %s...", index.name)
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

			// Get the mapping of old enum string values to the new integer value types.
			notificationMapping := notificationEnumMapping[old_gtsmodel.NotificationType]()

			// Migrate over old notifications table column over to new column type.
			if err := convertEnums(ctx, tx, "notifications", "notification_type", //nolint:revive
				notificationMapping, nil); err != nil {
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

// visibilityEnumMapping maps old Visibility enum values to their newer integer type.
func visibilityEnumMapping[T ~string]() map[T]new_gtsmodel.Visibility {
	return map[T]new_gtsmodel.Visibility{
		T(old_gtsmodel.VisibilityNone):          new_gtsmodel.VisibilityNone,
		T(old_gtsmodel.VisibilityPublic):        new_gtsmodel.VisibilityPublic,
		T(old_gtsmodel.VisibilityUnlocked):      new_gtsmodel.VisibilityUnlocked,
		T(old_gtsmodel.VisibilityFollowersOnly): new_gtsmodel.VisibilityFollowersOnly,
		T(old_gtsmodel.VisibilityMutualsOnly):   new_gtsmodel.VisibilityMutualsOnly,
		T(old_gtsmodel.VisibilityDirect):        new_gtsmodel.VisibilityDirect,
	}
}

// notificationEnumMapping maps old NotificationType enum values to their newer integer type.
func notificationEnumMapping[T ~string]() map[T]new_gtsmodel.NotificationType {
	return map[T]new_gtsmodel.NotificationType{
		T(old_gtsmodel.NotificationFollow):        new_gtsmodel.NotificationFollow,
		T(old_gtsmodel.NotificationFollowRequest): new_gtsmodel.NotificationFollowRequest,
		T(old_gtsmodel.NotificationMention):       new_gtsmodel.NotificationMention,
		T(old_gtsmodel.NotificationReblog):        new_gtsmodel.NotificationReblog,
		T(old_gtsmodel.NotificationFave):          new_gtsmodel.NotificationFavourite,
		T(old_gtsmodel.NotificationPoll):          new_gtsmodel.NotificationPoll,
		T(old_gtsmodel.NotificationStatus):        new_gtsmodel.NotificationStatus,
		T(old_gtsmodel.NotificationSignup):        new_gtsmodel.NotificationAdminSignup,
		T(old_gtsmodel.NotificationPendingFave):   new_gtsmodel.NotificationPendingFave,
		T(old_gtsmodel.NotificationPendingReply):  new_gtsmodel.NotificationPendingReply,
		T(old_gtsmodel.NotificationPendingReblog): new_gtsmodel.NotificationPendingReblog,
	}
}
