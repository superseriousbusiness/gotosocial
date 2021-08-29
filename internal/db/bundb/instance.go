/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

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
	config *config.Config
	conn   *DBConn
}

func (i *instanceDB) CountInstanceUsers(ctx context.Context, domain string) (int, db.Error) {
	q := i.conn.
		NewSelect().
		Model(&[]*gtsmodel.Account{}).
		Where("username != ?", domain).
		Where("? IS NULL", bun.Ident("suspended_at"))

	if domain == i.config.Host {
		// if the domain is *this* domain, just count where the domain field is null
		q = q.WhereGroup(" AND ", whereEmptyOrNull("domain"))
	} else {
		q = q.Where("domain = ?", domain)
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
		Model(&[]*gtsmodel.Status{})

	if domain == i.config.Host {
		// if the domain is *this* domain, just count where local is true
		q = q.Where("local = ?", true)
	} else {
		// join on the domain of the account
		q = q.Join("JOIN accounts AS account ON account.id = status.account_id").
			Where("account.domain = ?", domain)
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
		Model(&[]*gtsmodel.Instance{})

	if domain == i.config.Host {
		// if the domain is *this* domain, just count other instances it knows about
		// exclude domains that are blocked
		q = q.
			Where("domain != ?", domain).
			Where("? IS NULL", bun.Ident("suspended_at"))
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

func (i *instanceDB) GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, db.Error) {
	i.conn.log.Debug("GetAccountsForInstance")

	accounts := []*gtsmodel.Account{}

	q := i.conn.NewSelect().
		Model(&accounts).
		Where("domain = ?", domain).
		Order("id DESC")

	if maxID != "" {
		q = q.Where("id < ?", maxID)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, i.conn.ProcessError(err)
	}
	return accounts, nil
}
