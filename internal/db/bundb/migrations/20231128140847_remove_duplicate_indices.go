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

	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			type spec struct {
				old     string
				new     string
				table   string
				columns []string
			}

			if db.Dialect().Name() == dialect.PG {
				log.Info(ctx, "renaming misnamed postgres constraints; this may take some time, please be patient and don't interrupt this!")

				// Some constraints got kept around
				// in weird versions due to migration
				// issues; rename these for consistency
				// (this will also drop and recreate
				// indexes supporting the constraints).
				for _, spec := range []spec{
					{
						old:     "new_accounts_pkey",
						new:     "accounts_pkey",
						table:   "public.accounts",
						columns: []string{"id"},
					},
					{
						old:     "new_accounts_uri_key",
						new:     "accounts_uri_key",
						table:   "public.accounts",
						columns: []string{"uri"},
					},
					{
						old:     "new_accounts_url_key",
						new:     "accounts_url_key",
						table:   "public.accounts",
						columns: []string{"url"},
					},
					{
						old:     "new_accounts_inbox_uri_key",
						new:     "accounts_inbox_uri_key",
						table:   "public.accounts",
						columns: []string{"inbox_uri"},
					},
					{
						old:     "new_accounts_outbox_uri_key",
						new:     "accounts_outbox_uri_key",
						table:   "public.accounts",
						columns: []string{"outbox_uri"},
					},
					{
						old:     "new_accounts_following_uri_key",
						new:     "accounts_following_uri_key",
						table:   "public.accounts",
						columns: []string{"following_uri"},
					},
					{
						old:     "new_accounts_followers_uri_key",
						new:     "accounts_followers_uri_key",
						table:   "public.accounts",
						columns: []string{"followers_uri"},
					},
					{
						old:     "new_accounts_featured_collection_uri_key",
						new:     "accounts_featured_collection_uri_key",
						table:   "public.accounts",
						columns: []string{"featured_collection_uri"},
					},
					{
						old:     "new_accounts_public_key_uri_key",
						new:     "accounts_public_key_uri_key",
						table:   "public.accounts",
						columns: []string{"public_key_uri"},
					},
					{
						old:     "new_emojis_pkey1",
						new:     "emojis_pkey",
						table:   "public.emojis",
						columns: []string{"id"},
					},
					{
						old:     "new_emojis_uri_key1",
						new:     "emojis_uri_key",
						table:   "public.emojis",
						columns: []string{"uri"},
					},
					{
						old:     "new_status_faves_pkey",
						new:     "status_faves_pkey",
						table:   "public.status_faves",
						columns: []string{"id"},
					},
					{
						old:     "new_status_faves_uri_key",
						new:     "status_faves_uri_key",
						table:   "public.status_faves",
						columns: []string{"uri"},
					},
				} {
					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
						bun.Ident(spec.table),
						bun.Safe(spec.old),
					); err != nil {
						return err
					}

					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? ADD CONSTRAINT ? UNIQUE(?)",
						bun.Ident(spec.table),
						bun.Safe(spec.new),
						bun.Safe(strings.Join(spec.columns, ",")),
					); err != nil {
						return err
					}
				}
			}

			log.Info(ctx, "removing duplicate indexes; this may take some time, please be patient and don't interrupt this!")

			// Remove all indexes which duplicate
			// or are covered by other indexes,
			// including unique constraint indexes
			// created automatically by the db.
			for _, index := range []string{
				"account_notes_account_id_target_account_id_idx",
				"accounts_username_domain_idx",
				"accounts_id_idx",
				"accounts_inbox_uri_idx",
				"accounts_outbox_uri_idx",
				"accounts_uri_idx",
				"accounts_url_idx",
				"accounts_followers_uri_idx",
				"accounts_following_uri_idx",
				"accounts_public_key_uri_idx",
				"account_actions_id_idx",
				"blocks_account_id_target_account_id_idx",
				"emojis_id_idx",
				"emojis_uri_idx",
				"instances_domain_idx",
				"list_entries_id_idx",
				"lists_id_idx",
				"markers_account_id_name_idx",
				"media_attachments_id_idx",
				"status_faves_id_idx",
				"statuses_uri_idx",
				"tags_name_idx",
				"thread_mutes_id_idx",
				"thread_mutes_thread_id_account_id_idx",
				"threads_id_idx",
				"tombstone_uri_idx",
			} {
				if _, err := tx.
					NewDropIndex().
					Index(index).
					IfExists().
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
