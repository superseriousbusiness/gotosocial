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

	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
)

type ruleDB struct {
	db    *WrappedDB
	state *state.State
}

func (r *ruleDB) newRuleQ(rule interface{}) *bun.SelectQuery {
	return r.db.NewSelect().Model(rule)
}

// func (r *ruleDB) GetRuleByID(ctx context.Context, id string) (*gtsmodel.Rule, error) {
// 	return r.getRule(
// 		ctx,
// 		"ID",
// 		func(rule *gtsmodel.Rule) error {
// 			return r.newRuleQ(rule).Where("? = ?", bun.Ident("rule.id"), id).Scan(ctx)
// 		},
// 		id,
// 	)
// }

func (r *ruleDB) GetRules(ctx context.Context) ([]gtsmodel.Rule, error) {
	rules := make([]gtsmodel.Rule, 0)

	q := r.db.
		NewSelect().
		Model(new(gtsmodel.Rule)).
		Where("? IS ?", bun.Ident("rule.deleted"), util.Ptr(false)).
		Order("rule.order ASC")

	if err := q.Scan(ctx, &rules); err != nil {
		return nil, r.db.ProcessError(err)
	}

	return rules, nil
}

// func (r *ruleDB) getRule(ctx context.Context, lookup string, dbQuery func(*gtsmodel.Rule) error, keyParts ...any) (*gtsmodel.Rule, error) {

// 	// Fetch rule from database cache with loader callback
// 	rule, err := r.state.Caches.GTS.Rule().Load(lookup, func() (*gtsmodel.Rule, error) {
// 		var rule gtsmodel.Rule

// 		// Not cached! Perform database query
// 		if err := dbQuery(&rule); err != nil {
// 			return nil, r.db.ProcessError(err)
// 		}

// 		return &rule, nil
// 	}, keyParts...)
// 	if err != nil {
// 		// error already processed
// 		return nil, err
// 	}

// 	if gtscontext.Barebones(ctx) {
// 		// Only a barebones model was requested.
// 		return rule, nil
// 	}

// 	if err := r.state.DB.PopulateRule(ctx, rule); err != nil {
// 		return nil, err
// 	}

// 	return rule, nil
// }

// func (r *ruleDB) PopulateRule(ctx context.Context, rule *gtsmodel.Rule) error {
// 	var (
// 		err  error
// 		errs = gtserror.NewMultiError(4)
// 	)

// 	if rule.Account == nil {
// 		// Rule account is not set, fetch from the database.
// 		rule.Account, err = r.state.DB.GetAccountByID(
// 			gtscontext.SetBarebones(ctx),
// 			rule.AccountID,
// 		)
// 		if err != nil {
// 			errs.Appendf("error populating rule account: %w", err)
// 		}
// 	}

// 	if rule.TargetAccount == nil {
// 		// Rule target account is not set, fetch from the database.
// 		rule.TargetAccount, err = r.state.DB.GetAccountByID(
// 			gtscontext.SetBarebones(ctx),
// 			rule.TargetAccountID,
// 		)
// 		if err != nil {
// 			errs.Appendf("error populating rule target account: %w", err)
// 		}
// 	}

// 	if l := len(rule.StatusIDs); l > 0 && l != len(rule.Statuses) {
// 		// Rule target statuses not set, fetch from the database.
// 		rule.Statuses, err = r.state.DB.GetStatusesByIDs(
// 			gtscontext.SetBarebones(ctx),
// 			rule.StatusIDs,
// 		)
// 		if err != nil {
// 			errs.Appendf("error populating rule statuses: %w", err)
// 		}
// 	}

// 	if rule.ActionTakenByAccountID != "" &&
// 		rule.ActionTakenByAccount == nil {
// 		// Rule action account is not set, fetch from the database.
// 		rule.ActionTakenByAccount, err = r.state.DB.GetAccountByID(
// 			gtscontext.SetBarebones(ctx),
// 			rule.ActionTakenByAccountID,
// 		)
// 		if err != nil {
// 			errs.Appendf("error populating rule action taken by account: %w", err)
// 		}
// 	}

// 	return errs.Combine()
// }

// func (r *ruleDB) PutRule(ctx context.Context, rule *gtsmodel.Rule) error {
// 	return r.state.Caches.GTS.Rule().Store(rule, func() error {
// 		_, err := r.db.NewInsert().Model(rule).Exec(ctx)
// 		return r.db.ProcessError(err)
// 	})
// }

// func (r *ruleDB) UpdateRule(ctx context.Context, rule *gtsmodel.Rule, columns ...string) (*gtsmodel.Rule, error) {
// 	// Update the rule's last-updated
// 	rule.UpdatedAt = time.Now()
// 	if len(columns) != 0 {
// 		columns = append(columns, "updated_at")
// 	}

// 	if _, err := r.db.
// 		NewUpdate().
// 		Model(rule).
// 		Where("? = ?", bun.Ident("rule.id"), rule.ID).
// 		Column(columns...).
// 		Exec(ctx); err != nil {
// 		return nil, r.db.ProcessError(err)
// 	}

// 	r.state.Caches.GTS.Rule().Invalidate("ID", rule.ID)
// 	return rule, nil
// }

// func (r *ruleDB) DeleteRuleByID(ctx context.Context, id string) error {
// 	defer r.state.Caches.GTS.Rule().Invalidate("ID", id)

// 	// Load status into cache before attempting a delete,
// 	// as we need it cached in order to trigger the invalidate
// 	// callback. This in turn invalidates others.
// 	_, err := r.GetRuleByID(gtscontext.SetBarebones(ctx), id)
// 	if err != nil {
// 		if errors.Is(err, db.ErrNoEntries) {
// 			// not an issue.
// 			err = nil
// 		}
// 		return err
// 	}

// 	// Finally delete rule from DB.
// 	_, err = r.db.NewDelete().
// 		TableExpr("? AS ?", bun.Ident("rules"), bun.Ident("rule")).
// 		Where("? = ?", bun.Ident("rule.id"), id).
// 		Exec(ctx)
// 	return r.db.ProcessError(err)
// }
