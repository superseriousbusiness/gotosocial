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

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type DomainPermissionSubscriptionTestSuite struct {
	BunDBStandardTestSuite
}

func (suite *DomainPermissionSubscriptionTestSuite) TestCount() {
	var (
		ctx         = context.Background()
		testAccount = suite.testAccounts["admin_account"]
		permSub     = &gtsmodel.DomainPermissionSubscription{
			ID:                 "01JGV3VZ72K58BYW8H5GEVBZGN",
			PermissionType:     gtsmodel.DomainPermissionBlock,
			CreatedByAccountID: testAccount.ID,
			CreatedByAccount:   testAccount,
			URI:                "https://example.org/whatever.csv",
			ContentType:        gtsmodel.DomainPermSubContentTypeCSV,
		}
		perms = []*gtsmodel.DomainBlock{
			{
				ID:                 "01JGV42G72YCKN06AC51RZPFES",
				Domain:             "whatever.com",
				CreatedByAccountID: testAccount.ID,
				CreatedByAccount:   testAccount,
				SubscriptionID:     permSub.ID,
			},
			{
				ID:                 "01JGV43ZQKYPHM2M0YBQDFDSD1",
				Domain:             "aaaa.example.org",
				CreatedByAccountID: testAccount.ID,
				CreatedByAccount:   testAccount,
				SubscriptionID:     permSub.ID,
			},
			{
				ID:                 "01JGV444KDDC4WFG6MZQVM0N2Z",
				Domain:             "bbbb.example.org",
				CreatedByAccountID: testAccount.ID,
				CreatedByAccount:   testAccount,
				SubscriptionID:     permSub.ID,
			},
			{
				ID:                 "01JGV44AFEMBWS6P6S72BQK376",
				Domain:             "cccc.example.org",
				CreatedByAccountID: testAccount.ID,
				CreatedByAccount:   testAccount,
				SubscriptionID:     permSub.ID,
			},
		}
	)

	// Whack the perm sub in the DB.
	if err := suite.state.DB.PutDomainPermissionSubscription(ctx, permSub); err != nil {
		suite.FailNow(err.Error())
	}

	// Whack the perms in the db.
	for _, perm := range perms {
		if err := suite.state.DB.CreateDomainBlock(ctx, perm); err != nil {
			suite.FailNow(err.Error())
		}
	}

	// Count 'em.
	count, err := suite.state.DB.CountDomainPermissionSubscriptionPerms(ctx, permSub.ID)
	if err != nil {
		suite.FailNow(err.Error())
	}

	suite.Equal(4, count)
}

func TestDomainPermissionSubscriptionTestSuite(t *testing.T) {
	suite.Run(t, new(DomainPermissionSubscriptionTestSuite))
}
