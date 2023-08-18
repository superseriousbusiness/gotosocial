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
	"errors"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/log"
	"github.com/superseriousbusiness/gotosocial/internal/state"
	"github.com/superseriousbusiness/gotosocial/internal/util"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect"
)

type ruleDB struct {
	db    *DB
	state *state.State
}

func (r *ruleDB) GetRuleByID(ctx context.Context, id string) (*gtsmodel.Rule, error) {
	var rule gtsmodel.Rule

	q := r.db.
		NewSelect().
		Model(&rule).
		Where("? = ?", bun.Ident("rule.id"), bun.Ident(id)).
		Where("? = ?", bun.Ident("rule.deleted"), util.Ptr(false))

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return &rule, nil
}

func (r *ruleDB) GetRulesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Rule, error) {
	rules := make([]*gtsmodel.Rule, 0, len(ids))

	for _, id := range ids {
		// Attempt to fetch status from DB.
		rule, err := r.GetRuleByID(ctx, id)
		if err != nil {
			log.Errorf(ctx, "error getting rule %q: %v", id, err)
			continue
		}

		// Append status to return slice.
		rules = append(rules, rule)
	}

	return rules, nil
}

func (r *ruleDB) GetRules(ctx context.Context) ([]gtsmodel.Rule, error) {
	rules := make([]gtsmodel.Rule, 0)

	q := r.db.
		NewSelect().
		Model(&rules).
		Where("? = ?", bun.Ident("rule.deleted"), util.Ptr(false)).
		Order("rule.order ASC")

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return rules, nil
}

func (r *ruleDB) PutRule(ctx context.Context, rule *gtsmodel.Rule) error {
	if r.db.Dialect().Name() == dialect.SQLite {
		// sqlite does not support bun's autoincrement, so we need to set rule.Order ourselves

		var lastRule gtsmodel.Rule

		q := r.db.
			NewSelect().
			Model(rule).
			Order("rule.order DESC").
			Limit(1)

		if err := q.Scan(ctx, &lastRule); err != nil {
			dbErr := err

			if errors.Is(dbErr, db.ErrNoEntries) {
				rule.Order = 0
			} else {
				return dbErr
			}
		} else {
			rule.Order = lastRule.Order + 1
		}
	}

	if _, err := r.db.NewInsert().Model(rule).Exec(ctx); err != nil {
		return err
	}

	// invalidate cached local instance response, so it gets updated with the new rules
	r.state.Caches.GTS.Instance().Invalidate("Domain", config.GetHost())

	return nil
}

func (r *ruleDB) UpdateRule(ctx context.Context, rule *gtsmodel.Rule) (*gtsmodel.Rule, error) {
	// Update the rule's last-updated
	rule.UpdatedAt = time.Now()

	if _, err := r.db.
		NewUpdate().
		Model(rule).
		WherePK().
		Exec(ctx); err != nil {
		return nil, err
	}

	// invalidate cached local instance response, so it gets updated with the new rules
	r.state.Caches.GTS.Instance().Invalidate("Domain", config.GetHost())

	return rule, nil
}
