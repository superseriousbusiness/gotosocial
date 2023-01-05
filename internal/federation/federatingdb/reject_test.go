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
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/activity/streams"
	"github.com/superseriousbusiness/gotosocial/internal/db"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
	"github.com/superseriousbusiness/gotosocial/internal/uris"
	"github.com/superseriousbusiness/gotosocial/testrig"
)

type RejectTestSuite struct {
	FederatingDBTestSuite
}

func (suite *RejectTestSuite) TestRejectFollowRequest() {
	// local_account_1 sent a follow request to remote_account_2;
	// remote_account_2 rejects the follow request
	followingAccount := suite.testAccounts["local_account_1"]
	followedAccount := suite.testAccounts["remote_account_2"]
	ctx := createTestContext(followingAccount, followedAccount)

	// put the follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             uris.GenerateURIForFollow(followingAccount.Username, "01FJ1S8DX3STJJ6CEYPMZ1M0R3"),
		AccountID:       followingAccount.ID,
		TargetAccountID: followedAccount.ID,
	}
	err := suite.db.Put(ctx, fr)
	suite.NoError(err)

	asFollow, err := suite.tc.FollowToAS(ctx, suite.tc.FollowRequestToFollow(ctx, fr), followingAccount, followedAccount)
	suite.NoError(err)

	rejectingAccountURI := testrig.URLMustParse(followedAccount.URI)
	requestingAccountURI := testrig.URLMustParse(followingAccount.URI)

	// create a Reject
	reject := streams.NewActivityStreamsReject()

	// set the rejecting actor on it
	acceptActorProp := streams.NewActivityStreamsActorProperty()
	acceptActorProp.AppendIRI(rejectingAccountURI)
	reject.SetActivityStreamsActor(acceptActorProp)

	// Set the recreated follow as the 'object' property.
	acceptObject := streams.NewActivityStreamsObjectProperty()
	acceptObject.AppendActivityStreamsFollow(asFollow)
	reject.SetActivityStreamsObject(acceptObject)

	// Set the To of the reject as the originator of the follow
	acceptTo := streams.NewActivityStreamsToProperty()
	acceptTo.AppendIRI(requestingAccountURI)
	reject.SetActivityStreamsTo(acceptTo)

	// process the reject in the federating database
	err = suite.federatingDB.Reject(ctx, reject)
	suite.NoError(err)

	// there should be nothing in the federator channel since nothing needs to be passed
	suite.Empty(suite.fromFederator)

	// the follow request should not be in the database anymore -- it's been rejected
	err = suite.db.GetByID(ctx, fr.ID, &gtsmodel.FollowRequest{})
	suite.ErrorIs(err, db.ErrNoEntries)
}

func TestRejectTestSuite(t *testing.T) {
	suite.Run(t, &RejectTestSuite{})
}
