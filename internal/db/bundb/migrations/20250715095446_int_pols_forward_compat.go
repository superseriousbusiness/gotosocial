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

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"

	new_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250715095446_int_pols_forward_compat/new"
	old_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250715095446_int_pols_forward_compat/old"
)

func init() {
	up := func(ctx context.Context, bdb *bun.DB) error {
		// Count number of interaction
		// requests we need to update.
		total, err := bdb.
			NewSelect().
			Table("interaction_requests").
			Count(ctx)
		if err != nil && !errors.Is(err, db.ErrNoEntries) {
			return err
		}

		log.Infof(ctx, "converting %d interaction requests to new model...", total)

		return bdb.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			var (
				// ID for paging.
				maxID string

				// Batch size for
				// selecting + updating.
				batchsz = 100

				// Number of int reqs
				// updated so far.
				updated int
			)

			// Create the new table.
			if _, err := tx.
				NewCreateTable().
				ModelTableExpr("new_interaction_requests").
				Model((*new_gtsmodel.InteractionRequest)(nil)).
				Exec(ctx); err != nil {
				return err
			}

			for {
				// Batch of old model int reqs to select.
				oldIntReqs := make([]*old_gtsmodel.InteractionRequest, 0, batchsz)

				// Start building scan query.
				scanQ := tx.
					NewSelect().
					Model(&oldIntReqs).
					OrderExpr("? DESC", bun.Ident("id")).
					Limit(batchsz)

				// Return only int reqs with ID
				// lower than the maxID (paging
				// down from newest to oldest).
				if maxID != "" {
					scanQ = scanQ.Where("? < ?", bun.Ident("id"), maxID)
				}

				// Select this batch
				err := scanQ.Scan(ctx)
				if err != nil && !errors.Is(err, db.ErrNoEntries) {
					return err
				}

				l := len(oldIntReqs)
				if len(oldIntReqs) == 0 {
					// Nothing left
					// to update.
					break
				}

				// Convert old int reqs into new ones.
				newIntReqs := make([]*new_gtsmodel.InteractionRequest, 0, l)
				for _, oldIntReq := range oldIntReqs {

					newIntReqs = append(newIntReqs, &new_gtsmodel.InteractionRequest{
						ID:                    oldIntReq.ID,
						TargetStatusID:        oldIntReq.StatusID,
						TargetAccountID:       oldIntReq.TargetAccountID,
						InteractingAccountID:  oldIntReq.InteractingAccountID,
						InteractionRequestURI: "", // This wasn't supported yet by old int reqs.
						InteractionURI:        oldIntReq.InteractionURI,
						InteractionType:       int16(oldIntReq.InteractionType),
						AcceptedAt:            oldIntReq.AcceptedAt,
						RejectedAt:            oldIntReq.RejectedAt,
						ResponseURI:           oldIntReq.URI,
					})
				}

				// Insert this batch.
				res, err := tx.
					NewInsert().
					Model(&newIntReqs).
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
				maxID = oldIntReqs[l-1].ID

				// Log helpful message to admin.
				log.Infof(ctx,
					"migrated %d of %d interaction requests",
					updated, total,
				)
			}

			if total != int(updated) {
				// Return error here in order to rollback the whole transaction.
				return fmt.Errorf("total=%d does not match updated=%d", total, updated)
			}

			log.Infof(ctx, "finished migrating %d interaction requests", total)

			// Drop the old table.
			log.Info(ctx, "dropping old interaction_requests table")
			if _, err := tx.
				NewDropTable().
				Table("interaction_requests").
				Exec(ctx); err != nil {
				return err
			}

			// Rename new table to old table.
			log.Info(ctx, "renaming new interaction requests table")
			if _, err := tx.
				ExecContext(
					ctx,
					"ALTER TABLE ? RENAME TO ?",
					bun.Ident("new_interaction_requests"),
					bun.Ident("interaction_requests"),
				); err != nil {
				return err
			}

			// Add all indexes to the new table.
			log.Info(ctx, "recreating indexes on new interaction requests table")
			for index, columns := range map[string][]string{
				"interaction_requests_target_status_id_idx":       {"target_status_id"},
				"interaction_requests_interacting_account_id_idx": {"interacting_account_id"},
				"interaction_requests_target_account_id_idx":      {"target_account_id"},
				"interaction_requests_accepted_at_idx":            {"accepted_at"},
				"interaction_requests_rejected_at_idx":            {"rejected_at"},
			} {
				if _, err := tx.
					NewCreateIndex().
					Table("interaction_requests").
					Index(index).
					Column(columns...).
					Exec(ctx); err != nil {
					return err
				}
			}

			if tx.Dialect().Name() == dialect.PG {
				log.Info(ctx, "moving postgres constraints from old table to new table")

				type spec struct {
					old     string
					new     string
					columns []string
				}

				// Rename uniqueness constraints from
				// "new_interaction_requests_*" to "interaction_requests_*".
				for _, spec := range []spec{
					{
						old:     "new_interaction_requests_pkey",
						new:     "interaction_requests_pkey",
						columns: []string{"id"},
					},
					{
						old:     "new_interaction_requests_interaction_request_uri_key",
						new:     "interaction_requests_interaction_request_uri_key",
						columns: []string{"interaction_request_uri"},
					},
					{
						old:     "new_interaction_requests_interaction_uri_key",
						new:     "interaction_requests_interaction_uri_key",
						columns: []string{"interaction_uri"},
					},
					{
						old:     "new_interaction_requests_response_uri_key",
						new:     "interaction_requests_response_uri_key",
						columns: []string{"response_uri"},
					},
				} {
					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? DROP CONSTRAINT IF EXISTS ?",
						bun.Ident("interaction_requests"),
						bun.Safe(spec.old),
					); err != nil {
						return err
					}

					if _, err := tx.ExecContext(
						ctx,
						"ALTER TABLE ? ADD CONSTRAINT ? UNIQUE(?)",
						bun.Ident("interaction_requests"),
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
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
