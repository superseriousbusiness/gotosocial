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

package processing_test

import (
	"context"
	"errors"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"code.superseriousbusiness.org/gotosocial/internal/processing"
	"github.com/stretchr/testify/suite"
)

type ParseMentionTestSuite struct {
	ProcessingStandardTestSuite
}

func (suite *ParseMentionTestSuite) TestParseMentionFunc() {
	var (
		ctx          = context.Background()
		parseMention = processing.GetParseMentionFunc(&suite.state, suite.federator)
		originAcctID = suite.testAccounts["local_account_1"].ID
		statusID     = id.NewULID()
	)

	type testStruct struct {
		namestring         string
		expectedTargetAcct *gtsmodel.Account
		err                error
	}

	for i, test := range []testStruct{
		{
			namestring:         "@1happyturtle",
			expectedTargetAcct: suite.testAccounts["local_account_2"],
		},
		{
			namestring:         "@1happyturtle@localhost:8080",
			expectedTargetAcct: suite.testAccounts["local_account_2"],
		},
		{
			namestring:         "@foss_satan@fossbros-anonymous.io",
			expectedTargetAcct: suite.testAccounts["remote_account_1"],
		},
		{
			namestring: "@foss_satan",
			err:        errors.New("db error getting mention local target account foss_satan: sql: no rows in result set"),
		},
		{
			namestring: "@foss_satan@aaaaaaaaaaaaaaaaaaa.example.org",
			err:        errors.New("error fetching mention remote target account: enrichAccount: error webfingering account: fingerRemoteAccount: error webfingering @foss_satan@aaaaaaaaaaaaaaaaaaa.example.org: failed to discover webfinger URL fallback for: aaaaaaaaaaaaaaaaaaa.example.org through host-meta: GET request for https://aaaaaaaaaaaaaaaaaaa.example.org/.well-known/host-meta failed: "),
		},
		{
			namestring: "pee pee poo poo",
			err:        errors.New("error extracting mention target: couldn't match namestring pee pee poo poo"),
		},
	} {
		mention, err := parseMention(ctx, test.namestring, originAcctID, statusID)
		if test.err != nil {
			suite.EqualError(err, test.err.Error())
			continue
		}

		if err != nil {
			suite.Fail(err.Error())
			continue
		}

		if mention.OriginAccount == nil {
			suite.Failf("nil origin account", "test %d, namestring %s", i+1, test.namestring)
			continue
		}

		if mention.TargetAccount == nil {
			suite.Failf("nil target account", "test %d, namestring %s", i+1, test.namestring)
			continue
		}

		suite.NotEmpty(mention.ID)
		suite.Equal(originAcctID, mention.OriginAccountID)
		suite.Equal(originAcctID, mention.OriginAccount.ID)
		suite.Equal(test.expectedTargetAcct.ID, mention.TargetAccountID)
		suite.Equal(test.expectedTargetAcct.ID, mention.TargetAccount.ID)
		suite.Equal(test.expectedTargetAcct.URI, mention.TargetAccountURI)
		suite.Equal(test.expectedTargetAcct.URL, mention.TargetAccountURL)
	}
}

func TestParseMentionTestSuite(t *testing.T) {
	suite.Run(t, &ParseMentionTestSuite{})
}
