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

package account_test

import (
	"context"
	"slices"
	"testing"

	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"github.com/stretchr/testify/suite"
)

type AliasTestSuite struct {
	AccountStandardTestSuite
}

func (suite *AliasTestSuite) TestAliasAccount() {
	for _, test := range []struct {
		newAliases      []string
		expectedAliases []string
		expectedErr     string
	}{
		// Alias zork to turtle.
		{
			newAliases: []string{
				"http://localhost:8080/users/1happyturtle",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/1happyturtle",
			},
		},
		// Alias zork to admin.
		{
			newAliases: []string{
				"http://localhost:8080/users/admin",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/admin",
			},
		},
		// Alias zork to turtle AND admin.
		{
			newAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
			},
		},
		// Same again (noop).
		{
			newAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
			},
		},
		// Remove admin alias.
		{
			newAliases: []string{
				"http://localhost:8080/users/1happyturtle",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/1happyturtle",
			},
		},
		// Clear aliases.
		{
			newAliases:      []string{},
			expectedAliases: []string{},
		},
		// Set bad alias.
		{
			newAliases:  []string{"oh no"},
			expectedErr: "invalid also_known_as_uri (oh no) provided in account alias request: uri must not be empty and scheme must be http or https",
		},
		// Try to alias to self (won't do anything).
		{
			newAliases: []string{
				"http://localhost:8080/users/the_mighty_zork",
			},
			expectedAliases: []string{},
		},
		// Try to alias to self and admin
		// (only non-self alias will work).
		{
			newAliases: []string{
				"http://localhost:8080/users/the_mighty_zork",
				"http://localhost:8080/users/admin",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/admin",
			},
		},
		// Alias zork to turtle AND admin,
		// duplicates should be removed.
		{
			newAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
				"http://localhost:8080/users/admin",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/users/admin",
			},
		},
		// Alias zork to turtle using both URI and URL
		// for turtle. Only URI should end up being used.
		{
			newAliases: []string{
				"http://localhost:8080/users/1happyturtle",
				"http://localhost:8080/@1happyturtle",
			},
			expectedAliases: []string{
				"http://localhost:8080/users/1happyturtle",
			},
		},
	} {
		var (
			ctx      = context.Background()
			testAcct = new(gtsmodel.Account)
		)

		// Copy zork test account.
		*testAcct = *suite.testAccounts["local_account_1"]

		apiAcct, err := suite.accountProcessor.Alias(ctx, testAcct, test.newAliases)
		if err != nil {
			if err.Error() != test.expectedErr {
				suite.FailNow("", "unexpected error: %s", err)
			} else {
				continue
			}
		}

		if !slices.Equal(apiAcct.Source.AlsoKnownAsURIs, test.expectedAliases) {
			suite.FailNow("", "unexpected aliases: %+v", apiAcct.Source.AlsoKnownAsURIs)
		}
	}
}

func TestAliasTestSuite(t *testing.T) {
	suite.Run(t, new(AliasTestSuite))
}
