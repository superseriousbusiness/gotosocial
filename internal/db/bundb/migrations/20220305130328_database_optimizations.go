/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package migrations

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// ACCOUNTS are often selected by URI, username, or domain
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_uri_idx").
				Column("uri").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_username_idx").
				Column("username").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_domain_idx").
				Column("domain").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Account{}).
				Index("accounts_username_domain_idx"). // for selecting local accounts by username
				Column("username", "domain").
				Exec(ctx); err != nil {
				return err
			}

			// NOTIFICATIONS are commonly selected by their target_account_id
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Notification{}).
				Index("notifications_target_account_id_idx").
				Column("target_account_id").
				Exec(ctx); err != nil {
				return err
			}

			// STATUSES are selected in many different ways, so we need quite few indexes
			// to avoid queries becoming painfully slow as more statuses get stored
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_uri_idx").
				Column("uri").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_account_id_idx").
				Column("account_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_in_reply_to_account_id_idx").
				Column("in_reply_to_account_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_boost_of_account_id_idx").
				Column("boost_of_account_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_in_reply_to_id_idx").
				Column("in_reply_to_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_boost_of_id_idx").
				Column("boost_of_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_visibility_idx").
				Column("visibility").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Status{}).
				Index("statuses_local_idx").
				Column("local").
				Exec(ctx); err != nil {
				return err
			}

			// DOMAIN BLOCKS are almost always selected by their domain
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.DomainBlock{}).
				Index("domain_blocks_domain_idx").
				Column("domain").
				Exec(ctx); err != nil {
				return err
			}

			// INSTANCES are usually selected by their domain
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Instance{}).
				Index("instances_domain_idx").
				Column("domain").
				Exec(ctx); err != nil {
				return err
			}

			// STATUS FAVES are almost always selected by their target status
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.StatusFave{}).
				Index("status_faves_status_id_idx").
				Column("status_id").
				Exec(ctx); err != nil {
				return err
			}

			// MENTIONS are almost always selected by their originating status
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Mention{}).
				Index("mentions_status_id_idx").
				Column("status_id").
				Exec(ctx); err != nil {
				return err
			}

			// FOLLOW_REQUESTS and FOLLOWS are usually selected by who they originate from, and who they target
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.FollowRequest{}).
				Index("follow_requests_account_id_idx").
				Column("account_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.FollowRequest{}).
				Index("follow_requests_target_account_id_idx").
				Column("target_account_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Follow{}).
				Index("follows_account_id_idx").
				Column("account_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Follow{}).
				Index("follows_target_account_id_idx").
				Column("target_account_id").
				Exec(ctx); err != nil {
				return err
			}

			// BLOCKS are usually selected simultaneously by who they originate from and who they target
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Block{}).
				Index("blocks_account_id_target_account_id_idx").
				Column("account_id", "target_account_id").
				Exec(ctx); err != nil {
				return err
			}

			// MEDIA ATTACHMENTS are often selected by status ID, the account that owns the attachment
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.MediaAttachment{}).
				Index("media_attachments_status_id_idx").
				Column("status_id").
				Exec(ctx); err != nil {
				return err
			}

			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.MediaAttachment{}).
				Index("media_attachments_account_id_idx").
				Column("account_id").
				Exec(ctx); err != nil {
				return err
			}

			// for media cleanup jobs, attachments will be selected on a bunch of fields so make an index of this...
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.MediaAttachment{}).
				Index("media_attachments_cleanup_idx").
				Column("cached", "avatar", "header", "created_at", "remote_url").
				Exec(ctx); err != nil {
				return err
			}

			// TOKENS are usually selected via Access field for user-level tokens
			if _, err := tx.
				NewCreateIndex().
				Model(&gtsmodel.Token{}).
				Index("tokens_access_idx").
				Column("access").
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
