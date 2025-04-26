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

package admin

import (
	"context"
	"errors"
	"fmt"

	apimodel "code.superseriousbusiness.org/gotosocial/internal/api/model"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtserror"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/typeutils"
	"code.superseriousbusiness.org/gotosocial/internal/util"
)

// RulesGet returns all rules stored on this instance.
func (p *Processor) RulesGet(
	ctx context.Context,
) ([]*apimodel.AdminInstanceRule, gtserror.WithCode) {
	rules, err := p.state.DB.GetActiveRules(ctx)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	apiRules := make([]*apimodel.AdminInstanceRule, len(rules))

	for i := range rules {
		apiRules[i] = typeutils.InstanceRuleToAdminAPIRule(&rules[i])
	}

	return apiRules, nil
}

// RuleGet returns one rule, with the given ID.
func (p *Processor) RuleGet(ctx context.Context, id string) (*apimodel.AdminInstanceRule, gtserror.WithCode) {
	rule, err := p.state.DB.GetRuleByID(ctx, id)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	return typeutils.InstanceRuleToAdminAPIRule(rule), nil
}

// RuleCreate adds a new rule to the instance.
func (p *Processor) RuleCreate(ctx context.Context, form *apimodel.InstanceRuleCreateRequest) (*apimodel.AdminInstanceRule, gtserror.WithCode) {
	ruleID, err := id.NewRandomULID()
	if err != nil {
		return nil, gtserror.NewErrorInternalError(fmt.Errorf("error creating id for new instance rule: %s", err), "error creating rule ID")
	}

	rule := &gtsmodel.Rule{
		ID:   ruleID,
		Text: form.Text,
	}

	if err = p.state.DB.PutRule(ctx, rule); err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return typeutils.InstanceRuleToAdminAPIRule(rule), nil
}

// RuleUpdate updates text for an existing rule.
func (p *Processor) RuleUpdate(ctx context.Context, id string, form *apimodel.InstanceRuleCreateRequest) (*apimodel.AdminInstanceRule, gtserror.WithCode) {
	rule, err := p.state.DB.GetRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("RuleUpdate: no rule with id %s found in the db", id)
			return nil, gtserror.NewErrorNotFound(err)
		}
		err := fmt.Errorf("RuleUpdate: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	rule.Text = form.Text

	updatedRule, err := p.state.DB.UpdateRule(ctx, rule)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return typeutils.InstanceRuleToAdminAPIRule(updatedRule), nil
}

// RuleDelete deletes an existing rule.
func (p *Processor) RuleDelete(ctx context.Context, id string) (*apimodel.AdminInstanceRule, gtserror.WithCode) {
	rule, err := p.state.DB.GetRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, db.ErrNoEntries) {
			err = fmt.Errorf("RuleUpdate: no rule with id %s found in the db", id)
			return nil, gtserror.NewErrorNotFound(err)
		}
		err := fmt.Errorf("RuleUpdate: db error: %s", err)
		return nil, gtserror.NewErrorInternalError(err)
	}

	rule.Deleted = util.Ptr(true)
	deletedRule, err := p.state.DB.UpdateRule(ctx, rule)
	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return typeutils.InstanceRuleToAdminAPIRule(deletedRule), nil
}
