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

	"github.com/superseriousbusiness/gotosocial/internal/log"

	oldmodel "github.com/superseriousbusiness/gotosocial/internal/db/bundb/migrations/20240620074530_interaction_policy"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		log.Info(ctx, "migrating statuses and account settings to interaction policy model, please wait...")
		log.Warn(ctx, "**WITH A LARGE DATABASE / LOWER SPEC MACHINE, THIS MIGRATION MAY TAKE A VERY LONG TIME (an hour or even longer); DO NOT INTERRUPT IT!**")
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			// Add new columns for interaction
			// policies + related fields.
			type spec struct {
				table      string
				column     string
				columnType string
				extra      string
			}
			for _, spec := range []spec{
				// Statuses.
				{
					table:      "statuses",
					column:     "interaction_policy",
					columnType: "JSONB",
					extra:      "",
				},
				{
					table:      "statuses",
					column:     "pending_approval",
					columnType: "BOOLEAN",
					extra:      "NOT NULL DEFAULT false",
				},
				{
					table:      "statuses",
					column:     "approved_by_uri",
					columnType: "varchar",
					extra:      "",
				},

				// Status faves.
				{
					table:      "status_faves",
					column:     "pending_approval",
					columnType: "BOOLEAN",
					extra:      "NOT NULL DEFAULT false",
				},
				{
					table:      "status_faves",
					column:     "approved_by_uri",
					columnType: "varchar",
					extra:      "",
				},

				// Columns that must be added to the
				// `account_settings` table to populate
				// default interaction policies for
				// different status visibilities.
				{
					table:      "account_settings",
					column:     "interaction_policy_direct",
					columnType: "JSONB",
					extra:      "",
				},
				{
					table:      "account_settings",
					column:     "interaction_policy_mutuals_only",
					columnType: "JSONB",
					extra:      "",
				},
				{
					table:      "account_settings",
					column:     "interaction_policy_followers_only",
					columnType: "JSONB",
					extra:      "",
				},
				{
					table:      "account_settings",
					column:     "interaction_policy_unlocked",
					columnType: "JSONB",
					extra:      "",
				},
				{
					table:      "account_settings",
					column:     "interaction_policy_public",
					columnType: "JSONB",
					extra:      "",
				},
			} {
				exists, err := doesColumnExist(ctx, tx,
					spec.table, spec.column,
				)
				if err != nil {
					// Real error.
					return err
				} else if exists {
					// Already created.
					continue
				}

				args := []any{
					bun.Ident(spec.table),
					bun.Ident(spec.column),
					bun.Safe(spec.columnType),
				}

				qStr := "ALTER TABLE ? ADD COLUMN ? ?"
				if spec.extra != "" {
					qStr += " ?"
					args = append(args, bun.Safe(spec.extra))
				}

				log.Infof(ctx, "adding column '%s' to '%s'...", spec.column, spec.table)
				if _, err := tx.ExecContext(ctx, qStr, args...); err != nil {
					return err
				}
			}

			// Select each locally-created status
			// with non-default old flags set.
			oldStatuses := []oldmodel.Status{}

			log.Info(ctx, "migrating existing statuses to new visibility model...")
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
					// Only author can like.
					policy.CanLike = gtsmodel.PolicyRules{
						Always: gtsmodel.PolicyValues{
							gtsmodel.PolicyValueAuthor,
						},
						WithApproval: make(gtsmodel.PolicyValues, 0),
					}
				}

				if !*oldStatus.Replyable {
					// Only author + mentioned can Reply.
					policy.CanReply = gtsmodel.PolicyRules{
						Always: gtsmodel.PolicyValues{
							gtsmodel.PolicyValueAuthor,
							gtsmodel.PolicyValueMentioned,
						},
						WithApproval: make(gtsmodel.PolicyValues, 0),
					}
				}

				if !*oldStatus.Boostable {
					// Only author can Announce.
					policy.CanAnnounce = gtsmodel.PolicyRules{
						Always: gtsmodel.PolicyValues{
							gtsmodel.PolicyValueAuthor,
						},
						WithApproval: make(gtsmodel.PolicyValues, 0),
					}
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
				log.Infof(ctx, "dropping now-unused status column '%s'; this may take a while if you have lots of statuses in your database...", column)
				if _, err := tx.
					NewDropColumn().
					Table("statuses").
					Column(column).
					Exec(ctx); err != nil {
					return err
				}
			}

			// Add new indexes.
			log.Info(ctx, "adding new index 'statuses_pending_approval_idx' to 'statuses'...")
			if _, err := tx.
				NewCreateIndex().
				Table("statuses").
				Index("statuses_pending_approval_idx").
				Column("pending_approval").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			log.Info(ctx, "adding new index 'status_faves_pending_approval_idx' to 'status_faves'...")
			if _, err := tx.
				NewCreateIndex().
				Table("status_faves").
				Index("status_faves_pending_approval_idx").
				Column("pending_approval").
				IfNotExists().
				Exec(ctx); err != nil {
				return err
			}

			log.Info(ctx, "committing transaction, almost done...")
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
