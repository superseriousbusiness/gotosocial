/*
   GoToSocial
   Copyright (C) 2021-2023 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package federatingdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type InboxTestSuite struct {
	FederatingDBTestSuite
}

func (suite *InboxTestSuite) TestInboxesForFollowersIRI() {
	ctx := context.Background()
	testAccount := suite.testAccounts["local_account_1"]

	inboxIRIs, err := suite.federatingDB.InboxesForIRI(ctx, testrig.URLMustParse(testAccount.FollowersURI))
	suite.NoError(err)

	asStrings := []string{}
	for _, i := range inboxIRIs {
		asStrings = append(asStrings, i.String())
	}

	suite.Len(asStrings, 2)
	suite.Contains(asStrings, suite.testAccounts["local_account_2"].InboxURI)
	suite.Contains(asStrings, suite.testAccounts["admin_account"].InboxURI)
}

func (suite *InboxTestSuite) TestInboxesForAccountIRI() {
	ctx := context.Background()
	testAccount := suite.testAccounts["local_account_1"]

	inboxIRIs, err := suite.federatingDB.InboxesForIRI(ctx, testrig.URLMustParse(testAccount.URI))
	suite.NoError(err)

	asStrings := []string{}
	for _, i := range inboxIRIs {
		asStrings = append(asStrings, i.String())
	}

	suite.Len(asStrings, 1)
	suite.Contains(asStrings, suite.testAccounts["local_account_1"].InboxURI)
}

func (suite *InboxTestSuite) TestInboxesForAccountIRIWithSharedInbox() {
	ctx := context.Background()
	testAccount := suite.testAccounts["local_account_1"]
	sharedInbox := "http://some-inbox-iri/weeeeeeeeeeeee"
	testAccount.SharedInboxURI = &sharedInbox
	if err := suite.db.UpdateAccount(ctx, testAccount); err != nil {
		suite.FailNow("error updating account")
	}

	inboxIRIs, err := suite.federatingDB.InboxesForIRI(ctx, testrig.URLMustParse(testAccount.URI))
	suite.NoError(err)

	asStrings := []string{}
	for _, i := range inboxIRIs {
		asStrings = append(asStrings, i.String())
	}

	suite.Len(asStrings, 1)
	suite.Contains(asStrings, "http://some-inbox-iri/weeeeeeeeeeeee")
}

func TestInboxTestSuite(t *testing.T) {
	suite.Run(t, &InboxTestSuite{})
}
