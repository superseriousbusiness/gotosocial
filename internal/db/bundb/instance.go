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

package bundb

import (
	"context"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/uptrace/bun"
)

type instanceDB struct {
	conn *DBConn
}

func (i *instanceDB) CountInstanceUsers(ctx context.Context, domain string) (int, db.Error) {
	q := i.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		Column("account.id").
		Where("? != ?", bun.Ident("account.username"), domain).
		Where("? IS NULL", bun.Ident("account.suspended_at"))

	if domain == config.GetHost() || domain == config.GetAccountDomain() {
		// if the domain is *this* domain, just count where the domain field is null
		q = q.WhereGroup(" AND ", whereEmptyOrNull("account.domain"))
	} else {
		q = q.Where("? = ?", bun.Ident("account.domain"), domain)
	}

	count, err := q.Count(ctx)
	if err != nil {
		return 0, i.conn.ProcessError(err)
	}
	return count, nil
}

func (i *instanceDB) CountInstanceStatuses(ctx context.Context, domain string) (int, db.Error) {
	q := i.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("statuses"), bun.Ident("status"))

	if domain == config.GetHost() || domain == config.GetAccountDomain() {
		// if the domain is *this* domain, just count where local is true
		q = q.Where("? = ?", bun.Ident("status.local"), true)
	} else {
		// join on the domain of the account
		q = q.
			Join("JOIN ? AS ? ON ? = ?", bun.Ident("accounts"), bun.Ident("account"), bun.Ident("account.id"), bun.Ident("status.account_id")).
			Where("? = ?", bun.Ident("account.domain"), domain)
	}

	count, err := q.Count(ctx)
	if err != nil {
		return 0, i.conn.ProcessError(err)
	}
	return count, nil
}

func (i *instanceDB) CountInstanceDomains(ctx context.Context, domain string) (int, db.Error) {
	q := i.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("instances"), bun.Ident("instance"))

	if domain == config.GetHost() {
		// if the domain is *this* domain, just count other instances it knows about
		// exclude domains that are blocked
		q = q.
			Where("? != ?", bun.Ident("instance.domain"), domain).
			Where("? IS NULL", bun.Ident("instance.suspended_at"))
	} else {
		// TODO: implement federated domain counting properly for remote domains
		return 0, nil
	}

	count, err := q.Count(ctx)
	if err != nil {
		return 0, i.conn.ProcessError(err)
	}
	return count, nil
}

func (i *instanceDB) GetInstancePeers(ctx context.Context, includeSuspended bool) ([]*gtsmodel.Instance, db.Error) {
	instances := []*gtsmodel.Instance{}

	q := i.conn.
		NewSelect().
		Model(&instances).
		Where("? != ?", bun.Ident("instance.domain"), config.GetHost())

	if !includeSuspended {
		q = q.Where("? IS NULL", bun.Ident("instance.suspended_at"))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, i.conn.ProcessError(err)
	}

	return instances, nil
}

func (i *instanceDB) GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, db.Error) {
	accounts := []*gtsmodel.Account{}

	q := i.conn.NewSelect().
		Model(&accounts).
		Where("? = ?", bun.Ident("account.domain"), domain).
		Order("account.id DESC")

	if maxID != "" {
		q = q.Where("? < ?", bun.Ident("account.id"), maxID)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, i.conn.ProcessError(err)
	}

	if len(accounts) == 0 {
		return nil, db.ErrNoEntries
	}

	return accounts, nil
}
