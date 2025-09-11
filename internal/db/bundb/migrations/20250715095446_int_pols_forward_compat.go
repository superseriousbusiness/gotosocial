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
	"net/url"
	"strings"

	"code.superseriousbusiness.org/gotosocial/internal/config"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/log"
	"code.superseriousbusiness.org/gotosocial/internal/util"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"

	new_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250715095446_int_pols_forward_compat/new"
	old_gtsmodel "code.superseriousbusiness.org/gotosocial/internal/db/bundb/migrations/20250715095446_int_pols_forward_compat/old"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		const tmpTableName = "new_interaction_requests"
		const tableName = "interaction_requests"
		var host = config.GetHost()
		var accountDomain = config.GetAccountDomain()

		// Count number of interaction
		// requests we need to update.
		total, err := db.NewSelect().
			Table(tableName).
			Count(ctx)
		if err != nil {
			return gtserror.Newf("error geting interaction requests table count: %w", err)
		}

		// Create new interaction_requests table and convert all existing into it.
		if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			log.Info(ctx, "creating new interaction_requests table")
			if _, err := tx.NewCreateTable().
				ModelTableExpr(tmpTableName).
				Model((*new_gtsmodel.InteractionRequest)(nil)).
				Exec(ctx); err != nil {
				return gtserror.Newf("error creating new interaction requests table: %w", err)
			}

			// Conversion batch size.
			const batchsz = 1000

			var maxID string
			var count int

			// Start at largest
			// possible ULID value.
			maxID = id.Highest

			// Preallocate interaction request slices to maximum possible size.
			oldRequests := make([]*old_gtsmodel.InteractionRequest, 0, batchsz)
			newRequests := make([]*new_gtsmodel.InteractionRequest, 0, batchsz)

			for {
				// Reset slices *without*
				// clearing element memory,
				// so we can reuse already
				// allocated request models.
				oldRequests = oldRequests[:0]
				newRequests = newRequests[:0]

				// Select next batch of
				// interaction requests.
				if err := tx.NewSelect().
					Model(&oldRequests).
					Where("? < ?", bun.Ident("id"), maxID).
					OrderExpr("? DESC", bun.Ident("id")).
					Limit(batchsz).
					Scan(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
					return gtserror.Newf("error selecting interaction requests: %w", err)
				}

				// Reached end of requests.
				if len(oldRequests) == 0 {
					break
				}

				// Set next maxID value from old requests.
				maxID = oldRequests[len(oldRequests)-1].ID

				// Reslice new to equal that of old requests.
				newRequests = newRequests[:len(oldRequests)]
				if len(newRequests) != len(oldRequests) {
					panic(gtserror.New("BCD"))
				}

				// Convert old request models to new.
				for i, oldRequest := range oldRequests {
					if newRequests[i] == nil {
						newRequests[i] = new(new_gtsmodel.InteractionRequest)
					}
					newRequests[i].ID = oldRequest.ID
					newRequests[i].TargetStatusID = oldRequest.StatusID
					newRequests[i].TargetAccountID = oldRequest.TargetAccountID
					newRequests[i].InteractingAccountID = oldRequest.InteractingAccountID
					newRequests[i].InteractionURI = oldRequest.InteractionURI
					newRequests[i].InteractionType = int16(oldRequest.InteractionType) // #nosec G115
					newRequests[i].AcceptedAt = oldRequest.AcceptedAt
					newRequests[i].RejectedAt = oldRequest.RejectedAt
					newRequests[i].ResponseURI = oldRequest.URI
					newRequests[i].Polite = util.Ptr(false) // old requests were always impolite

					// If the request was accepted, and targeted
					// a status on our instance, then generate an
					// authorization URI for it, and reuse the
					// interaction URI to build a fake interaction
					// request URI, in order to be able to serve
					// an Authorization for it that makes sense.
					if oldRequest.AcceptedAt.IsZero() || oldRequest.URI == "" {
						// Not accepted,
						// nothing to do.
						continue
					}

					// Check if one of our accepts.
					acceptURI, err := url.Parse(oldRequest.URI)
					if err != nil {
						// Weird, skip this one.
						log.Warnf(ctx,
							"could not parse oldRequest.URI for interaction request %s,"+
								" skipping forward-compat hack (don't worry, this is not a big deal): %v",
							oldRequest.ID, err,
						)
						continue
					}

					if acceptURI.Host != host && acceptURI.Host != accountDomain {
						// Not one of ours.
						continue
					}

					// Do the nasty hack to bodge an authorization URI.
					// Creates `https://example.org/users/aaa/authorizations/[ID]`
					// from `https://example.org/users/aaa/authorizations/[ID]`.
					authorizationURI := strings.ReplaceAll(
						oldRequest.URI,
						"/accepts/"+oldRequest.ID,
						"/authorizations/"+oldRequest.ID,
					)
					newRequests[i].AuthorizationURI = authorizationURI

					// Update the corresponding interaction
					// with generated authorization URI.
					var intUpdateQTable string
					if oldRequest.InteractionType == old_gtsmodel.InteractionLike {
						intUpdateQTable = "status_faves"
					} else {
						intUpdateQTable = "statuses"
					}
					if _, err := tx.
						NewUpdate().
						Table(intUpdateQTable).
						Set("? = ?", bun.Ident("approved_by_uri"), authorizationURI).
						Where("? = ?", bun.Ident("approved_by_uri"), oldRequest.URI).
						Exec(ctx); err != nil {
						return gtserror.Newf("error updating approved_by_uri: %w", err)
					}

					// Re-use the original interaction URI to create
					// a mock interaction request URI on the new model.
					switch oldRequest.InteractionType {
					case old_gtsmodel.InteractionLike:
						newRequests[i].InteractionRequestURI = oldRequest.InteractionURI + new_gtsmodel.LikeRequestSuffix
					case old_gtsmodel.InteractionReply:
						newRequests[i].InteractionRequestURI = oldRequest.InteractionURI + new_gtsmodel.ReplyRequestSuffix
					case old_gtsmodel.InteractionAnnounce:
						newRequests[i].InteractionRequestURI = oldRequest.InteractionURI + new_gtsmodel.AnnounceRequestSuffix
					}
				}

				// Insert the converted interaction
				// request models to new table.
				if _, err := tx.NewInsert().
					Table(tmpTableName).
					Model(&newRequests).
					Exec(ctx); err != nil {
					return gtserror.Newf("error inserting interaction requests: %w", err)
				}

				// Increment insert count.
				count += len(newRequests)

				log.Infof(ctx, "[%d of %d] converting interaction requests", count, total)
			}

			return nil
		}); err != nil {
			return err
		}

		// Ensure that the above transaction
		// has gone ahead without issues.
		//
		// Also placing this here might make
		// breaking this into piecemeal steps
		// easier if turns out necessary.
		newTotal, err := db.NewSelect().
			Table(tmpTableName).
			Count(ctx)
		if err != nil {
			return gtserror.Newf("error geting new interaction requests table count: %w", err)
		} else if total != newTotal {
			return gtserror.Newf("new interaction requests table contains unexpected count %d, want %d", newTotal, total)
		}

		// Attempt to merge any sqlite write-ahead-log.
		if err := doWALCheckpoint(ctx, db); err != nil {
			return err
		}

		// Drop the old interaction requests table and rename new one to replace it.
		if err := db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {

			log.Info(ctx, "dropping old interaction_requests table")
			if _, err := tx.NewDropTable().
				Table(tableName).
				Exec(ctx); err != nil {
				return gtserror.Newf("error dropping old interaction requests table: %w", err)
			}

			log.Info(ctx, "renaming new interaction_requests table to old")
			if _, err := tx.NewRaw("ALTER TABLE ? RENAME TO ?",
				bun.Ident(tmpTableName),
				bun.Ident(tableName),
			).Exec(ctx); err != nil {
				return gtserror.Newf("error renaming interaction requests table: %w", err)
			}

			// Create necessary indices on the new table.
			for index, columns := range map[string][]string{
				"interaction_requests_target_status_id_idx":       {"target_status_id"},
				"interaction_requests_interacting_account_id_idx": {"interacting_account_id"},
				"interaction_requests_target_account_id_idx":      {"target_account_id"},
				"interaction_requests_accepted_at_idx":            {"accepted_at"},
				"interaction_requests_rejected_at_idx":            {"rejected_at"},
			} {
				log.Infof(ctx, "recreating %s index", index)
				if _, err := tx.NewCreateIndex().
					Table(tableName).
					Index(index).
					Column(columns...).
					Exec(ctx); err != nil {
					return err
				}
			}

			if tx.Dialect().Name() == dialect.PG {
				// Rename postgres uniqueness constraints:
				// "new_interaction_requests_*" -> "interaction_requests_*"
				log.Info(ctx, "renaming interaction_requests constraints on new table")
				for _, spec := range []struct {
					old string
					new string
				}{
					{
						old: "new_interaction_requests_pkey",
						new: "interaction_requests_pkey",
					},
					{
						old: "new_interaction_requests_interaction_request_uri_key",
						new: "interaction_requests_interaction_request_uri_key",
					},
					{
						old: "new_interaction_requests_interaction_uri_key",
						new: "interaction_requests_interaction_uri_key",
					},
					{
						old: "new_interaction_requests_response_uri_key",
						new: "interaction_requests_response_uri_key",
					},
				} {
					if _, err := tx.NewRaw("ALTER TABLE ? RENAME CONSTRAINT ? TO ?",
						bun.Ident(tableName),
						bun.Safe(spec.old),
						bun.Safe(spec.new),
					).Exec(ctx); err != nil {
						return gtserror.Newf("error renaming postgres interaction requests constraint %s: %w", spec.new, err)
					}
				}
			}

			return nil
		}); err != nil {
			return err
		}

		// Final sqlite write-ahead-log merge.
		return doWALCheckpoint(ctx, db)
	}

	down := func(ctx context.Context, db *bun.DB) error {
		return nil
	}

	if err := Migrations.Register(up, down); err != nil {
		panic(err)
	}
}
