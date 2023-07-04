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

package bundb

import (
	"context"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtscontext"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/id"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type instanceDB struct {
	conn  *DBConn
	state *state.State
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

func (i *instanceDB) GetInstance(ctx context.Context, domain string) (*gtsmodel.Instance, db.Error) {
	// Normalize the domain as punycode
	var err error
	domain, err = util.Punify(domain)
	if err != nil {
		return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
	}

	return i.getInstance(
		ctx,
		"Domain",
		func(instance *gtsmodel.Instance) error {
			return i.conn.NewSelect().
				Model(instance).
				Where("? = ?", bun.Ident("instance.domain"), domain).
				Scan(ctx)
		},
		domain,
	)
}

func (i *instanceDB) GetInstanceByID(ctx context.Context, id string) (*gtsmodel.Instance, error) {
	return i.getInstance(
		ctx,
		"ID",
		func(instance *gtsmodel.Instance) error {
			return i.conn.NewSelect().
				Model(instance).
				Where("? = ?", bun.Ident("instance.id"), id).
				Scan(ctx)
		},
		id,
	)
}

func (i *instanceDB) getInstance(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Instance) error, keyParts ...any) (*gtsmodel.Instance, db.Error) {
	// Fetch instance from database cache with loader callback
	instance, err := i.state.Caches.GTS.Instance().Load(lookup, func() (*gtsmodel.Instance, error) {
		var instance gtsmodel.Instance

		// Not cached! Perform database query.
		if err := dbQuery(&instance); err != nil {
			return nil, i.conn.ProcessError(err)
		}

		return &instance, nil
	}, keyParts...)
	if err != nil {
		return nil, err
	}

	if gtscontext.Barebones(ctx) {
		// no need to fully populate.
		return instance, nil
	}

	// Further populate the instance fields where applicable.
	if err := i.populateInstance(ctx, instance); err != nil {
		return nil, err
	}

	return instance, nil
}

func (i *instanceDB) populateInstance(ctx context.Context, instance *gtsmodel.Instance) error {
	var (
		err  error
		errs = make(gtserror.MultiError, 0, 2)
	)

	if instance.DomainBlockID != "" && instance.DomainBlock == nil {
		// Instance domain block is not set, fetch from database.
		instance.DomainBlock, err = i.state.DB.GetDomainBlock(
			gtscontext.SetBarebones(ctx),
			instance.Domain,
		)
		if err != nil {
			errs.Append(gtserror.Newf("error populating instance domain block: %w", err))
		}
	}

	if instance.ContactAccountID != "" && instance.ContactAccount == nil {
		// Instance domain block is not set, fetch from database.
		instance.ContactAccount, err = i.state.DB.GetAccountByID(
			gtscontext.SetBarebones(ctx),
			instance.ContactAccountID,
		)
		if err != nil {
			errs.Append(gtserror.Newf("error populating instance contact account: %w", err))
		}
	}

	return errs.Combine()
}

func (i *instanceDB) PutInstance(ctx context.Context, instance *gtsmodel.Instance) error {
	// Normalize the domain as punycode
	var err error
	instance.Domain, err = util.Punify(instance.Domain)
	if err != nil {
		return gtserror.Newf("error punifying domain %s: %w", instance.Domain, err)
	}

	return i.state.Caches.GTS.Instance().Store(instance, func() error {
		_, err := i.conn.NewInsert().Model(instance).Exec(ctx)
		return i.conn.ProcessError(err)
	})
}

func (i *instanceDB) UpdateInstance(ctx context.Context, instance *gtsmodel.Instance, columns ...string) error {
	// Normalize the domain as punycode
	var err error
	instance.Domain, err = util.Punify(instance.Domain)
	if err != nil {
		return gtserror.Newf("error punifying domain %s: %w", instance.Domain, err)
	}

	// Update the instance's last-updated
	instance.UpdatedAt = time.Now()
	if len(columns) != 0 {
		columns = append(columns, "updated_at")
	}

	return i.state.Caches.GTS.Instance().Store(instance, func() error {
		_, err := i.conn.
			NewUpdate().
			Model(instance).
			Where("? = ?", bun.Ident("instance.id"), instance.ID).
			Column(columns...).
			Exec(ctx)
		return i.conn.ProcessError(err)
	})
}

func (i *instanceDB) GetInstancePeers(ctx context.Context, includeSuspended bool) ([]*gtsmodel.Instance, db.Error) {
	instanceIDs := []string{}

	q := i.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("instances"), bun.Ident("instance")).
		// Select just the IDs of each instance.
		Column("instance.id").
		// Exclude our own instance.
		Where("? != ?", bun.Ident("instance.domain"), config.GetHost())

	if !includeSuspended {
		q = q.Where("? IS NULL", bun.Ident("instance.suspended_at"))
	}

	if err := q.Scan(ctx, &instanceIDs); err != nil {
		return nil, i.conn.ProcessError(err)
	}

	if len(instanceIDs) == 0 {
		return make([]*gtsmodel.Instance, 0), nil
	}

	instances := make([]*gtsmodel.Instance, 0, len(instanceIDs))

	for _, id := range instanceIDs {
		// Select each instance by its ID.
		instance, err := i.GetInstanceByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting instance %q: %v", id, err)
			continue
		}

		// Append to return slice.
		instances = append(instances, instance)
	}

	return instances, nil
}

