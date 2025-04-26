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

	gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20230328203024_migration_fix"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		// To update not null constraint on public key, we need to migrate accounts into a new table.
		// See section 7 here: https://www.sqlite.org/lang_altertable.html

		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Create the new accounts table.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_accounts").
				Model(&gtsmodel.Account{}).
				Exec(ctx); err != nil {
				return err
			}

			// If we don't specify columns explicitly,
			// Postgres gives the following error when
			// transferring accounts to new_accounts:
			//
			//	ERROR:  column "fetched_at" is of type timestamp with time zone but expression is of type character varying at character 35
			//	HINT:  You will need to rewrite or cast the expression.
			//
			// Rather than do funky casting to fix this,
			// it's simpler to just specify all columns.
			columns := []string{
				"id",
				"created_at",
				"updated_at",
				"fetched_at",
				"username",
				"domain",
				"avatar_media_attachment_id",
				"avatar_remote_url",
				"header_media_attachment_id",
				"header_remote_url",
				"display_name",
				"emojis",
				"fields",
				"note",
				"note_raw",
				"memorial",
				"also_known_as",
				"moved_to_account_id",
				"bot",
				"reason",
				"locked",
				"discoverable",
				"privacy",
				"sensitive",
				"language",
				"status_content_type",
				"custom_css",
				"uri",
				"url",
				"inbox_uri",
				"shared_inbox_uri",
				"outbox_uri",
				"following_uri",
				"followers_uri",
				"featured_collection_uri",
				"actor_type",
				"private_key",
				"public_key",
				"public_key_uri",
				"sensitized_at",
				"silenced_at",
				"suspended_at",
				"hide_collections",
				"suspension_origin",
				"enable_rss",
			}

			// Copy all accounts to the new table.
			if _, err := tx.
				NewInsert().
				Table("new_accounts").
				Table("accounts").
				Column(columns...).
				Exec(ctx); err != nil {
				return err
			}

			// Drop the old table.
			if _, err := tx.
				NewDropTable().
				Table("accounts").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			if _, err := tx.
				ExecContext(
					ctx,
					"ALTER TABLE ? RENAME TO ?",
					bun.Ident("new_accounts"),
					bun.Ident("accounts"),
				); err != nil {
				return err
			}

			// Add all account indexes to the new table.
			for index, columns := range map[string][]string{
				// Standard indices.
				"accounts_id_idx":              {"id"},
				"accounts_suspended_at_idx":    {"suspended_at"},
				"accounts_domain_idx":          {"domain"},
				"accounts_username_domain_idx": {"username", "domain"},
				// URI indices.
				"accounts_uri_idx":            {"uri"},
				"accounts_url_idx":            {"url"},
				"accounts_inbox_uri_idx":      {"inbox_uri"},
				"accounts_outbox_uri_idx":     {"outbox_uri"},
				"accounts_followers_uri_idx":  {"followers_uri"},
				"accounts_following_uri_idx":  {"following_uri"},
				"accounts_public_key_uri_idx": {"public_key_uri"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("accounts").
					Index(index).
					Column(columns...).
					Exec(ctx); err != nil {
					return err
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
