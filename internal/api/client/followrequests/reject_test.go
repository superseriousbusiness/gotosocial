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

package followrequests_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/api/client/followrequests"
	"github.com/superseriousbusiness/gotosocial/internal/gtsmodel"
)

type RejectTestSuite struct {
	FollowRequestStandardTestSuite
}

func (suite *RejectTestSuite) TestReject() {
	requestingAccount := suite.testAccounts["remote_account_2"]
	targetAccount := suite.testAccounts["local_account_1"]

	// put a follow request in the database
	fr := &gtsmodel.FollowRequest{
		ID:              "01FJ1S8DX3STJJ6CEYPMZ1M0R3",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		URI:             fmt.Sprintf("%s/follow/01FJ1S8DX3STJJ6CEYPMZ1M0R3", requestingAccount.URI),
		AccountID:       requestingAccount.ID,
		TargetAccountID: targetAccount.ID,
	}

	err := suite.db.Put(context.Background(), fr)
	suite.NoError(err)

	recorder := httptest.NewRecorder()
	ctx := suite.newContext(recorder, http.MethodPost, []byte{}, fmt.Sprintf("/api/v1/follow_requests/%s/reject", requestingAccount.ID), "")

	ctx.Params = gin.Params{
		gin.Param{
			Key:   followrequests.IDKey,
			Value: requestingAccount.ID,
		},
	}

	// call the handler
	suite.followRequestModule.FollowRequestRejectPOSTHandler(ctx)

	// 1. we should have OK because our request was valid
	suite.Equal(http.StatusOK, recorder.Code)

	// 2. we should have no error message in the result body
	result := recorder.Result()
	defer result.Body.Close()

	// check the response
	b, err := ioutil.ReadAll(result.Body)
	assert.NoError(suite.T(), err)

	suite.Equal(`{"id":"01FHMQX3GAABWSM0S2VZEC2SWC","following":false,"showing_reblogs":false,"notifying":false,"followed_by":false,"blocking":false,"blocked_by":false,"muting":false,"muting_notifications":false,"requested":false,"domain_blocking":false,"endorsed":false,"note":""}`, string(b))
}

func TestRejectTestSuite(t *testing.T) {
	suite.Run(t, &RejectTestSuite{})
}