func (i *instanceDB) GetInstanceAccounts(ctx context.Context, domain string, maxID string, limit int) ([]*gtsmodel.Account, db.Error) {
	// Ensure reasonable
	if limit < 0 {
		limit = 0
	}

	// Normalize the domain as punycode.
	var err error
	domain, err = util.Punify(domain)
	if err != nil {
		return nil, gtserror.Newf("error punifying domain %s: %w", domain, err)
	}

	// Make educated guess for slice size
	accountIDs := make([]string, 0, limit)

	q := i.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("accounts"), bun.Ident("account")).
		// Select just the account ID.
		Column("account.id").
		// Select accounts belonging to given domain.
		Where("? = ?", bun.Ident("account.domain"), domain).
		Order("account.id DESC")

	if maxID == "" {
		maxID = id.Highest
	}
	q = q.Where("? < ?", bun.Ident("account.id"), maxID)

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx, &accountIDs); err != nil {
		return nil, i.conn.ProcessError(err)
	}

	// Catch case of no accounts early.
	count := len(accountIDs)
	if count == 0 {
		return nil, db.ErrNoEntries
	}

	// Select each account by its ID.
	accounts := make([]*gtsmodel.Account, 0, count)
	for _, id := range accountIDs {
		account, err := i.state.DB.GetAccountByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting account %q: %v", id, err)
			continue
		}

		// Append to return slice.
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (i *instanceDB) GetInstanceModeratorAddresses(ctx context.Context) ([]string, db.Error) {
	addresses := []string{}

	// Select email addresses of approved, confirmed,
	// and enabled moderators or admins.

	q := i.conn.
		NewSelect().
		TableExpr("? AS ?", bun.Ident("users"), bun.Ident("user")).
		Column("user.email").
		Where("? = ?", bun.Ident("user.approved"), true).
		Where("? IS NOT NULL", bun.Ident("user.confirmed_at")).
		Where("? = ?", bun.Ident("user.disabled"), false).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				Where("? = ?", bun.Ident("user.moderator"), true).
				WhereOr("? = ?", bun.Ident("user.admin"), true)
		}).
		OrderExpr("? ASC", bun.Ident("user.email"))

	if err := q.Scan(ctx, &addresses); err != nil {
		return nil, i.conn.ProcessError(err)
	}

	if len(addresses) == 0 {
		return nil, db.ErrNoEntries
	}

	return addresses, nil
}
