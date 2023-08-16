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

	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtserror"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

// RulesGet returns all rules stored on this instance.
func (p *Processor) RulesGet(
	ctx context.Context,
) ([]gtsmodel.Rule, gtserror.WithCode) {
	rules, err := p.state.DB.GetRules(ctx)

	if err != nil {
		return nil, gtserror.NewErrorInternalError(err)
	}

	return rules, nil
}

// RuleGet returns one rule, with the given ID.
func (p *Processor) RuleGet(ctx context.Context, id string) (*gtsmodel.Rule, gtserror.WithCode) {
	rule, err := p.state.DB.GetRuleByID(ctx, id)
	if err != nil {
		if err == db.ErrNoEntries {
			return nil, gtserror.NewErrorNotFound(err)
		}
		return nil, gtserror.NewErrorInternalError(err)
	}

	return rule, nil
}
