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

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/gotosocial/internal/ap"
	"code.superseriousbusiness.org/gotosocial/internal/db"
	"code.superseriousbusiness.org/gotosocial/internal/gtsmodel"
	"code.superseriousbusiness.org/gotosocial/internal/uris"
	"code.superseriousbusiness.org/gotosocial/testrig"
	"github.com/stretchr/testify/suite"
)

type RejectTestSuite struct {
	FederatingDBTestSuite
}

func (suite *RejectTestSuite) TestRejectFollowRequest() {
	// local_account_1 sent a follow request to remote_account_2;
	// remote_account_2 rejects the follow request
	followingAccount := suite.testAccounts["local_account_1"]
	followedAccount := suite.testAccounts["remote_account_2"]
	ctx := createTestContext(suite.T(), followingAccount, followedAccount)

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

	asFollow, err := suite.tc.FollowToAS(ctx, suite.tc.FollowRequestToFollow(ctx, fr))
	suite.NoError(err)

	rejectingAccountURI := testrig.URLMustParse(followedAccount.URI)
	requestingAccountURI := testrig.URLMustParse(followingAccount.URI)

	// create a Reject
	reject := streams.NewActivityStreamsReject()

	// set an ID on it
	ap.SetJSONLDId(reject, testrig.URLMustParse("https://example.org/some/reject/id"))

	// set the rejecting actor on it
	ap.AppendActorIRIs(reject, rejectingAccountURI)

	// Set the recreated follow as the 'object' property.
	acceptObject := streams.NewActivityStreamsObjectProperty()
	acceptObject.AppendActivityStreamsFollow(asFollow)
	reject.SetActivityStreamsObject(acceptObject)

	// Set the To of the reject as the originator of the follow
	ap.AppendTo(reject, requestingAccountURI)

	// process the reject in the federating database
	err = suite.federatingDB.Reject(ctx, reject)
	suite.NoError(err)

	// there should be nothing in the federator channel since nothing needs to be passed
	_, ok := suite.getFederatorMsg(time.Second)
	suite.False(ok)

	// the follow request should not be in the database anymore -- it's been rejected
	err = suite.db.GetByID(ctx, fr.ID, &gtsmodel.FollowRequest{})
	suite.ErrorIs(err, db.ErrNoEntries)
}

func TestRejectTestSuite(t *testing.T) {
	suite.Run(t, &RejectTestSuite{})
}
