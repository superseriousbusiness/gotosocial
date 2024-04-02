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
	"net/url"

	"github.com/uptrace/bun"
)

func init() {
	up := func(ctx context.Context, db *bun.DB) error {
		return db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
			// Select URI of each friendica account
			// with an empty domain that doesn't have
			// a corresponding user (ie., not local).
			// Query looks like:
			//
			//	SELECT "uri" FROM "accounts"
			//	WHERE ("username" = 'friendica')
			//	AND ("actor_type" = 'Application')
			//	AND ("domain" IS NULL)
			//	AND ("id" NOT IN (SELECT "account_id" FROM "users"))
			URIStrs := []string{}
			if err := tx.
				NewSelect().
				Table("accounts").
				Column("uri").
				Where("? = ?", bun.Ident("username"), "friendica").
				Where("? = ?", bun.Ident("actor_type"), "Application").
				Where("? IS NULL", bun.Ident("domain")).
				Where("? NOT IN (?)", bun.Ident("id"), tx.NewSelect().Table("users").Column("account_id")).
				Scan(ctx, &URIStrs); err != nil {
				return err
			}

			// For each URI found this way, parse
			// out the Host part and update the
			// domain of the domain-less account.
			for _, uriStr := range URIStrs {
				uri, err := url.Parse(uriStr)
				if err != nil {
					return err
				}

				domain := uri.Host
				if _, err := tx.
					NewUpdate().
					Table("accounts").
					Set("? = ?", bun.Ident("domain"), domain).
					Where("? = ?", bun.Ident("uri"), uriStr).
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
