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

package db

import (
	"context"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
)

// Rule handles getting/creation/deletion/updating of instance rules.
type Rule interface {
	// GetRuleByID gets one rule by its db id.
	GetRuleByID(ctx context.Context, id string) (*gtsmodel.Rule, error)

	// GetRulesByIDs gets multiple rules by their db idd.
	GetRulesByIDs(ctx context.Context, ids []string) ([]*gtsmodel.Rule, error)

	// GetRules gets all active (not deleted) rules.
	GetActiveRules(ctx context.Context) ([]gtsmodel.Rule, error)

	// PutRule puts the given rule in the database.
	PutRule(ctx context.Context, rule *gtsmodel.Rule) error

	// UpdateRule updates one rule by its db id.
	UpdateRule(ctx context.Context, rule *gtsmodel.Rule) (*gtsmodel.Rule, error)
}
