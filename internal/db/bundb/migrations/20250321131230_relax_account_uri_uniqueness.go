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
	"fmt"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	new_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250321131230_relax_account_uri_uniqueness/new"
	old_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250321131230_relax_account_uri_uniqueness/old"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

func init() {
	up := func(ctx context.Context, bdb *bun.DB) error {
		log.Info(ctx, "converting accounts to new model; this may take a while, please don't interrupt!")

		return bdb.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			var (
				// We have to use different
				// syntax for this query
				// depending on dialect.
				dbDialect = tx.Dialect().Name()

				// ID for paging.
				maxID string

				// Batch size for
				// selecting + updating.
				batchsz = 100

				// Number of accounts
				// updated so far.
				updated int

				// We need to know our own host
				// for updating instance account.
				host = config.GetHost()
			)

			// Create the new accounts table.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_accounts").
				Model(&new_gtsmodel.Account{}).
				Exec(ctx); err != nil {
				return err
			}

			// Count number of accounts
			// we need to update.
			total, err := tx.
				NewSelect().
				Table("accounts").
				Count(ctx)
			if err != nil && !errors.Is(err, db.ErrNoEntries) {
				return err
			}

			// Create a subquery for
			// Postgres to reuse.
			var orderQPG *bun.RawQuery
			if dbDialect == dialect.PG {
				orderQPG = tx.NewRaw(
					"(COALESCE(?, ?) || ? || ?) COLLATE ?",
					bun.Ident("domain"), "",
					"/@",
					bun.Ident("username"),
					bun.Ident("C"),
				)
			}

			var orderQSqlite *bun.RawQuery
			if dbDialect == dialect.SQLite {
				orderQSqlite = tx.NewRaw(
					"(COALESCE(?, ?) || ? || ?)",
					bun.Ident("domain"), "",
					"/@",
					bun.Ident("username"),
				)
			}

			for {
				// Batch of old model account IDs to select.
				oldAccountIDs := make([]string, 0, batchsz)

				// Start building IDs query.
				idsQ := tx.
					NewSelect().
					Table("accounts").
					Column("id").
					Limit(batchsz)

				if dbDialect == dialect.SQLite {
					// For SQLite we can just select
					// our indexed expression once
					// as a column alias.
					idsQ = idsQ.
						ColumnExpr(
							"(COALESCE(?, ?) || ? || ?) AS ?",
							bun.Ident("domain"), "",
							"/@",
							bun.Ident("username"),
							bun.Ident("domain_username"),
						)
				}

				// Return only accounts with `[domain]/@[username]`
				// later in the alphabet (a-z) than provided maxID.
				if maxID != "" {
					if dbDialect == dialect.SQLite {
						idsQ = idsQ.Where("? > ?", bun.Ident("domain_username"), maxID)
					} else {
						idsQ = idsQ.Where("? > ?", orderQPG, maxID)
					}
				}

				// Page down.
				// It's counterintuitive because it
				// says ASC in the query, but we're
				// going forwards in the alphabet,
				// and z > a in a string comparison.
				if dbDialect == dialect.SQLite {
					idsQ = idsQ.OrderExpr("? ASC", bun.Ident("domain_username"))
				} else {
					idsQ = idsQ.OrderExpr("? ASC", orderQPG)
				}

				// Select this batch, providing a
				// slice to throw away username_domain.
				err := idsQ.Scan(ctx, &oldAccountIDs, new([]string))
				if err != nil {
					return err
				}

				l := len(oldAccountIDs)
				if len(oldAccountIDs) == 0 {
					// Nothing left
					// to update.
					break
				}

				// Get ready to select old accounts by their IDs.
				oldAccounts := make([]*old_gtsmodel.Account, 0, l)
				batchQ := tx.
					NewSelect().
					Model(&oldAccounts).
					Where("? IN (?)", bun.Ident("id"), bun.In(oldAccountIDs))

				// Order batch by usernameDomain
				// to ensure paging consistent.
				if dbDialect == dialect.SQLite {
					batchQ = batchQ.OrderExpr("? ASC", orderQSqlite)
				} else {
					batchQ = batchQ.OrderExpr("? ASC", orderQPG)
				}

				// Select old accounts.
				if err := batchQ.Scan(ctx); err != nil {
					return err
				}

				// Convert old accounts into new accounts.
				newAccounts := make([]*new_gtsmodel.Account, 0, l)
				for _, oldAccount := range oldAccounts {

					var actorType new_gtsmodel.AccountActorType
					switch {

					case oldAccount.Domain != "":
						// Not our account, just parse new actor type.
						actorType = new_gtsmodel.ParseAccountActorType(oldAccount.ActorType)

					case oldAccount.Username == host:
						// This is our instance account, override actor
						// type to Service, as previously it was just person.
						actorType = new_gtsmodel.AccountActorTypeService

					default:
						// Not our instance account. Use old
						// *Bot flag to determine actor type.
						if util.PtrOrZero(oldAccount.Bot) {
							// It's a bot.
							actorType = new_gtsmodel.AccountActorTypeApplication
						} else {
							// Just normal men, just innocent men.
							actorType = new_gtsmodel.AccountActorTypePerson
						}
					}

					if actorType == new_gtsmodel.AccountActorTypeUnknown {
						// This should not really happen, but it if does
						// just warn + set to person rather than failing.
						log.Warnf(ctx,
							"account %s actor type %s was not a recognized actor type, falling back to Person",
							oldAccount.ID, oldAccount.ActorType,
						)
						actorType = new_gtsmodel.AccountActorTypePerson
					}

					newAccount := &new_gtsmodel.Account{
						ID:                      oldAccount.ID,
						CreatedAt:               oldAccount.CreatedAt,
						UpdatedAt:               oldAccount.UpdatedAt,
						FetchedAt:               oldAccount.FetchedAt,
						Username:                oldAccount.Username,
						Domain:                  oldAccount.Domain,
						AvatarMediaAttachmentID: oldAccount.AvatarMediaAttachmentID,
						AvatarRemoteURL:         oldAccount.AvatarRemoteURL,
						HeaderMediaAttachmentID: oldAccount.HeaderMediaAttachmentID,
						HeaderRemoteURL:         oldAccount.HeaderRemoteURL,
						DisplayName:             oldAccount.DisplayName,
						EmojiIDs:                oldAccount.EmojiIDs,
						Fields:                  oldAccount.Fields,
						FieldsRaw:               oldAccount.FieldsRaw,
						Note:                    oldAccount.Note,
						NoteRaw:                 oldAccount.NoteRaw,
						AlsoKnownAsURIs:         oldAccount.AlsoKnownAsURIs,
						MovedToURI:              oldAccount.MovedToURI,
						MoveID:                  oldAccount.MoveID,
						Locked:                  oldAccount.Locked,
						Discoverable:            oldAccount.Discoverable,
						URI:                     oldAccount.URI,
						URL:                     oldAccount.URL,
						InboxURI:                oldAccount.InboxURI,
						SharedInboxURI:          oldAccount.SharedInboxURI,
						OutboxURI:               oldAccount.OutboxURI,
						FollowingURI:            oldAccount.FollowingURI,
						FollowersURI:            oldAccount.FollowersURI,
						FeaturedCollectionURI:   oldAccount.FeaturedCollectionURI,
						ActorType:               actorType,
						PrivateKey:              oldAccount.PrivateKey,
						PublicKey:               oldAccount.PublicKey,
						PublicKeyURI:            oldAccount.PublicKeyURI,
						PublicKeyExpiresAt:      oldAccount.PublicKeyExpiresAt,
						SensitizedAt:            oldAccount.SensitizedAt,
						SilencedAt:              oldAccount.SilencedAt,
						SuspendedAt:             oldAccount.SuspendedAt,
						SuspensionOrigin:        oldAccount.SuspensionOrigin,
					}

					newAccounts = append(newAccounts, newAccount)
				}

				// Insert this batch of accounts.
				res, err := tx.
					NewInsert().
					Model(&newAccounts).
					Returning("").
					Exec(ctx)
				if err != nil {
					return err
				}

				rowsAffected, err := res.RowsAffected()
				if err != nil {
					return err
				}

				// Add to updated count.
				updated += int(rowsAffected)
				if updated == total {
					// Done.
					break
				}

				// Set next page.
				fromAcct := oldAccounts[l-1]
				maxID = fromAcct.Domain + "/@" + fromAcct.Username

				// Log helpful message to admin.
				log.Infof(ctx,
					"migrated %d of %d accounts (next page will be from %s)",
					updated, total, maxID,
				)
			}

			if total != int(updated) {
				// Return error here in order to rollback the whole transaction.
				return fmt.Errorf("total=%d does not match updated=%d", total, updated)
			}

			log.Infof(ctx, "finished migrating %d accounts", total)

			// Drop the old table.
			log.Info(ctx, "dropping old accounts table")
			if _, err := tx.
				NewDropTable().
				Table("accounts").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			log.Info(ctx, "renaming new accounts table")
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
			log.Info(ctx, "recreating indexes on new accounts table")
			for index, columns := range map[string][]string{
				"accounts_domain_idx":        {"domain"},
				"accounts_uri_idx":           {"uri"},
				"accounts_url_idx":           {"url"},
				"accounts_inbox_uri_idx":     {"inbox_uri"},
				"accounts_outbox_uri_idx":    {"outbox_uri"},
				"accounts_followers_uri_idx": {"followers_uri"},
				"accounts_following_uri_idx": {"following_uri"},
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

			if dbDialect == dialect.PG {
				log.Info(ctx, "moving postgres constraints from old table to new table")

				type spec struct {
					old     string
					new     string
					columns []string
				}

				// Rename uniqueness constraints from
				// "new_accounts_*" to "accounts_*".
				for _, spec := range []spec{
					{
						old:     "new_accounts_pkey",
						new:     "accounts_pkey",
						columns: []string{"id"},
					},
					{
						old:     "new_accounts_uri_key",
						new:     "accounts_uri_key",
						columns: []string{"uri"},
					},
					{
						old:     "new_accounts_public_key_uri_key",
						new:     "accounts_public_key_uri_key",
						columns: []string{"public_key_uri"},
					},
				} {
					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
						bun.Ident("public.accounts"),
						bun.Safe(spec.old),
					); err != nil {
						return err
					}

					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? ADD CONSTRAINT ? UNIQUE(?)",
						bun.Ident("public.accounts"),
						bun.Safe(spec.new),
						bun.Safe(strings.Join(spec.columns, ",")),
					); err != nil {
						return err
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
