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
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams/vocab"
	"github.com/superseriousbusiness/gotosocial/internal/ap"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type AnnounceTestSuite struct {
	FederatingDBTestSuite
}

func (suite *AnnounceTestSuite) TestNewAnnounce() {
	receivingAccount1 := suite.testAccounts["local_account_1"]
	announcingAccount := suite.testAccounts["remote_account_1"]

	ctx := createTestContext(receivingAccount1, announcingAccount)
	announce1 := suite.testActivities["announce_forwarded_1_zork"]

	err := suite.federatingDB.Announce(ctx, announce1.Activity.(vocab.ActivityStreamsAnnounce))
	suite.NoError(err)

	// should be a message heading to the processor now, which we can intercept here
	msg := <-suite.fromFederator
	suite.Equal(ap.ActivityAnnounce, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)

	boost, ok := msg.GTSModel.(*gtsmodel.Status)
	suite.True(ok)
	suite.Equal(announcingAccount.ID, boost.AccountID)

	// only the URI will be set on the boosted status because it still needs to be dereferenced
	suite.NotEmpty(boost.BoostOf.URI)
}

func (suite *AnnounceTestSuite) TestAnnounceTwice() {
	receivingAccount1 := suite.testAccounts["local_account_1"]
	receivingAccount2 := suite.testAccounts["local_account_2"]

	announcingAccount := suite.testAccounts["remote_account_1"]

	ctx1 := createTestContext(receivingAccount1, announcingAccount)
	announce1 := suite.testActivities["announce_forwarded_1_zork"]

	err := suite.federatingDB.Announce(ctx1, announce1.Activity.(vocab.ActivityStreamsAnnounce))
	suite.NoError(err)

	// should be a message heading to the processor now, which we can intercept here
	msg := <-suite.fromFederator
	suite.Equal(ap.ActivityAnnounce, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)
	boost, ok := msg.GTSModel.(*gtsmodel.Status)
	suite.True(ok)
	suite.Equal(announcingAccount.ID, boost.AccountID)

	// only the URI will be set on the boosted status because it still needs to be dereferenced
	suite.NotEmpty(boost.BoostOf.URI)

	ctx2 := createTestContext(receivingAccount2, announcingAccount)
	announce2 := suite.testActivities["announce_forwarded_1_turtle"]

	err = suite.federatingDB.Announce(ctx2, announce2.Activity.(vocab.ActivityStreamsAnnounce))
	suite.NoError(err)

	// since this is a repeat announce with the same URI, just delivered to a different inbox,
	// we should have nothing in the messages channel...
	suite.Empty(suite.fromFederator)
}

func TestAnnounceTestSuite(t *testing.T) {
	suite.Run(t, &AnnounceTestSuite{})
}
