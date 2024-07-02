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
	"strings"

	"github.com/superseriousbusiness/gotosocial/internal/log"

	oldmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20240620074530_interaction_policy"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(ctx, "migrating statuses and account settings to interaction policy model, please wait...")

		// Add interaction_policy
		// column to statuses table.
		_, err := db.ExecContext(ctx,
			"ALTER TABLE ? ADD COLUMN ? JSONB",
			bun.Ident("statuses"),
			bun.Ident("interaction_policy"),
		)
		if err != nil {
			e := err.Error()
			if !(strings.Contains(e, "already exists") ||
				strings.Contains(e, "duplicate column name") ||
				strings.Contains(e, "SQLSTATE 42701")) {
				return err
			}
		}

		// Add pending_approval and approved_by_uri
		// columns to statuses and faves tables.
		type spec struct {
			table      string
			column     string
			columnType string
			defaultVal string
		}
		for _, spec := range []spec{
			{
				table:      "statuses",
				column:     "pending_approval",
				columnType: "BOOLEAN",
				defaultVal: "DEFAULT false",
			},
			{
				table:      "status_faves",
				column:     "pending_approval",
				columnType: "BOOLEAN",
				defaultVal: "DEFAULT false",
			},
			{
				table:      "statuses",
				column:     "approved_by_uri",
				columnType: "varchar",
				defaultVal: "",
			},
			{
				table:      "status_faves",
				column:     "approved_by_uri",
				columnType: "varchar",
				defaultVal: "",
			},
		} {
			_, err := db.ExecContext(ctx,
				"ALTER TABLE ? ADD COLUMN ? ? ?",
				bun.Ident(spec.table),
				bun.Ident(spec.column),
				bun.Safe(spec.columnType),
				bun.Safe(spec.defaultVal),
			)
			if err != nil {
				e := err.Error()
				if !(strings.Contains(e, "already exists") ||
					strings.Contains(e, "duplicate column name") ||
					strings.Contains(e, "SQLSTATE 42701")) {
					return err
				}
			}
		}

		// Columns that must be added to the
		// `account_settings` table to populate
		// default interaction policies for
		// different status visibilities.
		newSettingsColumns := []string{
			"interaction_policy_direct",
			"interaction_policy_mutuals_only",
			"interaction_policy_followers_only",
			"interaction_policy_unlocked",
			"interaction_policy_public",
		}

		for _, column := range newSettingsColumns {
			_, err := db.ExecContext(ctx,
				"ALTER TABLE ? ADD COLUMN ? JSONB",
				bun.Ident("account_settings"),
				bun.Ident(column),
			)
			if err != nil {
				e := err.Error()
				if !(strings.Contains(e, "already exists") ||
					strings.Contains(e, "duplicate column name") ||
					strings.Contains(e, "SQLSTATE 42701")) {
					return err
				}
			}
		}

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Select each locally-created status
			// with non-default old flags set.
			oldStatuses := []oldmodel.Status{}

			if err := tx.
				NewSelect().
				Model(&oldStatuses).
				Column("id", "likeable", "replyable", "boostable", "visibility").
				Where("? = ?", bun.Ident("local"), true).
				WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
					return sq.
						Where("? = ?", bun.Ident("likeable"), false).
						WhereOr("? = ?", bun.Ident("replyable"), false).
						WhereOr("? = ?", bun.Ident("boostable"), false)
				}).
				Scan(ctx); err != nil {
				return err
			}

			// For each status found in this way, update
			// to new version of interaction policy.
			for _, oldStatus := range oldStatuses {
				// Start with default policy for this visibility.
				v := gtsmodel.Visibility(oldStatus.Visibility)
				policy := gtsmodel.DefaultInteractionPolicyFor(v)

				if !*oldStatus.Likeable {
					// Nobody can Like.
					policy.CanLike = gtsmodel.PolicyRules{}
				}

				if !*oldStatus.Replyable {
					// Nobody can Reply.
					policy.CanReply = gtsmodel.PolicyRules{}
				}

				if !*oldStatus.Boostable {
					// Nobody can Announce.
					policy.CanAnnounce = gtsmodel.PolicyRules{}
				}

				// Update status with the new interaction policy.
				newStatus := &gtsmodel.Status{
					ID:                oldStatus.ID,
					InteractionPolicy: policy,
				}
				if _, err := tx.
					NewUpdate().
					Model(newStatus).
					Column("interaction_policy").
					Where("? = ?", bun.Ident("id"), newStatus.ID).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Drop now unused columns from statuses table.
			oldColumns := []string{
				"likeable",
				"replyable",
				"boostable",
			}
			for _, column := range oldColumns {
				if _, err := tx.
					NewDropColumn().
					Table("statuses").
					Column(column).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Add new indexes.
			if _, err := tx.
				NewCreateIndex().
				Table("statuses").
				Index("statuses_pending_approval_idx").
				Column("pending_approval").
				IfNotExists().
				Exec(ctx); err != nil {
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
