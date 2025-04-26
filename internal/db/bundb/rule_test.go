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

package bundb_test

import (
	"context"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/stretchr/testify/suite"
)

type RuleTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *RuleTestSuite) TestPutRuleWithExisting() {
	r := &gtsmodel.Rule{
		ID:   id.NewULID(),
		Text: "Pee pee poo poo",
	}

	if err := suite.state.DB.PutRule(context.Background(), r); err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(uint(len(suite.testRules)), *r.Order)
}

func (suite *RuleTestSuite) TestPutRuleNoExisting() {
	var (
		ctx      = context.Background()
		whereAny = []db.Where{{Key: "id", Value: "", Not: true}}
	)

	// Wipe all existing rules from the DB.
	if err := suite.state.DB.DeleteWhere(
		ctx,
		whereAny,
		&[]*gtsmodel.Rule{},
	); err != nil {
		suite.FailNow(err.Error())
	}

	r := &gtsmodel.Rule{
		ID:   id.NewULID(),
		Text: "Pee pee poo poo",
	}

	if err := suite.state.DB.PutRule(ctx, r); err != nil {
		suite.FailNow(err.Error())
	}

	// New rule is now only rule.
	suite.EqualValues(uint(0), *r.Order)
}

func (suite *RuleTestSuite) TestGetRuleByID() {
	rule, err := suite.state.DB.GetRuleByID(
		context.Background(),
		suite.testRules["rule1"].ID,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.NotNil(rule)
}

func (suite *RuleTestSuite) TestGetRulesByID() {
	ruleIDs := make([]string, 0, len(suite.testRules))
	for _, rule := range suite.testRules {
		ruleIDs = append(ruleIDs, rule.ID)
	}

	rules, err := suite.state.DB.GetRulesByIDs(
		context.Background(),
		ruleIDs,
	)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(rules, len(suite.testRules))
}

func (suite *RuleTestSuite) TestGetActiveRules() {
	var activeRules int
	for _, rule := range suite.testRules {
		if !*rule.Deleted {
			activeRules++
		}
	}

	rules, err := suite.state.DB.GetActiveRules(context.Background())
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Len(rules, activeRules)
}

func TestRuleTestSuite(t *testing.T) {
	suite.Run(t, new(RuleTestSuite))
}
