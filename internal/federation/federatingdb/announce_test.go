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

package federatingdb_test

import (
	"testing"
	"time"

	"code.superseriousbusiness.org/activity/streams/vocab"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/id"
	"github.com/stretchr/testify/suite"
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
	msg, _ := suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ActivityAnnounce, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)

	boost, ok := msg.GTSModel.(*gtsmodel.Status)
	suite.True(ok)
	suite.Equal(announcingAccount.ID, boost.AccountID)

	// only the URI will be set for the boosted status
	// because it still needs to be dereferenced
	suite.Nil(boost.BoostOf)
	suite.Equal("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1", boost.BoostOfURI)
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
	msg, _ := suite.getFederatorMsg(5 * time.Second)
	suite.Equal(ap.ActivityAnnounce, msg.APObjectType)
	suite.Equal(ap.ActivityCreate, msg.APActivityType)
	boost, ok := msg.GTSModel.(*gtsmodel.Status)
	suite.True(ok)
	suite.Equal(announcingAccount.ID, boost.AccountID)

	// Insert the boost-of status into the
	// DB cache to emulate processor handling
	boost.ID = id.NewULIDFromTime(boost.CreatedAt)
	suite.state.Caches.DB.Status.Put(boost)

	// only the URI will be set for the boosted status
	// because it still needs to be dereferenced
	suite.Nil(boost.BoostOf)
	suite.Equal("http://example.org/users/Some_User/statuses/afaba698-5740-4e32-a702-af61aa543bc1", boost.BoostOfURI)

	ctx2 := createTestContext(receivingAccount2, announcingAccount)
	announce2 := suite.testActivities["announce_forwarded_1_turtle"]

	err = suite.federatingDB.Announce(ctx2, announce2.Activity.(vocab.ActivityStreamsAnnounce))
	suite.NoError(err)

	// since this is a repeat announce with the same URI, just delivered to a different inbox,
	// we should have nothing in the messages channel...
	_, ok = suite.getFederatorMsg(time.Second)
	suite.False(ok)
}

func TestAnnounceTestSuite(t *testing.T) {
	suite.Run(t, &AnnounceTestSuite{})
}
